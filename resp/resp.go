package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var (
	ErrProtocol = errors.New("resp: protocol error") // ErrProtocol is wrapped by every error ReadCommand returns for malformed input.
)

const (
	maxArrayLen = 1024
	maxBulkLen  = 512 * 1024 * 1024 // 512 MiB
)

// ReadCommand reads one command, an array of bulk strings.
// It returns io.EOF if the input ends cleanly before a command starts,
// io.ErrUnexpectedEOF if it ends partway through one, and an error
// wrapping protocol if the input is malformed.
func ReadCommand(r *bufio.Reader) ([][]byte, error) {
	count, err := readHeader(r, '*')
	if err != nil {
		return nil, err
	}

	if count <= 0 || count > maxArrayLen {
		return nil, fmt.Errorf("%w: bad array length", ErrProtocol)
	}

	res := make([][]byte, count)
	for i := range count {
		length, err := readHeader(r, '$')
		if err != nil {
			return nil, unexpectedEOF(err)
		}

		if length < 0 || length > maxBulkLen {
			return nil, fmt.Errorf("%w: bad bulk length", ErrProtocol)
		}

		buf := make([]byte, length)
		if _, err := io.ReadFull(r, buf); err != nil { // read the value by length
			return nil, unexpectedEOF(err)
		}

		var crlf [2]byte
		if _, err := io.ReadFull(r, crlf[:]); err != nil { // read the CRLF '\r\n'
			return nil, unexpectedEOF(err)
		}

		if !(crlf[0] == '\r' && crlf[1] == '\n') {
			return nil, fmt.Errorf("%w: line not terminated by CRLF", ErrProtocol)
		}

		res[i] = buf
	}

	return res, nil
}

func readHeader(r *bufio.Reader, prefix byte) (int, error) {
	line, err := r.ReadSlice('\n') // read the header
	if err != nil {
		if errors.Is(err, io.EOF) && len(line) > 0 {
			return 0, io.ErrUnexpectedEOF
		}

		if errors.Is(err, bufio.ErrBufferFull) {
			return 0, fmt.Errorf("%w: header too long", ErrProtocol)
		}

		return 0, err
	}

	if len(line) < 3 { // the minimum must have: prefix, '\r', '\n'
		return 0, fmt.Errorf("%w: header too short", ErrProtocol)
	}

	if line[0] != prefix { // check prefix
		return 0, fmt.Errorf("%w: expected %q", ErrProtocol, prefix)
	}

	if line[len(line)-2] != '\r' { // check crlf
		return 0, fmt.Errorf("%w: line not terminated by CRLF", ErrProtocol)
	}

	line = line[1 : len(line)-2] // get the header value only

	val, err := strconv.Atoi(string(line))
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrProtocol, err)
	}
	return val, nil
}

func unexpectedEOF(err error) error {
	if errors.Is(err, io.EOF) {
		return io.ErrUnexpectedEOF
	}
	return err
}

// WriteSimpleString writes s as a RESP simple string, e.g. "+OK\r\n".
// It returns an error, and writes nothing, if s contains '\r' or '\n'.
func WriteSimpleString(w io.Writer, s string) error {
	if strings.ContainsAny(s, "\r\n") {
		return fmt.Errorf("resp: simple string contains CR or LF: %q", s)
	}
	_, err := fmt.Fprintf(w, "+%s\r\n", s)
	return err
}

// WriteError writes msg as a RESP error, e.g. "-ERR unknown command\r\n".
// msg should include its error prefix, such as "ERR".
// It returns an error, and writes nothing, if msg contains '\r' or '\n'.
func WriteError(w io.Writer, msg string) error {
	if strings.ContainsAny(msg, "\r\n") {
		return fmt.Errorf("resp: error message contains CR or LF: %q", msg)
	}

	_, err := fmt.Fprintf(w, "-%s\r\n", msg)
	return err
}

// WriteInteger writes n as a RESP integer, e.g. ":42\r\n".
func WriteInteger(w io.Writer, n int64) error {
	_, err := fmt.Fprintf(w, ":%d\r\n", n)
	return err
}

// WriteBulkString writes b as a RESP bulk string, e.g. "$5\r\nhello\r\n".
// b may contain any bytes, including '\r' and '\n'.
func WriteBulkString(w io.Writer, b []byte) error {
	_, err := fmt.Fprintf(w, "$%d\r\n", len(b))
	if err != nil {
		return err
	}

	_, err = w.Write(b)
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, "\r\n")
	return err
}

// WriteNull writes a RESP null bulk string, "$-1\r\n",
// which clients read as a missing value.
func WriteNull(w io.Writer) error {
	_, err := fmt.Fprint(w, "$-1\r\n")
	return err
}
