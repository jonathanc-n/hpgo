// cmd/start.go
//
// This is all making GET requests to the same url, this
// process can be simplified using the 'sync' library with the WaitGroup
// function.

package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [url] [numTimes]",
	Short: "Make GET request",
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

		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
            url = "http://" + url
        }
	
		var wg sync.WaitGroup

		for i := 0; i < times; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				resp, err := http.Get(url)
				if err != nil {
					fmt.Println("error making GET request: ", err)
					return
				}
				defer resp.Body.Close()
	
				fmt.Println("Status:", resp.Status)
				fmt.Println("Header: ", resp.Header)
				fmt.Println("Body:")
				_, err = io.Copy(os.Stdout, resp.Body)
				if err != nil {
					fmt.Println("error reading response body: ", err)
					return
				}
			}()
		}
		wg.Wait()
		fmt.Println("All requests completed.")
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
