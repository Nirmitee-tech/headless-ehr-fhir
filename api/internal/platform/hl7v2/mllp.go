package hl7v2

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	// MLLPStartBlock is the MLLP start-of-message byte (VT / vertical tab).
	MLLPStartBlock = 0x0B

	// MLLPEndBlock is the MLLP end-of-message byte (FS / file separator).
	MLLPEndBlock = 0x1C

	// MLLPCarriageReturn is the trailing CR after the end block.
	MLLPCarriageReturn = 0x0D

	// mllpMaxMessageSize is the maximum buffer size for a single MLLP message (1 MB).
	mllpMaxMessageSize = 1 << 20

	// mllpReadTimeout is the read deadline applied to each connection.
	mllpReadTimeout = 30 * time.Second
)

// MessageHandler is called for each received HL7v2 message.
// It receives the parsed message and returns an ACK/NAK message to send back.
// Return nil to send no response.
type MessageHandler func(msg *Message) *Message

// MLLPServer listens for HL7v2 messages over MLLP/TCP.
type MLLPServer struct {
	addr     string
	handler  MessageHandler
	listener net.Listener
	mu       sync.Mutex
	conns    map[net.Conn]struct{}
	done     chan struct{}
	wg       sync.WaitGroup
	logger   func(format string, args ...interface{})
}

// NewMLLPServer creates a new MLLP server that will listen on the given
// address and dispatch parsed messages to handler.
func NewMLLPServer(addr string, handler MessageHandler) *MLLPServer {
	return &MLLPServer{
		addr:    addr,
		handler: handler,
		conns:   make(map[net.Conn]struct{}),
		done:    make(chan struct{}),
	}
}

// Start begins listening for connections. It is non-blocking: the accept loop
// runs in a background goroutine.
func (s *MLLPServer) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("mllp: failed to listen on %s: %w", s.addr, err)
	}
	s.listener = ln

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.acceptLoop()
	}()

	return nil
}

// Stop gracefully shuts down the server. It closes the listener, then closes
// all tracked connections, and waits for all goroutines to finish.
func (s *MLLPServer) Stop() error {
	close(s.done)

	// Close the listener so acceptLoop unblocks.
	var err error
	if s.listener != nil {
		err = s.listener.Close()
	}

	// Close every tracked connection.
	s.mu.Lock()
	for conn := range s.conns {
		conn.Close()
	}
	s.mu.Unlock()

	// Wait for all goroutines (accept loop + connection handlers) to exit.
	s.wg.Wait()

	return err
}

// Addr returns the listener address string. This is especially useful when the
// server was started with port 0 (OS-assigned port).
func (s *MLLPServer) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.addr
}

// acceptLoop runs in its own goroutine, accepting new TCP connections until
// the listener is closed.
func (s *MLLPServer) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// Check if we are shutting down.
			select {
			case <-s.done:
				return
			default:
			}
			s.logf("mllp: accept error: %v", err)
			return
		}

		s.trackConn(conn, true)

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			defer s.trackConn(conn, false)
			defer conn.Close()
			s.handleConnection(conn)
		}()
	}
}

// trackConn adds or removes a connection from the tracked set.
func (s *MLLPServer) trackConn(conn net.Conn, add bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if add {
		s.conns[conn] = struct{}{}
	} else {
		delete(s.conns, conn)
	}
}

// handleConnection reads MLLP-framed messages from conn, parses them,
// dispatches to the handler, and writes back any response.
func (s *MLLPServer) handleConnection(conn net.Conn) {
	buf := make([]byte, 0, 4096)
	readBuf := make([]byte, 4096)

	for {
		// Check for shutdown.
		select {
		case <-s.done:
			return
		default:
		}

		// Set a read deadline so we don't block forever.
		conn.SetReadDeadline(time.Now().Add(mllpReadTimeout))

		n, err := conn.Read(readBuf)
		if n > 0 {
			buf = append(buf, readBuf[:n]...)

			// Guard against oversized messages.
			if len(buf) > mllpMaxMessageSize {
				s.logf("mllp: message exceeds max size, closing connection")
				return
			}

			// Process all complete messages in the buffer.
			for {
				msgBytes, rest, found := UnframeMessage(buf)
				if !found {
					break
				}
				buf = rest

				s.processMessage(conn, msgBytes)
			}
		}

		if err != nil {
			// Timeout or EOF is normal when idle or the client disconnects.
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// On timeout with no pending data, close the connection.
				if len(buf) == 0 {
					return
				}
				// Otherwise keep reading to finish the partial message.
				continue
			}
			// Connection closed or other error.
			return
		}
	}
}

// processMessage parses a single message, calls the handler, and writes
// the response (if any) back to conn.
func (s *MLLPServer) processMessage(conn net.Conn, raw []byte) {
	msg, err := Parse(raw)
	if err != nil {
		s.logf("mllp: parse error: %v", err)
		return
	}

	resp := s.handler(msg)
	if resp == nil {
		return
	}

	// Serialize the response message and frame it.
	respBytes := SerializeMessage(resp)
	framed := FrameMessage(respBytes)

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if _, err := conn.Write(framed); err != nil {
		s.logf("mllp: write error: %v", err)
	}
}

// logf logs a formatted message if a logger is configured.
func (s *MLLPServer) logf(format string, args ...interface{}) {
	if s.logger != nil {
		s.logger(format, args...)
	}
}

// ---------------------------------------------------------------------------
// MLLP framing helpers
// ---------------------------------------------------------------------------

// FrameMessage wraps raw HL7v2 bytes in MLLP framing:
//
//	<0x0B> + message + <0x1C><0x0D>
func FrameMessage(data []byte) []byte {
	frame := make([]byte, 0, len(data)+3)
	frame = append(frame, MLLPStartBlock)
	frame = append(frame, data...)
	frame = append(frame, MLLPEndBlock, MLLPCarriageReturn)
	return frame
}

// UnframeMessage extracts HL7v2 bytes from an MLLP frame. It looks for the
// first start block byte, then reads until end block + CR. It returns the
// extracted message, any remaining bytes after the frame, and whether a
// complete frame was found.
func UnframeMessage(data []byte) (message []byte, rest []byte, found bool) {
	// Find start block.
	startIdx := bytes.IndexByte(data, MLLPStartBlock)
	if startIdx == -1 {
		return nil, data, false
	}

	// Find end block sequence (0x1C 0x0D) after the start block.
	endSeq := []byte{MLLPEndBlock, MLLPCarriageReturn}
	endIdx := bytes.Index(data[startIdx+1:], endSeq)
	if endIdx == -1 {
		return nil, data, false
	}

	// Adjust endIdx to be relative to the full data slice.
	endIdx = startIdx + 1 + endIdx

	message = data[startIdx+1 : endIdx]
	rest = data[endIdx+2:]
	found = true
	return
}

// ---------------------------------------------------------------------------
// ACK generation
// ---------------------------------------------------------------------------

// GenerateACK creates an HL7v2 ACK message for the given incoming message.
// ackCode should be "AA" (accept), "AE" (error), or "AR" (reject).
//
// The ACK swaps the sending and receiving application/facility from the
// original message and references the original control ID in MSA-2.
func GenerateACK(incoming *Message, ackCode string) *Message {
	// Extract the trigger event from the incoming message type.
	// incoming.Type is something like "ADT^A01"; we want "A01".
	trigger := ""
	if parts := strings.SplitN(incoming.Type, "^", 2); len(parts) == 2 {
		trigger = parts[1]
	}

	now := time.Now().UTC()
	timestamp := now.Format("20060102150405")
	controlID := fmt.Sprintf("ACK%s", now.Format("20060102150405.000"))

	ack := &Message{
		Type:         "ACK^" + trigger,
		ControlID:    controlID,
		Version:      incoming.Version,
		Timestamp:    now,
		SendingApp:   incoming.ReceivingApp,
		SendingFac:   incoming.ReceivingFac,
		ReceivingApp: incoming.SendingApp,
		ReceivingFac: incoming.SendingFac,
	}

	// Build MSH segment.
	msh := Segment{
		Name: "MSH",
		Fields: []Field{
			{Value: "|", Components: []string{"|"}},                                // MSH-1
			{Value: "^~\\&", Components: []string{"^~\\&"}},                        // MSH-2
			{Value: ack.SendingApp, Components: []string{ack.SendingApp}},           // MSH-3
			{Value: ack.SendingFac, Components: []string{ack.SendingFac}},           // MSH-4
			{Value: ack.ReceivingApp, Components: []string{ack.ReceivingApp}},       // MSH-5
			{Value: ack.ReceivingFac, Components: []string{ack.ReceivingFac}},       // MSH-6
			{Value: timestamp, Components: []string{timestamp}},                     // MSH-7
			{Value: "", Components: []string{""}},                                   // MSH-8 (security)
			{Value: "ACK^" + trigger, Components: []string{"ACK", trigger}},         // MSH-9
			{Value: controlID, Components: []string{controlID}},                     // MSH-10
			{Value: "P", Components: []string{"P"}},                                 // MSH-11
			{Value: incoming.Version, Components: []string{incoming.Version}},       // MSH-12
		},
	}

	// Build MSA segment.
	msa := Segment{
		Name: "MSA",
		Fields: []Field{
			{Value: ackCode, Components: []string{ackCode}},                         // MSA-1
			{Value: incoming.ControlID, Components: []string{incoming.ControlID}},   // MSA-2
		},
	}

	ack.Segments = []Segment{msh, msa}

	return ack
}

// ---------------------------------------------------------------------------
// Message serialization
// ---------------------------------------------------------------------------

// SerializeMessage converts a Message struct back into raw HL7v2 bytes
// with \r segment separators.
func SerializeMessage(msg *Message) []byte {
	var segments []string
	for _, seg := range msg.Segments {
		segments = append(segments, serializeSegment(seg))
	}
	return []byte(strings.Join(segments, "\r"))
}

// serializeSegment converts a Segment back into its HL7v2 string form.
func serializeSegment(seg Segment) string {
	if seg.Name == "MSH" {
		// MSH is special: Fields[0] is the field separator itself (|),
		// and Fields[1] is the encoding characters. We reconstruct as:
		// MSH|^~\&|field3|field4|...
		if len(seg.Fields) < 2 {
			return "MSH|"
		}
		parts := make([]string, 0, len(seg.Fields)-1)
		// Start from Fields[1] (MSH-2) onward.
		for i := 1; i < len(seg.Fields); i++ {
			parts = append(parts, seg.Fields[i].Value)
		}
		return "MSH|" + strings.Join(parts, "|")
	}

	parts := make([]string, len(seg.Fields))
	for i, f := range seg.Fields {
		parts[i] = f.Value
	}
	return seg.Name + "|" + strings.Join(parts, "|")
}

// ---------------------------------------------------------------------------
// Default handler
// ---------------------------------------------------------------------------

// DefaultHandler returns a MessageHandler that always ACKs with "AA".
func DefaultHandler() MessageHandler {
	return func(msg *Message) *Message {
		return GenerateACK(msg, "AA")
	}
}
