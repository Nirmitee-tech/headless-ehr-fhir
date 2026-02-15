package hl7v2

import (
	"bytes"
	"net"
	"sync"
	"testing"
	"time"
)

// testADT is a minimal ADT^A01 message used across MLLP tests.
var testADT = "MSH|^~\\&|SendApp|SendFac|RecvApp|RecvFac|20240115120000||ADT^A01|MSG001|P|2.5.1\rPID|||12345||Smith^John||19800101|M"

// =========== Framing Tests ===========

func TestFrameMessage(t *testing.T) {
	raw := []byte("MSH|^~\\&|A|B|||20240115||ADT^A01|C1|P|2.5.1")
	framed := FrameMessage(raw)

	if framed[0] != MLLPStartBlock {
		t.Errorf("expected first byte 0x0B, got 0x%02X", framed[0])
	}
	if framed[len(framed)-2] != MLLPEndBlock {
		t.Errorf("expected second-to-last byte 0x1C, got 0x%02X", framed[len(framed)-2])
	}
	if framed[len(framed)-1] != MLLPCarriageReturn {
		t.Errorf("expected last byte 0x0D, got 0x%02X", framed[len(framed)-1])
	}

	inner := framed[1 : len(framed)-2]
	if !bytes.Equal(inner, raw) {
		t.Errorf("inner bytes do not match original")
	}
}

func TestUnframeMessage_Valid(t *testing.T) {
	raw := []byte("MSH|test")
	framed := FrameMessage(raw)

	msg, rest, found := UnframeMessage(framed)
	if !found {
		t.Fatal("expected found=true")
	}
	if !bytes.Equal(msg, raw) {
		t.Errorf("expected %q, got %q", raw, msg)
	}
	if len(rest) != 0 {
		t.Errorf("expected empty rest, got %d bytes", len(rest))
	}
}

func TestUnframeMessage_NoStart(t *testing.T) {
	data := []byte("no start block here")
	_, _, found := UnframeMessage(data)
	if found {
		t.Error("expected found=false when no start block present")
	}
}

func TestUnframeMessage_Partial(t *testing.T) {
	// Start block present but no end block sequence.
	data := []byte{MLLPStartBlock}
	data = append(data, []byte("MSH|partial")...)

	_, _, found := UnframeMessage(data)
	if found {
		t.Error("expected found=false for partial frame")
	}
}

func TestUnframeMessage_MultipleMessages(t *testing.T) {
	msg1 := []byte("MSG_ONE")
	msg2 := []byte("MSG_TWO")
	combined := append(FrameMessage(msg1), FrameMessage(msg2)...)

	first, rest, found := UnframeMessage(combined)
	if !found {
		t.Fatal("expected found=true for first message")
	}
	if !bytes.Equal(first, msg1) {
		t.Errorf("first message: expected %q, got %q", msg1, first)
	}

	second, rest2, found2 := UnframeMessage(rest)
	if !found2 {
		t.Fatal("expected found=true for second message")
	}
	if !bytes.Equal(second, msg2) {
		t.Errorf("second message: expected %q, got %q", msg2, second)
	}
	if len(rest2) != 0 {
		t.Errorf("expected empty rest after second message, got %d bytes", len(rest2))
	}
}

// =========== ACK Tests ===========

func TestGenerateACK_AA(t *testing.T) {
	msg := parseTestMessage(t, testADT)
	ack := GenerateACK(msg, "AA")

	if ack.SendingApp != "RecvApp" {
		t.Errorf("expected SendingApp 'RecvApp', got %q", ack.SendingApp)
	}
	if ack.SendingFac != "RecvFac" {
		t.Errorf("expected SendingFac 'RecvFac', got %q", ack.SendingFac)
	}
	if ack.ReceivingApp != "SendApp" {
		t.Errorf("expected ReceivingApp 'SendApp', got %q", ack.ReceivingApp)
	}
	if ack.ReceivingFac != "SendFac" {
		t.Errorf("expected ReceivingFac 'SendFac', got %q", ack.ReceivingFac)
	}

	msa := ack.GetSegment("MSA")
	if msa == nil {
		t.Fatal("expected MSA segment in ACK")
	}
	if msa.GetField(1) != "AA" {
		t.Errorf("expected MSA-1 'AA', got %q", msa.GetField(1))
	}
	if msa.GetField(2) != "MSG001" {
		t.Errorf("expected MSA-2 'MSG001', got %q", msa.GetField(2))
	}
}

func TestGenerateACK_AE(t *testing.T) {
	msg := parseTestMessage(t, testADT)
	ack := GenerateACK(msg, "AE")

	msa := ack.GetSegment("MSA")
	if msa == nil {
		t.Fatal("expected MSA segment in ACK")
	}
	if msa.GetField(1) != "AE" {
		t.Errorf("expected MSA-1 'AE', got %q", msa.GetField(1))
	}
}

func TestGenerateACK_PreservesControlID(t *testing.T) {
	msg := parseTestMessage(t, testADT)
	ack := GenerateACK(msg, "AA")

	msa := ack.GetSegment("MSA")
	if msa == nil {
		t.Fatal("expected MSA segment in ACK")
	}

	// MSA-2 must contain the original message's control ID.
	if msa.GetField(2) != msg.ControlID {
		t.Errorf("expected MSA-2 to be %q, got %q", msg.ControlID, msa.GetField(2))
	}
}

// =========== Server Integration Tests ===========

func TestMLLPServer_StartStop(t *testing.T) {
	s := NewMLLPServer("127.0.0.1:0", DefaultHandler())
	if err := s.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	addr := s.Addr()
	if addr == "" {
		t.Fatal("Addr() returned empty string")
	}

	if err := s.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestMLLPServer_ReceiveMessage(t *testing.T) {
	received := make(chan *Message, 1)
	handler := func(msg *Message) *Message {
		received <- msg
		return GenerateACK(msg, "AA")
	}

	s := NewMLLPServer("127.0.0.1:0", handler)
	if err := s.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer s.Stop()

	conn, err := net.DialTimeout("tcp", s.Addr(), 2*time.Second)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	// Send the MLLP-framed ADT message.
	framed := FrameMessage([]byte(testADT))
	if _, err := conn.Write(framed); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	select {
	case msg := <-received:
		if msg.Type != "ADT^A01" {
			t.Errorf("expected message type 'ADT^A01', got %q", msg.Type)
		}
		if msg.ControlID != "MSG001" {
			t.Errorf("expected control ID 'MSG001', got %q", msg.ControlID)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for message")
	}
}

func TestMLLPServer_SendsACK(t *testing.T) {
	s := NewMLLPServer("127.0.0.1:0", DefaultHandler())
	if err := s.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer s.Stop()

	conn, err := net.DialTimeout("tcp", s.Addr(), 2*time.Second)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	// Send a framed message.
	framed := FrameMessage([]byte(testADT))
	if _, err := conn.Write(framed); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Read back the ACK.
	ackBytes := readMLLPResponse(t, conn, 5*time.Second)

	ack, err := Parse(ackBytes)
	if err != nil {
		t.Fatalf("failed to parse ACK: %v", err)
	}

	msa := ack.GetSegment("MSA")
	if msa == nil {
		t.Fatal("ACK missing MSA segment")
	}
	if msa.GetField(1) != "AA" {
		t.Errorf("expected MSA-1 'AA', got %q", msa.GetField(1))
	}
	if msa.GetField(2) != "MSG001" {
		t.Errorf("expected MSA-2 'MSG001', got %q", msa.GetField(2))
	}
}

func TestMLLPServer_MultipleMessages(t *testing.T) {
	var mu sync.Mutex
	var received []string

	handler := func(msg *Message) *Message {
		mu.Lock()
		received = append(received, msg.ControlID)
		mu.Unlock()
		return GenerateACK(msg, "AA")
	}

	s := NewMLLPServer("127.0.0.1:0", handler)
	if err := s.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer s.Stop()

	conn, err := net.DialTimeout("tcp", s.Addr(), 2*time.Second)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	// Send first message.
	msg1 := "MSH|^~\\&|A|B|C|D|20240115120000||ADT^A01|CTRL1|P|2.5.1\rPID|||111||One^First||19900101|M"
	if _, err := conn.Write(FrameMessage([]byte(msg1))); err != nil {
		t.Fatalf("Write msg1 failed: %v", err)
	}
	// Read ACK for first message.
	readMLLPResponse(t, conn, 5*time.Second)

	// Send second message on the same connection.
	msg2 := "MSH|^~\\&|A|B|C|D|20240115120001||ADT^A01|CTRL2|P|2.5.1\rPID|||222||Two^Second||19910202|F"
	if _, err := conn.Write(FrameMessage([]byte(msg2))); err != nil {
		t.Fatalf("Write msg2 failed: %v", err)
	}
	// Read ACK for second message.
	readMLLPResponse(t, conn, 5*time.Second)

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(received))
	}
	if received[0] != "CTRL1" {
		t.Errorf("expected first control ID 'CTRL1', got %q", received[0])
	}
	if received[1] != "CTRL2" {
		t.Errorf("expected second control ID 'CTRL2', got %q", received[1])
	}
}

func TestMLLPServer_MultipleConnections(t *testing.T) {
	var mu sync.Mutex
	var received []string

	handler := func(msg *Message) *Message {
		mu.Lock()
		received = append(received, msg.ControlID)
		mu.Unlock()
		return GenerateACK(msg, "AA")
	}

	s := NewMLLPServer("127.0.0.1:0", handler)
	if err := s.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer s.Stop()

	var wg sync.WaitGroup
	for i, ctrlID := range []string{"CONN1", "CONN2"} {
		wg.Add(1)
		go func(idx int, id string) {
			defer wg.Done()

			conn, err := net.DialTimeout("tcp", s.Addr(), 2*time.Second)
			if err != nil {
				t.Errorf("Dial failed for conn %d: %v", idx, err)
				return
			}
			defer conn.Close()

			msg := "MSH|^~\\&|A|B|C|D|20240115120000||ADT^A01|" + id + "|P|2.5.1\rPID|||999||Test^User||19850101|M"
			if _, err := conn.Write(FrameMessage([]byte(msg))); err != nil {
				t.Errorf("Write failed for conn %d: %v", idx, err)
				return
			}
			readMLLPResponse(t, conn, 5*time.Second)
		}(i, ctrlID)
	}

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 2 {
		t.Fatalf("expected 2 messages from 2 connections, got %d", len(received))
	}
}

func TestMLLPServer_InvalidMessage(t *testing.T) {
	s := NewMLLPServer("127.0.0.1:0", DefaultHandler())
	if err := s.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer s.Stop()

	conn, err := net.DialTimeout("tcp", s.Addr(), 2*time.Second)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	// Send garbage data framed in MLLP. The server should not crash.
	garbage := FrameMessage([]byte("THIS IS NOT HL7"))
	if _, err := conn.Write(garbage); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Give the server a moment to process (it should just log and continue).
	time.Sleep(200 * time.Millisecond)

	// Now send a valid message to confirm the server is still alive.
	framed := FrameMessage([]byte(testADT))
	if _, err := conn.Write(framed); err != nil {
		t.Fatalf("Write valid message failed: %v", err)
	}

	ackBytes := readMLLPResponse(t, conn, 5*time.Second)
	ack, err := Parse(ackBytes)
	if err != nil {
		t.Fatalf("failed to parse ACK after invalid message: %v", err)
	}

	msa := ack.GetSegment("MSA")
	if msa == nil {
		t.Fatal("ACK missing MSA segment after invalid message")
	}
	if msa.GetField(1) != "AA" {
		t.Errorf("expected MSA-1 'AA', got %q", msa.GetField(1))
	}
}

func TestMLLPServer_Addr(t *testing.T) {
	s := NewMLLPServer("127.0.0.1:0", DefaultHandler())
	if err := s.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer s.Stop()

	addr := s.Addr()
	if addr == "" {
		t.Fatal("Addr() returned empty string")
	}

	// Verify we can connect to the reported address.
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Fatalf("could not connect to reported address %q: %v", addr, err)
	}
	conn.Close()
}

// =========== Helpers ===========

// parseTestMessage is a test helper that parses an HL7v2 string and fails
// the test on error.
func parseTestMessage(t *testing.T, raw string) *Message {
	t.Helper()
	msg, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("failed to parse test message: %v", err)
	}
	return msg
}

// readMLLPResponse reads an MLLP-framed response from a connection.
// It returns the unframed message bytes.
func readMLLPResponse(t *testing.T, conn net.Conn, timeout time.Duration) []byte {
	t.Helper()

	conn.SetReadDeadline(time.Now().Add(timeout))
	buf := make([]byte, 0, 4096)
	readBuf := make([]byte, 4096)

	for {
		n, err := conn.Read(readBuf)
		if n > 0 {
			buf = append(buf, readBuf[:n]...)
		}

		msg, _, found := UnframeMessage(buf)
		if found {
			return msg
		}

		if err != nil {
			t.Fatalf("error reading MLLP response: %v (buf so far: %d bytes)", err, len(buf))
		}
	}
}
