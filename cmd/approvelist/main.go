package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/bufbuild/connect-go"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/web/auth"
)

const (
	srcSuffix = "-original.xlsx"
	dstSuffix = "-approve-list.xlsx"
	pass      = "Godkjent"
	fail      = "Ikke godkjent"
)

func NewQuickFeed(serverURL string) qfconnect.QuickFeedServiceClient {
	return qfconnect.NewQuickFeedServiceClient(
		http.DefaultClient,
		serverURL,
		connect.WithGRPC(),
		// connect.WithGRPCWeb(),
	)
}

func main() {
	var (
		serverURL  = flag.String("server", "https://uis.itest.run", "UiS' QuickFeed server URL")
		passLimit  = flag.Int("limit", 6, "number of assignments required to pass")
		ignorePass = flag.Bool("ignore", false, "ignore assignments that pass; only insert failed")
		showAll    = flag.Bool("all", false, "show all students")
		courseCode = flag.String("course", "DAT320", "course code to query (case sensitive)")
		userName   = flag.String("user", "meling", "user name to request courses for")
		year       = flag.Int("year", time.Now().Year(), "year of course to fetch from QuickFeed")
	)
	flag.Parse()

	studentRowMap, sheetName, err := loadApproveSheet(*courseCode)
	if err != nil {
		log.Fatal(err)
	}

	submissions, err := getSubmissions(*serverURL, *userName, *courseCode, *year)
	if err != nil {
		log.Fatal(err)
	}

	tw := tabwriter.NewWriter(os.Stdout, 2, 8, 2, ' ', 0)
	fmt.Fprint(tw, "Student\tFS\tRow#\tQuickFeed\t#Approved\tApproved\n")

	// map of students found in QuickFeed; the int value is ignored (set to 1),
	// but used to allow use of the lookupRow() function.
	quickfeedStudents := make(map[string]int)
	approvedMap := make(map[string]string)
	numPass, numIgnored := 0, 0
	for _, el := range submissions.GetLinks() {
		enroll := el.GetEnrollment()
		// ignore course admins and teachers
		if enroll.IsAdmin() || enroll.IsTeacher() {
			continue
		}
		student := enroll.Name()
		approved := make([]bool, len(el.Submissions))
		for i, s := range el.Submissions {
			approved[i] = s.GetSubmission().IsApproved()
		}
		quickfeedStudents[student] = 1

		approvedValue := fail
		if isApproved(*passLimit, approved) {
			approvedValue = pass
			numPass++
			if *ignorePass {
				numIgnored++
				continue
			}
		}
		rowNum, err := lookupRow(student, studentRowMap)
		if err != nil {
			// not found in FS database, but has approved assignments
			fmt.Fprintf(tw, "%s\t-\t\t✓\t%d\t%s\n", student, numApproved(approved), approvedValue)
			continue
		}
		cell := fmt.Sprintf("B%d", rowNum)
		approvedMap[cell] = approvedValue
		if *showAll {
			fmt.Fprintf(tw, "%s\t✓\t%d\t✓\t%d\t%s\n", student, rowNum, numApproved(approved), approvedValue)
		}
	}

	// find students signed up to course, but not found in QuickFeed
	for student, rowNum := range studentRowMap {
		_, err := lookupRow(student, quickfeedStudents)
		if err != nil {
			// not found in QuickFeed, but is signed up in FS
			fmt.Fprintf(tw, "%s\t✓\t%d\t-\t\t\n", student, rowNum)
			cell := fmt.Sprintf("B%d", rowNum)
			approvedMap[cell] = fail
		}
	}
	tw.Flush()
	fmt.Println("----------")
	fmt.Printf("Total: %d, passed: %d, fail: %d\n", len(approvedMap)+numIgnored, numPass, len(approvedMap)+numIgnored-numPass)
	if err = saveApproveSheet(*courseCode, sheetName, approvedMap); err != nil {
		log.Fatal(err)
	}
}

func getSubmissions(serverURL, userName, courseCode string, year int) (*qf.CourseSubmissions, error) {
	// TODO(meling) how to get the cookie
	cookie := "secret"

	client := NewQuickFeed(serverURL)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// TODO(meling) Do we need all these RPCs just to get the submissions? See issue #724.
	courseUserRequest := &qf.CourseUserRequest{
		CourseCode: courseCode,
		CourseYear: uint32(year),
		UserLogin:  userName,
	}
	userResp, err := client.GetUserByCourse(ctx, requestWithCookie(courseUserRequest, cookie))
	if err != nil {
		return nil, fmt.Errorf("failed to get user %s in course %s: %w", userName, courseCode, err)
	}

	enrollStatusRequest := &qf.EnrollmentStatusRequest{UserID: userResp.Msg.GetID()}
	coursesResp, err := client.GetCoursesByUser(ctx, requestWithCookie(enrollStatusRequest, cookie))
	if err != nil {
		return nil, fmt.Errorf("failed to get courses for user %s: %w", userName, err)
	}
	var courseID uint64
	for _, c := range coursesResp.Msg.GetCourses() {
		if c.GetCode() == courseCode {
			courseID = c.GetID()
		}
	}
	if courseID == 0 {
		return nil, fmt.Errorf("course %s not found", courseCode)
	}

	submissionCourseRequest := &qf.SubmissionsForCourseRequest{
		CourseID: courseID,
		Type:     qf.SubmissionsForCourseRequest_ALL,
	}
	submissions, err := client.GetSubmissionsByCourse(ctx, requestWithCookie(submissionCourseRequest, cookie))
	if err != nil {
		return nil, fmt.Errorf("failed to get submissions for course %s: %w", courseCode, err)
	}
	return submissions.Msg, err
}

func requestWithCookie[T any](message *T, cookie string) *connect.Request[T] {
	request := connect.NewRequest(message)
	request.Header().Set(auth.Cookie, cookie)
	return request
}

func lookupRow(name string, studentMap map[string]int) (int, error) {
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
		return 0, fmt.Errorf("not found: %s", name)
	case len(possibleNames[name]) > 1:
		return 0, fmt.Errorf("multiple possibilities found for: %s --> %v", name, possibleNames[name])
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

func numApproved(approved []bool) (numApproved int) {
	for _, a := range approved {
		if a {
			numApproved++
		}
	}
	return
}

func fileName(courseCode, suffix string) string {
	return strings.ToLower(courseCode) + suffix
}

func loadApproveSheet(courseCode string) (approveMap map[string]int, sheetName string, err error) {
	f, err := excelize.OpenFile(fileName(courseCode, srcSuffix))
	if err != nil {
		return nil, "", err
	}
	if f.SheetCount != 1 {
		return nil, "", fmt.Errorf("expected a single sheet in %s, got %d", fileName(courseCode, srcSuffix), f.SheetCount)
	}
	// we expect only a single sheet; assume that is the active sheet
	sheetName = f.GetSheetName(f.GetActiveSheetIndex())
	approveMap = make(map[string]int)
	for i, row := range f.GetRows(sheetName) {
		if i > 0 && row[0] != "" {
			approveMap[row[0]] = i + 1
		}
	}
	return approveMap, sheetName, nil
}

func saveApproveSheet(courseCode, sheetName string, approveMap map[string]string) error {
	f, err := excelize.OpenFile(fileName(courseCode, srcSuffix))
	if err != nil {
		return err
	}
	for cell, approved := range approveMap {
		f.SetCellValue(sheetName, cell, approved)
	}
	return f.SaveAs(fileName(courseCode, dstSuffix))
}
