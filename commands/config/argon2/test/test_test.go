package test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTest(t *testing.T) {
	cases := []struct {
		desc       string
		iterations uint32
		memory     uint32
		threads    uint8
	}{
		{
			desc:       "Test 1",
			iterations: 1,
			memory:     400000,
			threads:    uint8(runtime.NumCPU()),
		},
		{
			desc:       "Test 2",
			iterations: 15,
			memory:     3000,
			threads:    uint8(runtime.NumCPU()),
		},
		{
			desc:       "Test 3",
			iterations: 2,
			memory:     716500,
			threads:    uint8(runtime.NumCPU() - 1),
		},
	}

	cmd := NewCmd()

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			f := cmd.Flags()
			f.Set("iterations", fmt.Sprintf("%d", tc.iterations))
			f.Set("memory", fmt.Sprintf("%d", tc.memory))
			f.Set("threads", fmt.Sprintf("%d", tc.threads))

			err := cmd.Execute()
			assert.NoError(t, err)
		})
	}
}

func TestTestInvalid(t *testing.T) {
	cases := []struct {
		desc       string
		iterations string
		memory     string
		threads    string
	}{
		{
			desc:       "Invalid iterations",
			iterations: "0",
			memory:     "1",
			threads:    "1",
		},
		{
			desc:       "Invalid memory",
			iterations: "1",
			memory:     "0",
			threads:    "1",
		},
		{
			desc:       "Invalid threads",
			iterations: "1",
			memory:     "1",
			threads:    "0",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd := NewCmd()
			f := cmd.Flags()
			f.Set("iterations", tc.iterations)
			f.Set("memory", tc.memory)
			f.Set("threads", tc.threads)

			err := cmd.RunE(nil, nil)
			assert.Error(t, err)
		})
	}
}

func TestPostRun(t *testing.T) {
	NewCmd().PostRun(nil, nil)
}
