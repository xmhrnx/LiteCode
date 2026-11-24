package lsp

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/textproto"
	"strconv"
	"sync"
)

// ErrClosed means the server went away.
var ErrClosed = errors.New("language server stopped")

// conn is a JSON-RPC 2.0 connection with Content-Length framing.
type conn struct {
	w       io.Writer
	wmu     sync.Mutex
	mu      sync.Mutex
	next    int
	pending map[int]chan response
	closed  bool
}

type response struct {
	result json.RawMessage
	err    error
}

type message struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method,omitempty"`
	Params  any              `json:"params,omitempty"`
	Result  json.RawMessage  `json:"result,omitempty"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func newConn(w io.Writer) *conn {
	return &conn{w: w, pending: map[int]chan response{}}
}

func (c *conn) send(m message) error {
	m.JSONRPC = "2.0"
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	c.wmu.Lock()
	defer c.wmu.Unlock()
	if _, err := fmt.Fprintf(c.w, "Content-Length: %d\r\n\r\n", len(data)); err != nil {
		return err
	}
	_, err = c.w.Write(data)
	return err
}

// call sends a request and returns a channel for its response.
func (c *conn) call(method string, params any) (int, <-chan response) {
	ch := make(chan response, 1)
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		ch <- response{err: ErrClosed}
		return 0, ch
	}
	c.next++
	id := c.next
	c.pending[id] = ch
	c.mu.Unlock()
	raw := json.RawMessage(strconv.Itoa(id))
	if err := c.send(message{ID: &raw, Method: method, Params: params}); err != nil {
		c.finish(id, response{err: err})
	}
	return id, ch
}

// cancel tells the server a request isn't wanted any more.
func (c *conn) cancel(id int) {
	c.finish(id, response{err: errors.New("cancelled")})
	_ = c.notify("$/cancelRequest", map[string]int{"id": id})
}

func (c *conn) notify(method string, params any) error {
	return c.send(message{Method: method, Params: params})
}

func (c *conn) finish(id int, r response) {
	c.mu.Lock()
	ch := c.pending[id]
	delete(c.pending, id)
	c.mu.Unlock()
	if ch != nil {
		ch <- r
	}
}

// read handles what the server sends until it closes: responses go to
// their callers, requests from the server get an empty answer.
func (c *conn) read(r io.Reader) {
	br := bufio.NewReader(r)
	tp := textproto.NewReader(br)
	for {
		h, err := tp.ReadMIMEHeader()
		if err != nil {
			break
		}
		n, err := strconv.Atoi(h.Get("Content-Length"))
		if err != nil || n < 0 {
			break
		}
		data := make([]byte, n)
		if _, err := io.ReadFull(br, data); err != nil {
			break
		}
		var m message
		if json.Unmarshal(data, &m) != nil {
			continue
		}
		switch {
		case m.ID != nil && m.Method != "":
			// A request from the server (configuration, progress, …):
			// answer it so the server doesn't wait.
			_ = c.send(message{ID: m.ID, Result: json.RawMessage("null")})
		case m.ID != nil:
			id, err := strconv.Atoi(string(*m.ID))
			if err != nil {
				continue
			}
			resp := response{result: m.Result}
			if m.Error != nil {
				resp.err = fmt.Errorf("%s (%d)", m.Error.Message, m.Error.Code)
			}
			c.finish(id, resp)
		}
	}
	c.mu.Lock()
	c.closed = true
	pending := c.pending
	c.pending = map[int]chan response{}
	c.mu.Unlock()
	for _, ch := range pending {
		ch <- response{err: ErrClosed}
	}
}
