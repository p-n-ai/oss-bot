package pipeline

import (
	"fmt"
	"sync"
)

// ProgressReporter receives real-time updates during bulk import operations.
// Implementations: NoopReporter (tests), CLIReporter (terminal), BotReporter (GitHub comment), SSEReporter (web).
type ProgressReporter interface {
	// OnStart is called when processing begins.
	OnStart(totalItems int)
	// OnProgress is called when an item completes or advances a stage.
	OnProgress(current, total int, itemName string, status string)
	// OnComplete is called when all processing finishes.
	OnComplete(result *BulkResult)
	// OnError is called when an item fails.
	OnError(itemName string, err error)
}

// NoopReporter is a no-op ProgressReporter for use in tests or when progress
// reporting is not needed.
type NoopReporter struct{}

func (n *NoopReporter) OnStart(totalItems int)                              {}
func (n *NoopReporter) OnProgress(current, total int, name, status string) {}
func (n *NoopReporter) OnComplete(result *BulkResult)                      {}
func (n *NoopReporter) OnError(name string, err error)                     {}

// CLIReporter prints progress to the terminal and tracks counts for inspection.
type CLIReporter struct {
	mu         sync.Mutex
	current    int
	total      int
	errorCount int
}

// NewCLIReporter creates a CLIReporter.
func NewCLIReporter() *CLIReporter {
	return &CLIReporter{}
}

func (r *CLIReporter) OnStart(totalItems int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.total = totalItems
	r.current = 0
	fmt.Printf("Starting bulk import of %d items\n", totalItems)
}

func (r *CLIReporter) OnProgress(current, total int, itemName, status string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.current = current
	r.total = total
	fmt.Printf("[%d/%d] %s: %s\n", current, total, itemName, status)
}

func (r *CLIReporter) OnComplete(result *BulkResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if result != nil {
		fmt.Printf("Bulk import complete: %d topics processed in %s\n",
			len(result.Topics), result.Duration)
	} else {
		fmt.Println("Bulk import complete")
	}
}

func (r *CLIReporter) OnError(itemName string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errorCount++
	if err != nil {
		fmt.Printf("ERROR processing %s: %v\n", itemName, err)
	} else {
		fmt.Printf("ERROR processing %s\n", itemName)
	}
}

// Current returns the number of items processed so far.
func (r *CLIReporter) Current() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.current
}

// Total returns the total number of items.
func (r *CLIReporter) Total() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.total
}

// ErrorCount returns the number of errors recorded.
func (r *CLIReporter) ErrorCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.errorCount
}
