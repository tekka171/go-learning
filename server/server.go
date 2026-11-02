package server

import (
	"bufio"
	"context"
	"errors"
	"net"
	"sync"

	"github.com/tekka171/go-learning/kvstore"
	"github.com/tekka171/go-learning/resp"
)

var (
	ErrServerClosed = errors.New("server: closed")
)

type Server struct {
	store *kvstore.Store

	mu      sync.Mutex
	ln      net.Listener
	closing bool

	wg sync.WaitGroup
}

func New(store *kvstore.Store) *Server {
	return &Server{
		store: store,
	}
}

// Serve accepts connections on ln and handles each in its own goroutine.
// It returns ErrServerClosed after Shutdown is called, or the error from
// Accept if the listener fails for another reason.
func (s *Server) Serve(ln net.Listener) error {
	s.mu.Lock()
	if s.closing { // if Shutdown before Serve
		s.mu.Unlock()
		ln.Close()
		return ErrServerClosed
	}
	s.ln = ln
	s.mu.Unlock()

	for {
		conn, err := ln.Accept()
		if err != nil {
			s.mu.Lock()
			closing := s.closing
			s.mu.Unlock()

			if closing { // if Shutdown when there is active connection
				return ErrServerClosed
			}
			return err
		}

		s.mu.Lock()
		if s.closing { // if Shutdown right after a connection is made
			s.mu.Unlock()
			conn.Close()
			return ErrServerClosed
		}
		s.wg.Go(func() {
			s.serveConn(conn)
		})
		s.mu.Unlock()
	}
}

// serveConn handles one client connection until it disconnects.
func (s *Server) serveConn(conn net.Conn) {
	defer conn.Close()

	r := bufio.NewReader(conn)

	for {
		args, err := resp.ReadCommand(r)
		if err != nil {
			if errors.Is(err, resp.ErrProtocol) {
				_ = resp.WriteError(conn, "ERR protocol error") // best effort; closing anyway
			}
			return
		}

		err = handle(s.store, args, conn)
		if err != nil {
			return
		}
	}
}

// Shutdown stops accepting connections and waits for active ones to finish.
// If ctx ends first, it returns ctx.Err().
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	s.closing = true
	if s.ln != nil {
		s.ln.Close()
	}
	s.mu.Unlock()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
