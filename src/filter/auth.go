package filter

import (
	"fmt"
	"github.com/studygolang/mux"
	"net/http"
	"github.com/ugorji/go/codec"
	"net"
	"log"
	"net/rpc"
	"strconv"
)

type Args struct {
	authToken string
}

type Reply struct {
	Ok bool
}

type AuthTokenVerifier int

type AuthFilter struct {
	*mux.EmptyFilter
	RPCServerHost string
	RPCServerPort int
}

func (this *AuthFilter) PreFilter(rw http.ResponseWriter, req *http.Request) bool {
	authToken := req.Header.Get("X-Auth-Token")

	ok := this.checkAuthToken(authToken)
	if ! ok {
		rw.WriteHeader(http.StatusForbidden)
		fmt.Fprint(rw, "403 Not Authorized\n")
		return false
	}

	return true
}

func (this *AuthFilter) checkAuthToken(authToken string) bool {
	port := strconv.Itoa(this.RPCServerPort)
	conn, err := net.Dial("tcp", this.RPCServerHost + ":" + port)
	if err != nil {
		log.Print("error dialing:", err)
		return false
	}

	rpcCodec := codec.GoRpc.ClientCodec(conn, &codec.MsgpackHandle{})
	client := rpc.NewClientWithCodec(rpcCodec)
	defer client.Close()

	reply := new(Reply)
	err = client.Call("AuthTokenVerifier.Add", authToken, reply)
	if err != nil {
		log.Printf("rpc error: Add: expected no error but got string %q", err.Error())
		return false
	}

	return reply.Ok
}
