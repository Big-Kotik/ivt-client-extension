package proxy

import (
	"bytes"
	"fmt"
	"inv-client-extension/ivt/types"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ReqType uint

const (
	SingleHttp ReqType = iota
	Connect
)

type SingleRequest struct {
	ID   uuid.UUID
	r    *http.Request
	w    http.ResponseWriter
	done chan struct{}
}

func (s *SingleRequest) Write(p []byte) (n int, err error) {
	return s.w.Write(p)
}

func (s *SingleRequest) WriteHeaders(headers http.Header) {
	copyHeader(s.w.Header(), headers)
}

func (s *SingleRequest) Complete() {
	s.done <- struct{}{}
}

func (s *SingleRequest) GetUuid() uuid.UUID {
	return s.ID
}

func (s *SingleRequest) ToRequestWrapper() *types.RequestWrapper {
	headers := make(map[string][]string)
	for k, v := range s.r.Header {
		headers[k] = v
	}

	buf := bytes.Buffer{}
	buf.ReadFrom(s.r.Body)

	return &types.RequestWrapper{
		ID:      s.ID,
		Url:     s.r.URL.String(),
		Method:  s.r.Method,
		Headers: headers,
		Body:    buf.Bytes(),
	}
}

type Proxy struct {
	requestChan chan types.Requests
	logger      *zap.Logger
	port        int
	server      *http.Server
}

func NewProxy(port int, logger *zap.Logger) *Proxy {
	return &Proxy{requestChan: make(chan types.Requests, 100), logger: logger, port: port}
}

func (p *Proxy) GetChan() <-chan types.Requests {
	return p.requestChan
}

func (p *Proxy) handleConnect(w http.ResponseWriter, r *http.Request) {
	p.logger.Sugar().Debug("Handeling connect request")

	return

	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go transfer(destConn, clientConn)
	go transfer(clientConn, destConn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func (p *Proxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	id := uuid.New()

	logger := p.logger.With(zap.String("request_id", id.String()))
	logger.Info("Handeling simple request")
	logger.Sugar().Debugf("Request: %v", r)
	defer logger.Info("Request handled")

	req := SingleRequest{
		ID:   id,
		r:    r,
		w:    w,
		done: make(chan struct{}, 1),
	}

	p.requestChan <- &req

	<-req.done

	p.logger.Sugar().Debugf("Finish request: %d", id)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			log.Print(k, v)
			dst.Add(k, v)
		}
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		p.handleConnect(w, r)
	} else {
		p.handleRequest(w, r)
	}
}

func (p *Proxy) ListenAndServe() error {
	p.server = &http.Server{Addr: fmt.Sprintf(":%d", p.port), Handler: p}
	return p.server.ListenAndServe()
}

func (p *Proxy) Shutdown() error {
	return p.server.Shutdown(nil)
}
