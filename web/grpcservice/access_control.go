package grpcservice

import (
	"context"
	"errors"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/scm"
	"github.com/markbates/goth"
)

func getCurrentUser(ctx context.Context, db database.Database) (*pb.User, error) {
	// process user id from context
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("malformed request")
	}
	userValues := meta.Get("user")
	if len(userValues) == 0 {
		return nil, errors.New("no user metadata in context")
	}
	if len(userValues) != 1 || userValues[0] == "" {
		return nil, errors.New("invalid user payload in context")
	}
	userID, err := strconv.ParseUint(userValues[0], 10, 64)
	if err != nil {
		return nil, err
	}
	// return the user corresponding to userID, or an error.
	return db.GetUser(userID)
}

func (s *AutograderService) getSCM(ctx context.Context, provider string) (scm.SCM, error) {
	if _, err := goth.GetProvider(provider); err != nil {
		return nil, status.Errorf(codes.NotFound, "invalid provider")
	}
	user, err := getCurrentUser(ctx, s.db)
	if err != nil {
		return nil, err
	}
	for _, remoteID := range user.RemoteIdentities {
		if remoteID.Provider == provider {
			scm, ok := s.scms.GetSCM(remoteID.GetAccessToken())
			if !ok {
				return nil, status.Errorf(codes.PermissionDenied, "invalid token")
			}
			return scm, nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "no SCM found")
}
