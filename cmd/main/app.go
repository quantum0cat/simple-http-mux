package main

import (
	"context"
	"flag"
	"github.com/quantum0cat/simple-http-mux/internal/http_mux"
	"github.com/quantum0cat/simple-http-mux/pkg/logging"
	"log"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const defaultPort = 10000
const defaultMaxConns = 100

func main() {
	logging.Init()
	//better to move it to config, but we got no external modules limitation
	portVal := flag.Uint("p", defaultPort, "port to listen on")
	maxConns := flag.Uint("m", defaultMaxConns, "max connections limit (0 -> no limit)")
	flag.Parse()

	var port uint16
	if *portVal > math.MaxUint16 {
		port = defaultPort
	} else {
		port = uint16(*portVal)
	}

	//propagate context to stop requests from being processed
	serverCtx, serverStop := context.WithCancel(context.Background())
	defer serverStop()

	mux := http_mux.NewHttpMux(serverCtx, port, *maxConns)

	go func() { _ = mux.Run() }()

	//perform graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	serverStop()

	//give the server a few seconds to close listeners in a gentle way
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := mux.Shutdown(ctx); err != nil {
		log.Printf("Error occured during server shutdown: %s", err.Error())
	}

}
