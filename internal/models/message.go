package models

// Message is the internal representation passed from TelegramBot to App.
// Not persisted — used only for allowlist check and agent dispatch.
type Message struct {
	ChatID          int64
	MessageID       int64
	Sender          string
	Content         string
	FilePath        string
	ReplyToSender   string
	ReplyToContent  string
	ReplyToFilePath string
}
