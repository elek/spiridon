package satellite

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/pkg/errors"
	"net"
	"storj.io/common/identity"
	"storj.io/common/peertls/tlsopts"
	"storj.io/drpc/drpcmigrate"
	"storj.io/drpc/drpcmux"
	"storj.io/drpc/drpcserver"
)

type RPCServer struct {
	Port       int
	Mux        *drpcmux.Mux
	listenMux  *drpcmigrate.ListenMux
	tlsOptions *tlsopts.Options
}

func NewRPCServer(ident *identity.FullIdentity, port int) (*RPCServer, error) {

	tlsConfig := tlsopts.Config{
		UsePeerCAWhitelist: false,
		PeerIDVersions:     "0",
	}
	tlsOptions, err := tlsopts.NewOptions(ident, tlsConfig, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	address := fmt.Sprintf("0.0.0.0:%d", port)
	fmt.Println("Listening on", address)
	tcpListener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	listenMux := drpcmigrate.NewListenMux(tcpListener, len(drpcmigrate.DRPCHeader))

	m := drpcmux.New()

	return &RPCServer{
		Mux:        m,
		listenMux:  listenMux,
		tlsOptions: tlsOptions,
	}, nil
}

func (r *RPCServer) Run(ctx context.Context) error {
	go r.listenMux.Run(ctx)
	serv := drpcserver.NewWithOptions(r.Mux, drpcserver.Options{})
	tlsListener := tls.NewListener(r.listenMux.Route(drpcmigrate.DRPCHeader), r.tlsOptions.ServerTLSConfig())
	return serv.Serve(ctx, tlsListener)
}
