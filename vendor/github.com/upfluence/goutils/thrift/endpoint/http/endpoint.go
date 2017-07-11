package http

import (
	"fmt"
	"net/http"

	"github.com/upfluence/base/base_service"
	"github.com/upfluence/thrift-http-go/http_thrift"
	stdThrift "github.com/upfluence/thrift/lib/go/thrift"
	"github.com/upfluence/goutils/httputil"
	"github.com/upfluence/goutils/thrift"
	"github.com/upfluence/goutils/thrift/handler"
)

type Endpoint struct {
	servers []*thrift.Server
	port    int
	mux     *http.ServeMux
}

func NewEndpoint(baseHandler *handler.Base, port int) (*Endpoint, error) {
	mux := httputil.NewMux()
	trans, err := http_thrift.NewTHTTPServerFromMux(mux, "/base")

	if err != nil {
		return nil, err
	}

	return &Endpoint{
		servers: []*thrift.Server{
			thrift.NewServer(
				base_service.NewBaseServiceProcessor(baseHandler),
				trans,
			),
		},
		port: port,
		mux:  mux,
	}, nil
}

func (e *Endpoint) Mount(processor stdThrift.TProcessor, path string) error {
	trans, err := http_thrift.NewTHTTPServerFromMux(e.mux, path)

	if err != nil {
		return err
	}

	e.servers = append(e.servers, thrift.NewServer(processor, trans))

	return nil
}

func (e *Endpoint) Serve() error {
	for _, s := range e.servers {
		go func(server *thrift.Server) { server.Start() }(s)
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", e.port), e.mux)
}
