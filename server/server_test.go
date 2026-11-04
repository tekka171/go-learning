package server

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/tekka171/go-learning/kvstore"
)

func send(t *testing.T, conn net.Conn, s string) {
	t.Helper()

	_, err := conn.Write([]byte(s))
	if err != nil {
		t.Fatalf("write: %q, err: %v", s, err)
	}
}

func expectLine(t *testing.T, r *bufio.Reader, want string) {
	t.Helper()

	reply, err := r.ReadString('\n')
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if reply != want {
		t.Fatalf("reply: %q, want: %q", reply, want)
	}
}

func startServer(t *testing.T) (*Server, net.Listener, <-chan error) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	srv := New(kvstore.NewStore(kvstore.Config{}))
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- srv.Serve(ln)
	}()
	t.Cleanup(func() { srv.Shutdown(context.Background()) }) // stops Serve when the test ends
	return srv, ln, serveErr
}

func dial(t *testing.T, ln net.Listener) net.Conn {
	t.Helper()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	conn.SetDeadline(time.Now().Add(time.Second))

	return conn
}

func TestServeConnPing(t *testing.T) {
	srv := New(kvstore.NewStore(kvstore.Config{}))

	// 1. create the two ends
	client, server := net.Pipe()
	defer client.Close()

	// 2. give the "server" end to your code, running in the background
	go srv.serveConn(server)

	// 3. the test acts as the client, using the "client" end
	client.SetDeadline(time.Now().Add(time.Second))

	// send a command, as raw RESP bytes
	send(t, client, "*1\r\n$4\r\nPING\r\n")

	// read the reply
	r := bufio.NewReader(client)
	expectLine(t, r, "+PONG\r\n")
}

func TestServeConnProtocolError(t *testing.T) {
	srv := New(kvstore.NewStore(kvstore.Config{}))

	// 1. create the two ends
	client, server := net.Pipe()
	defer client.Close()

	// 2. give the "server" end to your code, running in the background
	go srv.serveConn(server)

	// 3. the test acts as the client, using the "client" end
	client.SetDeadline(time.Now().Add(time.Second))

	// send a command, with incorrect protocol
	send(t, client, "hello\r\n")

	// read the reply
	r := bufio.NewReader(client)
	expectLine(t, r, "-ERR protocol error\r\n")

	reply, err := r.ReadString('\n')
	if !errors.Is(err, io.EOF) {
		t.Fatalf("second read: %q %v; want: io.EOF", reply, err)
	}
}

func TestServeConnClientDisconnect(t *testing.T) {
	srv := New(kvstore.NewStore(kvstore.Config{}))

	client, server := net.Pipe()

	done := make(chan struct{})
	go func() {
		srv.serveConn(server)
		close(done)
	}()

	client.Close()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("serveConn did not return after the client disconnected")
	}
}

func TestServeConnSetGet(t *testing.T) {
	store := kvstore.NewStore(kvstore.Config{})

	// 1. create the two ends
	client, server := net.Pipe()
	defer client.Close()

	// 2. give the "server" end to your code, running in the background
	srv := New(store)
	go srv.serveConn(server)

	// 3. the test acts as the client, using the "client" end
	client.SetDeadline(time.Now().Add(time.Second))

	// send a command, as raw RESP bytes
	send(t, client, "*3\r\n$3\r\nSET\r\n$4\r\nkey1\r\n$4\r\nval1\r\n")

	r := bufio.NewReader(client)

	expectLine(t, r, "+OK\r\n")

	send(t, client, "*2\r\n$3\r\nGET\r\n$4\r\nkey1\r\n")

	// first part reply
	expectLine(t, r, "$4\r\n")

	// second part of reply
	expectLine(t, r, "val1\r\n")
}

func TestServeTCPPing(t *testing.T) {
	_, ln, _ := startServer(t)

	conn := dial(t, ln)

	send(t, conn, "*1\r\n$4\r\nPING\r\n")

	r := bufio.NewReader(conn)
	expectLine(t, r, "+PONG\r\n")
}

func TestServeTCPSetGet(t *testing.T) {
	_, ln, _ := startServer(t)

	conn1 := dial(t, ln) // SET with 1st conn

	send(t, conn1, "*3\r\n$3\r\nSET\r\n$5\r\nkey_1\r\n$5\r\nval_1\r\n")

	r1 := bufio.NewReader(conn1)
	expectLine(t, r1, "+OK\r\n")

	conn2 := dial(t, ln) // GET with 2nd conn

	send(t, conn2, "*2\r\n$3\r\nGET\r\n$5\r\nkey_1\r\n")

	r2 := bufio.NewReader(conn2)
	expectLine(t, r2, "$5\r\n")
	expectLine(t, r2, "val_1\r\n")
}

func TestShutdownNoClient(t *testing.T) {
	srv, _, serveErr := startServer(t)
	if err := srv.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	// Serve should close itself after Shutdown
	select {
	case err := <-serveErr:
		if !errors.Is(err, ErrServerClosed) {
			t.Fatalf("Serve: %v, want ErrServerClosed", err)
		}

	case <-time.After(time.Second):
		t.Fatalf("Serve did not return after Shutdown")
	}

}

func TestShutdownWithActiveClient(t *testing.T) {
	srv, ln, _ := startServer(t)
	conn := dial(t, ln)

	// make sure the connection is really being served
	send(t, conn, "*1\r\n$4\r\nPING\r\n")
	expectLine(t, bufio.NewReader(conn), "+PONG\r\n")

	// call Shutdown in the background; its result arrives on the channel
	result := make(chan error, 1)
	go func() {
		result <- srv.Shutdown(context.Background())
	}()

	// 1. it must NOT return while the client is still connected
	select {
	case err := <-result:
		t.Fatalf("Shutdown returned %v while a client was still connected", err)

	case <-time.After(50 * time.Millisecond):
		//still waiting: ok
	}

	// 2. the client leaves...
	conn.Close()

	// 3. ...so Shutdown should now return nil
	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("Shutdown: %v, want nil", err)
		}
		// no error: ok

	case <-time.After(time.Second):
		t.Fatalf("Shutdown did not return after the client disconnected")
	}
}

func TestShutdownBeforeServe(t *testing.T) {
	srv := New(kvstore.NewStore(kvstore.Config{}))

	if err := srv.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}

	if err := srv.Serve(ln); !errors.Is(err, ErrServerClosed) {
		t.Fatalf("Serve: %v, want ErrServerClosed", err)
	}
}

func TestShutdownTimeout(t *testing.T) {
	srv, ln, _ := startServer(t)
	conn := dial(t, ln)

	send(t, conn, "*1\r\n$4\r\nPING\r\n")
	r := bufio.NewReader(conn)
	expectLine(t, r, "+PONG\r\n")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if err := srv.Shutdown(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Shutdown: %v, want context.DeadlineExceeded", err)
	}

	_, err := r.Read(make([]byte, 1))
	if err == nil {
		t.Fatal("read succeeded; want the connection to be closed by the server")
	}

	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		t.Fatal("read timed out; the server did not close the connection")
	}
}
