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
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				// do stuff
				cluster = data.HandleDataUpdates()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	handleFunc := web.BuildHandleFunc(&cluster)
	http.HandleFunc("/", handleFunc)
	http.ListenAndServe(":8000", nil)
}
