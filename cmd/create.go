// cmd/create.go
package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"strings"
	"path/filepath"
	"bufio"
)

var createCmd = &cobra.Command{
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

		filePath := filepath.Join("executable", fileName)

		if _, err := os.Stat(filePath); err == nil {
			fmt.Printf("File '%s' already exists. Do you want to replace it? (y/n): ", fileName)
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			if response != "y\n" && response != "Y\n" {
				fmt.Println("File not replaced.")
				return
			}
		} else if !os.IsNotExist(err) {
			fmt.Println("Error checking file:", err)
			return
		}
	
		file, err := os.Create(filePath)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		defer file.Close()
		
		
		fmt.Println("File created successfully: ", fileName)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
