// cmd/head.go
package cmd

import (
	"fmt"
	// "io"
	// "net/http"
	// "os"
	// "strconv"
	"strings"
	// "sync"
	"github.com/spf13/cobra"
)

var headCmd = &cobra.Command{
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
        var err error

        if len(args) == 2 {
            if err != nil {
                fmt.Println("Error converting numberOfTimes to integer:", err)
                return
            }
        }

		// URL must be in the format of 'example.com' or 'http://example.com'
		// Ensures the input url contains http at the front
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
            url = "http://" + url
		}
	},
}

func init() {
	rootCmd.AddCommand(headCmd)
}
