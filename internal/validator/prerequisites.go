package validator

import "fmt"

// DetectCycles finds circular dependencies in a prerequisite graph using DFS.
// Returns a list of cycle descriptions (empty if no cycles).
func DetectCycles(graph map[string][]string) []string {
	var cycles []string

	const (
		white = 0 // unvisited
		gray  = 1 // in current path
		black = 2 // fully processed
	)

	colors := make(map[string]int)
	parent := make(map[string]string)

	var dfs func(node string) bool
	dfs = func(node string) bool {
		colors[node] = gray

		for _, dep := range graph[node] {
			if colors[dep] == gray {
				// Found a cycle — reconstruct it
				cycle := reconstructCycle(parent, node, dep)
				cycles = append(cycles, fmt.Sprintf("cycle detected: %s", cycle))
				return true
			}
			if colors[dep] == white {
				parent[dep] = node
				if dfs(dep) {
					return false // Continue looking for more cycles
				}
			}
		}

		colors[node] = black
		return false
	}

	for node := range graph {
		if colors[node] == white {
			dfs(node)
		}
	}

	return cycles
}

// reconstructCycle builds a human-readable cycle string.
func reconstructCycle(parent map[string]string, from, to string) string {
	path := []string{to, from}
	current := from
	for current != to {
		p, ok := parent[current]
		if !ok {
			break
		}
		path = append(path, p)
		current = p
	}
	// Reverse
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	result := ""
	for i, p := range path {
		if i > 0 {
			result += " → "
		}
		result += p
	}
	return result
}

// FindMissingPrereqs returns prerequisites that reference non-existent topic IDs.
func FindMissingPrereqs(graph map[string][]string) []string {
	var missing []string
	for topic, prereqs := range graph {
		for _, prereq := range prereqs {
			if _, exists := graph[prereq]; !exists {
				missing = append(missing, fmt.Sprintf("%s requires %s (not found)", topic, prereq))
			}
		}
	}
	return missing
}
