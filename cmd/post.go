package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"
)

var postFlags struct {
	NumWorkers          int
    ShowSingleProcesses bool
}

func init() {
	postCmd.Flags().IntVarP(&postFlags.NumWorkers, "workers", "w", 5, "Number of concurrent go workers")
    postCmd.Flags().BoolVar(&postFlags.ShowSingleProcesses, "s", false, "Shows single processes")
	rootCmd.AddCommand(postCmd)
}

var postCmd = &cobra.Command{
	Use:   "post [url] [data]",
	Short: "Sends POST requests to a URL",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		data := args[1]

		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "http://" + url
		}

		valid := json.Valid([]byte(data))
		if !valid {
			fmt.Println("data provided is not in JSON format")
			return
		}

		payload := strings.NewReader(data)

		resp, err := http.Post(url, "application/json", payload)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		fmt.Println("Response status:", resp.Status)
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			panic(err)
		}
	},
}
