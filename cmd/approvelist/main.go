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
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/web/interceptor"
)

const (
	srcSuffix = "-original.xlsx"
	dstSuffix = "-approve-list.xlsx"
	pass      = "Godkjent"
	fail      = "Ikke godkjent"
)

func NewQuickFeed(serverURL, token string) qfconnect.QuickFeedServiceClient {
	return qfconnect.NewQuickFeedServiceClient(
		http.DefaultClient,
		serverURL,
		connect.WithInterceptors(
			interceptor.NewTokenAuthClientInterceptor(token),
		),
	)
}

func main() {
	var (
		serverURL  = flag.String("server", "https://uis.itest.run", "UiS' QuickFeed server URL")
		passLimit  = flag.Int("limit", 6, "number of assignments required to pass")
		ignorePass = flag.Bool("ignore", false, "ignore assignments that pass; only insert failed")
		showAll    = flag.Bool("all", false, "show all students")
		courseCode = flag.String("course", "DAT320", "course code to query (case sensitive)")
		year       = flag.Int("year", time.Now().Year(), "year of course to fetch from QuickFeed")
	)
	flag.Parse()

	studentRowMap, sheetName, err := loadApproveSheet(*courseCode)
	if err != nil {
		log.Fatal(err)
	}

	submissions, enrollments, err := getSubmissions(*serverURL, *courseCode, uint32(*year))
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
	for _, enroll := range enrollments {
		// ignore course admins and teachers
		if enroll.IsAdmin() || enroll.IsTeacher() {
			continue
		}
		student := enroll.Name()
		approved := make([]bool, len(submissions.For(enroll.ID)))
		for i, s := range submissions.For(enroll.ID) {
			approved[i] = s.IsApproved()
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

func getSubmissions(serverURL, courseCode string, year uint32) (*qf.CourseSubmissions, []*qf.Enrollment, error) {
	token, err := env.GetAccessToken()
	if err != nil {
		return nil, nil, err
	}

	client := NewQuickFeed(serverURL, token)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := client.GetUser(ctx, connect.NewRequest(&qf.Void{}))
	if err != nil {
		return nil, nil, err
	}
	var courseID uint64
	for _, enrollment := range user.Msg.GetEnrollments() {
		course := enrollment.GetCourse()
		if course.GetCode() == courseCode && course.GetYear() == year {
			courseID = course.GetID()
			break
		}
	}
	if courseID == 0 {
		return nil, nil, fmt.Errorf("course %s-%d not found", courseCode, year)
	}

	submissionCourseRequest := &qf.SubmissionRequest{
		CourseID: courseID,
		FetchMode: &qf.SubmissionRequest_Type{
			Type: qf.SubmissionRequest_ALL,
		},
	}
	submissions, err := client.GetSubmissionsByCourse(ctx, connect.NewRequest(submissionCourseRequest))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get submissions for course %s: %w", courseCode, err)
	}
	enrollments, err := client.GetEnrollments(ctx, connect.NewRequest(&qf.EnrollmentRequest{
		FetchMode: &qf.EnrollmentRequest_CourseID{
			CourseID: courseID,
		},
	}))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get enrollments for course %s: %w", courseCode, err)
	}
	return submissions.Msg, enrollments.Msg.Enrollments, err
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
	// The approve sheet is a single sheet Excel file with five columns:
	//
	// 		First name | Last name | Student number    | Candidate number | Approval
	// 		-----------+-----------+-------------------+------------------+----------
	// 		<first>    | <last>    | <student_no>      | <candidate_no>   | <approved>
	//      John       | Doe       | 123456            |                  |
	//
	// Approval and candidate number columns are empty by default.
	// The approval column should be filled with either "Godkjent" or "Ikke godkjent".
	// The candidate number column is irrelevant for approval and can be ignored.
	//
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
		if i > 0 && row[0] != "" && row[1] != "" {
			fullName := fmt.Sprintf("%s %s", row[0], row[1])
			approveMap[fullName] = i + 1
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
