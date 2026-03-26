package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/p-n-ai/oss-bot/internal/ai"
	gh "github.com/p-n-ai/oss-bot/internal/github"
	"github.com/p-n-ai/oss-bot/internal/output"
	"github.com/p-n-ai/oss-bot/internal/parser"
	"github.com/p-n-ai/oss-bot/internal/pipeline"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Required env vars
	appIDStr := os.Getenv("OSS_GITHUB_APP_ID")
	keyPath := os.Getenv("OSS_GITHUB_PRIVATE_KEY_PATH")
	webhookSecret := os.Getenv("OSS_GITHUB_WEBHOOK_SECRET")

	if appIDStr == "" || keyPath == "" || webhookSecret == "" {
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

	// AI provider
	aiProvider, err := ai.NewProvider(
		getEnvOr("OSS_AI_PROVIDER", "openai"),
		os.Getenv("OSS_AI_API_KEY"),
	)
	if err != nil {
		slog.Error("failed to initialize AI provider", "error", err)
		os.Exit(1)
	}

	// GitHub output writer
	repoOwner := getEnvOr("OSS_REPO_OWNER", "p-n-ai")
	repoName := getEnvOr("OSS_REPO_NAME", "oss")
	writer := output.NewGitHubWriter(repoOwner, repoName)

	// Shared pipeline (all bot commands route through this)
	p := pipeline.New(aiProvider, writer, "prompts/", os.Getenv("OSS_REPO_PATH"))

	srv := &botServer{
		app:       app,
		pipeline:  p,
		repoOwner: repoOwner,
		repoName:  repoName,
	}

	port := getEnvOr("OSS_PORT", "8090")
	webhookHandler := gh.NewWebhookHandler(webhookSecret, srv.handleCommand)

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

// botServer holds shared state for the bot HTTP handlers.
type botServer struct {
	app       *gh.App
	pipeline  *pipeline.Pipeline
	repoOwner string
	repoName  string
}

// handleCommand is called by the webhook handler for every parsed @oss-bot command.
// It routes to the shared pipeline and posts the PR link back to the issue.
func (s *botServer) handleCommand(cmd parser.BotCommand) error {
	ctx := context.Background()

	contribType := cmdContribType(cmd)
	if contribType == "" {
		slog.Warn("unhandled command", "action", cmd.Action, "content_type", cmd.ContentType)
		return nil
	}

	slog.Info("executing pipeline",
		"action", cmd.Action,
		"type", contribType,
		"topic", cmd.TopicPath,
		"user", cmd.User,
		"issue", cmd.IssueNum,
	)

	result, err := s.pipeline.Execute(ctx, pipeline.Request{
		TopicPath:        cmd.TopicPath,
		ContributionType: contribType,
		Mode:             pipeline.ModeCreatePR,
		Source:           "bot",
		Options:          cmd.Options,
	})
	if err != nil {
		slog.Error("pipeline failed", "error", err, "topic", cmd.TopicPath)
		return fmt.Errorf("pipeline failed: %w", err)
	}

	// Build the response comment.
	var msg string
	if result.PRNumber > 0 {
		msg = fmt.Sprintf(
			"I've generated %s for `%s` and opened #%d for review. Please check for accuracy.",
			contribType, cmd.TopicPath, result.PRNumber,
		)
	} else {
		msg = fmt.Sprintf("Generated %s for `%s`. (PR creation in progress)", contribType, cmd.TopicPath)
	}

	slog.Info("posting comment to issue", "issue", cmd.IssueNum, "pr", result.PRNumber)

	// Post the comment via the GitHub API using an installation token.
	token, err := s.installationToken(ctx, cmd.RepoFullName)
	if err != nil {
		// Don't fail the whole command if we can't post the comment.
		slog.Warn("failed to get installation token — skipping comment", "error", err)
		return nil
	}

	owner, repo := splitFullName(cmd.RepoFullName, s.repoOwner, s.repoName)
	if err := postIssueComment(token, owner, repo, cmd.IssueNum, msg); err != nil {
		slog.Warn("failed to post comment", "error", err, "issue", cmd.IssueNum)
	}

	return nil
}

// installationToken returns a short-lived GitHub App installation access token.
func (s *botServer) installationToken(ctx context.Context, repoFullName string) (string, error) {
	jwtToken, err := s.app.GenerateJWT()
	if err != nil {
		return "", fmt.Errorf("generating JWT: %w", err)
	}

	installID, err := repoInstallationID(ctx, jwtToken, repoFullName)
	if err != nil {
		return "", fmt.Errorf("getting installation: %w", err)
	}

	return createInstallationToken(ctx, jwtToken, installID)
}

// postIssueComment posts a comment to a GitHub issue using the REST API.
func postIssueComment(token, owner, repo string, issueNum int, body string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments", owner, repo, issueNum)
	payload, err := json.Marshal(map[string]string{"body": body})
	if err != nil {
		return fmt.Errorf("marshaling comment: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("posting comment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("GitHub API returned %s", resp.Status)
	}
	return nil
}

// repoInstallationID looks up the GitHub App installation ID for a repository.
func repoInstallationID(ctx context.Context, jwtToken, repoFullName string) (int64, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/installation", repoFullName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return 0, fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var result struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

// createInstallationToken exchanges a JWT for a GitHub App installation access token.
func createInstallationToken(ctx context.Context, jwtToken string, installationID int64) (string, error) {
	url := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", installationID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Token, nil
}

// cmdContribType maps a parsed bot command to a pipeline contribution type string.
func cmdContribType(cmd parser.BotCommand) string {
	if cmd.Action != "add" {
		return ""
	}
	switch cmd.ContentType {
	case "teaching notes":
		return "teaching_notes"
	case "assessments":
		return "assessments"
	case "examples":
		return "examples"
	}
	return ""
}

// splitFullName splits "owner/repo" into (owner, repo), falling back to defaults.
func splitFullName(fullName, defaultOwner, defaultRepo string) (string, string) {
	if idx := strings.Index(fullName, "/"); idx > 0 {
		return fullName[:idx], fullName[idx+1:]
	}
	return defaultOwner, defaultRepo
}

func getEnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
