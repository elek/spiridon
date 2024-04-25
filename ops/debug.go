package ops

import (
	"context"
	"github.com/spacemonkeygo/monkit/v3"
	"github.com/spacemonkeygo/monkit/v3/present"
	"net"
	"net/http"
)

type Debug struct {
	listener net.Listener
}

func NewDebug() (*Debug, error) {
	tcpListener, err := net.Listen("tcp", "127.0.0.1:9000")
	if err != nil {
		return nil, err
	}
	return &Debug{
		listener: tcpListener,
	}, nil
}

func (d *Debug) Run(ctx context.Context) error {
	return http.Serve(d.listener, present.HTTP(monkit.Default))
}
