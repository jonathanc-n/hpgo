// cmd/custom_post.go
package cmd

import (
	"fmt"
	"net/http"
	"io"
	"github.com/spf13/cobra"
)

var customPostFlags struct {
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
	customPostCmd.Flags().IntVarP(&customPostFlags.NumWorkers, "workers", "w", 5, "Number of concurrent go workers")
    customPostCmd.Flags().BoolVar(&customPostFlags.ShowSingleProcesses, "s", false, "Shows single processes")
	customPostCmd.Flags().StringVarP(&customPostFlags.Format, "format", "f", "", "Request content type")
	customPostCmd.Flags().StringVarP(&customPostFlags.UserAgent, "user-agent", "u", "", "Custom User-Agent")
	customPostCmd.Flags().StringVarP(&customPostFlags.AuthToken, "auth-token", "a", "", "Authorization Token")
	customPostCmd.Flags().StringVarP(&customPostFlags.ClientVersion, "client-version", "v", "", "Client version")
	customPostCmd.Flags().StringVarP(&customPostFlags.ApiKey, "api-key", "k", "", "API key")
	customPostCmd.Flags().StringVarP(&customPostFlags.CorrelationID, "correlation-id", "c", "", "Correlation ID")
	customPostCmd.Flags().StringVarP(&customPostFlags.CustomHeader, "custom-header", "d", "", "Custom header")

	rootCmd.AddCommand(customPostCmd)
}

var customPostCmd = &cobra.Command{
	Use:   "postc [url]",
	Short: "Send a POST requests to a URL (customizable)",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]

		req, err := http.NewRequest("Post", url, nil)
		if err != nil {
			fmt.Println("error creating Post request")
			return
		}

		headers := map[string]string{
			"Accept":            customPostFlags.Format,
			"User-Agent":        customPostFlags.UserAgent,
			"Authorization":     "Bearer " + customPostFlags.AuthToken,
			"X-Client-Version":  customPostFlags.ClientVersion,
			"X-Api-Key":         customPostFlags.ApiKey,
			"X-Correlation-ID":  customPostFlags.CorrelationID,
			"X-Custom-Header":   customPostFlags.CustomHeader,
		}

		// Iterates through the headers and checks if there is a value in the flag
		for key, value := range headers {
			if value != "" {
				fmt.Println("Format: ", value)
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
