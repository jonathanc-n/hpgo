package cmd

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var postFlags struct {
	NumWorkers          int
    ShowSingleProcesses bool
}

func init() {
	postCmd.Flags().IntVarP(&postFlags.NumWorkers, "workers", "w", 5, "Number of concurrent go workers")
    postCmd.Flags().BoolVar(&postFlags.ShowSingleProcesses, "s", false, "Shows single processes")
	rootCmd.AddCommand(postCmd)
}

var postCmd = &cobra.Command{
	Use:   "post [url] [numTimes]",
	Short: "Sends POST requests to a URL",
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
				fmt.Println("Error converting numTimes to integer:", err)
				return
			}
		}

		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "http://" + url
		}

		var wg sync.WaitGroup
		ch := make(chan string, times)

		for i := 0; i < times; i++ {
			wg.Add(1)
			go postRequest(url, &wg, ch)
		}

		go func() {
			wg.Wait()
			close(ch)
		}()

        for result := range ch {
			fmt.Println(result)
		}

		fmt.Println("Number of Requests:", times)
		fmt.Println("Method: 'POST'")
	},
}

func postRequest(url string, wg *sync.WaitGroup, ch chan<- string) {
	defer wg.Done()
	req, _ := http.NewRequest("POST", url, nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ch <- fmt.Sprintf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	ch <- fmt.Sprintf("Status: %s, Headers: %v", resp.Status, resp.Header)
}
