package command

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rakuv3r/codi/internal/generator"
)

// promptAITools prompts user to select AI tools interactively.
func promptAITools() []generator.AIToolConfig {
	available := generator.AvailableAITools()

	// Build options
	options := make([]string, len(available))
	for i, tool := range available {
		options[i] = fmt.Sprintf("%s (%s)", tool.Name, tool.OutputPath)
	}

	// Prompt multi-select
	selected := promptMultiSelect("Which AI tools do you use?", options)

	// If user didn't select anything, default to Claude
	if len(selected) == 0 {
		selected = []int{0} // Default to first option (claude)
	}

	// Build result
	result := make([]generator.AIToolConfig, 0, len(selected))
	for _, idx := range selected {
		tool := available[idx]
		tool.Enabled = true
		result = append(result, tool)
	}

	return result
}

// promptMultiSelect prompts user to select multiple options.
// Returns slice of selected indices.
func promptMultiSelect(question string, options []string) []int {
	fmt.Println(question)
	for i, opt := range options {
		fmt.Printf("  %d) %s\n", i+1, opt)
	}
	fmt.Print("→ (comma-separated, e.g., 1,2): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return []int{}
	}

	parts := strings.Split(input, ",")
	selected := []int{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if num, err := strconv.Atoi(part); err == nil && num >= 1 && num <= len(options) {
			selected = append(selected, num-1) // Convert to 0-based index
		}
	}

	return selected
}
