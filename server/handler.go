package server

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/tekka171/go-learning/kvstore"
	"github.com/tekka171/go-learning/resp"
)

const (
	// maxTTLSeconds is the largest number of seconds that can be converted to a
	// time.Duration without overflowing.
	maxTTLSeconds = math.MaxInt64 / int64(time.Second)
)

var (
	// lineBreaks replaces CR and LF with spaces, so client-supplied text can be
	// embedded safely in a single-line reply.
	lineBreaks = strings.NewReplacer("\r", " ", "\n", " ")
)

// sanitize returns b as a string with any '\r' or '\n' replaced by a space,
// making it safe to include in an error reply.
func sanitize(b []byte) string {
	return lineBreaks.Replace(string(b))
}

// handle executes one parsed command against the store and writes exactly one
// reply to w. Client mistakes (unknown command, wrong arguments, bad values)
// are reported to the client as error replies; the returned error is non-nil
// only if writing the reply fails.
func handle(store *kvstore.Store, args [][]byte, w io.Writer) error {
	switch strings.ToUpper(string(args[0])) {
	case "PING":
		return handlePing(args, w)
	case "SET":
		return handleSet(store, args, w)
	case "GET":
		return handleGet(store, args, w)
	case "DEL":
		return handleDel(store, args, w)
	case "EXPIRE":
		return handleExpire(store, args, w)
	case "TTL":
		return handleTTL(store, args, w)
	default:
		return resp.WriteError(w, fmt.Sprintf("ERR unknown command '%s'", sanitize(args[0])))
	}
}

func handlePing(args [][]byte, w io.Writer) error {
	return resp.WriteSimpleString(w, "PONG")
}

func handleSet(store *kvstore.Store, args [][]byte, w io.Writer) error {
	if len(args) != 3 && len(args) != 5 {
		return wrongArgs(w, string(args[0]))
	}

	var ttl time.Duration
	if len(args) == 5 {
		if !strings.EqualFold(string(args[3]), "EX") {
			return resp.WriteError(w, fmt.Sprintf("ERR unknown option '%s'", sanitize(args[3])))
		}

		s, ok := parseSeconds(string(args[4]))
		if !ok || s <= 0 {
			return resp.WriteError(w, "ERR value is not an integer or out of range")
		}

		ttl = time.Duration(s) * time.Second
	}

	store.Set(string(args[1]), args[2], ttl)
	return resp.WriteSimpleString(w, "OK")
}

func handleGet(store *kvstore.Store, args [][]byte, w io.Writer) error {
	if len(args) != 2 {
		return wrongArgs(w, string(args[0]))
	}

	val, ok := store.Get(string(args[1]))
	if !ok {
		return resp.WriteNull(w)
	}

	return resp.WriteBulkString(w, val)
}

func handleDel(store *kvstore.Store, args [][]byte, w io.Writer) error {
	if len(args) < 2 {
		return wrongArgs(w, string(args[0]))
	}

	var count int64
	for i := 1; i < len(args); i++ {
		if store.Delete(string(args[i])) {
			count++
		}
	}
	return resp.WriteInteger(w, count)
}

func handleExpire(store *kvstore.Store, args [][]byte, w io.Writer) error {
	if len(args) != 3 {
		return wrongArgs(w, string(args[0]))
	}

	s, ok := parseSeconds(string(args[2]))
	if !ok {
		return resp.WriteError(w, "ERR value is not an integer or out of range")
	}

	if s <= 0 {
		var n int64
		if store.Delete(string(args[1])) {
			n = 1
		}
		return resp.WriteInteger(w, n)
	}

	ttl := time.Duration(s) * time.Second

	var n int64
	if store.Expire(string(args[1]), ttl) {
		n = 1
	}
	return resp.WriteInteger(w, n)
}

func handleTTL(store *kvstore.Store, args [][]byte, w io.Writer) error {
	if len(args) != 2 {
		return wrongArgs(w, string(args[0]))
	}

	remaining, exists, hasTTL := store.TTL(string(args[1]))
	if !exists {
		return resp.WriteInteger(w, -2)
	}

	if !hasTTL {
		return resp.WriteInteger(w, -1)
	}

	return resp.WriteInteger(w, int64(remaining.Seconds()))
}

func wrongArgs(w io.Writer, name string) error {
	return resp.WriteError(w, fmt.Sprintf("ERR wrong number of arguments for '%s' command", strings.ToLower(name)))
}

func parseSeconds(n string) (int64, bool) {
	s, err := strconv.ParseInt(n, 10, 64)
	if err != nil || s > maxTTLSeconds {
		return 0, false
	}

	return s, true
}
