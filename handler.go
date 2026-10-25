package bmur2

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const maxUpdateBodyBytes = 4 << 20

func writeJSON(w http.ResponseWriter, status int, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	cfg, err := loadConfig()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "invalid configuration"})
		return
	}

	incoming := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
	if subtle.ConstantTimeCompare([]byte(incoming), []byte(cfg.WebhookSecret)) != 1 {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxUpdateBodyBytes+1))
	if err != nil || len(body) > maxUpdateBodyBytes {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid update payload"})
		return
	}

	var update models.Update
	if err := json.Unmarshal(body, &update); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid update payload"})
		return
	}

	b, err := getBot()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to initialize bot"})
		return
	}

	if err := processUpdate(r.Context(), b, &update); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to process update"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func processUpdate(ctx context.Context, b *tgbot.Bot, update *models.Update) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("panic while processing update %d: %v", update.ID, rec)
		}
	}()
	b.ProcessUpdate(ctx, update)
	return nil
}
