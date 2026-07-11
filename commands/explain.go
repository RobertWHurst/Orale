package commands

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/RobertWHurst/orale"
	"github.com/spf13/cobra"
)

var explainCmd = &cobra.Command{
	Use:                "explain <app-name> [flags...]",
	Short:              "Explain loaded configuration values",
	Long:               "Load configuration for an application and display all values with their sources in a table. Any flags after the app name will be passed to the config loader.",
	Args:               cobra.MinimumNArgs(1),
	ValidArgsFunction:  cobra.NoFileCompletions,
	DisableFlagParsing: true,
	Run: func(_ *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Error: app-name required\n")
			os.Exit(1)
		}
		appName := args[0]
		configArgs := args[1:]

		if err := runExplain(appName, configArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runExplain(appName string, configArgs []string) error {
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	envVarPrefix := convertAppNameToEnvPrefix(appName)
	configName := convertAppNameToConfigName(appName)

	loader, err := orale.LoadFromValues(
		configArgs,
		envVarPrefix,
		os.Environ(),
		workingDir,
		[]string{configName},
	)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	entries := loader.Explain()

	if len(entries) == 0 {
		fmt.Println("No configuration values loaded")
		return nil
	}

	printTable(entries)

	return nil
}

func printTable(entries []orale.ExplainEntry) {
	headers := []string{"PATH", "VALUE", "SOURCE"}
	colWidths := calculateColumnWidths(entries, headers)

	printSeparator(colWidths)
	printRow(headers, colWidths)
	printSeparator(colWidths)

	for _, entry := range entries {
		row := []string{entry.Path, entry.Value, entry.Source}
		printRow(row, colWidths)
	}

	printSeparator(colWidths)
}

func calculateColumnWidths(entries []orale.ExplainEntry, headers []string) []int {
	widths := []int{len(headers[0]), len(headers[1]), len(headers[2])}

	for _, entry := range entries {
		if len(entry.Path) > widths[0] {
			widths[0] = len(entry.Path)
		}
		if len(entry.Value) > widths[1] {
			widths[1] = len(entry.Value)
		}
		if len(entry.Source) > widths[2] {
			widths[2] = len(entry.Source)
		}
	}

	maxWidths := []int{40, 50, 30}
	for i := range widths {
		if widths[i] > maxWidths[i] {
			widths[i] = maxWidths[i]
		}
	}

	return widths
}

func printSeparator(widths []int) {
	parts := make([]string, len(widths))
	for i, w := range widths {
		parts[i] = strings.Repeat("-", w+2)
	}
	fmt.Println("+" + strings.Join(parts, "+") + "+")
}

func printRow(cells []string, widths []int) {
	wrapped := wrapCells(cells, widths)

	maxLines := 0
	for _, lines := range wrapped {
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}

	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
		parts := make([]string, len(cells))
		for colIdx := range cells {
			if lineIdx < len(wrapped[colIdx]) {
				parts[colIdx] = padRight(wrapped[colIdx][lineIdx], widths[colIdx])
			} else {
				parts[colIdx] = strings.Repeat(" ", widths[colIdx])
			}
		}
		fmt.Println("| " + strings.Join(parts, " | ") + " |")
	}
}

func wrapCells(cells []string, widths []int) [][]string {
	result := make([][]string, len(cells))

	for i, cell := range cells {
		result[i] = wrapText(cell, widths[i])
	}

	return result
}

func wrapText(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		for i := 0; i < len(text); i += width {
			end := i + width
			if end > len(text) {
				end = len(text)
			}
			lines = append(lines, text[i:end])
		}
		return lines
	}

	currentLine := ""
	for _, word := range words {
		if len(word) > width {
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = ""
			}
			for i := 0; i < len(word); i += width {
				end := i + width
				if end > len(word) {
					end = len(word)
				}
				lines = append(lines, word[i:end])
			}
			continue
		}

		if currentLine == "" {
			currentLine = word
		} else if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	if len(lines) == 0 {
		return []string{""}
	}

	return lines
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func convertAppNameToEnvPrefix(applicationName string) string {
	applicationNameRunes := []rune(applicationName)
	envPrefixRunes := []rune{}

	for i := 0; i < len(applicationNameRunes); i += 1 {
		currentChar := applicationNameRunes[i]
		var nextChar rune
		if i+1 < len(applicationNameRunes) {
			nextChar = applicationNameRunes[i+1]
		}
		if currentChar == '-' {
			continue
		}
		if unicode.IsLower(currentChar) {
			envPrefixRunes = append(envPrefixRunes, unicode.ToUpper(currentChar))
			if unicode.IsUpper(nextChar) {
				envPrefixRunes = append(envPrefixRunes, '_')
			}
		} else {
			envPrefixRunes = append(envPrefixRunes, currentChar)
		}
	}

	return string(envPrefixRunes)
}

func convertAppNameToConfigName(applicationName string) string {
	applicationNameRunes := []rune(applicationName)
	configNameRunes := []rune{}

	for i := 0; i < len(applicationNameRunes); i += 1 {
		currentChar := applicationNameRunes[i]
		var nextChar rune
		if i+1 < len(applicationNameRunes) {
			nextChar = applicationNameRunes[i+1]
		}
		if currentChar == '_' {
			continue
		}
		if unicode.IsUpper(currentChar) {
			configNameRunes = append(configNameRunes, unicode.ToLower(currentChar))
		} else {
			configNameRunes = append(configNameRunes, currentChar)
			if unicode.IsLower(currentChar) && unicode.IsUpper(nextChar) {
				configNameRunes = append(configNameRunes, '-')
			}
		}
	}

	return string(configNameRunes)
}
