package routes

import (
	"context"
	. "generated"
	"log"

	"db"
	pb "services"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// RegisterServers registers gRPC services to the server.
func RegisterServers(server *grpc.Server, client *db.PrismaClient, sugar *zap.SugaredLogger) {

	RegisterAuthServer(server, &pb.AuthServiceServer{
		PrismaClient: client,
		Logger:       sugar,
	})

}

// RegisterHandlers registers gRPC-Gateway handlers.
func RegisterHandlers(gwmux *runtime.ServeMux, conn *grpc.ClientConn) {
	var err error

	err = RegisterAuthHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

}
