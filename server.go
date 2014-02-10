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

type serverCodec struct {
	sync.Mutex
	conn io.ReadWriteCloser
	r    *bufio.Reader
	w    *bufio.Writer
}

func NewServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	return &serverCodec{
		conn: conn,
		r:    bufio.NewReader(conn),
		w:    bufio.NewWriter(conn),
	}
}

func (c *serverCodec) ReadRequestHeader(req *rpc.Request) error {
	line, err := c.r.ReadSlice('\n')
	if err != nil {
		return err
	}
	if len(line) <= 2 {
		return fmt.Errorf("incomplete request header")
	}
	llen, err := strconv.Atoi(string(line[:len(line)-2])) // ignore trailing CRLF
	if err != nil {
		return err
	}
	if llen < 0 {
		return fmt.Errorf("negative request header length")
	}
	b := make([]byte, llen+2) // 2 bytes for trailing CRLF
	_, err = io.ReadFull(c.r, b)
	if err != nil {
		return err
	}

	var header struct {
		Method string `json:"interface_name"`
		//Seq    uint64 `json:"seq"`
	}
	err = json.Unmarshal(b, &header)
	if err != nil {
		return err
	}
	req.ServiceMethod = header.Method
	// req.Seq = header.Seq
	return nil
}

func (c *serverCodec) ReadRequestBody(x interface{}) error {
	if x == nil {
		return nil
	}
	line, err := c.r.ReadSlice('\n')
	if err != nil {
		return err
	}
	if len(line) <= 2 {
		return fmt.Errorf("incomplete request body")
	}
	llen, err := strconv.Atoi(string(line[:len(line)-2])) // ignore trailing CRLF
	if err != nil {
		return err
	}
	if llen < 0 {
		return fmt.Errorf("negative request body length")
	}
	b := make([]byte, llen+4) // 4 bytes for two consecutive trailing CRLFs
	_, err = io.ReadFull(c.r, b)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, x)
}

func (c *serverCodec) WriteResponse(rsp *rpc.Response, x interface{}) error {
	c.Lock()
	defer c.Unlock()
	var header struct {
		Code int    `json:"return_code"`
		Msg  string `json:"message",omitempty`
	}

	if header.Msg == "" {
		header.Code = 200
		header.Msg = "OK"
	} else {
		header.Code = 500
		header.Msg = rsp.Error
	}

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

func (c *serverCodec) Close() error {
	return c.conn.Close()
}

func ServeConn(conn io.ReadWriteCloser) {
	rpc.ServeCodec(NewServerCodec(conn))
}

// SnowRPC server.
type Server struct {
	*rpc.Server
}

// Create a SnowRPC server.
func NewServer() *Server {
	return &Server{rpc.NewServer()}
}

// Run the server at the specified protocol/address.
func (srv *Server) ListenAndServe(network, address string) error {
	l, err := net.Listen(network, address)
	if err != nil {
		return err
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go srv.ServeCodec(NewServerCodec(conn))
	}
}
