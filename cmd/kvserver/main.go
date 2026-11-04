package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tekka171/go-learning/kvstore"
	"github.com/tekka171/go-learning/server"
)

func main() {
	if err := run(); err != nil {
		log.Printf("run error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	// Define flags
	addr := flag.String("addr", ":6379", "address to listen on")
	cleanup := flag.Duration("cleanup", time.Second, "interval between expired-key sweeps (0 disables)")
	maxEntries := flag.Int("max-entries", 0, "maximum number of keys (0 means unlimited)")
	flag.Parse()

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		return fmt.Errorf("listen on %s:%w", *addr, err)
	}
	log.Printf("listening on %v", ln.Addr())

	store := kvstore.NewStore(kvstore.Config{
		CleanupInterval: *cleanup,
		MaxEntries:      *maxEntries,
	})
	defer store.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := server.New(store)
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- srv.Serve(ln)
	}()

	var serveFailure error
	select {
	case <-ctx.Done():
		log.Printf("signal received, shutting down")
	case err := <-serveErr:
		serveFailure = err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}

	if serveFailure != nil {
		return fmt.Errorf("serve: %w", serveFailure)
	}

	return nil
}
