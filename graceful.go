// Package graceful provides a net/http compatible Server that can be
// crafeully shut down. When closed, open connections are allowed to
// finish.
package graceful

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"
)

// Server is net/http compatible graceful server.
type Server struct {
	s  *http.Server
	wg sync.WaitGroup
	l  net.Listener
}

// NewServer turns the given net/http server into a graceful server.
func NewServer(srv *http.Server) *Server {
	return &Server{
		s: srv,
	}
}

// ListenAndServe works like net/http.Server.ListenAndServe except
// that it gracefully shuts down when Close() is called. When that
// occurs, no new connections will be allowed and existing connections
// will be allowed to finish. This will not return until all existing
// connections have closed.
func (s *Server) ListenAndServe() error {
	addr := s.s.Addr
	if addr == "" {
		addr = ":http"
	}
	var err error
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return s.Serve(l)
}

// ListenAndServe works like net/http.Server.ListenAndServeTLS except
// that it gracefully shuts down when Close() is called. When that
// occurs, no new connections will be allowed and existing connections
// will be allowed to finish. This will not return until all existing
// connections have closed.
func (s *Server) ListenAndServeTLS(certFile, keyFile string) error {
	addr := s.s.Addr
	if addr == "" {
		addr = ":https"
	}
	config := &tls.Config{}
	if s.s.TLSConfig != nil {
		*config = *s.s.TLSConfig
	}
	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	config.Certificates = []tls.Certificate{cert}
	l, err := tls.Listen("tcp", addr, config)
	if err != nil {
		return err
	}
	return s.Serve(l)
}

// Server works like net/http.Server.Serve except that it gracefully
// shuts down when Close() is called. When that occurs, no new
// connections will be allowed and existing connections will be
// allowed to finish. This will not return until all existing
// connections have closed.
func (s *Server) Serve(l net.Listener) error {
	s.l = l
	err := s.s.Serve(&gracefulListener{s.l, s})
	s.wg.Wait()
	return err
}

// Close gracefully shuts down the listener. This should be called
// when the server should stop listening for new connection.
func (s *Server) Close() error {
	err := s.l.Close()
	return err
}

// gracefulListener implements the net.Listener interface. When accept
// for the underlying listener returns a connection, it adds 1 to the
// servers wait group. The connection will be a gracefulConn which
// will call Done() when it finished.
type gracefulListener struct {
	net.Listener
	s *Server
}

func (g *gracefulListener) Accept() (net.Conn, error) {
	c, err := g.Listener.Accept()
	if err != nil {
		return nil, err
	}
	g.s.wg.Add(1)
	return &gracefulConn{c, g.s}, nil
}

// gracefulConn implements the net.Conn interface. When it closes, it
// calls Done() on the servers waitgroup.
type gracefulConn struct {
	net.Conn
	s *Server
}

func (g *gracefulConn) Close() error {
	err := g.Conn.Close()
	g.s.wg.Done()
	return err
}
