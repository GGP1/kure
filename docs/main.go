package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/GGP1/kure/commands/root"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Utils to help developers update Kure's documentation.
// Remember to update the wiki if necessary.
// build "go build main.go" and execute "./main [flags] [args]"

func main() {
	cmd := flag.Bool("cmd", false, "specific command documentation")
	comp := flag.Bool("completion", false, "generate completion")
	summ := flag.Bool("summary", false, "commands summary")
	flag.Parse()

	if *cmd {
		if err := cmdDocs(os.Args); err != nil {
			log.Fatalf("failed generating %s documentation: %v", os.Args[1], err)
		}
	} else if *comp {
		if err := completion(os.Args); err != nil {
			log.Fatalf("failed generating %s completion: %v", os.Args[1], err)
		}
	} else if *summ {
		if err := summary(os.Args); err != nil {
			log.Fatalf("failed generating commands summary: %v", err)
		}
	}
}

// Generate a command's documentation.
//
// The output generated by this command is exactly what we would expect
// in most cases but not in all of them, as some commands' documentation
// contain extra information that can't be extracted from cobra (or little tweaks
// depending on each case).
//
// Please make sure not to overwrite that information in those specific cases.
//
// Usage: main --cmd ls.
func cmdDocs(args []string) error {
	root := root.CmdForDocs()

	cmd, _, err := root.Find(args[2:])
	if err != nil {
		return err
	}

	return customMarkdown(cmd, os.Stdout)
}

// customMarkdown creates custom markdown output.
func customMarkdown(cmd *cobra.Command, w io.Writer) error {
	buf := new(bytes.Buffer)

	buf.WriteString("## Use\n\n")
	buf.WriteString("`")
	buf.WriteString(strings.Replace(cmd.UseLine(), "[flags]", "", 1))
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		var shorthand string
		if f.Shorthand != "" {
			shorthand = fmt.Sprintf("-%s ", f.Shorthand)
		}
		buf.WriteString(fmt.Sprintf("[%s%s] ", shorthand, f.Name))
	})
	buf.WriteString("`\n\n")

	if len(cmd.Aliases) > 0 {
		buf.WriteString("*Aliases*: ")
		for i, a := range cmd.Aliases {
			buf.WriteString(a)
			if i != len(cmd.Aliases)-1 {
				buf.WriteString(", ")
			}
		}
		buf.WriteString(".\n\n")
	}

	buf.WriteString("## Description\n\n")
	if cmd.Long != "" {
		buf.WriteString(cmd.Long)
		buf.WriteString("\n\n")
	} else {
		buf.WriteString(cmd.Short)
		buf.WriteString(".\n\n")
	}

	if cmd.HasSubCommands() {
		url := getURL(cmd)
		buf.WriteString("## Subcommands\n\n")
		for _, c := range cmd.Commands() {
			name := c.Name()
			if !c.HasSubCommands() {
				name += ".md"
			}
			buf.WriteString(fmt.Sprintf("- [`%s`](%s%s): %s.\n", c.CommandPath(), url, name, c.Short))
		}
	}
	buf.WriteString("\n")

	buf.WriteString("## Flags\n\n")
	flags := cmd.Flags()
	if flags.HasFlags() {
		buf.WriteString("| Name | Shorthand | Type | Default | Description |\n")
		buf.WriteString("|------|-----------|------|---------|-------------|\n")
		flags.VisitAll(func(f *pflag.Flag) {
			// Uppercase the first letter only
			usage := []byte(f.Usage)
			usage[0] = byte(unicode.ToUpper(rune(usage[0])))

			buf.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n", f.Name, f.Shorthand, f.Value.Type(), f.DefValue, usage))
		})
	} else {
		buf.WriteString("No flags.\n")
	}

	if cmd.Example != "" {
		buf.WriteString("\n### Examples\n\n")
		examples := strings.Split(cmd.Example, "\n")

		for _, e := range examples {
			if strings.HasPrefix(e, "*") {
				buf.WriteString(strings.Replace(e, "* ", "", 1) + ":\n")
			} else if e != "" {
				buf.WriteString("```\n" + e + "\n```\n\n")
			}
		}
	}

	_, err := buf.WriteTo(w)
	return err
}

func getURL(cmd *cobra.Command) string {
	url := "https://github.com/GGP1/kure/tree/master/docs/commands/"

	split := strings.Split(cmd.CommandPath(), " ")
	for _, s := range split[1:] {
		url += s + "/subcommands/"
	}

	return url
}

// Generate code completion files. By default it generates all the files.
//
// Usage: main --completion bash.
func completion(args []string) error {
	root := root.CmdForDocs()

	if len(args) < 3 {
		if err := root.GenBashCompletionFile("completion/bash.sh"); err != nil {
			return err
		}
		if err := root.GenFishCompletionFile("completion/fish.sh", true); err != nil {
			return err
		}
		if err := root.GenPowerShellCompletionFile("completion/powershell.ps1"); err != nil {
			return err
		}
		return root.GenZshCompletionFile("completion/zsh.sh")
	}

	switch args[2] {
	case "bash":
		return root.GenBashCompletionFile("completion/bash.sh")
	case "fish":
		return root.GenFishCompletionFile("completion/fish.sh", true)
	case "powershell":
		return root.GenPowerShellCompletionFile("completion/powershell.ps1")
	case "zsh":
		return root.GenZshCompletionFile("completion/zsh.sh")
	}

	return nil
}

// Generate the wiki commands summary page.
// https://www.github.com/kure/wiki/Commands-summary
//
// Usage: main --summary copy.
func summary(args []string) error {
	root := root.CmdForDocs()
	docs := fmtSummary(root)

	if len(args) == 3 {
		if args[2] == "copy" {
			if err := clipboard.WriteAll(docs); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprint(os.Stdout, docs); err != nil {
		return err
	}
	return nil
}

func fmtSummary(cmd *cobra.Command) string {
	var (
		sb      strings.Builder
		subCmds func(*cobra.Command, string)
	)

	cmdAndFlags := func(c *cobra.Command) {
		sb.WriteString(fmt.Sprintf("%s ", c.Use))

		c.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Shorthand != "" {
				f.Shorthand = fmt.Sprintf("-%s ", f.Shorthand)
			}
			sb.WriteString(fmt.Sprintf("[%s%s] ", f.Shorthand, f.Name))
		})
	}
	// Add subcommands and flags using recursion
	subCmds = func(cmd *cobra.Command, indent string) {
		// Add indent on each call
		indent += "    "
		for _, sub := range cmd.Commands() {
			sb.WriteString("\n" + indent)
			cmdAndFlags(sub)
			subCmds(sub, indent)
		}
	}

	sb.WriteString(`For further information about each each command, its flags and examples please visit the [commands folder](https://github.com/GGP1/kure/tree/master/docs/commands).
`)

	// Index
	for _, c := range cmd.Commands() {
		sb.WriteString(fmt.Sprintf("\n- [%s](#%s)", c.Name(), c.Name()))
	}
	// Separation
	sb.WriteString("\n\n\n")

	// Each command name and its flags
	for _, c := range cmd.Commands() {
		sb.WriteString(fmt.Sprintf("### %s\n```\n", c.Name()))
		cmdAndFlags(c)
		subCmds(c, "")
		sb.WriteString("\n```\n\n---\n\n")
	}

	return sb.String()
}
