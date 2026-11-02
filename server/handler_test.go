package server

import (
	"bytes"
	"testing"
	"time"

	"github.com/tekka171/go-learning/kvstore"
)

func TestHandle(t *testing.T) {
	tests := []struct {
		name  string
		args  [][]byte
		want  string
		setup func(s *kvstore.Store)               //optional: set up the store before the command
		check func(t *testing.T, s *kvstore.Store) //optional: inspect the store after the command
	}{
		{
			name: "unknown command",
			args: cmd("foo"),
			want: "-ERR unknown command 'foo'\r\n",
		},
		{
			name: "unknown command with CRLF",
			args: cmd("f\ro\no"),
			want: "-ERR unknown command 'f o o'\r\n",
		},
		{
			name: "PING",
			args: cmd("PING"),
			want: "+PONG\r\n",
		},
		{
			name: "PING lowercase",
			args: cmd("ping"),
			want: "+PONG\r\n",
		},
		{
			name:  "GET hit",
			args:  cmd("GET", "key_1"),
			want:  "$7\r\nvalue_1\r\n",
			setup: func(s *kvstore.Store) { s.Set("key_1", []byte("value_1"), 0) },
		},
		{
			name:  "GET hit mixed case",
			args:  cmd("gEt", "key_1"),
			want:  "$7\r\nvalue_1\r\n",
			setup: func(s *kvstore.Store) { s.Set("key_1", []byte("value_1"), 0) },
		},
		{
			name: "GET miss",
			args: cmd("GET", "key_1"),
			want: "$-1\r\n",
		},
		{
			name: "GET wrong args alone",
			args: cmd("GET"),
			want: "-ERR wrong number of arguments for 'get' command\r\n",
		},
		{
			name: "GET wrong args more",
			args: cmd("GET", "a", "b"),
			want: "-ERR wrong number of arguments for 'get' command\r\n",
		},
		{
			name: "SET stores the value",
			args: cmd("SET", "key_1", "value_1"),
			want: "+OK\r\n",
			check: func(t *testing.T, s *kvstore.Store) {
				if val, ok := s.Get("key_1"); !ok || string(val) != "value_1" {
					t.Fatalf("store has %q, %v; want \"value_1\", true", val, ok)
				}
			},
		},
		{
			name: "SET TTL simple",
			args: cmd("SET", "key_1", "val_1", "EX", "200"),
			want: "+OK\r\n",
			check: func(t *testing.T, s *kvstore.Store) {
				_, exists, hasTTL := s.TTL("key_1")
				if !exists {
					t.Fatalf("expect 'key_1' exist")
				}

				if !hasTTL {
					t.Fatalf("expect 'key_1' has TTL")
				}
			},
		},
		{
			name: "SET TTL lowercase",
			args: cmd("SET", "key_1", "val_1", "ex", "200"),
			want: "+OK\r\n",
			check: func(t *testing.T, s *kvstore.Store) {
				_, exists, hasTTL := s.TTL("key_1")
				if !exists {
					t.Fatalf("expect 'key_1' exist")
				}

				if !hasTTL {
					t.Fatalf("expect 'key_1' has TTL")
				}
			},
		},
		{
			name: "SET TTL huge value but valid",
			args: cmd("SET", "key_1", "val_1", "EX", "9223372036"),
			want: "+OK\r\n",
			check: func(t *testing.T, s *kvstore.Store) {
				_, exists, hasTTL := s.TTL("key_1")
				if !exists || !hasTTL {
					t.Fatalf("exists=%v hasTTL=%v, want both true", exists, hasTTL)
				}
			},
		},
		{
			name: "SET TTL overflow",
			args: cmd("SET", "key_1", "val_1", "EX", "9223372037"),
			want: "-ERR value is not an integer or out of range\r\n",
			check: func(t *testing.T, s *kvstore.Store) {
				if _, ok := s.Get("key_1"); ok {
					t.Fatalf("key should not be stored when EX is rejected")
				}
			},
		},
		{
			name: "SET no TTL",
			args: cmd("SET", "key_1", "val_1"),
			want: "+OK\r\n",
			check: func(t *testing.T, s *kvstore.Store) {
				_, exists, hasTTL := s.TTL("key_1")
				if !exists {
					t.Fatalf("expect 'key_1' exist")
				}

				if hasTTL {
					t.Fatalf("expect 'key_1' should have no TTL")
				}
			},
		},
		{
			name: "SET TTL wrong option",
			args: cmd("SET", "key_1", "val_1", "EX_WRONG", "200"),
			want: "-ERR unknown option 'EX_WRONG'\r\n",
		},
		{
			name: "SET TTL unknown option with CRLF",
			args: cmd("SET", "key_1", "val_1", "EX\nWRONG", "200"),
			want: "-ERR unknown option 'EX WRONG'\r\n",
		},
		{
			name: "SET TTL with expiry zero",
			args: cmd("SET", "key_1", "val_1", "EX", "0"),
			want: "-ERR value is not an integer or out of range\r\n",
		},
		{
			name: "SET TTL with negative expiry value",
			args: cmd("SET", "key_1", "val_1", "EX", "-123"),
			want: "-ERR value is not an integer or out of range\r\n",
		},
		{
			name: "SET TTL wrong invalid expiry value",
			args: cmd("SET", "key_1", "val_1", "EX", "abc"),
			want: "-ERR value is not an integer or out of range\r\n",
		},
		{
			name:  "SET simple overwrite",
			args:  cmd("SET", "key_1", "val_100"),
			want:  "+OK\r\n",
			setup: func(s *kvstore.Store) { s.Set("key_1", []byte("val_1"), 0) },
			check: func(t *testing.T, s *kvstore.Store) {
				val, ok := s.Get("key_1")
				if !ok {
					t.Fatalf("key 'key_1' not exists")
				}

				if string(val) != "val_100" {
					t.Fatalf("get %q, want 'val_100'", val)
				}
			},
		},
		{
			name: "SET wrong args",
			args: cmd("SET"),
			want: "-ERR wrong number of arguments for 'set' command\r\n",
		},
		{
			name: "SET wrong args with key only",
			args: cmd("SET", "key_1"),
			want: "-ERR wrong number of arguments for 'set' command\r\n",
		},
		{
			name: "SET wrong args with no option value",
			args: cmd("SET", "key_1", "val_1", "EX"),
			want: "-ERR wrong number of arguments for 'set' command\r\n",
		},
		{
			name: "DEL wrong args alone",
			args: cmd("DEL"),
			want: "-ERR wrong number of arguments for 'del' command\r\n",
		},
		{
			name: "DEL miss",
			args: cmd("DEL", "key_1"),
			want: ":0\r\n",
		},
		{
			name:  "DEL one exist",
			args:  cmd("DEL", "key_1"),
			want:  ":1\r\n",
			setup: func(s *kvstore.Store) { s.Set("key_1", []byte("val_1"), 0) },
		},
		{
			name: "DEL multiple mixed",
			args: cmd("DEL", "key_1", "key_2", "key_3"),
			want: ":2\r\n",
			setup: func(s *kvstore.Store) {
				s.Set("key_1", []byte("val_1"), 0)
				s.Set("key_3", []byte("val_3"), 0)
			},
		},
		{
			name:  "EXPIRE exists",
			args:  cmd("EXPIRE", "key_1", "20"),
			want:  ":1\r\n",
			setup: func(s *kvstore.Store) { s.Set("key_1", []byte("val_1"), 0) },
			check: func(t *testing.T, s *kvstore.Store) {
				_, exists, hasTTL := s.TTL("key_1")
				if !exists {
					t.Fatalf("expect 'key_1' exist")
				}

				if !hasTTL {
					t.Fatalf("expect 'key_1' has TTL")
				}
			},
		},
		{
			name: "EXPIRE not exists",
			args: cmd("EXPIRE", "key_1", "20"),
			want: ":0\r\n",
		},
		{
			name: "EXPIRE wrong args",
			args: cmd("EXPIRE"),
			want: "-ERR wrong number of arguments for 'expire' command\r\n",
		},
		{
			name: "EXPIRE non-integer value",
			args: cmd("EXPIRE", "key_1", "a"),
			want: "-ERR value is not an integer or out of range\r\n",
		},
		{
			name: "EXPIRE zero value on missing key",
			args: cmd("EXPIRE", "key_1", "0"),
			want: ":0\r\n",
		},
		{
			name:  "EXPIRE zero value",
			args:  cmd("EXPIRE", "key_1", "0"),
			want:  ":1\r\n",
			setup: func(s *kvstore.Store) { s.Set("key_1", []byte("val_1"), 100*time.Second) },
			check: func(t *testing.T, s *kvstore.Store) {
				if _, ok := s.Get("key_1"); ok {
					t.Fatalf("Expire with 0 should delete the key")
				}
			},
		},
		{
			name:  "EXPIRE negative value",
			args:  cmd("EXPIRE", "key_1", "-50"),
			want:  ":1\r\n",
			setup: func(s *kvstore.Store) { s.Set("key_1", []byte("val_1"), 100*time.Second) },
			check: func(t *testing.T, s *kvstore.Store) {
				if _, ok := s.Get("key_1"); ok {
					t.Fatalf("Expire with negative value should delete the key")
				}
			},
		},
		{
			name:  "EXPIRE overflow",
			args:  cmd("EXPIRE", "key_1", "9223372037"),
			want:  "-ERR value is not an integer or out of range\r\n",
			setup: func(s *kvstore.Store) { s.Set("key_1", []byte("val_1"), 0) },
			check: func(t *testing.T, s *kvstore.Store) {
				_, exists, hasTTL := s.TTL("key_1")
				if !exists || hasTTL {
					t.Fatalf("exists=%v hasTTL=%v, want true, false", exists, hasTTL)
				}
			},
		},
		{
			name:  "TTL non-expired key",
			args:  cmd("TTL", "key_1"),
			want:  ":-1\r\n",
			setup: func(s *kvstore.Store) { s.Set("key_1", []byte("value_1"), 0) },
		},
		{
			name: "TTL missing key",
			args: cmd("TTL", "key_1"),
			want: ":-2\r\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := kvstore.NewStore(kvstore.Config{})
			defer s.Close()
			if tc.setup != nil {
				tc.setup(s)
			}

			var buf bytes.Buffer
			if err := handle(s, tc.args, &buf); err != nil {
				t.Fatalf("err: %s", err)
			}

			if got := buf.String(); got != tc.want {
				t.Fatalf("get %q, want %q", got, tc.want)
			}

			if tc.check != nil {
				tc.check(t, s)
			}
		})
	}
}

func TestHandleTTLWithDeadline(t *testing.T) {
	s := kvstore.NewStore(kvstore.Config{})
	defer s.Close()

	s.Set("key_1", []byte("val_1"), 100*time.Second)

	var buf bytes.Buffer
	if err := handle(s, cmd("TTL", "key_1"), &buf); err != nil {
		t.Fatalf("err: %v", err)
	}

	if got := buf.String(); got != ":100\r\n" && got != ":99\r\n" {
		t.Fatalf("reply = %q, want \":100\\r\\n\" or \":99\\r\\n\"", got)
	}
}

func cmd(s ...string) [][]byte {
	res := make([][]byte, len(s))
	for i := range s {
		res[i] = []byte(s[i])
	}
	return res
}
