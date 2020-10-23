package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	pb "github.com/autograde/quickfeed/ag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func main() {
	currentUserID := os.Getenv("QUICKFEED_USER")
	if currentUserID == "" {
		log.Fatal("Requires a 'QUICKFEED_USER' environmental variable with a valid ID of a registered user")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, ":9090", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Connection failed, make sure server is running on :9090: %v", err)
	}
	defer conn.Close()

	client := pb.NewAutograderServiceClient(conn)

	convertedID := strings.TrimSpace(currentUserID)
	requestMetadata := metadata.New(map[string]string{"user": convertedID})
	reqCtx := metadata.NewOutgoingContext(context.Background(), requestMetadata)

	request := &pb.CourseUserRequest{
		CourseCode: "DAT320",
		CourseYear: 2020,
		UserLogin:  "meling",
	}
	userInfo, err := client.GetUserByCourse(reqCtx, request)
	if err != nil {
		log.Fatal(err)
	}

	courses, err := client.GetCoursesByUser(reqCtx, &pb.EnrollmentStatusRequest{UserID: userInfo.GetID()})
	if err != nil {
		log.Fatal(err)
	}
	var courseID uint64
	for _, c := range courses.GetCourses() {
		courseID = c.GetID()
	}

	gotSubmissions, err := client.GetSubmissionsByCourse(reqCtx, &pb.SubmissionsForCourseRequest{CourseID: courseID, Type: pb.SubmissionsForCourseRequest_ALL})
	if err != nil {
		log.Fatal(err)
	}
	for _, el := range gotSubmissions.GetLinks() {
		if el.Enrollment.User.IsAdmin || el.Enrollment.GetHasTeacherScopes() {
			continue
		}
		approved := make([]bool, len(el.Submissions))
		for i, s := range el.Submissions {
			approved[i] = s.GetSubmission().IsApproved()
		}
		fmt.Printf("%s\t%t\n", el.Enrollment.User.Name, isApproved(8, approved))
	}
}

func isApproved(requirements int, approved []bool) bool {
	for _, a := range approved {
		if a {
			requirements--
		}
	}
	return requirements <= 0
}
