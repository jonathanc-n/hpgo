// cmd/create.go
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	// "net/http"
)

var createCmd = &cobra.Command{
	Use:   "create [fileName]",
	Short: "Make POST request",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting the web server on port 8080...")
		
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
