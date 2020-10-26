package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	pb "github.com/autograde/quickfeed/ag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	srcFile      = "dat320-original.xlsx"
	approvedFile = "dat320-approve-list.xlsx"
	sheetName    = "DAT320 Operativsystemer og syst"
	pass         = "Godkjent"
	fail         = "Ikke godkjent"
)

func main() {
	studentMap := loadApproveSheet(srcFile, sheetName)
	currentUserID := os.Getenv("QUICKFEED_USER")
	if currentUserID == "" {
		log.Fatal("Requires a 'QUICKFEED_USER' environmental variable with a valid ID of a registered user")
	}
	convertedID := strings.TrimSpace(currentUserID)
	requestMetadata := metadata.New(map[string]string{"user": convertedID})
	reqCtx := metadata.NewOutgoingContext(context.Background(), requestMetadata)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, ":9090",
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(1024*1024*20),
			grpc.MaxCallSendMsgSize(1024*1024*20),
		),
	)
	if err != nil {
		log.Fatalf("Connection failed, make sure server is running on :9090: %v", err)
	}
	defer conn.Close()

	client := pb.NewAutograderServiceClient(conn)

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

	gotSubmissions, err := client.GetSubmissionsByCourse(
		reqCtx,
		&pb.SubmissionsForCourseRequest{CourseID: courseID, Type: pb.SubmissionsForCourseRequest_ALL},
	)
	if err != nil {
		log.Fatal(err)
	}
	approvedMap := make(map[string]string)
	agStudents := make(map[string]int)
	numPass := 0
	for _, el := range gotSubmissions.GetLinks() {
		if el.Enrollment.User.IsAdmin || el.Enrollment.IsTeacher() {
			//			log.Printf("%s: admin: %t, teacher: %t\n", el.Enrollment.GetUser().GetName(), el.Enrollment.User.IsAdmin, el.Enrollment.IsTeacher())
			continue
		}
		approved := make([]bool, len(el.Submissions))
		for i, s := range el.Submissions {
			approved[i] = s.GetSubmission().IsApproved()
		}
		agStudents[el.Enrollment.User.Name] = 1
		rowNum, err := lookup(el.Enrollment.User.Name, studentMap)
		if err != nil {
			log.Print(err)
			continue
		}
		approvedValue := fail
		if isApproved(6, approved) {
			approvedValue = pass
			numPass++
		}
		cell := fmt.Sprintf("B%d", rowNum)
		approvedMap[cell] = approvedValue
	}
	for student, rowNum := range studentMap {
		_, err := lookup(student, agStudents)
		if err != nil {
			fmt.Printf("%v in QuickFeed database; is signed up at row %d\n", err, rowNum)
			continue
		}
		approvedValue := fail
		cell := fmt.Sprintf("B%d", rowNum)
		approvedMap[cell] = approvedValue
	}
	fmt.Printf("Total: %d, passed: %d, fail: %d\n", len(approvedMap), numPass, len(approvedMap)-numPass)
	saveApproveSheet(srcFile, approvedFile, sheetName, approvedMap)
}

func lookup(name string, studentMap map[string]int) (int, error) {
	if rowNum, ok := studentMap[name]; ok {
		return rowNum, nil
	} else {
		return partialMatch(name, studentMap)
	}
}

func partialMatch(name string, studentMap map[string]int) (int, error) {
	nameParts := strings.Split(strings.ToLower(name), " ")
	possibleNames := make(map[string][]string)
	for expectedName := range studentMap {
		expectedNameParts := strings.Split(strings.ToLower(expectedName), " ")
		matchCount := 0
		for _, n := range nameParts {
			for _, m := range expectedNameParts {
				if n == m {
					matchCount++
				}
			}
		}
		if matchCount > 1 {
			// if at least two parts of the names match
			possibleNames[name] = append(possibleNames[name], expectedName)
			// fmt.Printf("Probable match found: %s = %s\n", name, expectedName)
		}
	}
	switch {
	case len(possibleNames[name]) == 0:
		return 0, fmt.Errorf("Not found: %s", name)
	case len(possibleNames[name]) > 1:
		return 0, fmt.Errorf("Multiple possibilities found for: %s --> %v", name, possibleNames[name])
	}
	return studentMap[possibleNames[name][0]], nil
}

func isApproved(requirements int, approved []bool) bool {
	for _, a := range approved {
		if a {
			requirements--
		}
	}
	return requirements <= 0
}

func loadApproveSheet(file, sheetName string) map[string]int {
	f, err := excelize.OpenFile(file)
	if err != nil {
		log.Fatal(err)
	}
	approveMap := make(map[string]int)
	for i, row := range f.GetRows(sheetName) {
		if row[0] != "" {
			approveMap[row[0]] = i + 1
		}
	}
	return approveMap
}

func saveApproveSheet(srcFile, dstFile, sheetName string, approveMap map[string]string) {
	f, err := excelize.OpenFile(srcFile)
	if err != nil {
		log.Fatal(err)
	}
	for cell, approved := range approveMap {
		f.SetCellValue(sheetName, cell, approved)
	}
	if err := f.SaveAs(dstFile); err != nil {
		log.Fatal(err)
	}
}
