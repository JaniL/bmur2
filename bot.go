package bmur2

import (
	"context"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var (
	apiServerURL = "https://api.telegram.org"

	botOnce sync.Once
	bot     *tgbot.Bot
	botErr  error

	runtimeConfig *Config
)

func getBot() (*tgbot.Bot, error) {
	botOnce.Do(func() {
		config, err := loadConfig()
		if err != nil {
			botErr = err
			return
		}
		runtimeConfig = config

		b, err := tgbot.New(config.BotToken,
			tgbot.WithNotAsyncHandlers(),
			tgbot.WithSkipGetMe(),
			tgbot.WithServerURL(apiServerURL),
			tgbot.WithHTTPClient(0, &http.Client{Timeout: 10 * time.Second}),
		)
		if err != nil {
			botErr = err
			return
		}

		b.RegisterHandlerRegexp(tgbot.HandlerTypeMessageText, regexp.MustCompile(`^/np(@\S+)?(\s|$)`), nowPlayingHandler)
		b.RegisterHandlerRegexp(tgbot.HandlerTypeMessageText, regexp.MustCompile(`^/recent(@\S+)?(\s|$)`), recentTracksHandler)

		bot = b
	})
	return bot, botErr
}

func nowPlayingHandler(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	msg := update.Message
	if msg == nil {
		return
	}
	track, err := lastfmClient().LatestTrack(ctx)
	if err != nil {
		log.Printf("np: lastfm error: %v", err)
		sendMessage(ctx, b, msg.Chat.ID, "Could not fetch current track right now.")
		return
	}
	if track == nil {
		sendMessage(ctx, b, msg.Chat.ID, "No recent tracks available right now.")
		return
	}
	sendHTMLMessage(ctx, b, msg.Chat.ID, FormatTrack(*track))
}

func recentTracksHandler(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	msg := update.Message
	if msg == nil {
		return
	}
	tracks, err := lastfmClient().RecentTracks(ctx, 15)
	if err != nil {
		log.Printf("recent: lastfm error: %v", err)
		sendMessage(ctx, b, msg.Chat.ID, "Could not fetch recent tracks right now.")
		return
	}
	if len(tracks) == 0 {
		sendMessage(ctx, b, msg.Chat.ID, "No recent tracks available right now.")
		return
	}
	sendHTMLMessage(ctx, b, msg.Chat.ID, FormatRecentTracks(tracks))
}

func lastfmClient() *LastfmClient {
	return NewLastfmClient(runtimeConfig.LastfmAPIKey, runtimeConfig.LastfmUser)
}

func sendMessage(ctx context.Context, b *tgbot.Bot, chatID int64, text string) {
	params := &tgbot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	}
	if _, err := b.SendMessage(ctx, params); err != nil {
		log.Printf("send message to chat %d: %v", chatID, err)
	}
}

func sendHTMLMessage(ctx context.Context, b *tgbot.Bot, chatID int64, text string) {
	isDisabled := true
	params := &tgbot.SendMessageParams{
		ChatID:             chatID,
		Text:               text,
		ParseMode:          models.ParseModeHTML,
		LinkPreviewOptions: &models.LinkPreviewOptions{IsDisabled: &isDisabled},
	}
	if _, err := b.SendMessage(ctx, params); err != nil {
		log.Printf("send html message to chat %d: %v", chatID, err)
	}
}
