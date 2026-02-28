package internal

import "testing"

func TestStatusTracker_Empty(t *testing.T) {
	s := newStatusTracker()
	if r := s.Render(); r != "" {
		t.Errorf("Render() on empty = %q, want empty", r)
	}
	if r := s.RenderDone(); r != "" {
		t.Errorf("RenderDone() on empty = %q, want empty", r)
	}
	if r := s.RenderFinal(); r != "" {
		t.Errorf("RenderFinal() on empty = %q, want empty", r)
	}
}

func TestStatusTracker_Add_ReturnValue(t *testing.T) {
	s := newStatusTracker()

	// First add returns true
	if got := s.Add("Read", "<code>main.go</code>"); !got {
		t.Error("first Add should return true")
	}

	// Second add with different entry returns false
	if got := s.Add("Bash", "<code>go test</code>"); got {
		t.Error("second Add should return false")
	}

	// Duplicate consecutive entry returns false
	if got := s.Add("Bash", "<code>go test</code>"); got {
		t.Error("duplicate Add should return false")
	}

	// Same emoji but different label is not a duplicate
	if got := s.Add("Bash", "<code>go build</code>"); got {
		t.Error("non-first Add should return false")
	}
}

func TestStatusTracker_Add_DeduplicateConsecutive(t *testing.T) {
	s := newStatusTracker()
	s.Add("Read", "<code>a.go</code>")
	s.Add("Read", "<code>a.go</code>")
	s.Add("Read", "<code>a.go</code>")

	// Should only have 1 entry
	want := "📄 <code>a.go</code> 🟡"
	if got := s.Render(); got != want {
		t.Errorf("Render() = %q, want %q", got, want)
	}
}

func TestStatusTracker_Add_ExitPlanMode(t *testing.T) {
	s := newStatusTracker()

	// ExitPlanMode on empty tracker returns false (no entries means len==0, returns true... let me re-check)
	// Actually: if toolName == "ExitPlanMode" { return len(s.entries) == 0 }
	// On empty: returns true (meaning "this is the first status update, send it")
	// But it doesn't add an entry.
	got := s.Add("ExitPlanMode", "")
	if !got {
		t.Error("ExitPlanMode on empty tracker should return true")
	}
	// Still empty
	if r := s.Render(); r != "" {
		t.Errorf("ExitPlanMode should not add entry, Render() = %q", r)
	}

	// ExitPlanMode on non-empty tracker returns false
	s.Add("Read", "file")
	if got := s.Add("ExitPlanMode", ""); got {
		t.Error("ExitPlanMode on non-empty tracker should return false")
	}
}

func TestStatusTracker_Add_TodoWriteEmptyLabel(t *testing.T) {
	s := newStatusTracker()

	// TodoWrite with empty label behaves like ExitPlanMode
	got := s.Add("TodoWrite", "")
	if !got {
		t.Error("TodoWrite with empty label on empty tracker should return true")
	}
	if r := s.Render(); r != "" {
		t.Errorf("TodoWrite with empty label should not add entry, Render() = %q", r)
	}
}

func TestStatusTracker_Add_UnknownTool(t *testing.T) {
	s := newStatusTracker()
	s.Add("SomeNewTool", "doing stuff")

	want := "⚙️ doing stuff 🟡"
	if got := s.Render(); got != want {
		t.Errorf("Render() = %q, want %q", got, want)
	}
}

func TestStatusTracker_Add_EmptyLabelUsesToolName(t *testing.T) {
	s := newStatusTracker()
	s.Add("Read", "")

	want := "📄 Read 🟡"
	if got := s.Render(); got != want {
		t.Errorf("Render() = %q, want %q", got, want)
	}
}

func TestStatusTracker_Render(t *testing.T) {
	s := newStatusTracker()
	s.Add("Read", "<code>main.go</code>")
	s.Add("Bash", "<code>go test</code>")
	s.Add("WebSearch", "golang errors")

	want := "📄 <code>main.go</code>\n⚡ <code>go test</code>\n🌐 golang errors 🟡"
	if got := s.Render(); got != want {
		t.Errorf("Render() =\n%s\nwant:\n%s", got, want)
	}
}

func TestStatusTracker_RenderDone(t *testing.T) {
	s := newStatusTracker()
	s.Add("Read", "<code>main.go</code>")
	s.Add("Bash", "<code>go test</code>")

	want := "📄 <code>main.go</code>\n⚡ <code>go test</code>\n"
	if got := s.RenderDone(); got != want {
		t.Errorf("RenderDone() =\n%q\nwant:\n%q", got, want)
	}
}

func TestStatusTracker_RenderFinal(t *testing.T) {
	s := newStatusTracker()
	s.Add("Read", "<code>main.go</code>")
	s.Add("Bash", "<code>go test</code>")

	// RenderFinal trims trailing newlines from RenderDone
	want := "📄 <code>main.go</code>\n⚡ <code>go test</code>"
	if got := s.RenderFinal(); got != want {
		t.Errorf("RenderFinal() =\n%q\nwant:\n%q", got, want)
	}
}
