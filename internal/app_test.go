package internal

import "testing"

func TestIsAllowed_EmptyList(t *testing.T) {
	a := &App{config: Config{AllowedChatIDs: nil}}

	if !a.isAllowed(12345) {
		t.Error("empty allowlist should permit all chats")
	}
}

func TestIsAllowed_Match(t *testing.T) {
	a := &App{config: Config{AllowedChatIDs: []int64{111, 222, 333}}}

	if !a.isAllowed(222) {
		t.Error("should allow chat ID in the list")
	}
}

func TestIsAllowed_NoMatch(t *testing.T) {
	a := &App{config: Config{AllowedChatIDs: []int64{111, 222, 333}}}

	if a.isAllowed(999) {
		t.Error("should reject chat ID not in the list")
	}
}

func TestIsAllowed_NegativeChatID(t *testing.T) {
	// Telegram group chat IDs are negative
	a := &App{config: Config{AllowedChatIDs: []int64{-100123456}}}

	if !a.isAllowed(-100123456) {
		t.Error("should allow negative chat IDs (group chats)")
	}
	if a.isAllowed(-100999999) {
		t.Error("should reject other negative chat IDs")
	}
}
