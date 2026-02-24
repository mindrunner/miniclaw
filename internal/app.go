package internal

import (
	"context"
	"log"
	"sync"
	"time"

	"goclaw/internal/models"
)

type App struct {
	config       Config
	bot          *TelegramBot
	agentRunner  *AgentRunner
	scheduler    *Scheduler
	activeAgents sync.Map // map[int64]context.CancelFunc — chatID → cancel func
}

func NewApp(cfg Config) *App {
	a := &App{config: cfg}

	sessions, err := NewSessionStore(cfg.DataDir + "/sessions.json")
	if err != nil {
		log.Fatalf("failed to load session store: %v", err)
	}

	a.agentRunner = NewAgentRunner(cfg, sessions)

	bot, err := NewTelegramBot(cfg.TelegramToken, a.onMessage)
	if err != nil {
		log.Fatalf("failed to create telegram bot: %v", err)
	}
	a.bot = bot
	a.bot.onCancel = a.cancelAgent

	a.scheduler = NewScheduler(cfg, a.agentRunner, a.bot)

	return a
}

func (a *App) Start(ctx context.Context) error {
	if err := a.bot.Start(); err != nil {
		return err
	}
	log.Println("telegram bot started")

	go a.scheduler.Start(ctx)
	log.Println("scheduler started")

	<-ctx.Done()

	a.bot.Stop()
	log.Println("shutting down")
	return nil
}

func (a *App) onMessage(msg models.Message) {
	if !a.isAllowed(msg.ChatID) {
		log.Printf("message from unauthorized chat %d, ignoring", msg.ChatID)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	if _, loaded := a.activeAgents.LoadOrStore(msg.ChatID, cancel); loaded {
		cancel()
		log.Printf("agent already running for chat %d, ignoring message", msg.ChatID)
		return
	}

	input := models.AgentInput{
		ChatID:         msg.ChatID,
		MessageID:      msg.MessageID,
		Prompt:         msg.Content,
		ReplyToSender:  msg.ReplyToSender,
		ReplyToContent: msg.ReplyToContent,
	}

	go a.startAgent(ctx, cancel, input)
}

func (a *App) startAgent(ctx context.Context, cancel context.CancelFunc, input models.AgentInput) {
	defer a.activeAgents.Delete(input.ChatID)
	defer cancel()

	// Send typing indicator every 4s until the agent finishes
	go func() {
		a.bot.SendTyping(input.ChatID)
		ticker := time.NewTicker(4 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.bot.SendTyping(input.ChatID)
			}
		}
	}()

	output, err := a.agentRunner.Run(ctx, input)
	if err != nil {
		if ctx.Err() == context.Canceled {
			log.Printf("agent cancelled for chat %d", input.ChatID)
			a.bot.SendReply(input.ChatID, input.MessageID, "Cancelled.")
			return
		}
		log.Printf("agent error for chat %d: %v", input.ChatID, err)
		a.bot.SendMessage(input.ChatID, "Sorry, I encountered an error. Check logs for details.")
		return
	}

	if output.Result != "" {
		if err := a.bot.SendMessage(input.ChatID, output.Result); err != nil {
			log.Printf("error sending message to chat %d: %v", input.ChatID, err)
		}
	}
}

func (a *App) cancelAgent(chatID int64) {
	val, loaded := a.activeAgents.Load(chatID)
	if !loaded {
		a.bot.SendMessage(chatID, "Nothing to cancel.")
		return
	}
	if cancel, ok := val.(context.CancelFunc); ok {
		cancel()
	}
}

func (a *App) isAllowed(chatID int64) bool {
	if len(a.config.AllowedChatIDs) == 0 {
		return true
	}
	for _, id := range a.config.AllowedChatIDs {
		if id == chatID {
			return true
		}
	}
	return false
}
