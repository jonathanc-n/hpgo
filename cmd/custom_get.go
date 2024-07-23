// cmd/custom_get.go
package cmd

import (
	"fmt"
	"net/http"
	"io"
	"github.com/spf13/cobra"
)

var customGetFlags struct {
	NumWorkers          int
    ShowSingleProcesses bool
	Format              string 
	UserAgent           string
	AuthToken           string
	ClientVersion       string
	ApiKey              string
	CorrelationID       string
	CustomHeader        string
}

func init() {
	postCmd.Flags().IntVarP(&customGetFlags.NumWorkers, "workers", "w", 5, "Number of concurrent go workers")
    postCmd.Flags().BoolVar(&customGetFlags.ShowSingleProcesses, "s", false, "Shows single processes")
	customGetCmd.Flags().StringVarP(&customGetFlags.Format, "format", "f", "", "Request content type")
	customGetCmd.Flags().StringVarP(&customGetFlags.UserAgent, "user-agent", "u", "", "Custom User-Agent")
	customGetCmd.Flags().StringVarP(&customGetFlags.AuthToken, "auth-token", "a", "", "Authorization Token")
	customGetCmd.Flags().StringVarP(&customGetFlags.ClientVersion, "client-version", "v", "", "Client version")
	customGetCmd.Flags().StringVarP(&customGetFlags.ApiKey, "api-key", "k", "", "API key")
	customGetCmd.Flags().StringVarP(&customGetFlags.CorrelationID, "correlation-id", "c", "", "Correlation ID")
	customGetCmd.Flags().StringVarP(&customGetFlags.CustomHeader, "custom-header", "d", "", "Custom header")

	rootCmd.AddCommand(customGetCmd)
}

var customGetCmd = &cobra.Command{
	Use:   "getc [url]",
	Short: "Sends GET requests to a URL (customizable)",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Println("error creating GET request")
			return
		}

		headers := map[string]string{
			"Accept":            customGetFlags.Format,
			"User-Agent":        customGetFlags.UserAgent,
			"Authorization":     "Bearer " + customGetFlags.AuthToken,
			"X-Client-Version":  customGetFlags.ClientVersion,
			"X-Api-Key":         customGetFlags.ApiKey,
			"X-Correlation-ID":  customGetFlags.CorrelationID,
			"X-Custom-Header":   customGetFlags.CustomHeader,
		}

		// Iterates through the headers and checks if there is a value in the flag
		for key, value := range headers {
			if value != "" {
				req.Header.Set(key, value)
			}
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		fmt.Println("Response status:", resp.Status)
		fmt.Println("Response body:", string(body))
	},
}
