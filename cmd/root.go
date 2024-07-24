/*
Copyright © 2024 NAME HERE chenleejonathan@gmail.com

*/
package cmd

import (
	"os"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hpgo",
	Short: "A http cli tool",
	Long: "HTTP CLIgo - a simple http cli tool in Go for basic/custom requests, api testing, debugging, etc.",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}


