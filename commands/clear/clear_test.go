package clear

import (
	"bufio"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/atotto/clipboard"
)

func TestClearClipboard(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("No clipboard utilities available")
	}

	cmd := NewCmd()
	if err := cmd.Flags().Set("clipboard", "true"); err != nil {
		t.Error(err)
	}

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	got, _ := clipboard.ReadAll()
	if got != "" {
		t.Errorf("Expected clipboard to be empty but got: %s", got)
	}
}

func TestClearTerminalScreen(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		t.Skip("linux and macOS return an exit status 1 when clearing the terminal")
	}

	cmd := NewCmd()
	cmd.SetOut(io.Discard)
	if err := cmd.Flags().Set("terminal", "true"); err != nil {
		t.Error(err)
	}

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Failed: %v", err)
	}
}

func TestClearTerminalHistory(t *testing.T) {
	mockHistoryPath := "./testdata/.history"

	originalContent, err := os.ReadFile(mockHistoryPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := clearHistoryFile(mockHistoryPath); err != nil {
		t.Error(err)
	}

	f, err := os.Open(mockHistoryPath)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.HasPrefix(strings.TrimSpace(scanner.Text()), "kure ") {
			t.Errorf("The history file contains kure commands: %s", scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		t.Error(err)
	}

	// Restore file content
	if err := os.WriteFile(mockHistoryPath, originalContent, 0600); err != nil {
		t.Error(err)
	}
}

func TestPostRun(t *testing.T) {
	NewCmd().PostRun(nil, nil)
}
