// cmd/stress_api.go
//
// Uses a Transport instance to create a keep alive connection
// for faster requests to an API

package cmd

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"math"
	"net/http/httptrace"
	"github.com/spf13/cobra"
	"io"
)

var stressAPIFlags struct {
	NumWorkers  		int
	ShowSingleProcesses bool
	ApiKey				string
}

func init() {
	stressAPICmd.Flags().IntVarP(&stressAPIFlags.NumWorkers, "workers", "w", 5, "Number of concurrent go workers")
	stressAPICmd.Flags().BoolVar(&stressAPIFlags.ShowSingleProcesses, "s", false, "Shows single processes")
	stressAPICmd.Flags().StringVarP(&stressAPIFlags.ApiKey, "api-key", "k", "", "API key")
	rootCmd.AddCommand(stressAPICmd)
}

func createHTTPClient() *http.Client {
	tr := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}
	return &http.Client{Transport: tr}
}


var stressAPICmd = &cobra.Command{
	Use:   "stressa [url] [numTimes]",
	Short: "Stress tests an api",
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
		result := record{ 
			URL : url, 
			Method : "GET", 
			Fastest : time.Duration(math.Inf(0)),
			Slowest : time.Duration(0),
			Status: make(map[string]int),
		} 

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

		ch := make(chan measuredResponse, stressFlags.NumWorkers)

		client := createHTTPClient()

		for i := 0; i < times; i++ {
			wg.Add(1)
			go stressAPIRequest(client, url, &wg, ch)
		}
		go func() {
			wg.Wait()
			close(ch)
		}()
	
		for response := range ch {
			result.TotalDNSTimeRecorded += response.DNS
			result.TotalConnectTimeRecorded += response.Connect
			result.TotalTLSTimeRecorded += response.TLS
			result.TotalTimeRecorded += response.TotalTime
			result.Status[response.Status]++
			if result.Fastest > response.TotalTime {
				result.Fastest = response.TotalTime
			}
			if result.Slowest < response.TotalTime {
				result.Slowest = response.TotalTime
			}
		}
		
		averageDNSTime := result.TotalDNSTimeRecorded / time.Duration(times)
		averageConnectTime := result.TotalConnectTimeRecorded / time.Duration(times)
		averageTLSTime := result.TotalTLSTimeRecorded / time.Duration(times)
		averageTime := result.TotalTimeRecorded / time.Duration(times)

		fmt.Println("Number of Requests:", times)
		fmt.Println("Method: 'GET'")
		fmt.Println("Number of concurrent workers:", stressFlags.NumWorkers)
		fmt.Println("Average DNS Runtime:", averageDNSTime)
		fmt.Println("Average Connect Runtime:", averageConnectTime)
		fmt.Println("Average TLS Runtime:", averageTLSTime)
		fmt.Println("Average Total Runtime:", averageTime)
		fmt.Println("Fastest Runtime: ", result.Fastest)
		fmt.Println("Slowest Runtime: ", result.Slowest)
		fmt.Println("Status Results: ")
		for status, count := range result.Status {
			fmt.Printf("%s: %d\n", status, count)
		}
	},
}
func stressAPIRequest(client *http.Client, url string, wg *sync.WaitGroup, ch chan<- measuredResponse) {
    defer wg.Done()

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        fmt.Println("Error creating request:", err)
        return
    }

    measured := measuredResponse{}
    var start, connect, dns, tlsHandshake time.Time

    trace := &httptrace.ClientTrace{
        DNSStart: func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
        DNSDone: func(ddi httptrace.DNSDoneInfo) {
            measured.DNS = time.Since(dns)
        },
        ConnectStart: func(network, addr string) { connect = time.Now() },
        ConnectDone: func(network, addr string, err error) {
            measured.Connect = time.Since(connect)
        },
        TLSHandshakeStart: func() { tlsHandshake = time.Now() },
        TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
            measured.TLS = time.Since(tlsHandshake)
        },
        GotFirstResponseByte: func() {
            measured.TotalTime = time.Since(start)
        },
    }

    req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
    start = time.Now()

    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("Error performing request:", err)
        return
    }
    measured.Start = start
    measured.Res = resp
    measured.Status = resp.Status
    defer resp.Body.Close()

    if stressAPIFlags.ShowSingleProcesses {
        fmt.Printf("Status: %s\nTotal Time: %v\n\n", resp.Status, measured.TotalTime)
        body, _ := io.ReadAll(resp.Body)
        fmt.Println("Body: ", string(body))
    }
    ch <- measured
}
