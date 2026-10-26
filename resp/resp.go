package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

var (
	ErrProtocol = errors.New("resp: protocol error")
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

// func WriteSimpleString(w io.Writer, s string) error // +OK\r\n
// func WriteError(w io.Writer, msg string) error      // -ERR message\r\n
// func WriteInteger(w io.Writer, n int64) error       // :42\r\n
// func WriteBulkString(w io.Writer, b []byte) error   // $5\r\nhello\r\n
// func WriteNull(w io.Writer) error                   // $-1\r\n
