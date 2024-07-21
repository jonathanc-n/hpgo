// cmd/create.go
package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"strings"
	"path/filepath"
)

var addCmd = &cobra.Command{
	Use:   "add [fileName] [url] [numTimes]",
	Short: "Creates txt file in current directory",
	Args: func(cmd *cobra.Command, args []string) error {
        if len(args) < 2 {
            return fmt.Errorf("requires at least one argument")
        }
        if len(args) > 3 {
            return fmt.Errorf("requires at most two arguments")
        }
        return nil
    },
	Run: func(cmd *cobra.Command, args []string) {
		fileName := args[0]
		times := "1"
		var err error

		if len(args) == 3 {
            times = args[2]
        }
		urlAdd := args[1] + " " + times
		
		if !strings.HasSuffix(fileName, ".txt") {
			fileName = fileName + ".txt"
		}

		filePath := filepath.Join("executable", fileName)

		file, err := os.Create(filePath)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		defer file.Close()

		_, err = file.WriteString(urlAdd)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}

		fmt.Println("File created successfully", fileName)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
