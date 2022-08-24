/*
 * Go program to gather information on a Kubernetes cluster
 * Gathers information on each node, then uses nvidia-smi and
 * dcgm-exporter to get GPU metrics for the cluster
 */

package main

import (
	"fmt"
	data "internal/datagathering"
	web "internal/webhosting"
	"net/http"
	"os"
	"strconv"
    "sync"
	"time"
)

func main() {
	// initial data gathering
	cluster := data.HandleDataUpdates()

	// periodic data gathering
	var rate int
	rate, err := strconv.Atoi(os.Getenv("RATE"))
	if err != nil {
		fmt.Println("Error reading rate variable as integer")
	}
	// default rate is 24 (hours)
	ticker := time.NewTicker(time.Duration(rate) * time.Hour)
	quit := make(chan bool)
    var wg sync.WaitGroup
    wg.Add(1)
	go func() {
        running := true
		for running {
			select {
			case <-ticker.C:
				cluster = data.HandleDataUpdates()
			case <-quit:
				running = false
                ticker.Stop()
			}
		}
        wg.Done()
	}()

	deployWeb, err := strconv.ParseBool(os.Getenv("WEB"))
	if err != nil {
		fmt.Println("Error reading web variable as bool")
	}

    if deployWeb {
        fmt.Println("Hosting Web")
    	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
    	handleFunc := web.BuildHandleFunc(&cluster)
    	http.HandleFunc("/", handleFunc)
    	http.ListenAndServe(":8080", nil)
    }
    wg.Wait()
}
