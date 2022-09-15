package route

import (
	"github.com/Blackmocca/go-clean-template/proto/proto_models"
	"google.golang.org/grpc"
)

type GrpcRoute struct {
	server *grpc.Server
}

func NewGRPCRoute(server *grpc.Server) *GrpcRoute {
	return &GrpcRoute{server}
}

func (g GrpcRoute) RegisterOrganize(handler proto_models.OrganizeServer) {
	// proto_models.RegisterOrganizeServer(g.server, handler)
}
