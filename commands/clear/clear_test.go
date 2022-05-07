package clear

import (
	"bufio"
	"bytes"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/atotto/clipboard"
)

const mockHistoryPath = "./testdata/.history"

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
	if runtime.GOOS == "darwin" {
		t.Skip("macOS returns an exit status 1 when clearing the terminal")
	}

	cmd := NewCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := cmd.Flags().Set("terminal", "true"); err != nil {
		t.Error(err)
	}

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Failed: %v", err)
	}
}

func TestClearTerminalHistory(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Apparently there's no persistent way to modify the powershell history file path in Windows
		// Setting it with `Set-PSReadLineOption -HistorySavePath` is not shared across sessions
		t.Skip()
	}

	os.Setenv("HISTFILE", mockHistoryPath)
	originalContent, err := os.ReadFile(mockHistoryPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.WriteFile(mockHistoryPath, originalContent, 0600); err != nil {
			t.Error(err)
		}
	})

	cmd := NewCmd()
	if err := cmd.Flags().Set("history", "true"); err != nil {
		t.Error(err)
	}

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	f, err := os.Open(mockHistoryPath)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "kure") {
			t.Errorf("The history file contains kure commands: %s", scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		t.Error(err)
	}
}

func TestPostRun(t *testing.T) {
	NewCmd().PostRun(nil, nil)
}
