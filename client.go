package snowrpc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"strconv"
	"sync"
)

type clientCodec struct {
	sync.Mutex
	conn io.ReadWriteCloser
	r    *bufio.Reader
	w    *bufio.Writer
}

func NewClientCodec(conn io.ReadWriteCloser) rpc.ClientCodec {
	return &clientCodec{
		conn: conn,
		r:    bufio.NewReader(conn),
		w:    bufio.NewWriter(conn),
	}
}

func (c *clientCodec) WriteRequest(req *rpc.Request, x interface{}) error {
	c.Lock()
	defer c.Unlock()
	var header struct {
		Method string `json:"interface_name"`
	}
	header.Method = req.ServiceMethod
	hb, err := json.Marshal(header)
	if err != nil {
		return err
	}
	bb, err := json.Marshal(x)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(c.w, "%d\r\n%s\r\n%d\r\n%s\r\n\r\n", len(hb), hb, len(bb), bb)
	if err != nil {
		return err
	}
	return c.w.Flush()
}

func (c *clientCodec) ReadResponseHeader(rsp *rpc.Response) error {
	line, err := c.r.ReadSlice('\n')
	if err != nil {
		return err
	}
	if len(line) <= 2 {
		return fmt.Errorf("incomplete response header")
	}
	llen, err := strconv.Atoi(string(line[:len(line)-2]))
	if err != nil {
		return err
	}
	if llen < 0 {
		return fmt.Errorf("negative request header length")
	}
	b := make([]byte, llen+2)
	_, err = io.ReadFull(c.r, b)
	if err != nil {
		return err
	}
	var header struct {
		Code int    `json:"return_code"`
		Msg  string `json:"message"`
	}
	err = json.Unmarshal(b, &header)
	if err != nil {
		return err
	}
	if header.Code != 200 {
		return fmt.Errorf("code %d: %s", header.Code, header.Msg)
	}
	return nil
}

func (c *clientCodec) ReadResponseBody(x interface{}) error {
	if x == nil {
		return nil
	}
	line, err := c.r.ReadSlice('\n')
	if err != nil {
		return err
	}
	if len(line) <= 2 {
		return fmt.Errorf("incomplete response body")
	}
	llen, err := strconv.Atoi(string(line[:len(line)-2])) // ignore trailing CRLF
	if err != nil {
		return err
	}
	if llen < 0 {
		return fmt.Errorf("negative response body length")
	}
	b := make([]byte, llen+4) // ignore two consecutive trailing CRLFs
	_, err = io.ReadFull(c.r, b)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, x)
}

func (c *clientCodec) Close() error {
	return c.conn.Close()
}

// Create a SnowRPC client to connect to a SnowRPC at the specified network address.
func Dial(network, address string) (*rpc.Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return NewClient(conn), nil
}

// Create a SnowRPC client from the given connection.
func NewClient(conn io.ReadWriteCloser) *rpc.Client {
	return rpc.NewClientWithCodec(NewClientCodec(conn))
}
