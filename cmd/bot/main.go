package main

import (
	"fmt"
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Placeholder — webhook server implemented in Week 5
	fmt.Println("oss-bot server — not yet implemented")
	fmt.Println("See: go run ./cmd/oss for CLI commands")
	os.Exit(0)
}
