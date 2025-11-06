package main

import (
	"log"
	"net/http"
	"os"

	"github.com/sorin-work/workout-log/internal/api"
)

func main() {
	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	srv := api.NewServer()
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatal(err)
	}
}
