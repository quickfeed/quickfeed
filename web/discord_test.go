package web_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	pb "github.com/autograde/quickfeed/ag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// the test expects a grpc server already running on port :9090
// and a valid user ID inside the DISCORD_USER environmental variable.
// The user has to have an active record in the database and either be an admin
// or a teacher of the requested course.
func TestDiscordClient(t *testing.T) {
	currentUserID := os.Getenv("DISCORD_USER")
	if currentUserID == "" {
		t.Skip("This test requires a 'DISCORD_USER' environmental variable with a valid ID of a registered user")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, ":9090", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		t.Skip("Connection failed, make sure the server is running on :9090")
	}
	defer conn.Close()

	client := pb.NewAutograderServiceClient(conn)

	convertedID := strings.TrimSpace(currentUserID)
	requestMetadata := metadata.New(map[string]string{"user": convertedID})
	requestContext := metadata.NewOutgoingContext(context.Background(), requestMetadata)

	request := &pb.CourseUserRequest{
		CourseCode: "DAT320",
		CourseYear: 2020,
		UserLogin:  "0xf8f8ff",
	}

	userInfo, err := client.GetUserByCourse(requestContext, request)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Found user: ", userInfo)
}
