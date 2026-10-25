package main

import (
	"log"
	"net/http"
	"os"

	"github.com/JaniL/bmur2"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/telegram/webhook", bmur2.Handle)
	log.Printf("webhook server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
