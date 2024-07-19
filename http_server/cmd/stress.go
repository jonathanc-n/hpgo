// cmd/stress.go
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"github.com/spf13/cobra"
)

var stressCmd = &cobra.Command{
	Use:   "head [url] [numberOfTimes]",
	Short: "Make HEAD request",
	Args: func(cmd *cobra.Command, args []string) error {
        if len(args) < 1 {
            return fmt.Errorf("requires at least one argument")
        }
        if len(args) > 2 {
            return fmt.Errorf("requires at most two arguments")
        }
        return nil
    },
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		times := 1
        var err error

        if len(args) == 2 {
            times, err = strconv.Atoi(args[1])
            if err != nil {
                fmt.Println("Error converting numberOfTimes to integer:", err)
                return
            }
        }

		// Ensures the input url contains http at the front
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
            url = "http://" + url
        }

		for i := 0; i < times; i++ {
			resp, err := http.Get(url)
			if err != nil {
				fmt.Println("error making GET request: ", err)
				return
			}
			defer resp.Body.Close()

			fmt.Println("Status:", resp.Status)

			fmt.Println("Body:")
			_, err = io.Copy(os.Stdout, resp.Body)
			if err != nil {
				fmt.Errorf("error reading response body: %w", err)
				return
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(stressCmd)
}
