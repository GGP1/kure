package test

import (
	"crypto/rand"
	"errors"
	"fmt"
	"runtime"
	"time"

	cmdutil "github.com/GGP1/kure/commands"

	"github.com/GGP1/atoll"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/argon2"
)

var example = `
kure config argon2 test -m 500000 -i 2 -t 4`

type testOptions struct {
	memory, iterations uint32
	threads            uint8
}

// NewCmd returns a new command.
func NewCmd() *cobra.Command {
	opts := testOptions{}

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test argon2 performance",
		Long: `Test the time taken by argon2 to derive the key with the parameters passed.
		
The Argon2id variant with 1 iteration and maximum available memory is recommended as a default setting for all environments. This setting is secure against side-channel attacks and maximizes adversarial costs on dedicated bruteforce hardware.

If one of the devices that will handle the database has lower than 1GB of memory, we recommend setting the memory value to the half of that device RAM availability. Otherwise, default values should be fine.

• Memory: amount of memory allowed for argon2 to use. There is no "insecure" value for this parameter, though clearly the more memory the better. The value is represented in kibibytes, 1 kibibyte = 1024 bytes. Default is 1048576 kibibytes (1024 MB).

• Iterations: number of passes over the memory. The running time depends linearly on this parameter. Again, there is no "insecure value". Default is 1.

• Threads: number of threads number in parallel. Default is the maximum number of logical CPUs usable.`,
		Example: example,
		RunE:    runTest(&opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = testOptions{
				memory:     1048576,
				iterations: 1,
				threads:    uint8(runtime.NumCPU()),
			}
		},
	}

	f := cmd.Flags()
	f.Uint32VarP(&opts.iterations, "iterations", "i", 1, "number of passes over the memory")
	f.Uint32VarP(&opts.memory, "memory", "m", 1048576, "amount of memory allowed for argon2 to use")
	f.Uint8VarP(&opts.threads, "threads", "t", uint8(runtime.NumCPU()), "number of threads running in parallel")

	return cmd
}

func runTest(opts *testOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		if opts.iterations < 1 || opts.memory < 1 {
			return errors.New("iterations and memory should be higher than 0")
		}
		if opts.threads < 1 {
			return errors.New("the number of threads must be higher than 0")
		}

		password, err := atoll.NewPassword(25, []atoll.Level{
			atoll.Lowercase,
			atoll.Uppercase,
			atoll.Digit,
			atoll.Space,
			atoll.Special})
		if err != nil {
			return err
		}

		salt := make([]byte, 32)
		if _, err = rand.Read(salt); err != nil {
			return errors.New("failed generating salt")
		}

		start := time.Now()

		argon2.IDKey([]byte(password), salt, opts.iterations, opts.memory, opts.threads, 32)

		fmt.Println(time.Since(start))
		return nil
	}
}
