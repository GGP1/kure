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
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cont := sessionCommand(tc.args, time.Time{}, &sessionOptions{})
			if cont != tc.cont {
				t.Errorf("Expected %v, got %v", tc.cont, cont)
			}
		})
	}
}

func TestCommands(t *testing.T) {
	dir, _ := os.Getwd()

	cases := []struct {
		desc        string
		args        []string
		timeout     time.Duration
		input       string
		expectedOut string
		expectedErr string
	}{
		{
			desc: "Empty",
			args: []string{""},
		},
		{
			desc:        "Block",
			args:        []string{"block"},
			expectedOut: "Press Enter to continue",
			input:       "\n",
		},
		{
			desc: "Sleep",
			args: []string{"sleep", "1ns"},
		},
		{
			desc:        "Sleep error",
			args:        []string{"sleep", "1", "ns"},
			expectedErr: "error: invalid duration \"1\"\n",
		},
		{
			desc:        "Show current directory",
			args:        []string{"pwd"},
			expectedOut: dir + "\n",
		},
		{
			desc:        "No timeout",
			args:        []string{"timeout"},
			timeout:     time.Duration(0),
			expectedOut: "The session has no timeout.\n",
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
				args:    tc.args,
				start:   time.Time{},
				timeout: tc.timeout,
			}

			if !cmd(params) {
				t.Errorf("Expected the command to return true and got false")
			}

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
