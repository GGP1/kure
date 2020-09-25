package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/GGP1/kure/db"
	"github.com/GGP1/kure/model/card"

	"github.com/atotto/clipboard"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var add, copy, delete, list, view bool
var field string

var cardCmd = &cobra.Command{
	Use:   "card <name> [-a add | -c copy | -d delete | -l list | -v view] [-t timeout] [-f field]",
	Short: "Add, copy, delete or list cards",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")

		if add {
			if err := addCard(); err != nil {
				must(err)
			}
			return
		}

		if copy {
			if err := copyCard(name, timeout); err != nil {
				must(err)
			}
			return
		}

		if delete {
			if err := deleteCard(name); err != nil {
				must(err)
			}
			return
		}

		if view {
			if err := viewCards(); err != nil {
				must(err)
			}
			return
		}

		if err := listCard(name); err != nil {
			must(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(cardCmd)
	cardCmd.Flags().BoolVarP(&add, "add", "a", false, "add a card")
	cardCmd.Flags().BoolVarP(&copy, "copy", "c", false, "copy card number")
	cardCmd.Flags().BoolVarP(&delete, "delete", "d", false, "delete a card")
	cardCmd.Flags().BoolVarP(&list, "list", "l", true, "list card/cards")
	cardCmd.Flags().BoolVarP(&view, "view", "v", false, "view cards")
	cardCmd.Flags().DurationVarP(&timeout, "timeout", "t", 0, "clipboard cleaning timeout")
	cardCmd.Flags().StringVarP(&field, "field", "f", "number", "choose which field to copy")
}

func addCard() error {
	card, err := cardInput()
	if err != nil {
		return err
	}

	if err := db.CreateCard(card); err != nil {
		return err
	}

	fmt.Print("\nSucessfully created the card.")
	return nil
}

func copyCard(name string, timeout time.Duration) error {
	if name == "" {
		return errInvalidName
	}

	card, err := db.GetCard(name)
	if err != nil {
		return err
	}

	field := strings.ToLower(field)

	if field == "number" {
		number := strconv.Itoa(int(card.Number))
		if err := clipboard.WriteAll(number); err != nil {
			return err
		}
	} else if field == "code" {
		cvc := strconv.Itoa(int(card.CVC))
		if err := clipboard.WriteAll(cvc); err != nil {
			return err
		}
	}

	if timeout > 0 {
		<-time.After(timeout)
		clipboard.WriteAll("")
		os.Exit(1)
	}

	return nil
}

func deleteCard(name string) error {
	if name == "" {
		return errInvalidName
	}

	_, err := db.GetCard(name)
	if err != nil {
		return errors.New("this card does not exist")
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Are you sure you want to proceed? [y/n]: ")

	scanner.Scan()
	text := scanner.Text()
	input := strings.ToLower(text)

	if strings.Contains(input, "y") || strings.Contains(input, "yes") {
		if err := db.DeleteCard(name); err != nil {
			must(err)
		}

		fmt.Printf("\nSuccessfully deleted %s card.", name)
	}

	return nil
}

func listCard(name string) error {
	if name != "" {
		card, err := db.GetCard(name)
		if err != nil {
			return err
		}

		printCard(card)
		return nil
	}

	cards, err := db.ListCards()
	if err != nil {
		return err
	}

	for _, card := range cards {
		printCard(card)
	}

	return nil
}

func viewCards() error {
	if port := viper.GetInt("http.port"); port != 0 {
		httpPort = uint16(port)
	}

	cards, err := db.ListCards()
	if err != nil {
		return err
	}

	for _, c := range cards {
		c.Name = strings.Title(c.Name)
	}

	http.HandleFunc("/", viewTemplate(cards))

	addr := fmt.Sprintf(":%d", httpPort)
	fmt.Printf("Serving cards on port %s\n", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		return err
	}

	return nil
}

func cardInput() (*card.Card, error) {
	var name, cType, expireDate, num, carVerCode string

	scanner := bufio.NewScanner(os.Stdin)

	name = scan(scanner, "Name", name)
	cType = scan(scanner, "Type", cType)
	expireDate = scan(scanner, "Expire date", expireDate)
	num = scan(scanner, "Number", num)
	carVerCode = scan(scanner, "CVC", carVerCode)

	number, err := strconv.Atoi(num)
	if err != nil {
		return nil, errors.Wrap(err, "invalid card number")
	}

	cvc, err := strconv.Atoi(carVerCode)
	if err != nil {
		return nil, errors.Wrap(err, "invalid card verification code")
	}

	name = strings.ToLower(name)

	card := card.New(name, cType, expireDate, int32(number), int32(cvc))

	return card, nil
}

func printCard(c *card.Card) {
	c.Name = strings.Title(c.Name)

	if hide {
		c.CVC = 0
	}

	str := fmt.Sprintf(
		`
+───────────────+─────────────────>
│ Name	        │ %s
│ Type      	│ %s
│ Number      	│ %d
│ CVC           │ %d
│ Expire Date   │ %s
+───────────────+─────────────────>`,
		c.Name, c.Type, c.Number, c.CVC, c.ExpireDate)
	fmt.Println(str)
}
