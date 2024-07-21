// cmd/create.go
package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"strings"
	"path/filepath"
)

var removeCmd = &cobra.Command{
	Use:   "remove [fileName]",
	Short: "Creates txt file in current directory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fileName := args[0]

		if !strings.HasSuffix(fileName, ".txt") {
			fileName = fileName + ".txt"
		}

		filePath := filepath.Join("executable", fileName)

		if _, err := os.Stat(filePath); err == nil {
			err := os.Remove(filePath)
			if err != nil {
				fmt.Println("Error removing file:", err)
				return
			}
			fmt.Println("File removed successfully:", fileName)
		} else if os.IsNotExist(err) {
			fmt.Println("File does not exist:", fileName)
		} else {
			fmt.Println("Error checking file:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
