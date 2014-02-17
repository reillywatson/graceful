// Copyright 2014 Joshua Marsh. All rights reserved. Use of this
// source code is governed by the MIT license that can be found in the
// LICENSE file.

package main

import (
	"fmt"
	"github.com/icub3d/graceful"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Create our server.
	s := graceful.NewServer(&http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Sleep for a bit so we can open some connections and send the
			// signal.
			time.Sleep(10 * time.Second)
			log.Printf("%v %v", r.Method, r.URL)
			fmt.Fprintln(w, r.Method, r.URL)
		}),
	})

	// Listen for the SIGTERM.
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Print("got SIGHUP, shutting down.")
		s.Close()
	}()

	// Start the server.
	fmt.Println("Using PID:", os.Getpid())
	log.Print(s.ListenAndServe())

	// At this point, try opening a few connection in another
	// terminal. Then in another, send a TERM kignal.
	// For example, in terminal one:
	//
	//   > go run main.go
	//   Using PID: 22191
	//
	// Then in terminal two:
	//
	//   > curl localhost:8080
	//
	// Then in terminal three:
	//
	//   > kill 22191
	//
	// Note that terminal two still gets a response and terminal one
	// remains open until the response is sent.
}
