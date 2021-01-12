package proto

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

func serve(srv ScoreServiceServer) {
	lis, err := net.Listen("tcp", ":8070")
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	RegisterScoreServiceServer(grpcServer, srv)
	fmt.Printf("Server is running at :8070.\n")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
