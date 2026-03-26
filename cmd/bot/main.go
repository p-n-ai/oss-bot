package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	gh "github.com/p-n-ai/oss-bot/internal/github"
	"github.com/p-n-ai/oss-bot/internal/parser"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	appIDStr := os.Getenv("OSS_GITHUB_APP_ID")
	keyPath := os.Getenv("OSS_GITHUB_PRIVATE_KEY_PATH")
	webhookSecret := os.Getenv("OSS_GITHUB_WEBHOOK_SECRET")
	port := os.Getenv("OSS_PORT")
	if port == "" {
		port = "8090"
	}

	if appIDStr == "" || keyPath == "" || webhookSecret == "" {
		slog.Error("missing required environment variables",
			"OSS_GITHUB_APP_ID", appIDStr != "",
			"OSS_GITHUB_PRIVATE_KEY_PATH", keyPath != "",
			"OSS_GITHUB_WEBHOOK_SECRET", webhookSecret != "",
		)
		fmt.Fprintln(os.Stderr, "Required: OSS_GITHUB_APP_ID, OSS_GITHUB_PRIVATE_KEY_PATH, OSS_GITHUB_WEBHOOK_SECRET")
		os.Exit(1)
	}

	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		slog.Error("invalid OSS_GITHUB_APP_ID", "value", appIDStr, "error", err)
		os.Exit(1)
	}

	app, err := gh.NewApp(appID, keyPath)
	if err != nil {
		slog.Error("failed to initialize GitHub App", "error", err)
		os.Exit(1)
	}
	slog.Info("GitHub App initialized", "app_id", app.AppID)

	webhookHandler := gh.NewWebhookHandler(webhookSecret, func(cmd parser.BotCommand) error {
		slog.Info("received bot command",
			"action", cmd.Action,
			"content_type", cmd.ContentType,
			"topic", cmd.TopicPath,
			"user", cmd.User,
			"issue", cmd.IssueNum,
			"repo", cmd.RepoFullName,
		)
		// Pipeline integration will be wired in Day 22+
		return nil
	})

	mux := http.NewServeMux()
	mux.Handle("POST /webhook", webhookHandler)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	addr := ":" + port
	slog.Info("oss-bot server starting", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
