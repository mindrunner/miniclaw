package internal

import (
	"strings"
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func TestSplitMessage_Short(t *testing.T) {
	text := "hello world"
	chunks := splitMessage(text)
	if len(chunks) != 1 || chunks[0] != text {
		t.Errorf("expected single chunk %q, got %v", text, chunks)
	}
}

func TestSplitMessage_ExactLimit(t *testing.T) {
	text := strings.Repeat("a", maxMessageLength)
	chunks := splitMessage(text)
	if len(chunks) != 1 {
		t.Errorf("expected 1 chunk, got %d", len(chunks))
	}
}

func TestSplitMessage_SplitsOnNewline(t *testing.T) {
	// Create a message that's over the limit with a newline near the boundary
	first := strings.Repeat("a", maxMessageLength-10)
	second := strings.Repeat("b", 20)
	text := first + "\n" + second

	chunks := splitMessage(text)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	if chunks[0] != first+"\n" {
		t.Errorf("first chunk should end at newline, got len=%d", len(chunks[0]))
	}
	if chunks[1] != second {
		t.Errorf("second chunk = %q, want %q", chunks[1], second)
	}
}

func TestSplitMessage_NoNewlineFallback(t *testing.T) {
	// A single long line with no newlines, must hard cut at maxMessageLength
	text := strings.Repeat("x", maxMessageLength+100)
	chunks := splitMessage(text)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	if len(chunks[0]) != maxMessageLength {
		t.Errorf("first chunk len = %d, want %d", len(chunks[0]), maxMessageLength)
	}
	if len(chunks[1]) != 100 {
		t.Errorf("second chunk len = %d, want 100", len(chunks[1]))
	}
}

func TestSplitMessage_ManyChunks(t *testing.T) {
	// 3x the limit, split across newlines
	line := strings.Repeat("z", maxMessageLength-1) + "\n"
	text := line + line + line
	chunks := splitMessage(text)
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}
	for i, c := range chunks {
		if c != line {
			t.Errorf("chunk %d: got len=%d, want len=%d", i, len(c), len(line))
		}
	}
}

func TestSplitMessage_Empty(t *testing.T) {
	chunks := splitMessage("")
	if len(chunks) != 1 || chunks[0] != "" {
		t.Errorf("expected single empty chunk, got %v", chunks)
	}
}

func TestSenderName(t *testing.T) {
	tests := []struct {
		name string
		user *gotgbot.User
		want string
	}{
		{"nil user", nil, "Unknown"},
		{"first name only", &gotgbot.User{FirstName: "Alice"}, "Alice"},
		{"first and last name", &gotgbot.User{FirstName: "Alice", LastName: "Smith"}, "Alice Smith"},
		{"username only", &gotgbot.User{Username: "alice123"}, "alice123"},
		{"first name takes precedence over username", &gotgbot.User{FirstName: "Alice", Username: "alice123"}, "Alice"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := senderName(tt.user); got != tt.want {
				t.Errorf("senderName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractFileID(t *testing.T) {
	tests := []struct {
		name       string
		msg        *gotgbot.Message
		wantFileID string
		wantName   string
	}{
		{
			name:       "photo takes highest resolution",
			msg:        &gotgbot.Message{Photo: []gotgbot.PhotoSize{{FileId: "small"}, {FileId: "large"}}},
			wantFileID: "large",
			wantName:   "",
		},
		{
			name:       "document with filename",
			msg:        &gotgbot.Message{Document: &gotgbot.Document{FileId: "doc1", FileName: "report.pdf"}},
			wantFileID: "doc1",
			wantName:   "report.pdf",
		},
		{
			name:       "video with filename",
			msg:        &gotgbot.Message{Video: &gotgbot.Video{FileId: "vid1", FileName: "clip.mp4"}},
			wantFileID: "vid1",
			wantName:   "clip.mp4",
		},
		{
			name:       "audio with filename",
			msg:        &gotgbot.Message{Audio: &gotgbot.Audio{FileId: "aud1", FileName: "song.mp3"}},
			wantFileID: "aud1",
			wantName:   "song.mp3",
		},
		{
			name:       "voice has no filename",
			msg:        &gotgbot.Message{Voice: &gotgbot.Voice{FileId: "voice1"}},
			wantFileID: "voice1",
			wantName:   "",
		},
		{
			name:       "empty message",
			msg:        &gotgbot.Message{},
			wantFileID: "",
			wantName:   "",
		},
		{
			name:       "photo takes priority over document",
			msg:        &gotgbot.Message{Photo: []gotgbot.PhotoSize{{FileId: "photo1"}}, Document: &gotgbot.Document{FileId: "doc1"}},
			wantFileID: "photo1",
			wantName:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotName := extractFileID(tt.msg)
			if gotID != tt.wantFileID {
				t.Errorf("fileID = %q, want %q", gotID, tt.wantFileID)
			}
			if gotName != tt.wantName {
				t.Errorf("fileName = %q, want %q", gotName, tt.wantName)
			}
		})
	}
}
