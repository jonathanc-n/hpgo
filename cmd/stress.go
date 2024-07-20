// cmd/start.go
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
	"net/http/httptrace"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stressCmd)
}

type record struct {
	URL		  		  		 string
	Method	  		  		 string
	TotalDNSTimeRecorded	 time.Duration
	TotalConnectTimeRecorded time.Duration
	TotalTLSTimeRecorded 	 time.Duration
	TotalTimeRecorded 		 time.Duration
}

var stressCmd = &cobra.Command{
	Use:   "stress [url] [numTimes]",
	Short: "Stress tests a url",
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
		result := record{ URL : url, Method : "GET", TotalTimeRecorded : 0}
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

		ch := make(chan measuredResponse)
		for i := 0; i < times; i++ {
			wg.Add(1)
			go getRequest(url, &wg, ch)
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
		}
		
		averageDNSTime := result.TotalDNSTimeRecorded / time.Duration(times)
		averageConnectTime := result.TotalConnectTimeRecorded / time.Duration(times)
		averageTLSTime := result.TotalTLSTimeRecorded / time.Duration(times)
		averageTime := result.TotalTimeRecorded / time.Duration(times)

		fmt.Println("Number of Requests: ", times)
		fmt.Println("Method: 'GET'")
		fmt.Println("Average DNS Runtime: ", averageDNSTime)
		fmt.Println("Average Connect Runtime: ", averageConnectTime)
		fmt.Println("Average TLS Runtime: ", averageTLSTime)
		fmt.Println("Average Total Runtime: ", averageTime)
	},
}

type measuredResponse struct {
	Res       *http.Response
	Start     time.Time
	DNS       time.Duration
	Connect   time.Duration
	TLS       time.Duration
	TotalTime time.Duration
}

func getRequest(url string, wg *sync.WaitGroup, ch chan <- measuredResponse) {
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

		// From when the first byte response
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
	defer resp.Body.Close()
	
	fmt.Printf("Status: %s\nDNS: %v \nTotal Time: %v\n\n", resp.Status, measured.DNS, measured.TotalTime)
	ch <- measured
}