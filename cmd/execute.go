// cmd/execute.go
package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"strings"
)

var executeCmd = &cobra.Command{
	Use:   "create [fileName]",
	Short: "Creates txt file in current directory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fileName := args[0]
		
		err := os.MkdirAll("executable", os.ModePerm)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}

		if !strings.HasSuffix(fileName, ".txt") {
			fileName = fileName + ".txt"
		}

		file, err := os.Create(fileName)
		if err != nil {
			fmt.Println("Error creating file: ", err)
			return 
		}
		defer file.Close()

		fmt.Println("File created successfully", fileName)
	},
}

func init() {
	rootCmd.AddCommand(executeCmd)
}
