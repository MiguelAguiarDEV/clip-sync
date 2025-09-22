package main

import (
	"log"
	"net/http"

	"clip-sync/server/internal/app"
)

func main() {
	mux := app.NewMux()
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
