package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/GGP1/kure/db"
	"github.com/GGP1/kure/pb"
	"github.com/GGP1/kure/tree"

	"github.com/atotto/clipboard"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

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

		fmt.Printf("\nSuccessfully created \"%s\" card.\n", name)
	},
}

var copyCard = &cobra.Command{
	Use:   "copy <name> [-t timeout] [-f field]",
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

		field = strings.ToLower(field)

		var f string
		switch field {
		case "number":
			f = card.Number
		case "code":
			f = card.CVC
		default:
			fatalf("invalid card field, use \"number\" or \"code\"")
		}

		if err := clipboard.WriteAll(f); err != nil {
			fatal(errWritingClipboard)
		}

		if timeout > 0 {
			<-time.After(timeout)
			clipboard.WriteAll("")
			os.Exit(1)
		}
	},
}

var lsCard = &cobra.Command{
	Use:   "ls <name> [-f filter]",
	Short: "List cards",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")

		if name == "" {
			cards, err := db.ListCards()
			if err != nil {
				fatal(err)
			}

			paths := make([]string, len(cards))

			for i, card := range cards {
				paths[i] = card.Name
			}

			tree.Print(paths)
			return
		}

		if filter {
			cards, err := db.CardsByName(name)
			if err != nil {
				fatal(err)
			}

			for _, card := range cards {
				fmt.Print("\n")
				printCard(card)
			}
			return
		}

		card, err := db.GetCard(name)
		if err != nil {
			fatal(err)
		}

		fmt.Print("\n")
		printCard(card)
	},
}

var rmCard = &cobra.Command{
	Use:   "rm <name>",
	Short: "Remove a card from the database",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")
		if name == "" {
			fatal(errInvalidName)
		}

		_, err := db.GetCard(name)
		if err != nil {
			fatalf("%s card does not exist", name)
		}

		if proceed() {
			if err := db.RemoveCard(name); err != nil {
				fatal(err)
			}

			fmt.Printf("\nSuccessfully removed \"%s\" card.\n", name)
		}
	},
}

func init() {
	rootCmd.AddCommand(cardCmd)

	cardCmd.AddCommand(addCard)
	cardCmd.AddCommand(copyCard)
	cardCmd.AddCommand(lsCard)
	cardCmd.AddCommand(rmCard)

	copyCard.Flags().DurationVarP(&timeout, "timeout", "t", 0, "clipboard cleaning timeout")
	copyCard.Flags().StringVarP(&field, "field", "f", "number", "choose which field to copy")

	lsCard.Flags().BoolVarP(&filter, "filter", "f", false, "filter cards")
}

func cardInput(name string) (*pb.Card, error) {
	var cType, expDate, num, CVC string

	if !strings.Contains(name, "/") && len(name) > 43 {
		return nil, errors.New("card name must contain 43 letters or less")
	}

	scanner := bufio.NewScanner(os.Stdin)

	scan(scanner, "Type", &cType)
	scan(scanner, "Expire date", &expDate)
	scan(scanner, "Number", &num)
	scan(scanner, "CVC", &CVC)

	name = strings.ToLower(name)

	card := &pb.Card{
		Name:       name,
		Type:       cType,
		ExpireDate: expDate,
		Number:     num,
		CVC:        CVC,
	}

	return card, nil
}

func printCard(c *pb.Card) {
	printObjectName(c.Name)

	if hide {
		c.CVC = "••••"
	}

	fmt.Printf(`│ Type       │ %s
│ Number     │ %s
│ CVC        │ %s
│ Expires    │ %s
`, c.Type, c.Number, c.CVC, c.ExpireDate)

	fmt.Println("+────────────+──────────────────────────────>")
}
