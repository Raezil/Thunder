package routes

import (
	"log"

	. "generated"

	"github.com/ysugimoto/grpc-graphql-gateway/runtime"
	"google.golang.org/grpc"
)

// RegisterGraphQLHandlers registers GraphQL Gateway handlers.
func RegisterGraphQLHandlers(mux *runtime.ServeMux, conn *grpc.ClientConn) {
	var err error

	err = RegisterAuthGraphqlHandler(mux, conn)
	if err != nil {
		log.Fatalln("Failed to register GraphQL gateway:", err)
	}

}
