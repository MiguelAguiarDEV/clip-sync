package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"clip-sync/server/internal/app"
)

func main() {
	a := app.NewApp()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: app.WithHTTPLogging(a.Mux),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()
	log.Println("clip-sync server listening on :8080")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	a.WSS.Shutdown(ctx)
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("http shutdown: %v", err)
	}
	log.Println("bye")
}
