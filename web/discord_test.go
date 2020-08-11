package web_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	_ "github.com/mattn/go-sqlite3"
)

// the test expects a grpc server already running on port :9090
// and a valid user ID inside the DISCORD_USER environmental variable.
// The user has to have an active record in the database and either be an admin
// or a teacher of the requested course.
func TestDiscordClient(t *testing.T) {
	conn, err := grpc.Dial(":9090", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewAutograderServiceClient(conn)

	currentUserID := os.Getenv("DISCORD_USER")
	if currentUserID == "" {
		t.Fatal("User ID is not set")
	}
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
