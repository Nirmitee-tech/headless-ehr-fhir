package fhir

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// ScheduleAvailability represents a time window of availability.
type ScheduleAvailability struct {
	Start       time.Time
	End         time.Time
	SlotType    string // e.g., "routine", "walkin", "urgent"
	Status      string // "free", "busy", "busy-unavailable", "busy-tentative", "entered-in-error"
	Comment     string
	ServiceType string // What service is offered
	Specialty   string // Specialty for this slot
	Actor       string // Who provides the service (Practitioner reference)
}

// SlotStatus represents FHIR Slot status values.
type SlotStatus string

const (
	SlotStatusFree            SlotStatus = "free"
	SlotStatusBusy            SlotStatus = "busy"
	SlotStatusBusyUnavailable SlotStatus = "busy-unavailable"
	SlotStatusBusyTentative   SlotStatus = "busy-tentative"
	SlotStatusEnteredInError  SlotStatus = "entered-in-error"
)

// AvailabilityRequest represents the parameters for a $find or $available query.
type AvailabilityRequest struct {
	Start           time.Time
	End             time.Time
	Duration        int    // Requested duration in minutes
	SlotType        string // Requested slot type
	ServiceType     string // Requested service type
	Specialty       string // Requested specialty
	Practitioner    string // Specific practitioner
	Location        string // Specific location
	Status          []SlotStatus
	IncludeSchedule bool // Include the parent Schedule in results
}

// AvailabilityResult is the response from availability queries.
type AvailabilityResult struct {
	Slots     []map[string]interface{} // Available Slot resources
	Schedules []map[string]interface{} // Associated Schedule resources (if requested)
	Total     int                      // Total available slots
}

// TimeSlot represents a discrete time slot for availability computation.
type TimeSlot struct {
	Start    time.Time
	End      time.Time
	Duration int // minutes
}

// AvailabilityRule defines recurring availability patterns.
type AvailabilityRule struct {
	DaysOfWeek   []time.Weekday
	StartTime    string // HH:MM format
	EndTime      string // HH:MM format
	SlotDuration int    // minutes
	BreakStart   string // Optional break start
	BreakEnd     string // Optional break end
}

// ConflictCheckResult contains the results of checking for scheduling conflicts.
type ConflictCheckResult struct {
	HasConflict bool
	Conflicts   []map[string]interface{} // Conflicting Slot/Appointment resources
	Message     string
}

// AvailabilityStore interface for slot/schedule storage.
type AvailabilityStore interface {
	FindSlots(ctx interface{}, req *AvailabilityRequest) ([]map[string]interface{}, error)
	GetSchedule(ctx interface{}, scheduleID string) (map[string]interface{}, error)
	CheckConflicts(ctx interface{}, start, end time.Time, actor string) (*ConflictCheckResult, error)
}

// InMemoryAvailabilityStore is a test/demo implementation.
type InMemoryAvailabilityStore struct {
	mu        sync.RWMutex
	slots     []map[string]interface{}
	schedules []map[string]interface{}
}

// NewInMemoryAvailabilityStore creates a new in-memory store.
func NewInMemoryAvailabilityStore() *InMemoryAvailabilityStore {
	return &InMemoryAvailabilityStore{
		slots:     make([]map[string]interface{}, 0),
		schedules: make([]map[string]interface{}, 0),
	}
}

// AddSlot adds a slot to the store.
func (s *InMemoryAvailabilityStore) AddSlot(slot map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.slots = append(s.slots, slot)
}

// AddSchedule adds a schedule to the store.
func (s *InMemoryAvailabilityStore) AddSchedule(schedule map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.schedules = append(s.schedules, schedule)
}

// FindSlots finds slots matching the request.
func (s *InMemoryAvailabilityStore) FindSlots(ctx interface{}, req *AvailabilityRequest) ([]map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []map[string]interface{}

	for _, slot := range s.slots {
		if !s.matchesRequest(slot, req) {
			continue
		}
		results = append(results, slot)
	}

	if results == nil {
		results = make([]map[string]interface{}, 0)
	}
	return results, nil
}

// matchesRequest checks whether a slot matches the given availability request.
func (s *InMemoryAvailabilityStore) matchesRequest(slot map[string]interface{}, req *AvailabilityRequest) bool {
	// Check date range.
	if !req.Start.IsZero() && !req.End.IsZero() {
		slotStart, _ := parseSlotTime(slot["start"])
		slotEnd, _ := parseSlotTime(slot["end"])
		if !slotStart.IsZero() && !slotEnd.IsZero() {
			if slotEnd.Before(req.Start) || slotStart.After(req.End) {
				return false
			}
		}
	}

	// Check status filter.
	if len(req.Status) > 0 {
		slotStatus, _ := slot["status"].(string)
		matched := false
		for _, s := range req.Status {
			if string(s) == slotStatus {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check practitioner filter via schedule reference.
	if req.Practitioner != "" {
		if !s.slotMatchesPractitioner(slot, req.Practitioner) {
			return false
		}
	}

	// Check service type filter.
	if req.ServiceType != "" {
		if !slotMatchesServiceType(slot, req.ServiceType) {
			return false
		}
	}

	return true
}

// slotMatchesPractitioner checks if a slot's schedule has the given practitioner.
func (s *InMemoryAvailabilityStore) slotMatchesPractitioner(slot map[string]interface{}, practitioner string) bool {
	schedRef := extractScheduleRef(slot)
	if schedRef == "" {
		return false
	}

	// Extract schedule ID from reference like "Schedule/sched-1".
	schedID := schedRef
	if idx := strings.LastIndex(schedRef, "/"); idx >= 0 {
		schedID = schedRef[idx+1:]
	}

	for _, sched := range s.schedules {
		if sched["id"] == schedID {
			return scheduleHasActor(sched, practitioner)
		}
	}
	return false
}

// extractScheduleRef extracts the schedule reference string from a slot.
func extractScheduleRef(slot map[string]interface{}) string {
	schedMap, ok := slot["schedule"].(map[string]interface{})
	if !ok {
		return ""
	}
	ref, _ := schedMap["reference"].(string)
	return ref
}

// scheduleHasActor checks if a schedule contains the given actor reference.
func scheduleHasActor(sched map[string]interface{}, actor string) bool {
	actors, ok := sched["actor"].([]interface{})
	if !ok {
		return false
	}
	for _, a := range actors {
		aMap, ok := a.(map[string]interface{})
		if !ok {
			continue
		}
		if aMap["reference"] == actor {
			return true
		}
	}
	return false
}

// slotMatchesServiceType checks if a slot has the given service type code.
func slotMatchesServiceType(slot map[string]interface{}, serviceType string) bool {
	serviceTypes, ok := slot["serviceType"].([]interface{})
	if !ok {
		return false
	}
	for _, st := range serviceTypes {
		stMap, ok := st.(map[string]interface{})
		if !ok {
			continue
		}
		codings, ok := stMap["coding"].([]interface{})
		if !ok {
			continue
		}
		for _, c := range codings {
			cMap, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			if cMap["code"] == serviceType {
				return true
			}
		}
	}
	return false
}

// parseSlotTime parses a time string from a slot resource.
func parseSlotTime(v interface{}) (time.Time, error) {
	s, ok := v.(string)
	if !ok {
		return time.Time{}, fmt.Errorf("not a string")
	}
	return parseFlexibleTime(s)
}

// GetSchedule gets a schedule by ID.
func (s *InMemoryAvailabilityStore) GetSchedule(ctx interface{}, scheduleID string) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, sched := range s.schedules {
		if sched["id"] == scheduleID {
			return sched, nil
		}
	}
	return nil, fmt.Errorf("schedule %s not found", scheduleID)
}

// CheckConflicts checks for scheduling conflicts.
func (s *InMemoryAvailabilityStore) CheckConflicts(ctx interface{}, start, end time.Time, actor string) (*ConflictCheckResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := &ConflictCheckResult{
		Conflicts: make([]map[string]interface{}, 0),
	}

	for _, slot := range s.slots {
		status, _ := slot["status"].(string)
		if status != "busy" && status != "busy-tentative" {
			continue
		}

		slotStart, err1 := parseSlotTime(slot["start"])
		slotEnd, err2 := parseSlotTime(slot["end"])
		if err1 != nil || err2 != nil {
			continue
		}

		if !OverlapsTimeRange(start, end, slotStart, slotEnd) {
			continue
		}

		// Check if this slot belongs to the specified actor.
		if actor != "" {
			schedRef := extractScheduleRef(slot)
			if schedRef == "" {
				continue
			}
			schedID := schedRef
			if idx := strings.LastIndex(schedRef, "/"); idx >= 0 {
				schedID = schedRef[idx+1:]
			}

			actorMatch := false
			for _, sched := range s.schedules {
				if sched["id"] == schedID && scheduleHasActor(sched, actor) {
					actorMatch = true
					break
				}
			}
			if !actorMatch {
				continue
			}
		}

		result.HasConflict = true
		result.Conflicts = append(result.Conflicts, slot)
	}

	if result.HasConflict {
		result.Message = fmt.Sprintf("found %d conflicting slot(s)", len(result.Conflicts))
	} else {
		result.Message = "no conflicts found"
	}

	return result, nil
}

// ParseAvailabilityRequest parses URL parameters into an AvailabilityRequest.
func ParseAvailabilityRequest(params url.Values) (*AvailabilityRequest, error) {
	req := &AvailabilityRequest{}

	// Parse start time.
	if startStr := params.Get("start"); startStr != "" {
		t, err := parseFlexibleTime(startStr)
		if err != nil {
			return nil, fmt.Errorf("invalid start date: %w", err)
		}
		req.Start = t
	}

	// Parse end time.
	if endStr := params.Get("end"); endStr != "" {
		t, err := parseFlexibleTime(endStr)
		if err != nil {
			return nil, fmt.Errorf("invalid end date: %w", err)
		}
		req.End = t
	}

	// Parse duration.
	if durStr := params.Get("duration"); durStr != "" {
		d, err := strconv.Atoi(durStr)
		if err != nil {
			return nil, fmt.Errorf("invalid duration: must be an integer")
		}
		req.Duration = d
	}

	req.SlotType = params.Get("slot-type")
	req.ServiceType = params.Get("service-type")
	req.Specialty = params.Get("specialty")
	req.Practitioner = params.Get("practitioner")
	req.Location = params.Get("location")

	// Parse status.
	if statusStr := params.Get("status"); statusStr != "" {
		statuses := strings.Split(statusStr, ",")
		for _, s := range statuses {
			s = strings.TrimSpace(s)
			if s != "" {
				req.Status = append(req.Status, SlotStatus(s))
			}
		}
	}

	// Parse _include.
	if includeStr := params.Get("_include"); includeStr != "" {
		if strings.Contains(includeStr, "Schedule") {
			req.IncludeSchedule = true
		}
	}

	return req, nil
}

// parseFlexibleTime parses a time string supporting RFC3339 and date-only formats.
func parseFlexibleTime(s string) (time.Time, error) {
	// Try RFC3339 first.
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}

	// Try date-only format.
	t, err2 := time.Parse("2006-01-02", s)
	if err2 == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("cannot parse %q as a date/time", s)
}

// ValidateAvailabilityRequest validates the request.
func ValidateAvailabilityRequest(req *AvailabilityRequest) []ValidationIssue {
	var issues []ValidationIssue

	if req.Start.IsZero() {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeRequired,
			Diagnostics: "start parameter is required",
			Location:    "start",
		})
	}

	if req.End.IsZero() {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeRequired,
			Diagnostics: "end parameter is required",
			Location:    "end",
		})
	}

	if !req.Start.IsZero() && !req.End.IsZero() && req.Start.After(req.End) {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeValue,
			Diagnostics: "start must be before end",
			Location:    "start",
		})
	}

	if req.Duration < 0 {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeValue,
			Diagnostics: "duration must be non-negative",
			Location:    "duration",
		})
	}

	return issues
}

// ParseTimeOfDay parses a HH:MM time string.
func ParseTimeOfDay(s string) (int, int, error) {
	if len(s) != 5 || s[2] != ':' {
		return 0, 0, fmt.Errorf("invalid time format %q: expected HH:MM", s)
	}

	hour, err := strconv.Atoi(s[:2])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid hour in %q", s)
	}
	minute, err := strconv.Atoi(s[3:])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid minute in %q", s)
	}

	if hour < 0 || hour > 23 {
		return 0, 0, fmt.Errorf("hour out of range in %q", s)
	}
	if minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("minute out of range in %q", s)
	}

	return hour, minute, nil
}

// OverlapsTimeRange checks if two time ranges overlap.
// Adjacent ranges (end1 == start2) are not considered overlapping.
// Zero-length ranges (start == end) are not considered overlapping.
func OverlapsTimeRange(start1, end1, start2, end2 time.Time) bool {
	if start1.Equal(end1) || start2.Equal(end2) {
		return false
	}
	return start1.Before(end2) && start2.Before(end1)
}

// GenerateTimeSlots generates discrete time slots from availability rules.
func GenerateTimeSlots(rule *AvailabilityRule, start, end time.Time) []TimeSlot {
	if rule.SlotDuration <= 0 {
		return nil
	}

	startHour, startMin, err := ParseTimeOfDay(rule.StartTime)
	if err != nil {
		return nil
	}
	endHour, endMin, err := ParseTimeOfDay(rule.EndTime)
	if err != nil {
		return nil
	}

	var breakStartH, breakStartM, breakEndH, breakEndM int
	hasBreak := false
	if rule.BreakStart != "" && rule.BreakEnd != "" {
		breakStartH, breakStartM, err = ParseTimeOfDay(rule.BreakStart)
		if err != nil {
			return nil
		}
		breakEndH, breakEndM, err = ParseTimeOfDay(rule.BreakEnd)
		if err != nil {
			return nil
		}
		hasBreak = true
	}

	var slots []TimeSlot

	// Iterate through each day in the range.
	current := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	endDay := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())

	for !current.After(endDay) {
		// Check if this day matches any day of week in the rule.
		dayMatches := false
		for _, dow := range rule.DaysOfWeek {
			if current.Weekday() == dow {
				dayMatches = true
				break
			}
		}

		if dayMatches {
			dayStart := time.Date(current.Year(), current.Month(), current.Day(), startHour, startMin, 0, 0, current.Location())
			dayEnd := time.Date(current.Year(), current.Month(), current.Day(), endHour, endMin, 0, 0, current.Location())

			var breakStart, breakEnd time.Time
			if hasBreak {
				breakStart = time.Date(current.Year(), current.Month(), current.Day(), breakStartH, breakStartM, 0, 0, current.Location())
				breakEnd = time.Date(current.Year(), current.Month(), current.Day(), breakEndH, breakEndM, 0, 0, current.Location())
			}

			slotStart := dayStart
			for slotStart.Add(time.Duration(rule.SlotDuration) * time.Minute).Before(dayEnd) ||
				slotStart.Add(time.Duration(rule.SlotDuration)*time.Minute).Equal(dayEnd) {
				slotEnd := slotStart.Add(time.Duration(rule.SlotDuration) * time.Minute)

				// Skip slots that overlap with break.
				if hasBreak && OverlapsTimeRange(slotStart, slotEnd, breakStart, breakEnd) {
					slotStart = slotEnd
					continue
				}

				slots = append(slots, TimeSlot{
					Start:    slotStart,
					End:      slotEnd,
					Duration: rule.SlotDuration,
				})
				slotStart = slotEnd
			}
		}

		current = current.AddDate(0, 0, 1)
	}

	return slots
}

// MergeAvailability merges multiple availability streams and removes conflicts.
func MergeAvailability(available []TimeSlot, busy []TimeSlot) []TimeSlot {
	if len(available) == 0 {
		return nil
	}

	var result []TimeSlot

	for _, avail := range available {
		conflicted := false
		for _, b := range busy {
			if OverlapsTimeRange(avail.Start, avail.End, b.Start, b.End) {
				conflicted = true
				break
			}
		}
		if !conflicted {
			result = append(result, avail)
		}
	}

	if result == nil {
		result = make([]TimeSlot, 0)
	}
	return result
}

// FilterSlotsByDuration filters slots that meet minimum duration requirement.
func FilterSlotsByDuration(slots []TimeSlot, minDuration int) []TimeSlot {
	if minDuration <= 0 {
		return slots
	}

	var result []TimeSlot
	for _, s := range slots {
		if s.Duration >= minDuration {
			result = append(result, s)
		}
	}
	if result == nil {
		result = make([]TimeSlot, 0)
	}
	return result
}

// BuildSlotResource creates a FHIR Slot resource from availability data.
func BuildSlotResource(avail *ScheduleAvailability, scheduleRef string) map[string]interface{} {
	slot := map[string]interface{}{
		"resourceType": "Slot",
		"status":       avail.Status,
		"start":        avail.Start.UTC().Format(time.RFC3339),
		"end":          avail.End.UTC().Format(time.RFC3339),
		"schedule": map[string]interface{}{
			"reference": scheduleRef,
		},
	}

	if avail.SlotType != "" {
		slot["appointmentType"] = map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://terminology.hl7.org/CodeSystem/v2-0276",
					"code":   avail.SlotType,
				},
			},
		}
	}

	if avail.ServiceType != "" {
		slot["serviceType"] = []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"code": avail.ServiceType,
					},
				},
			},
		}
	}

	if avail.Specialty != "" {
		slot["specialty"] = []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"code": avail.Specialty,
					},
				},
			},
		}
	}

	if avail.Comment != "" {
		slot["comment"] = avail.Comment
	}

	return slot
}

// BuildScheduleResource creates a FHIR Schedule resource.
func BuildScheduleResource(id, actorRef, serviceType, specialty string) map[string]interface{} {
	sched := map[string]interface{}{
		"resourceType": "Schedule",
		"id":           id,
		"active":       true,
		"actor": []interface{}{
			map[string]interface{}{
				"reference": actorRef,
			},
		},
	}

	if serviceType != "" {
		sched["serviceType"] = []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"code": serviceType,
					},
				},
			},
		}
	}

	if specialty != "" {
		sched["specialty"] = []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"code": specialty,
					},
				},
			},
		}
	}

	return sched
}

// BuildAvailabilityBundle creates a FHIR Bundle of available slots.
func BuildAvailabilityBundle(result *AvailabilityResult, baseURL string) map[string]interface{} {
	entries := make([]interface{}, 0, len(result.Slots)+len(result.Schedules))

	for _, slot := range result.Slots {
		id, _ := slot["id"].(string)
		entry := map[string]interface{}{
			"fullUrl":  fmt.Sprintf("%s/Slot/%s", baseURL, id),
			"resource": slot,
			"search": map[string]interface{}{
				"mode": "match",
			},
		}
		entries = append(entries, entry)
	}

	for _, sched := range result.Schedules {
		id, _ := sched["id"].(string)
		entry := map[string]interface{}{
			"fullUrl":  fmt.Sprintf("%s/Schedule/%s", baseURL, id),
			"resource": sched,
			"search": map[string]interface{}{
				"mode": "include",
			},
		}
		entries = append(entries, entry)
	}

	return map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "searchset",
		"total":        result.Total,
		"entry":        entries,
	}
}

// SlotSearchHandler returns an HTTP handler for GET /fhir/Slot with availability search.
func SlotSearchHandler(store AvailabilityStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		req, err := ParseAvailabilityRequest(c.QueryParams())
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeValue, err.Error(),
			))
		}

		// Default status to "free" if not specified for slot search.
		if len(req.Status) == 0 {
			req.Status = []SlotStatus{SlotStatusFree}
		}

		slots, err := store.FindSlots(c.Request().Context(), req)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, NewOperationOutcome(
				IssueSeverityError, IssueTypeProcessing, "slot search failed: "+err.Error(),
			))
		}

		result := &AvailabilityResult{
			Slots: slots,
			Total: len(slots),
		}

		bundle := BuildAvailabilityBundle(result, "")
		return c.JSON(http.StatusOK, bundle)
	}
}

// ScheduleAvailableHandler returns an HTTP handler for GET /fhir/Schedule/{id}/$available.
func ScheduleAvailableHandler(store AvailabilityStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		scheduleID := c.Param("id")
		if scheduleID == "" {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeRequired, "schedule id is required",
			))
		}

		// Verify the schedule exists.
		_, err := store.GetSchedule(c.Request().Context(), scheduleID)
		if err != nil {
			return c.JSON(http.StatusNotFound, NotFoundOutcome("Schedule", scheduleID))
		}

		req, err := ParseAvailabilityRequest(c.QueryParams())
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeValue, err.Error(),
			))
		}

		// Default status to "free".
		if len(req.Status) == 0 {
			req.Status = []SlotStatus{SlotStatusFree}
		}

		// Find all slots for this schedule.
		allSlots, err := store.FindSlots(c.Request().Context(), req)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, NewOperationOutcome(
				IssueSeverityError, IssueTypeProcessing, "availability search failed: "+err.Error(),
			))
		}

		// Filter to slots belonging to this schedule.
		schedRef := "Schedule/" + scheduleID
		var matchedSlots []map[string]interface{}
		for _, slot := range allSlots {
			ref := extractScheduleRef(slot)
			if ref == schedRef {
				matchedSlots = append(matchedSlots, slot)
			}
		}

		if matchedSlots == nil {
			matchedSlots = make([]map[string]interface{}, 0)
		}

		result := &AvailabilityResult{
			Slots: matchedSlots,
			Total: len(matchedSlots),
		}

		bundle := BuildAvailabilityBundle(result, "")
		return c.JSON(http.StatusOK, bundle)
	}
}

// FindAvailabilityHandler returns an HTTP handler for GET /fhir/Slot/$find.
func FindAvailabilityHandler(store AvailabilityStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		req, err := ParseAvailabilityRequest(c.QueryParams())
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeValue, err.Error(),
			))
		}

		issues := ValidateAvailabilityRequest(req)
		if len(issues) > 0 {
			return c.JSON(http.StatusBadRequest, MultiValidationOutcome(issues))
		}

		// Default status to "free".
		if len(req.Status) == 0 {
			req.Status = []SlotStatus{SlotStatusFree}
		}

		slots, err := store.FindSlots(c.Request().Context(), req)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, NewOperationOutcome(
				IssueSeverityError, IssueTypeProcessing, "find operation failed: "+err.Error(),
			))
		}

		result := &AvailabilityResult{
			Slots: slots,
			Total: len(slots),
		}

		// Include schedules if requested.
		if req.IncludeSchedule {
			seen := make(map[string]bool)
			for _, slot := range slots {
				schedRef := extractScheduleRef(slot)
				if schedRef == "" {
					continue
				}
				schedID := schedRef
				if idx := strings.LastIndex(schedRef, "/"); idx >= 0 {
					schedID = schedRef[idx+1:]
				}
				if seen[schedID] {
					continue
				}
				seen[schedID] = true
				sched, err := store.GetSchedule(c.Request().Context(), schedID)
				if err == nil {
					result.Schedules = append(result.Schedules, sched)
				}
			}
		}

		bundle := BuildAvailabilityBundle(result, "")
		return c.JSON(http.StatusOK, bundle)
	}
}

// conflictCheckBody is the expected JSON body for the $check-conflict operation.
type conflictCheckBody struct {
	Start string `json:"start"`
	End   string `json:"end"`
	Actor string `json:"actor"`
}

// CheckConflictHandler returns an HTTP handler for POST /fhir/Slot/$check-conflict.
func CheckConflictHandler(store AvailabilityStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		var body conflictCheckBody
		if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeValue, "invalid request body: "+err.Error(),
			))
		}

		if body.Start == "" || body.End == "" || body.Actor == "" {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeRequired, "start, end, and actor are required",
			))
		}

		start, err := parseFlexibleTime(body.Start)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeValue, "invalid start time: "+err.Error(),
			))
		}

		end, err := parseFlexibleTime(body.End)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeValue, "invalid end time: "+err.Error(),
			))
		}

		result, err := store.CheckConflicts(c.Request().Context(), start, end, body.Actor)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, NewOperationOutcome(
				IssueSeverityError, IssueTypeProcessing, "conflict check failed: "+err.Error(),
			))
		}

		params := []interface{}{
			map[string]interface{}{
				"name":         "hasConflict",
				"valueBoolean": result.HasConflict,
			},
			map[string]interface{}{
				"name":         "conflictCount",
				"valueInteger": len(result.Conflicts),
			},
			map[string]interface{}{
				"name":        "message",
				"valueString": result.Message,
			},
		}

		response := map[string]interface{}{
			"resourceType": "Parameters",
			"parameter":    params,
		}

		return c.JSON(http.StatusOK, response)
	}
}

// DefaultAvailabilityRules returns example availability rules for common scenarios.
func DefaultAvailabilityRules() []AvailabilityRule {
	weekdays := []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}

	return []AvailabilityRule{
		{
			DaysOfWeek:   weekdays,
			StartTime:    "08:00",
			EndTime:      "17:00",
			SlotDuration: 30,
			BreakStart:   "12:00",
			BreakEnd:     "13:00",
		},
		{
			DaysOfWeek:   weekdays,
			StartTime:    "09:00",
			EndTime:      "12:00",
			SlotDuration: 15,
		},
		{
			DaysOfWeek:   []time.Weekday{time.Saturday},
			StartTime:    "09:00",
			EndTime:      "13:00",
			SlotDuration: 30,
		},
	}
}
