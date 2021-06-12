package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"net/http/httputil"
)

var listen string
var contentType string

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start basic http server",
	Run: func(cmd *cobra.Command, args []string) {
		printCyan("Listening on:", listen)
		printCyan("Content type:", contentType)

		http.HandleFunc("/", handler)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", listen), nil))
	},
}

func handler(w http.ResponseWriter, r *http.Request) {
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Fatal(err)
	}

	printGreen("-------------------------RECEIVED REQUEST-------------------------")
	printCyan("received: ", string(b))
	w.Header().Set("Content-Type", fmt.Sprintf("%v; charset=utf-8", contentType))
	w.WriteHeader(http.StatusOK)
	printGreen("------------------------------------------------------------------")
	printGreen("Completed:", http.StatusOK)
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVar(&listen, "listen", "8080", "specify port to listen")
	serverCmd.Flags().StringVar(&contentType, "contentType", "application/json", "response content-type header")
}
