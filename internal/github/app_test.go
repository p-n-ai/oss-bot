package github_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	gh "github.com/p-n-ai/oss-bot/internal/github"
)

func TestNewApp(t *testing.T) {
	keyPath := generateTestKey(t)

	app, err := gh.NewApp(12345, keyPath)
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
	if app == nil {
		t.Fatal("NewApp() returned nil")
	}
}

func TestNewApp_MissingKey(t *testing.T) {
	_, err := gh.NewApp(12345, "/nonexistent/key.pem")
	if err == nil {
		t.Error("NewApp() should fail with missing key file")
	}
}

func TestGenerateJWT(t *testing.T) {
	keyPath := generateTestKey(t)
	app, err := gh.NewApp(12345, keyPath)
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}

	token, err := app.GenerateJWT()
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}
	if token == "" {
		t.Error("GenerateJWT() returned empty token")
	}
}

// generateTestKey creates a temporary RSA private key for testing.
func generateTestKey(t *testing.T) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	pemBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyBytes}

	path := filepath.Join(t.TempDir(), "test-key.pem")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if err := pem.Encode(f, pemBlock); err != nil {
		t.Fatal(err)
	}

	return path
}
