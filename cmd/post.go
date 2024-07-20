// cmd/post.go
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
)

var postCmd = &cobra.Command{
	Use:   "post",
	Short: "Make POST request",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting the web server on port 8080...")

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodHead {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
				return
			}
			fmt.Fprintf(w, "Hello, World!")
		})

		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Printf("Failed to start server: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(postCmd)
}
