package gocsi

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"
)

// PipeConn is an in-memory network connection that can be provided
// to a Serve function as a net.Listener and to gRPC/net.http clients
// as their dialer.
type PipeConn interface {
	net.Listener

	// DialGrpc is used by a grpc client.
	DialGrpc(raddr string, timeout time.Duration) (net.Conn, error)

	// DialHTTP is used by net.http clients.
	DialHTTP(ctx context.Context, network, addr string) (net.Conn, error)
}

// NewPipeConn returns a new pipe connection. The provided name
// is returned by PipeConn.Addr().String().
func NewPipeConn(name string) PipeConn {
	return &pipeConn{
		addr: &pipeAddr{name: name},
		chcn: make(chan net.Conn, 1),
	}
}

type pipeConn struct {
	sync.Once
	addr *pipeAddr
	chcn chan net.Conn
}

func (p *pipeConn) Dial(ctx context.Context) (net.Conn, error) {
	r, w := net.Pipe()
	go func() {
		p.chcn <- r
	}()
	return w, nil
}

func (p *pipeConn) DialGrpc(
	raddr string,
	timeout time.Duration) (net.Conn, error) {

	return p.Dial(context.Background())
}

func (p *pipeConn) DialHTTP(
	ctx context.Context, network, addr string) (net.Conn, error) {

	return p.Dial(ctx)
}

func (p *pipeConn) Accept() (net.Conn, error) {
	for c := range p.chcn {
		return c, nil
	}
	return nil, http.ErrServerClosed
}

func (p *pipeConn) Close() error {
	p.Do(func() { close(p.chcn) })
	return http.ErrServerClosed
}

func (p *pipeConn) Addr() net.Addr {
	return p.addr
}

type pipeAddr struct {
	name string
}

func (a *pipeAddr) Network() string {
	return "pipe"
}

func (a *pipeAddr) String() string {
	return a.name
}
