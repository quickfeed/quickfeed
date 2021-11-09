package main

import (
	"context"
	"flag"
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
	srcSuffix = "-original.xlsx"
	dstSuffix = "-approve-list.xlsx"
	pass      = "Godkjent"
	fail      = "Ikke godkjent"
)

var ignoredStudents = map[string]bool{
	"Hein Meling Student":      true,
	"Eivind Stavnes (student)": true,
	"John Ingve Olsen Test":    true,
	"Hein Meling Stud5":        true,
	"Hans Erik Frøyland":       true,
}

type QuickFeed struct {
	cc *grpc.ClientConn
	pb.AutograderServiceClient
	md metadata.MD
}

func (s *QuickFeed) Close() {
	s.cc.Close()
}

func NewQuickFeed(authToken string) (*QuickFeed, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cc, err := grpc.DialContext(ctx, "uis.itest.run:9090",
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(1024*1024*20),
			grpc.MaxCallSendMsgSize(1024*1024*20),
		),
	)
	if err != nil {
		return nil, err
	}
	return &QuickFeed{
		cc:                      cc,
		AutograderServiceClient: pb.NewAutograderServiceClient(cc),
		md:                      metadata.New(map[string]string{"cookie": authToken}),
	}, nil
}

func main() {
	var (
		passLimit  = flag.Int("limit", 6, "number of assignments required to pass")
		ignorePass = flag.Bool("ignore", false, "ignore assignments that pass; only insert failed")
		courseCode = flag.String("course", "DAT320", "course code to query (case sensitive)")
		userName   = flag.String("user", "meling", "user name to request courses for")
		year       = flag.Int("year", time.Now().Year(), "year of course to fetch from QuickFeed")
	)
	flag.Parse()

	studentMap, sheetName := loadApproveSheet(*courseCode)

	authToken := os.Getenv("QUICKFEED_AUTH_TOKEN")
	if authToken == "" {
		log.Fatalln("QUICKFEED_AUTH_TOKEN is not set")
	}

	client, err := NewQuickFeed(authToken)
	if err != nil {
		log.Fatalln("Failed to connect to quickfeed server:", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, client.md)

	// ERROR   web/autograder_service.go:541   GetSubmissionsByCourse failed: user quickfeed-uis is not teacher or submission author
	request := &pb.CourseUserRequest{
		CourseCode: *courseCode,
		CourseYear: uint32(*year),
		UserLogin:  *userName,
	}
	userInfo, err := client.GetUserByCourse(ctx, request)
	if err != nil {
		log.Fatal(err)
	}

	courses, err := client.GetCoursesByUser(ctx, &pb.EnrollmentStatusRequest{UserID: userInfo.GetID()})
	if err != nil {
		log.Fatal(err)
	}
	var courseID uint64
	for _, c := range courses.GetCourses() {
		if c.GetCode() == *courseCode {
			courseID = c.GetID()
			fmt.Printf("Found course ID for %s: %d\n", c.GetCode(), courseID)
		}
	}
	if courseID == 0 {
		log.Fatalf("Could not find course: %s", *courseCode)
	}

	gotSubmissions, err := client.GetSubmissionsByCourse(
		ctx,
		&pb.SubmissionsForCourseRequest{
			CourseID: courseID,
			Type:     pb.SubmissionsForCourseRequest_ALL,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	approvedMap := make(map[string]string)
	agStudents := make(map[string]int)
	numPass, numIgnored := 0, 0
	for _, el := range gotSubmissions.GetLinks() {
		fmt.Printf("st: %s\n", el.Enrollment.User.Name)
		if el.Enrollment.User.IsAdmin || el.Enrollment.IsTeacher() {
			// log.Printf("%s: admin: %t, teacher: %t\n", el.Enrollment.GetUser().GetName(), el.Enrollment.User.IsAdmin, el.Enrollment.IsTeacher())
			continue
		}
		if ignoredStudents[el.Enrollment.User.Name] {
			continue
		}
		approved := make([]bool, len(el.Submissions))
		for i, s := range el.Submissions {
			approved[i] = s.GetSubmission().IsApproved()
		}
		agStudents[el.Enrollment.User.Name] = 1

		rowNum, err := lookup(el.Enrollment.User.Name, studentMap)
		if err != nil {
			fmt.Printf("%v in FS database; but has QuickFeed account\n", err)
			continue
		}
		approvedValue := fail
		if isApproved(*passLimit, approved) {
			approvedValue = pass
			numPass++
			if *ignorePass {
				numIgnored++
				continue
			}
		}
		cell := fmt.Sprintf("B%d", rowNum)
		approvedMap[cell] = approvedValue
	}

	// find students signed up to course, but not found in QuickFeed
	for student, rowNum := range studentMap {
		_, err := lookup(student, agStudents)
		if err != nil {
			fmt.Printf("%v in QuickFeed database; is signed up at row %d\n", err, rowNum)
			cell := fmt.Sprintf("B%d", rowNum)
			approvedMap[cell] = fail
		}
	}
	fmt.Printf("Total: %d, passed: %d, fail: %d\n", len(approvedMap)+numIgnored, numPass, len(approvedMap)+numIgnored-numPass)
	saveApproveSheet(*courseCode, sheetName, approvedMap)
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

func fileName(courseCode, suffix string) string {
	return strings.ToLower(courseCode) + suffix
}

func loadApproveSheet(courseCode string) (approveMap map[string]int, sheetName string) {
	f, err := excelize.OpenFile(fileName(courseCode, srcSuffix))
	if err != nil {
		log.Fatal(err)
	}
	if f.SheetCount != 1 {
		log.Fatalf("Unexpected number of sheets: %d; only single-sheet files supported", f.SheetCount)
	}
	// we expect only a single sheet; assume that is the active sheet
	sheetName = f.GetSheetName(f.GetActiveSheetIndex())
	fmt.Println("Found sheet in file:", sheetName)
	approveMap = make(map[string]int)
	for i, row := range f.GetRows(sheetName) {
		if i > 0 && row[0] != "" {
			approveMap[row[0]] = i + 1
		}
	}
	return approveMap, sheetName
}

func saveApproveSheet(courseCode, sheetName string, approveMap map[string]string) {
	f, err := excelize.OpenFile(fileName(courseCode, srcSuffix))
	if err != nil {
		log.Fatal(err)
	}
	for cell, approved := range approveMap {
		f.SetCellValue(sheetName, cell, approved)
	}
	if err := f.SaveAs(fileName(courseCode, dstSuffix)); err != nil {
		log.Fatal(err)
	}
}
