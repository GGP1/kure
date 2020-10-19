package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/GGP1/kure/db"
	"github.com/GGP1/kure/model/card"

	"github.com/atotto/clipboard"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var field string

var cardCmd = &cobra.Command{
	Use:   "card",
	Short: "Card operations",
}

var addCard = &cobra.Command{
	Use:   "add <name>",
	Short: "Add card to the database",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")
		if name == "" {
			fatal(errInvalidName)
		}

		card, err := cardInput(name)
		if err != nil {
			fatal(err)
		}

		if err := db.CreateCard(card); err != nil {
			fatal(err)
		}

		fmt.Print("\nSucessfully created the card.")
	},
}

var copyCard = &cobra.Command{
	Use:   "copy <name> [-t timeout] [field]",
	Short: "Copy card number or CVC",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")
		if name == "" {
			fatal(errInvalidName)
		}

		card, err := db.GetCard(name)
		if err != nil {
			fatal(err)
		}

		field := strings.ToLower(field)

		if field == "number" {
			if err := clipboard.WriteAll(card.Number); err != nil {
				fatal(errors.New("failed writing to the clipboard"))
			}
		} else if field == "code" {
			if err := clipboard.WriteAll(card.CVC); err != nil {
				fatal(errors.New("failed writing to the clipboard"))
			}
		}

		if timeout > 0 {
			<-time.After(timeout)
			clipboard.WriteAll("")
			os.Exit(1)
		}
	},
}

var deleteCard = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a card from the database",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")
		if name == "" {
			fatal(errInvalidName)
		}

		_, err := db.GetCard(name)
		if err != nil {
			fatalf("%s card does not exist", name)
		}

		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Are you sure you want to proceed? [y/n]: ")

		scanner.Scan()
		text := scanner.Text()
		input := strings.ToLower(text)

		if strings.Contains(input, "y") || strings.Contains(input, "yes") {
			if err := db.DeleteCard(name); err != nil {
				fatal(err)
			}

			fmt.Printf("\nSuccessfully deleted %s card.", name)
		}
	},
}

var listCard = &cobra.Command{
	Use:   "list <name>",
	Short: "List a card or all the cards from the database",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")

		if name != "" {
			card, err := db.GetCard(name)
			if err != nil {
				fatal(err)
			}

			fmt.Print("\n")
			printCard(card)
			return
		}

		cards, err := db.ListCards()
		if err != nil {
			fatal(err)
		}

		for _, card := range cards {
			fmt.Print("\n")
			printCard(card)
		}
	},
}

func init() {
	rootCmd.AddCommand(cardCmd)

	cardCmd.AddCommand(addCard)
	cardCmd.AddCommand(copyCard)
	cardCmd.AddCommand(deleteCard)
	cardCmd.AddCommand(listCard)

	copyCard.Flags().DurationVarP(&timeout, "timeout", "t", 0, "clipboard cleaning timeout")
	copyCard.Flags().StringVarP(&field, "field", "f", "number", "choose which field to copy")
}

func cardInput(name string) (*card.Card, error) {
	var cType, expireDate, num, cardVerCode string

	if len(name) > 43 {
		return nil, errors.New("card name must contain 43 letters or less")
	}

	scanner := bufio.NewScanner(os.Stdin)

	scan(scanner, "Type", &cType)
	scan(scanner, "Expire date", &expireDate)
	scan(scanner, "Number", &num)
	scan(scanner, "CVC", &cardVerCode)

	name = strings.TrimSpace(strings.ToLower(name))

	card := card.New(name, cType, expireDate, num, cardVerCode)

	return card, nil
}

func printCard(c *card.Card) {
	c.Name = strings.Title(c.Name)

	if hide {
		c.CVC = "••••"
	}

	dashes := 43
	halfBar := ((dashes - len(c.Name)) / 2) - 1

	fmt.Print("+")
	for i := 0; i < halfBar; i++ {
		fmt.Print("─")
	}
	fmt.Printf(" %s ", c.Name)

	if (len(c.Name) % 2) == 0 {
		halfBar++
	}

	for i := 0; i < halfBar; i++ {
		fmt.Print("─")
	}
	fmt.Print(">\n")

	if c.Type != "" {
		fmt.Printf("│ Type       │ %s\n", c.Type)
	}

	if c.Number != "" {
		fmt.Printf("│ Number     │ %s\n", c.Number)
	}

	if c.CVC != "" {
		fmt.Printf("│ CVC        │ %s\n", c.CVC)
	}

	if c.ExpireDate != "" {
		fmt.Printf("│ Expires    │ %s\n", c.ExpireDate)
	}

	fmt.Println("+────────────+──────────────────────────────>")
}
