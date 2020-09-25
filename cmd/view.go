package cmd

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/GGP1/kure/db"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type viewData struct {
	Models interface{}
}

var viewCmd = &cobra.Command{
	Use:   "view [-p port]",
	Short: "Display all entries on a server",
	Run: func(cmd *cobra.Command, args []string) {
		if port := viper.GetInt("http.port"); port != 0 {
			httpPort = uint16(port)
		}

		entries, err := db.ListEntries()
		if err != nil {
			must(err)
		}

		for _, e := range entries {
			e.Title = strings.Title(e.Title)
		}

		http.HandleFunc("/", viewTemplate(entries))

		addr := fmt.Sprintf(":%d", httpPort)
		fmt.Printf("Serving entries on port %s\n", addr)

		if err := http.ListenAndServe(addr, nil); err != nil {
			must(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(viewCmd)
	viewCmd.Flags().Uint16VarP(&httpPort, "port", "p", 4000, "server port")
}

// viewTemplate serves a html file with all the database entries
func viewTemplate(list interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./model/template/model.html"))

		vData := viewData{
			Models: list,
		}

		if err := tmpl.Execute(w, vData); err != nil {
			must(err)
		}
	}
}
