package models

// Message is the internal representation passed from TelegramBot to App.
// Not persisted. Used only for allowlist check and agent dispatch.
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
