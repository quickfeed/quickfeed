package proto

import (
	context "context"
	"fmt"
	"log"
	"net"

	"github.com/autograde/quickfeed/kit/score"
	"google.golang.org/grpc"
)

func serve(srv score.ScoreServiceServer) {
	// TODO(meling) read port from flags
	lis, err := net.Listen("tcp", ":8070")
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	// NewScoreServiceServer()
	score.RegisterScoreServiceServer(grpcServer, srv)
	// TODO(meling) pass logger interface as input and replace this with logger output
	fmt.Printf("Server is running at :8070.\n")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

type scoreServer struct {
	userTests map[string]score.Tests
	// mustEmbedUnimplementedScoreServiceServer()
}

// Register the tests to be expected for this test run.
func (s *scoreServer) Register(context.Context, *score.Tests) (*score.Void, error) {
	return nil, nil
}

// Notify sends one score for each test.
func (s *scoreServer) Notify(score.ScoreService_NotifyServer) error {
	return nil
}
