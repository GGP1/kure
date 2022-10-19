package session

import (
	"bytes"
	"os"
	"testing"
	"time"
)

func TestSessionCommand(t *testing.T) {
	cases := []struct {
		desc string
		args []string
		cont bool
	}{
		{
			desc: "Session command",
			args: []string{"pwd"},
			cont: true,
		},
		{
			desc: "Non-existent command",
			args: []string{"jump"},
			cont: false,
		},
		{
			desc: "No args",
			args: []string{},
			cont: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cont := runSessionCommand(tc.args, &timeout{})
			if cont != tc.cont {
				t.Errorf("Expected %v, got %v", tc.cont, cont)
			}
		})
	}
}

func TestCommands(t *testing.T) {
	dir, _ := os.Getwd()

	cases := []struct {
		timeout     *timeout
		desc        string
		input       string
		expectedOut string
		expectedErr string
		args        []string
	}{
		{
			desc: "Empty",
			args: []string{""},
		},
		{
			desc:        "block",
			args:        []string{"block"},
			expectedOut: "Press Enter to continue",
			input:       "\n",
		},
		{
			desc: "sleep",
			args: []string{"sleep", "1ns"},
		},
		{
			desc:        "sleep no duration",
			args:        []string{"sleep"},
			expectedErr: "error: invalid duration, use sleep [duration]\n",
		},
		{
			desc:        "sleep invalid duration",
			args:        []string{"sleep", "1", "ns"},
			expectedErr: "error: invalid duration \"1\"\n",
		},
		{
			desc:        "pwd",
			args:        []string{"pwd"},
			expectedOut: dir + "\n",
		},
		{
			desc:        "No timeout",
			args:        []string{"timeout"},
			timeout:     &timeout{duration: 0},
			expectedOut: "The session has no timeout.\n",
		},
		{
			desc:    "ttadd",
			args:    []string{"ttadd", "15s"},
			timeout: &timeout{timer: time.NewTimer(0)},
		},
		{
			desc:        "ttadd no duration",
			args:        []string{"ttadd"},
			expectedErr: "error: invalid duration, use ttadd [duration]\n",
		},
		{
			desc:        "ttadd invalid duration",
			args:        []string{"ttadd", "s"},
			expectedErr: "error: invalid duration \"s\"\n",
		},
		{
			desc:    "ttset",
			args:    []string{"ttset", "15s"},
			timeout: &timeout{timer: time.NewTimer(0)},
		},
		{
			desc:        "ttset no duration",
			args:        []string{"ttset"},
			expectedErr: "error: invalid duration, use ttset [duration]\n",
		},
		{
			desc:        "ttset invalid duration",
			args:        []string{"ttset", "15"},
			expectedErr: "error: invalid duration \"15\"\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			in, out, outErr := new(bytes.Buffer), new(bytes.Buffer), new(bytes.Buffer)
			in.WriteString(tc.input)

			cmd, ok := commands[tc.args[0]]
			if !ok {
				t.Error("Command not found")
			}

			params := params{
				in:      in,
				out:     out,
				outErr:  outErr,
				args:    tc.args[1:],
				timeout: tc.timeout,
			}

			cmd(params)

			if tc.expectedErr != "" {
				gotErr := outErr.String()
				if gotErr != tc.expectedErr {
					t.Errorf("Expected %q, got %q", tc.expectedErr, gotErr)
				}
			}
			if tc.expectedOut != "" {
				gotOut := out.String()
				if gotOut != tc.expectedOut {
					t.Errorf("Expected %q, got %q", tc.expectedOut, gotOut)
				}
			}
		})
	}
}
