package grpcservice

import (
	"context"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/scm"
	"github.com/autograde/aguis/web"
)

func getCurrentUser(ctx context.Context, db database.Database) (*pb.User, error) {
	// process user id from context
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Aborted, "malformed request")
	}
	userValues := meta.Get("user")
	if len(userValues) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "no user metadata")
	}

	if len(userValues) != 1 {
		return nil, status.Errorf(codes.PermissionDenied, "invalid user payload")
	}
	currentUser := userValues[0]
	if currentUser == "" {
		return nil, status.Errorf(codes.Unauthenticated, "invalid user payload")
	}
	numID, err := strconv.ParseUint(currentUser, 10, 64)
	if err != nil {
		return nil, err
	}

	// check that user is a valid user in db. If not, return only the error
	usr, err := db.GetUser(numID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}
	return usr, nil
}

func getSCM(ctx context.Context, scms *web.Scms, db database.Database, provider string) (scm.SCM, error) {
	user, err := getCurrentUser(ctx, db)
	if err != nil {
		return nil, err
	}
	for _, identity := range user.RemoteIdentities {
		if identity.Provider == provider {
			scms.Mux.RLock()
			defer scms.Mux.Unlock()
			if _, ok := scms.Scms[identity.AccessToken]; !ok {
				return nil, status.Errorf(codes.PermissionDenied, "invalid token")
			}
			return scms.Scms[identity.AccessToken], nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "no SCM found")
}
