/*
Copyright (c) 2020 Sergey Anisimov

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/rb1980/smartthings-metrics/recording"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"os/signal"
	"strconv"
)

const (
	defaultPromAddr = ":9153"
)

var (
	exitFunc = os.Exit
	stopChan = make(chan os.Signal, 1)
)

func main() {
	var clientID, clientSecret, intervalStr, certFile, keyFile string

	flag.StringVar(&clientID, "client-id", os.Getenv("CLIENT_ID"), "The OAuth client ID.")
	flag.StringVar(&clientSecret, "client-secret", os.Getenv("CLIENT_SECRET"), "The OAuth client secret.")
	flag.StringVar(&intervalStr, "interval", os.Getenv("REFRESH_INTERVAL"), "The status refresh interval in seconds.")
	flag.StringVar(&certFile, "cert-file", os.Getenv("SSL_CERT_FILE"), "Path to SSL certificate file.")
	flag.StringVar(&keyFile, "key-file", os.Getenv("SSL_KEY_FILE"), "Path to SSL private key file.")

	flag.Parse()

	exitFunc(Run(clientID, clientSecret, intervalStr, certFile, keyFile))
}

func ParseInterval(intervalStr string) (int, error) {
	if len(intervalStr) > 0 {
		interval, err := strconv.Atoi(intervalStr)
		if err != nil {
			return 0, errors.New("the interval specified is not an integer")
		}
		if interval <= 0 {
			return 0, errors.New("the interval should be greater than zero")
		}
		return interval, nil
	}
	return 60, nil // Default interval is 60 seconds
}

func Run(clientID, clientSecret, intervalStr, certFile, keyFile string) int {
	if clientID == "" || clientSecret == "" {
		flag.PrintDefaults()
		return 1
	}

	interval, err := ParseInterval(intervalStr)
	if err != nil {
		_, _ = fmt.Fprint(flag.CommandLine.Output(), err)
		return 1
	}

	go servePrometheus(defaultPromAddr, certFile, keyFile)

	loop := recording.NewLoop(clientID, clientSecret, interval)
	loop.Start()

	signal.Notify(stopChan, os.Interrupt, os.Kill)
	<-stopChan

	return 0
}

func servePrometheus(addr, certFile, keyFile string) {
	http.Handle("/metrics", promhttp.Handler())
	
	var err error
	if certFile != "" && keyFile != "" {
		err = http.ListenAndServeTLS(addr, certFile, keyFile, nil)
	} else {
		err = http.ListenAndServe(addr, nil)
	}
	
	if err != nil {
		panic("unable to create HTTP server, error: " + err.Error())
	}
}
