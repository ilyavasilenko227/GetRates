package server

import (
	"net"
	"rates/internal/controller"
	pb "rates/internal/infrastructure/pb"
	"rates/pkg/logger"

	"google.golang.org/grpc"
)

var (
	log = logger.Logger().Named("server").Sugar()
)

type Server struct {
	controller *controller.Controller
}

func NewServer(controller *controller.Controller) *Server {
	return &Server{controller: controller}
}

func (s *Server) RunApp(host, port string) *grpc.Server {
	server := grpc.NewServer()
	pb.RegisterGetRateserServer(server, s.controller)

	addr := net.JoinHostPort(host, port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("error to listen: %s", err)
	}

	log.Infof("starting server on %s:%s", host, port)

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Errorf("error to listen server %s", err)
		}
	}()

	return server
}

