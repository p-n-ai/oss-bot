package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/p-n-ai/oss-bot/internal/parser"
)

// CommandHandler processes a parsed bot command.
type CommandHandler func(cmd parser.BotCommand) error

// WebhookHandler handles incoming GitHub webhook events.
type WebhookHandler struct {
	secret  string
	handler CommandHandler
}

// NewWebhookHandler creates a new webhook handler with HMAC verification.
func NewWebhookHandler(secret string, handler CommandHandler) *WebhookHandler {
	return &WebhookHandler{secret: secret, handler: handler}
}

// ServeHTTP implements the http.Handler interface.
func (wh *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	signature := r.Header.Get("X-Hub-Signature-256")
	if err := VerifySignature(body, signature, wh.secret); err != nil {
		slog.Warn("webhook signature verification failed", "error", err)
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	eventType := r.Header.Get("X-GitHub-Event")

	switch eventType {
	case "issue_comment":
		wh.handleIssueComment(w, body)
	case "ping":
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "pong")
	default:
		slog.Debug("ignoring event type", "type", eventType)
		w.WriteHeader(http.StatusOK)
	}
}

func (wh *WebhookHandler) handleIssueComment(w http.ResponseWriter, body []byte) {
	var event struct {
		Action  string `json:"action"`
		Comment struct {
			Body string `json:"body"`
			User struct {
				Login string `json:"login"`
			} `json:"user"`
		} `json:"comment"`
		Issue struct {
			Number int `json:"number"`
		} `json:"issue"`
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
	}

	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if event.Action != "created" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if !strings.Contains(event.Comment.Body, "@oss-bot") {
		w.WriteHeader(http.StatusOK)
		return
	}

	cmd, err := parser.ParseCommand(event.Comment.Body)
	if err != nil {
		slog.Warn("failed to parse command", "error", err, "body", event.Comment.Body)
		w.WriteHeader(http.StatusOK)
		return
	}

	cmd.User = event.Comment.User.Login
	cmd.IssueNum = event.Issue.Number
	cmd.RepoFullName = event.Repository.FullName
	cmd.CommentBody = event.Comment.Body

	if wh.handler != nil {
		if err := wh.handler(*cmd); err != nil {
			slog.Error("command handler failed", "error", err, "command", cmd.Action)
		}
	}

	w.WriteHeader(http.StatusOK)
}

// VerifySignature validates the HMAC-SHA256 signature of a webhook payload.
func VerifySignature(body []byte, signature, secret string) error {
	if !strings.HasPrefix(signature, "sha256=") {
		return fmt.Errorf("invalid signature format (expected sha256=...)")
	}

	expectedMAC := strings.TrimPrefix(signature, "sha256=")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	actualMAC := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(actualMAC), []byte(expectedMAC)) {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}
