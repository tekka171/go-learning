package resp

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestReadCommand(t *testing.T) {
	tests := []struct {
		name       string
		cmd        string
		wantErr    error
		wantValues [][]byte
	}{
		{
			name: "simple command",
			cmd: "*3\r\n" +
				"$3\r\nSET\r\n" +
				"$5\r\nkey_1\r\n" +
				"$5\r\nhello\r\n",
			wantValues: [][]byte{[]byte("SET"), []byte("key_1"), []byte("hello")},
		},
		{
			name: "values with CRLF",
			cmd: "*3\r\n" +
				"$3\r\nSET\r\n" +
				"$5\r\nkey_1\r\n" +
				"$4\r\na\r\nb\r\n",
			wantValues: [][]byte{[]byte("SET"), []byte("key_1"), []byte("a\r\nb")},
		},
		{
			name: "empty values",
			cmd: "*3\r\n" +
				"$3\r\nSET\r\n" +
				"$5\r\nkey_1\r\n" +
				"$0\r\n\r\n",
			wantValues: [][]byte{[]byte("SET"), []byte("key_1"), []byte("")},
		},
		{
			name:       "truncated command",
			cmd:        "*3\r\n",
			wantErr:    io.ErrUnexpectedEOF,
			wantValues: [][]byte{},
		},
		{
			name:    "empty input",
			cmd:     "",
			wantErr: io.EOF,
		},
		{
			name:    "cut off between elements",
			cmd:     "*2\r\n$3\r\nGET\r\n",
			wantErr: io.ErrUnexpectedEOF,
		},
		{
			name:    "wrong first byte",
			cmd:     "+OK\r\n",
			wantErr: ErrProtocol,
		},
		{
			name:    "non-numeric count",
			cmd:     "*abc\r\n",
			wantErr: ErrProtocol,
		},
		{
			name:    "zero count",
			cmd:     "*0\r\n",
			wantErr: ErrProtocol,
		},
		{
			name:    "short element header",
			cmd:     "*1\r\n$\n",
			wantErr: ErrProtocol,
		},
		{
			name:    "bare LF",
			cmd:     "*1\r\n$3\nGET\r\n",
			wantErr: ErrProtocol,
		},
		{
			name:    "negative length",
			cmd:     "*1\r\n$-1\r\n",
			wantErr: ErrProtocol,
		},
		{
			name:    "length over limit",
			cmd:     "*1\r\n$600000000\r\n",
			wantErr: ErrProtocol,
		},
		{
			name:    "payload without CRLF",
			cmd:     "*1\r\n$3\r\nGETxx",
			wantErr: ErrProtocol,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			strReader := strings.NewReader(tc.cmd)
			reader := bufio.NewReader(strReader)
			got, err := ReadCommand(reader)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err: %v, want %v", err, tc.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("err: %s", err)
			}

			if len(got) != len(tc.wantValues) {
				t.Fatalf("Get length = %v, want %v", len(got), len(tc.wantValues))
			}

			for i := range tc.wantValues {
				if !bytes.Equal(got[i], tc.wantValues[i]) {
					t.Fatalf("Get value = %v, want %v", got[i], tc.wantValues[i])
				}
			}
		})
	}
}

func TestReadMultiCommand(t *testing.T) {
	strReader := strings.NewReader("*3\r\n" +
		"$3\r\nSET\r\n" +
		"$5\r\nkey_1\r\n" +
		"$6\r\nhello1\r\n" +
		"*2\r\n" +
		"$3\r\nGET\r\n" +
		"$5\r\nkey_2\r\n")
	reader := bufio.NewReader(strReader)
	first, err := ReadCommand(reader)
	if err != nil {
		t.Fatalf("err: %s", err.Error())
	}

	if len(first) != 3 || string(first[0]) != "SET" || string(first[1]) != "key_1" || string(first[2]) != "hello1" {
		t.Fatalf("first call = %q", first)
	}

	second, err := ReadCommand(reader)
	if err != nil {
		t.Fatalf("err: %s", err.Error())
	}

	if len(second) != 2 || string(second[0]) != "GET" || string(second[1]) != "key_2" {
		t.Fatalf("second call = %q", second)
	}

	if _, err := ReadCommand(reader); !errors.Is(err, io.EOF) {
		t.Fatalf("third call err = %v, wants io.EOF", err)
	}
}
