package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"simple-http-mux/internal/http_mux"
	"syscall"
	"time"
)

func main() {

	mux := http_mux.NewHttpMux(10000, 100)

	go func() {
		_ = mux.Run()
	}()

	//perform graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := mux.Shutdown(ctx); err != nil {
		log.Printf("Error occured on server shutting down: %s", err.Error())
	}

}
