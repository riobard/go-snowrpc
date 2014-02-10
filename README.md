# SnowRPC: lightweight RPC framework

Package `snowrpc` implments a SnowRPC ClientCodec and ServerCodec for the
`net/rpc` package.

Documentation is availabe at http://godoc.org/github.com/riobard/go-snowrpc.

SnowRPC is a lightweight RPC framework developed at http://www.zhihu.com.
SnowRPC uses JSON (http://www.json.org) for object serialization and STP
(http://www.simpletp.org) for message framing. STP has a very simple grammar:

    MESSAGE = FRAME* CRLF
    FRAME = LENGTH CRLF DATA CRLF
    LENGTH = [0-9]+
    DATA = .*
    CRLF = "\r\n"

A minimal setup consists of a server and a client. The client sends a request to
the server, and the server sends a reply back. Requests and replies share the
same general structure: an STP message with a header frame and a body frame,
both of which are serialized JSON objects.

    [header][body]

A sample request:

    {"interface_name": "Calc.Add"}
    {"a": 1, "b": 2}

The `interface_name` parameter in the request header is required and it specifies which
remote method to invoke.

A sample response:

    {"return_code": 200, "message": "OK"}
    {"c": 3}

The `return_code` parameter in the response header is required. If the response
is successful, the `return_code` will be 200 and the `message` parameter will be
the string "OK"; otherwise, the `return_code` will be 500 and the `message`
parameter will be a string describing what is wrong. 
