// cmd/head.go
//
// This is all making GET requests to the same url, this
// process can be simplified using the 'sync' library with the WaitGroup
// function.

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
)

var maxStressFlags struct {
	NumWorkers  		int
	ShowSingleProcesses bool
	MaxTime				time.Duration
}

func init() {
	// maxStressCmd.Flags().IntVarP(&maxStressFlags.NumWorkers, "workers", "w", 5, "Number of concurrent go workers")
	maxStressCmd.Flags().BoolVar(&maxStressFlags.ShowSingleProcesses, "s", false, "Shows single processes")
	maxStressCmd.Flags().DurationVarP(&maxStressFlags.MaxTime, "max-time", "t", 1 * time.Millisecond, "Holds the max time for a variable")
	rootCmd.AddCommand(maxStressCmd)
}

var maxStressCmd = &cobra.Command{
	Use:   "stressm [url]",
	Short: "Stress tests a url",
	Args:  cobra.ExactArgs(1),
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

		incrementArray := [5]int{100, 50, 10, 5, 1}
		increment := 0
		checkDuration := time.Duration(0);
		ch := make(chan measuredResponse)

		for checkDuration < maxStressFlags.MaxTime && increment < 5{
			var wg sync.WaitGroup

			ch = make(chan measuredResponse)
			startTime := time.Now()
			for i := 0; i < times; i++ {
				wg.Add(1)
				go maxStressRequest(url, &wg, ch)
			}
			go func() {
				wg.Wait()
				close(ch)
			}()
			
			checkTime := time.Since(startTime)
			fmt.Println("pog", checkTime)
	
			if checkTime > maxStressFlags.MaxTime && increment == 4 {
				for response := range ch {
					result.TotalDNSTimeRecorded += response.DNS
					result.TotalConnectTimeRecorded += response.Connect
					result.TotalTLSTimeRecorded += response.TLS
					result.TotalTimeRecorded += response.TotalTime
					result.Status[response.Status]++
					if result.Fastest > time.Duration(response.TotalTime) {
						result.Fastest = time.Duration(response.TotalTime)
					}
					if result.Slowest < time.Duration(response.TotalTime) {
						result.Slowest = time.Duration(response.TotalTime)
					}
				}
				break
			}

			if checkTime > maxStressFlags.MaxTime {
				increment += 1
				time.Sleep(1 * time.Second)
				continue
			}
			checkDuration = checkTime
			times += incrementArray[increment]
			time.Sleep(1 * time.Second)
		}
		
		averageDNSTime := result.TotalDNSTimeRecorded / time.Duration(times)
		averageConnectTime := result.TotalConnectTimeRecorded / time.Duration(times)
		averageTLSTime := result.TotalTLSTimeRecorded / time.Duration(times)
		averageTime := result.TotalTimeRecorded / time.Duration(times)

		fmt.Println("Last stress test run results:")
		fmt.Println("Max number of requests: ", times)
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

func maxStressRequest(url string, wg *sync.WaitGroup, ch chan <- measuredResponse) {
	defer wg.Done()
	req, _ := http.NewRequest("GET", url, nil)

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

		// From when the first byte is registered back
		GotFirstResponseByte: func() {
			measured.TotalTime = time.Since(start)
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	start = time.Now()

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return
	}
	fmt.Println(start)
	measured.Start = start
	measured.Res = resp
	measured.Status = resp.Status
	defer resp.Body.Close()
	
	if stressFlags.ShowSingleProcesses {
		fmt.Printf("Status: %s\nTotal Time: %v\n\n", resp.Status, measured.TotalTime)
	}
	ch <- measured
}