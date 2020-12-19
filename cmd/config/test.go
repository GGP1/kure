package config

import (
	"crypto/rand"
	"errors"
	"fmt"
	"runtime"
	"time"

	cmdutil "github.com/GGP1/kure/cmd"

	"github.com/GGP1/atoll"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/argon2"
)

var (
	memory     uint32
	iterations uint32
	threads    uint8
)

var testExample = `
kure config test -m 500000 -i 2 -t 4`

func testSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test argon2 performance",
		Long: `The Argon2id variant with 1 iteration and maximum available memory is recommended as a default setting for all environments. This setting is secure against side-channel attacks and maximizes adversarial costs on dedicated bruteforce hardware.

If one of the devices that will handle the database has lower than 1GB of memory, we recommend setting the memory value to the half of that device RAM availability. Otherwise, default values should be fine.

• Memory: there is no "insecure" value for this parameter, though clearly the more memory the better. The value is represented in kibibytes, 1 kibibyte = 1024 bytes. Default is 1048576 kibibytes (1024 MB).

• Iterations: the running time depends linearly on this parameter. We expect that the user chooses this number according to the time constraints on the application. Again, there is no "insecure value". Default is 1.

• Threads: default is the maximum number of logical CPUs usable.`,
		Example: testExample,
		RunE:    runTest(),
		PostRun: func(cmd *cobra.Command, args []string) {
			memory, iterations = 1048576, 1
			threads = uint8(runtime.NumCPU())
		},
	}

	f := cmd.Flags()
	f.Uint32VarP(&iterations, "iterations", "i", 1, "number of passes over the memory")
	f.Uint32VarP(&memory, "memory", "m", 1048576, "amount of memory allowed for argon2 to use")
	f.Uint8VarP(&threads, "threads", "t", uint8(runtime.NumCPU()), "number of threads running in parallel")

	return cmd
}

func runTest() cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		if iterations|memory < 1 {
			return errors.New("iterations and memory should be higher than 0")
		}
		if threads < 1 {
			return errors.New("the number of threads must be higher than 0")
		}

		password, err := atoll.NewPassword(25, []uint8{1, 2, 3, 4, 5, 6})
		if err != nil {
			return err
		}

		salt := make([]byte, 32)
		_, err = rand.Read(salt)
		if err != nil {
			return errors.New("failed generating salt")
		}

		start := time.Now()

		argon2.IDKey([]byte(password), salt, iterations, memory, threads, 32)

		fmt.Println(time.Since(start))
		return nil
	}
}
