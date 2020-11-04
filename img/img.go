// Package img displays images on the terminal.
//
// Adapted from:
//
// github.com/trashhalo/imgcat.
//
// github.com/muesli/termenv.
//
// github.com/lucasb-eyer/go-colorful.
package img

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

type model struct {
	content []byte
	secret  string
	image   string
	height  uint
	err     error
}

// Display shows the image on the user terminal.
func Display(secret string, content []byte) error {
	p := tea.NewProgram(model{
		content: content,
		secret:  secret,
	})
	p.EnterAltScreen()
	defer p.ExitAltScreen()

	if err := p.Start(); err != nil {
		return errors.Wrap(err, "failed displaying image on the terminal")
	}

	return nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		if _, ok := msg.(tea.KeyMsg); ok {
			return m, tea.Quit
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = uint(msg.Height)
		return m, load(m.content)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "Q", "esc":
			return m, tea.Quit
		case "e", "E":
			if err := clipboard.WriteAll(m.secret); err != nil {
				m.err = errors.Errorf("couldn't copy the secret to the clipboard: %v", err)
			}
			return m, nil
		}
	case errMsg:
		m.err = msg
		return m, nil
	case loadMsg:
		img, err := readerToImage(m.height, m.secret, msg.content)
		if err != nil {
			return m, func() tea.Msg { return errMsg{err} }
		}
		m.image = img
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("failed creating the QR code: %v\n\nPress any key to exit.", m.err)
	}
	return m.image
}

type loadMsg struct {
	content *bytes.Reader
}

type errMsg struct{ error }

func load(content []byte) tea.Cmd {
	return func() tea.Msg {
		r := bytes.NewReader(content)
		return loadMsg{content: r}
	}
}

func readerToImage(height uint, password string, r io.Reader) (string, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return "", errors.Wrap(err, "couldn't decode the image")
	}
	b := img.Bounds()
	w := b.Max.X
	h := b.Max.Y
	p := colorProfile()
	str := strings.Builder{}
	for y := 0; y < h; y += 2 {
		for x := 0; x < w; x++ {
			c1, _ := makeColor(img.At(x, y))
			color1 := p.color(c1.hex())
			c2, _ := makeColor(img.At(x, y+1))
			color2 := p.color(c2.hex())
			str.WriteString(newStyle("▀").
				foreground(color1).
				background(color2).
				String())
		}
		str.WriteString("\n")
	}
	if len(password) > 80 {
		str.WriteString("If you are having trouble scanning your code, please increase your terminal size and try again.\n")
	}
	s := fmt.Sprintf(`
Secret: %s

Press e to copy the secret.
Press ESC|q|Ctrl+C to quit.
`, password)
	str.WriteString(s)
	return str.String(), nil
}

// enableAnsiColors enables support for ANSI color sequences in Windows
// default console. Note that this only works with Windows 10.
func enableAnsiColors() {
	stdout := windows.Handle(os.Stdout.Fd())
	var originalMode uint32

	windows.GetConsoleMode(stdout, &originalMode)
	windows.SetConsoleMode(stdout, originalMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
}

func init() {
	enableAnsiColors()
}
