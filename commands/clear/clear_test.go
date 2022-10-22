package clear

import (
	"bufio"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/atotto/clipboard"
	"github.com/stretchr/testify/assert"
)

func TestClearClipboard(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("No clipboard utilities available")
	}

	cmd := NewCmd()
	err := cmd.Flags().Set("clipboard", "true")
	assert.NoError(t, err)

	err = cmd.Execute()
	assert.NoError(t, err)

	got, _ := clipboard.ReadAll()
	assert.Empty(t, got)
}

func TestClearTerminalScreen(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		t.Skip("linux and macOS return an exit status 1 when clearing the terminal")
	}

	cmd := NewCmd()
	cmd.SetOut(io.Discard)
	err := cmd.Flags().Set("terminal", "true")
	assert.NoError(t, err)

	err = cmd.Execute()
	assert.NoError(t, err)
}

func TestClearTerminalHistory(t *testing.T) {
	mockHistoryPath := "./testdata/.history"

	originalContent, err := os.ReadFile(mockHistoryPath)
	assert.NoError(t, err)

	err = clearHistoryFile(mockHistoryPath)
	assert.NoError(t, err)

	f, err := os.Open(mockHistoryPath)
	assert.NoError(t, err)
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.HasPrefix(strings.TrimSpace(scanner.Text()), "kure ") {
			t.Errorf("The history file contains kure commands: %s", scanner.Text())
		}
	}

	err = scanner.Err()
	assert.NoError(t, err)

	// Restore file content
	err = os.WriteFile(mockHistoryPath, originalContent, 0o600)
	assert.NoError(t, err)
}

func TestPostRun(t *testing.T) {
	NewCmd().PostRun(nil, nil)
}
