// Package github provides GitHub App authentication and API integration.
package github

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// App represents a GitHub App for authentication.
type App struct {
	AppID      int64
	PrivateKey *rsa.PrivateKey
}

// NewApp creates a GitHub App instance from an app ID and private key file.
func NewApp(appID int64, keyPath string) (*App, error) {
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("reading private key: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from %s", keyPath)
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing private key: %w", err)
	}

	return &App{
		AppID:      appID,
		PrivateKey: key,
	}, nil
}

// InstallationToken returns a short-lived GitHub App installation access token for the given repo.
// It generates a JWT, looks up the installation ID, then exchanges it for an access token.
func (a *App) InstallationToken(ctx context.Context, repoFullName string) (string, error) {
	jwtToken, err := a.GenerateJWT()
	if err != nil {
		return "", fmt.Errorf("generating JWT: %w", err)
	}

	installID, err := repoInstallationID(ctx, jwtToken, repoFullName)
	if err != nil {
		return "", fmt.Errorf("getting installation ID for %s: %w", repoFullName, err)
	}

	return createInstallationToken(ctx, jwtToken, installID)
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
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

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
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

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

// GenerateJWT creates a signed JWT for GitHub App authentication.
// The JWT is valid for 10 minutes (GitHub's maximum).
func (a *App) GenerateJWT() (string, error) {
	now := time.Now()

	claims := jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now.Add(-60 * time.Second)),
		ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
		Issuer:    fmt.Sprintf("%d", a.AppID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(a.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("signing JWT: %w", err)
	}

	return signed, nil
}
