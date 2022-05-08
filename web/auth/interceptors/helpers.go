package interceptors

import (
	"context"
	"fmt"
	"strings"

	pb "github.com/autograde/quickfeed/ag/types"
	"github.com/autograde/quickfeed/database"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var ErrAccessDenied = status.Errorf(codes.Unauthenticated, "access denied")

// GetFromMetadata extracts a value from a filed of incoming metadata
// by the given key. Used to extract JWT tokens.
func GetFromMetadata(ctx context.Context, field, key string) (string, error) {
	if field == "" {
		return "", fmt.Errorf("missing metadata field name (%s)", field)
	}
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("failed to read metadata")
	}
	content := meta.Get(field)
	// if there is no key, a field is expected to have only one element
	if key == "" {
		if len(content) != 1 {
			return "", fmt.Errorf("incorrect metadata content length: %d", len(content))
		}
		return content[0], nil
	}
	for _, c := range meta.Get(field) {
		_, content, ok := strings.Cut(c, key+"=")
		if !ok {
			return "", fmt.Errorf("missing %s cookie", key)
		}
		return strings.TrimSpace(content), nil
	}
	return "", fmt.Errorf("missing metadata field %s", field)
}

// setToMetadata sets a new metadata field to the incoming context.
func setToMetadata(ctx context.Context, field, value string) (context.Context, error) {
	if field == "" || value == "" {
		return nil, fmt.Errorf("missing metadata field name (%s) or value (%s)", field, value)
	}
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to read metadata")
	}
	meta.Set(field, value)
	return metadata.NewIncomingContext(ctx, meta), nil
}

// setCookie sets a "Set-Cookie" header with JWT token to the outgoing context.
func setCookie(ctx context.Context, cookie string) error {
	if cookie == "" {
		return fmt.Errorf("empty cookie")
	}
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("failed to read metadata")
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "Set-Cookie", cookie)
	if err := grpc.SetHeader(ctx, meta); err != nil {
		return fmt.Errorf("failed to set grpc header: %w", err)
	}
	return nil
}

func hasCourseAccess(db database.Database, courseID, userID uint64, status pb.Enrollment_UserStatus) bool {
	enrol, err := db.GetEnrollmentByCourseAndUser(courseID, userID)
	if err != nil {
		return false
	}
	return enrol.Status == status
}
