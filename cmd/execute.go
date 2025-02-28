// cmd/execute.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"strings"
	"path/filepath"
	"strconv"
	"sync"
	"time"
	"math"
	"net/http"
	"net/http/httptrace"
	"crypto/tls"
)

var executeFlags struct {
	NumWorkers  		int
	ShowSingleProcesses bool
}

func init() {
	executeCmd.Flags().IntVarP(&executeFlags.NumWorkers, "workers", "w", 5, "Number of concurrent go workers")
	executeCmd.Flags().BoolVar(&executeFlags.ShowSingleProcesses, "s", false, "Shows single processes")
	rootCmd.AddCommand(executeCmd)
}

type lines struct {
	URL			string
	NumTimes	int
}

type urlMeasuredResponse struct {
	URL				 	string
	Method			 	string
	NumberOfRequests 	int
	Fastest 			time.Duration
	Slowest				time.Duration
	AverageDNSTime 		time.Duration
	AverageConnectTime 	time.Duration
	AverageTLSTime 		time.Duration
	AverageTotalTime 	time.Duration
	Status 				map[string]int			
}

var executeCmd = &cobra.Command{
	Use:   "execute [fileName]",
	Short: "Executes all lines in file concurrently",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fileName := args[0]
		if !strings.HasSuffix(fileName, ".txt") {
			fileName = fileName + ".txt"
		}

		filePath := filepath.Join("executable", fileName)


		if _, err := os.Stat(filePath); err == nil {
			file, err := os.Open(filePath)
			if err != nil {
				fmt.Println("Error creating file:", err)
				return
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			
			ch := make(chan urlMeasuredResponse)
	
			var waitGroupLine sync.WaitGroup
			for scanner.Scan() {
				waitGroupLine.Add(1)
				line := scanner.Text() 
				parts := strings.SplitN(line, " ", 2)
				
				numTimes, err := strconv.Atoi(parts[1])
				if err != nil {
					fmt.Println("Error converting numTimes to integer:", err)
					continue
				}
				url := parts[0]
				if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
					url = "http://" + url
				}

				addLine := lines {
					URL:      url,
					NumTimes: numTimes,
				}
				go executeLine(addLine, ch, &waitGroupLine)
			}

			go func() {
				waitGroupLine.Wait()
				close(ch)
			}()

			for response := range ch {
				fmt.Println("\nURL: ", response.URL)
				fmt.Println("Number of requests: ", response.NumberOfRequests)
				fmt.Println("Average DNS Time: ", response.AverageDNSTime)
				fmt.Println("Average Connect Time: ", response.AverageConnectTime)
				fmt.Println("Average TLS Time: ", response.AverageTLSTime)
				fmt.Println("Average Runtime: ", response.AverageTotalTime)
				fmt.Println("Fastest Runtime: ", response.Fastest)
				fmt.Println("Slowest Runtiem: ", response.Slowest)
				fmt.Println("Status Results: ")
				for status, count := range response.Status {
					fmt.Printf("%s: %d\n", status, count)
				}
			}
		} else if os.IsNotExist(err) {
			fmt.Println("File does not exist:", fileName)
			return
		} else {
			fmt.Println("Error checking file:", err)
			return
		}
	},
}

func executeLine(line lines, ch chan <- urlMeasuredResponse, waitGroupLine *sync.WaitGroup) {
	defer waitGroupLine.Done()
	var wg sync.WaitGroup
	result := record{
		Fastest: time.Duration(math.Inf(1)),
		Slowest: time.Duration(0),
		Status: make(map[string]int),
	}
	
	chMeasured := make(chan measuredResponse, executeFlags.NumWorkers)
	times := line.NumTimes
	for i := 0; i < line.NumTimes; i++ {
		wg.Add(1)
		go executeRequest(line.URL, &wg, chMeasured)
	}
	go func() {
		wg.Wait()
		close(chMeasured)
	}()


	for response := range chMeasured {
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
	
	newURLMeasuredResponse := urlMeasuredResponse{
		URL : 					line.URL,
		Method : 				"GET",
		Fastest : 				result.Fastest,
		Slowest : 				result.Slowest,
		NumberOfRequests : 		line.NumTimes,
		AverageDNSTime : 		result.TotalDNSTimeRecorded / time.Duration(times),
		AverageConnectTime : 	result.TotalConnectTimeRecorded / time.Duration(times),
		AverageTLSTime : 		result.TotalTLSTimeRecorded / time.Duration(times),
		AverageTotalTime : 		result.TotalTimeRecorded / time.Duration(times),
		Status : 				result.Status,
	}
	ch <- newURLMeasuredResponse
}	


func executeRequest(url string, wg *sync.WaitGroup, chMeasured chan <- measuredResponse) {
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
	measured.Start = start
	measured.Res = resp
	measured.Status = resp.Status
	defer resp.Body.Close()
	
	if executeFlags.ShowSingleProcesses {
		fmt.Printf("Status: %s\nTotal Time: %v\n\n", resp.Status, measured.TotalTime)
	}
	chMeasured <- measured
}
