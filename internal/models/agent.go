package models

type AgentInput struct {
	ChatID         int64
	MessageID      int64  // telegram message ID of the user's message
	Prompt         string
	ReplyToSender  string // who sent the message being replied to (empty if not a reply)
	ReplyToContent string // content of the message being replied to (empty if not a reply)
}

type AgentOutput struct {
	Result string
	Status string // "success" or "error"
	Error  string
}
