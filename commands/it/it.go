package it

import (
	"strings"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var example = `
* No arguments
kure it

* Command without flags
kure it ls

* Command with flags
kure it ls -s -q

* Only the name
kure sample`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "it <command|flags|name>",
		Short: "Interactive prompt",
		Long: `Interactive prompt.
This commands behave depending on the arguments received, it requests the missing information.

Given 				Requests
command 			flags and name
command and flags 		name
name 				command and flags`,
		Example:            example,
		DisableFlagParsing: true,
		PreRunE:            auth.Login(db),
		RunE:               runIt(db),
	}

	return cmd
}

func runIt(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		var (
			arguments []string
			err       error
		)

		root := cmd.Root()
		// Get rid of unnecessary information and reset in case we are inside a session
		defer root.SetUsageTemplate(root.UsageTemplate())
		root.SetUsageTemplate(template)

		// Evaluate whether we received:
		// Nothing
		// Commands
		// Commands + flags
		// Name
		switch len(args) {
		case 0:
			arguments, err = requestCommands(db, root, nil)

		default:
			var hasFlags bool
			for _, arg := range args {
				if strings.Contains(arg, "-") {
					hasFlags = true
					break
				}
			}

			found, _, err := root.Find(args)
			if err != nil {
				// If the command does not exist, assume the user passed a name
				arguments, err = gotName(db, root, strings.Join(args, " "))
				break
			}

			// Flags should be along with the command
			// If we received flags we don't have to look for subcommands
			if hasFlags {
				arguments, err = requestName(db, args)
			} else {
				// Pass on received cmds and look for subcommands
				arguments, err = requestCommands(db, found, args)
			}
		}
		if err != nil {
			return err
		}

		// Discard empty arguments as some commands will fail if we don't
		// eg. file cat
		var filteredArgs []string
		for _, arg := range arguments {
			if arg != "" {
				filteredArgs = append(filteredArgs, arg)
			}
		}

		root.SetArgs(filteredArgs)
		return root.Execute()
	}
}

func requestCommands(db *bolt.DB, root *cobra.Command, receivedCmds []string) ([]string, error) {
	commands, err := selectCommands(root)
	if err != nil {
		return nil, err
	}

	flags, err := selectFlags(root, commands)
	if err != nil {
		return nil, err
	}

	instructions := append(commands, flags...)
	// Preprend the received commands if there is any
	// We would have [received commands] [commands] [flags]
	if len(receivedCmds) > 0 {
		instructions = append(receivedCmds, instructions...)
	}
	return requestName(db, instructions)
}

// Instructions contains commands and flags.
func requestName(db *bolt.DB, instructions []string) ([]string, error) {
	var (
		name string
		err  error
	)

	search := strings.Join(instructions, " ")
	// contains reports whether s is within search
	contains := func(s string) bool {
		return strings.Contains(search, s)
	}

	// Behave depending on which command the user is executing
	switch {
	case contains("add"),
		contains("ls") && contains("-f"), // Filter
		contains("rm") && contains("-d"): // Remove directory
		name, err = inputName()

	case contains("import"), contains("export"):
		name, err = selectManager(db)

	case contains("file cat"), contains("file touch"):
		names, err := fileMultiselect(db)
		if err != nil {
			return nil, err
		}
		return append(instructions, names...), nil

	case contains("file mv"):
		names, err := fileMvNames(db)
		if err != nil {
			return nil, err
		}
		return append(instructions, names...), nil

	default:
		list := []string{"2fa", "copy", "edit", "ls", "rm"}
		// Request the name depending on the command
		for _, cmd := range list {
			if contains(cmd) {
				// Skip "config edit" as it doesn't need a name
				if instructions[0] != "config" {
					name, err = selectName(db, instructions)
					break
				}
			}
		}
	}
	if err != nil {
		return nil, err
	}

	// Remember: the flags are inside the commands slice
	result := append(instructions, name)
	return result, nil
}

// gotName is executed when the user already provided the name, commands and flags are requested only.
func gotName(db *bolt.DB, root *cobra.Command, name string) ([]string, error) {
	commands, err := selectCommands(root)
	if err != nil {
		return nil, err
	}

	flags, err := selectFlags(root, commands)
	if err != nil {
		return nil, err
	}

	result := append(commands, flags...)
	result = append(result, name)

	return result, nil
}
