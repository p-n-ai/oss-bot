package github_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gh "github.com/p-n-ai/oss-bot/internal/github"
	"github.com/p-n-ai/oss-bot/internal/parser"
)

func TestVerifySignature_Valid(t *testing.T) {
	secret := "test-secret"
	body := `{"action":"created","comment":{"body":"@oss-bot add teaching notes for F1-01"}}`
	signature := computeHMAC(body, secret)

	err := gh.VerifySignature([]byte(body), "sha256="+signature, secret)
	if err != nil {
		t.Errorf("VerifySignature() error = %v", err)
	}
}

func TestVerifySignature_Invalid(t *testing.T) {
	err := gh.VerifySignature([]byte("body"), "sha256=invalid", "secret")
	if err == nil {
		t.Error("VerifySignature() should fail with invalid signature")
	}
}

func TestVerifySignature_MissingPrefix(t *testing.T) {
	err := gh.VerifySignature([]byte("body"), "invalidsig", "secret")
	if err == nil {
		t.Error("VerifySignature() should fail with missing sha256= prefix")
	}
}

func TestWebhookHandler_IssueComment(t *testing.T) {
	var got parser.BotCommand
	handler := gh.NewWebhookHandler("test-secret", func(cmd parser.BotCommand) error {
		got = cmd
		return nil
	})

	body := `{
		"action": "created",
		"comment": {
			"body": "@oss-bot add teaching notes for F1-01",
			"user": {"login": "testuser"}
		},
		"issue": {
			"number": 42
		},
		"repository": {
			"full_name": "p-n-ai/oss"
		}
	}`

	secret := "test-secret"
	signature := "sha256=" + computeHMAC(body, secret)

	req := httptest.NewRequest("POST", "/webhook", strings.NewReader(body))
	req.Header.Set("X-Hub-Signature-256", signature)
	req.Header.Set("X-GitHub-Event", "issue_comment")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		b, _ := io.ReadAll(rr.Body)
		t.Errorf("Status = %d, want %d. Body: %s", rr.Code, http.StatusOK, string(b))
	}

	if got.Action != "add" {
		t.Errorf("Action = %q, want %q", got.Action, "add")
	}
	if got.User != "testuser" {
		t.Errorf("User = %q, want %q", got.User, "testuser")
	}
	if got.IssueNum != 42 {
		t.Errorf("IssueNum = %d, want 42", got.IssueNum)
	}
	if got.RepoFullName != "p-n-ai/oss" {
		t.Errorf("RepoFullName = %q, want %q", got.RepoFullName, "p-n-ai/oss")
	}
}

func TestWebhookHandler_InvalidSignature(t *testing.T) {
	handler := gh.NewWebhookHandler("test-secret", nil)

	req := httptest.NewRequest("POST", "/webhook", strings.NewReader("{}"))
	req.Header.Set("X-Hub-Signature-256", "sha256=invalid")
	req.Header.Set("X-GitHub-Event", "issue_comment")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestWebhookHandler_Ping(t *testing.T) {
	handler := gh.NewWebhookHandler("test-secret", nil)

	body := `{"zen":"Keep it logically awesome."}`
	signature := "sha256=" + computeHMAC(body, "test-secret")

	req := httptest.NewRequest("POST", "/webhook", strings.NewReader(body))
	req.Header.Set("X-Hub-Signature-256", signature)
	req.Header.Set("X-GitHub-Event", "ping")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestWebhookHandler_NoBotMention(t *testing.T) {
	called := false
	handler := gh.NewWebhookHandler("test-secret", func(cmd parser.BotCommand) error {
		called = true
		return nil
	})

	body := `{
		"action": "created",
		"comment": {
			"body": "Just a normal comment with no bot mention",
			"user": {"login": "someone"}
		},
		"issue": {"number": 1},
		"repository": {"full_name": "p-n-ai/oss"}
	}`

	signature := "sha256=" + computeHMAC(body, "test-secret")
	req := httptest.NewRequest("POST", "/webhook", strings.NewReader(body))
	req.Header.Set("X-Hub-Signature-256", signature)
	req.Header.Set("X-GitHub-Event", "issue_comment")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusOK)
	}
	if called {
		t.Error("handler should not be called when no @oss-bot mention")
	}
}

func computeHMAC(message, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
