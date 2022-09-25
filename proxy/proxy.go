package proxy

import (
	"bytes"
	"inv-client-extension/ivt/client"
	"io"
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

type Requests interface {
	io.Writer
	client.Wrapper
	WriteHeaders(headers http.Header)
	Complete()
	GetUuid() uuid.UUID
}

type SingleRequest struct {
	id   uuid.UUID
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
	return s.id
}

func (s *SingleRequest) ToRequestWrapper() *client.RequestWrapper {

	headers := make(map[string][]string)
	for k, v := range s.r.Header {
		headers[k] = v
	}

	buf := bytes.Buffer{}
	buf.ReadFrom(s.r.Body)

	return &client.RequestWrapper{
		ID:      s.id,
		Url:     s.r.URL.String(),
		Method:  s.r.Method,
		Headers: headers,
		Body:    buf.Bytes(),
	}
}

type Proxy struct {
	requestChan chan Requests
	logger      *zap.Logger
}

func NewProxy(logger *zap.Logger) *Proxy {
	return &Proxy{requestChan: make(chan Requests, 100), logger: logger}
}

func (p *Proxy) handleConnect(w http.ResponseWriter, r *http.Request) {
	p.logger.Sugar().Debug("Handeling connect request")

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
	p.logger.Sugar().Debug("Handeling simple request")

	p.logger.Sugar().Debugf("Request: %v", r)

	req := SingleRequest{
		id:   uuid.New(),
		r:    r,
		w:    w,
		done: make(chan struct{}, 1),
	}

	p.requestChan <- &req

	<-req.done
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
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
	server := &http.Server{Addr: ":8080", Handler: p}
	return server.ListenAndServe()
}
