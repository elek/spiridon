package dns

import (
	"google.golang.org/grpc"
	"net"
)

type Endpoint func(s grpc.ServiceRegistrar)

type Server struct {
	listener net.Listener
	server   *grpc.Server
}

func NewServer(service *Service) (*Server, error) {
	lis, err := net.Listen("tcp", "127.0.0.1:8053")
	if err != nil {
		return nil, err
	}
	grpcServer := grpc.NewServer()
	RegisterDnsServiceServer(grpcServer, service)

	return &Server{
		listener: lis,
		server:   grpcServer,
	}, nil
}

func (s *Server) Run() error {
	return s.server.Serve(s.listener)
}
