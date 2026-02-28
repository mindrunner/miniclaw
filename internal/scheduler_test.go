package internal

import (
	"encoding/json"
	"testing"
	"time"

	"miniclaw/internal/models"
)

func TestCalculateNextRun_Once(t *testing.T) {
	s := &Scheduler{}
	task := models.Task{ScheduleType: "once", ScheduleValue: ""}

	if got := s.calculateNextRun(task); got != nil {
		t.Errorf("once task should return nil, got %q", *got)
	}
}

func TestCalculateNextRun_Cron(t *testing.T) {
	s := &Scheduler{}
	task := models.Task{ScheduleType: "cron", ScheduleValue: "0 9 * * *"} // daily at 9am

	got := s.calculateNextRun(task)
	if got == nil {
		t.Fatal("cron task should return non-nil next run")
	}

	parsed, err := time.Parse(time.RFC3339, *got)
	if err != nil {
		t.Fatalf("should be valid RFC3339: %v", err)
	}

	if !parsed.After(time.Now()) {
		t.Error("next run should be in the future")
	}
}

func TestCalculateNextRun_CronInvalid(t *testing.T) {
	s := &Scheduler{}
	task := models.Task{ScheduleType: "cron", ScheduleValue: "not a cron"}

	if got := s.calculateNextRun(task); got != nil {
		t.Errorf("invalid cron should return nil, got %q", *got)
	}
}

func TestCalculateNextRun_Interval(t *testing.T) {
	s := &Scheduler{}
	task := models.Task{ScheduleType: "interval", ScheduleValue: "30m"}

	before := time.Now()
	got := s.calculateNextRun(task)
	if got == nil {
		t.Fatal("interval task should return non-nil next run")
	}

	parsed, err := time.Parse(time.RFC3339, *got)
	if err != nil {
		t.Fatalf("should be valid RFC3339: %v", err)
	}

	expected := before.Add(30 * time.Minute)
	diff := parsed.Sub(expected)
	if diff < -2*time.Second || diff > 2*time.Second {
		t.Errorf("next run should be ~30m from now, got %v (diff=%v)", parsed, diff)
	}
}

func TestCalculateNextRun_IntervalInvalid(t *testing.T) {
	s := &Scheduler{}
	task := models.Task{ScheduleType: "interval", ScheduleValue: "not a duration"}

	if got := s.calculateNextRun(task); got != nil {
		t.Errorf("invalid interval should return nil, got %q", *got)
	}
}

func TestCalculateNextRun_UnknownType(t *testing.T) {
	s := &Scheduler{}
	task := models.Task{ScheduleType: "weekly", ScheduleValue: "monday"}

	if got := s.calculateNextRun(task); got != nil {
		t.Errorf("unknown type should return nil, got %q", *got)
	}
}

func TestTaskJSON_RoundTrip(t *testing.T) {
	nextRun := "2025-01-15T09:00:00Z"
	expires := "2025-12-31T23:59:59Z"
	task := models.Task{
		Prompt:        "check weather",
		ChatID:        12345,
		ScheduleType:  "cron",
		ScheduleValue: "0 9 * * *",
		Status:        "active",
		NextRun:       &nextRun,
		Expires:       &expires,
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded models.Task
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.Prompt != task.Prompt {
		t.Errorf("Prompt = %q, want %q", decoded.Prompt, task.Prompt)
	}
	if decoded.ChatID != task.ChatID {
		t.Errorf("ChatID = %d, want %d", decoded.ChatID, task.ChatID)
	}
	if decoded.ScheduleType != task.ScheduleType {
		t.Errorf("ScheduleType = %q, want %q", decoded.ScheduleType, task.ScheduleType)
	}
	if decoded.ScheduleValue != task.ScheduleValue {
		t.Errorf("ScheduleValue = %q, want %q", decoded.ScheduleValue, task.ScheduleValue)
	}
	if decoded.Status != task.Status {
		t.Errorf("Status = %q, want %q", decoded.Status, task.Status)
	}
	if decoded.NextRun == nil || *decoded.NextRun != nextRun {
		t.Errorf("NextRun = %v, want %q", decoded.NextRun, nextRun)
	}
	if decoded.Expires == nil || *decoded.Expires != expires {
		t.Errorf("Expires = %v, want %q", decoded.Expires, expires)
	}
}

func TestTaskJSON_OmitsFilename(t *testing.T) {
	task := models.Task{
		Filename:      "should-not-appear.json",
		Prompt:        "test",
		ChatID:        1,
		ScheduleType:  "once",
		ScheduleValue: "",
		Status:        "active",
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal to map error: %v", err)
	}

	if _, exists := raw["filename"]; exists {
		t.Error("Filename should be omitted from JSON (json:\"-\")")
	}
}

func TestTaskJSON_ExpiresOmittedWhenNil(t *testing.T) {
	task := models.Task{
		Prompt:       "test",
		ChatID:       1,
		ScheduleType: "once",
		Status:       "active",
		NextRun:      nil,
		Expires:      nil,
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if _, exists := raw["expires"]; exists {
		t.Error("Expires should be omitted when nil (omitempty)")
	}

	// next_run is NOT omitempty, so it should be present even when nil
	if _, exists := raw["next_run"]; !exists {
		t.Error("NextRun should be present even when nil (no omitempty)")
	}
}

func TestTaskJSON_FromExternalFormat(t *testing.T) {
	// Simulate what a task file actually looks like on disk
	raw := `{
		"prompt": "check server health",
		"chat_id": 99887766,
		"type": "interval",
		"value": "1h",
		"status": "active",
		"next_run": "2025-06-01T12:00:00Z",
		"expires": "2025-12-31T00:00:00Z"
	}`

	var task models.Task
	if err := json.Unmarshal([]byte(raw), &task); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if task.Prompt != "check server health" {
		t.Errorf("Prompt = %q", task.Prompt)
	}
	if task.ChatID != 99887766 {
		t.Errorf("ChatID = %d", task.ChatID)
	}
	if task.ScheduleType != "interval" {
		t.Errorf("ScheduleType = %q (should map from 'type')", task.ScheduleType)
	}
	if task.ScheduleValue != "1h" {
		t.Errorf("ScheduleValue = %q (should map from 'value')", task.ScheduleValue)
	}
	if task.Status != "active" {
		t.Errorf("Status = %q", task.Status)
	}
	if task.NextRun == nil || *task.NextRun != "2025-06-01T12:00:00Z" {
		t.Errorf("NextRun = %v", task.NextRun)
	}
	if task.Expires == nil || *task.Expires != "2025-12-31T00:00:00Z" {
		t.Errorf("Expires = %v", task.Expires)
	}
	if task.Filename != "" {
		t.Errorf("Filename should be empty after unmarshal, got %q", task.Filename)
	}
}
