package telegram

import (
	"errors"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramNotifier struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

const (
	telegramRetryCount = 3
	telegramRetryDelay = 2 * time.Second
)

func NewTelegramNotifier(token string, chatID int64) (*TelegramNotifier, error) {
	bot, err := tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, &http.Client{
		Timeout: 10 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return &TelegramNotifier{
		bot:    bot,
		chatID: chatID,
	}, nil
}

func (n *TelegramNotifier) SendText(text string) error {
	if n == nil || n.bot == nil {
		return errors.New("telegram notifier not initialized")
	}
	msgText := strings.TrimSpace(text)
	if msgText == "" {
		return nil
	}
	if len(msgText) > 3800 {
		msgText = msgText[:3800] + "\n..."
	}

	var lastErr error
	for i := 0; i < telegramRetryCount; i++ {
		msg := tgbotapi.NewMessage(n.chatID, msgText)
		_, err := n.bot.Send(msg)
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(telegramRetryDelay)
	}
	return lastErr
}
