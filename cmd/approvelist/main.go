package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"text/tabwriter"
	"time"

	"connectrpc.com/connect"
	"github.com/360EntSecGroup-Skylar/excelize"
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
		showAll    = flag.Bool("all", false, "show all students")
		courseCode = flag.String("course", "DAT320", "course code to query (case sensitive)")
		year       = flag.Int("year", time.Now().Year(), "year of course to fetch from QuickFeed")
	)
	flag.Parse()

	as, err := loadApproveSheet(*courseCode)
	if err != nil {
		log.Fatal(err)
	}
	courseSubmissions, enrollments, err := getSubmissions(*serverURL, *courseCode, uint32(*year))
	if err != nil {
		log.Fatal(err)
	}

	numPass := 0
	buf := newOutput()
	quickfeedStudents := make(map[string]string) // map of students found on quickfeed: student id -> student name
	for _, enroll := range enrollments {
		// ignore course admins and teachers
		if enroll.IsAdmin() || enroll.IsTeacher() {
			continue
		}
		studID := enroll.GetUser().GetStudentID()
		student := enroll.Name()
		if quickfeedStudents[studID] != "" {
			fmt.Printf("Duplicate student ID: %s, %q and %q\n", studID, quickfeedStudents[studID], student)
		}
		quickfeedStudents[studID] = student

		submissions := courseSubmissions.For(enroll.ID)
		numApproved := numApproved(submissions)
		approvedValue := fail
		if approved(numApproved, *passLimit) {
			approvedValue = pass
			numPass++
		}

		rowNum, err := as.lookupRow(studID)
		if err != nil {
			// student ID not found in FS database, but has approved assignments
			rowNum, err = as.lookupRowByName(student)
			if err != nil {
				// student name not found in FS database, but has approved assignments
				buf.addQF(student, studID, approvedValue, numApproved)
				continue
			}
		}
		as.setApproveCell(rowNum, approvedValue)
		// use student name from FS
		student = as.lookupStudentByRow(rowNum)
		buf.addBoth(rowNum, student, studID, approvedValue, numApproved)
	}

	// find students signed up to course, but not found in QuickFeed
	for studID, rowNum := range as.approveStudMap {
		_, ok := quickfeedStudents[studID]
		if !ok {
			// student found in FS, but not in QuickFeed
			buf.addFS(rowNum, as.lookupStudentByRow(rowNum), studID, fail, 0)
			as.setApproveCell(rowNum, fail)
		}
	}

	buf.Print(*showAll)
	fmt.Printf("Total: %d, passed: %d, fail: %d\n", len(as.approveMap), numPass, len(as.approveMap)-numPass)
	if err = saveApproveSheet(*courseCode, as.sheetName, as.approveMap); err != nil {
		log.Fatal(err)
	}
}

type output struct {
	fs     map[int]string // row -> student data
	qf     map[int]string
	both   map[int]string
	negRow int
}

func newOutput() *output {
	return &output{
		fs:     make(map[int]string),
		qf:     make(map[int]string),
		both:   make(map[int]string),
		negRow: -1,
	}
}

func (o *output) addFS(row int, student, studID, approveValue string, numApproved int) {
	o.fs[row] = out(row, student, studID, approveValue, numApproved, true, false)
}

func (o *output) addQF(student, studID, approveValue string, numApproved int) {
	// we use a negative row number for students found in QuickFeed, but not in FS
	o.qf[o.negRow] = outNoRow(student, studID, approveValue, numApproved, false, true)
	o.negRow--
}

func (o *output) addBoth(row int, student, studID, approveValue string, numApproved int) {
	o.both[row] = out(row, student, studID, approveValue, numApproved, true, true)
}

func (o *output) Print(showAll bool) {
	tw := tabwriter.NewWriter(os.Stdout, 2, 8, 2, ' ', 0)
	fmt.Fprint(tw, head())

	rows := Keys(o.both)
	if showAll {
		slices.Sort(rows)
		for _, r := range rows {
			fmt.Fprint(tw, o.both[r])
		}
	}
	rows = Keys(o.fs)
	slices.Sort(rows)
	for _, r := range rows {
		fmt.Fprint(tw, o.fs[r])
	}
	rows = Keys(o.qf)
	slices.Sort(rows)
	for _, r := range rows {
		fmt.Fprint(tw, o.qf[r])
	}
	tw.Flush()
	fmt.Println("----------")
	fmt.Printf("FS: %d, QF: %d, Both: %d\n", len(o.fs), len(o.qf), len(o.both))
}

func numApproved(submissions []*qf.Submission) int {
	numApproved := 0
	for _, s := range submissions {
		if s.IsApproved() {
			numApproved++
		}
	}
	return numApproved
}

func approved(numApproved, passLimit int) bool {
	return numApproved >= passLimit
}

func head() string {
	return "Row#\tStudent\tStudID\tApproved\tFS\tQF\t#Approved\n"
}

func out(row int, student, studID, approvedValue string, numApproved int, fs, qf bool) string {
	return fmt.Sprintf("%d\t%s\t%s\t%s\t%s\t%s\t%d\n", row, student, studID, approvedValue, mark(fs), mark(qf), numApproved)
}

func outNoRow(student, studID, approvedValue string, numApproved int, fs, qf bool) string {
	return fmt.Sprintf("\t%s\t%s\t%s\t%s\t%s\t%d\n", student, studID, approvedValue, mark(fs), mark(qf), numApproved)
}

func mark(b bool) string {
	if b {
		return "✓"
	}
	return "x"
}

func Keys[K comparable, V any](m map[K]V) []K {
	ks := make([]K, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
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

func fileName(courseCode, suffix string) string {
	return strings.ToLower(courseCode) + suffix
}

const (
	firstNameColumn    = "Fornavn"
	lastNameColumn     = "Etternavn"
	studentNumColumn   = "Studentnr."
	candidateNumColumn = "Kandidatnr."
	approvedColumn     = "Godkjenning"
)

type approveSheet struct {
	sheetName      string
	headerLabels   map[string]string
	headerIndexes  map[string]int
	rows           [][]string
	approveNameMap map[string]int
	approveStudMap map[string]int
	approveMap     map[string]string
}

func newApproveSheet(sheetName string, rows [][]string) (*approveSheet, error) {
	as := &approveSheet{
		sheetName: sheetName,
		headerLabels: map[string]string{
			firstNameColumn:    "A",
			lastNameColumn:     "B",
			studentNumColumn:   "C",
			candidateNumColumn: "D",
			approvedColumn:     "E",
		},
		headerIndexes: map[string]int{
			firstNameColumn:    0,
			lastNameColumn:     1,
			studentNumColumn:   2,
			candidateNumColumn: 3,
			approvedColumn:     4,
		},
		rows:           rows[1:],                // skip header row
		approveNameMap: make(map[string]int),    // map of full names to row numbers
		approveStudMap: make(map[string]int),    // map of student numbers to row numbers
		approveMap:     make(map[string]string), // map of approve cells to approval status
	}
	for i, row := range as.rows { // skip header row
		rowNum := i + 2 // since we skip the header row
		fn := as.fullName(row)
		sn := as.studentNum(row)
		as.approveNameMap[fn] = rowNum
		as.approveStudMap[sn] = rowNum
	}
	return as, nil
}

func (a *approveSheet) fullName(row []string) string {
	fi, li := a.headerIndexes[firstNameColumn], a.headerIndexes[lastNameColumn]
	first, last := row[fi], row[li]
	if first == "" && last == "" {
		return "MISSING NAME"
	}
	return fmt.Sprintf("%s %s", first, last)
}

func (a *approveSheet) studentNum(row []string) string {
	return row[a.headerIndexes[studentNumColumn]]
}

func (a *approveSheet) lookupStudentByRow(rowNum int) string {
	return a.fullName(a.rows[rowNum-2])
}

func (a *approveSheet) lookupRow(studNum string) (int, error) {
	if rowNum, ok := a.approveStudMap[studNum]; ok {
		return rowNum, nil
	}
	return 0, fmt.Errorf("not found: %s", studNum)
}

func (a *approveSheet) lookupRowByName(name string) (int, error) {
	if rowNum, ok := a.approveNameMap[name]; ok {
		return rowNum, nil
	}
	return partialMatch(name, a.approveNameMap)
}

func (a *approveSheet) setApproveCell(rowNum int, approveValue string) {
	a.approveMap[a.approveCell(rowNum)] = approveValue
}

func (a *approveSheet) approveCell(rowNum int) string {
	return fmt.Sprintf("%s%d", a.headerLabels[approvedColumn], rowNum)
}

func loadApproveSheet(courseCode string) (*approveSheet, error) {
	// The approve sheet is a single sheet Excel file with five columns:
	//
	//		First name | Last name | Student number    | Candidate number | Approval
	// 		-----------+-----------+-------------------+------------------+----------
	// 		<first>    | <last>    | <student_no>      | <candidate_no>   | <approved>
	//		John       | Doe       | 123456            |                  |
	//
	// Approval and candidate number columns are empty by default.
	// The approval column should be filled with either "Godkjent" or "Ikke godkjent".
	// The candidate number column is irrelevant for approval and can be ignored.
	//
	f, err := excelize.OpenFile(fileName(courseCode, srcSuffix))
	if err != nil {
		return nil, err
	}
	if f.SheetCount != 1 {
		return nil, fmt.Errorf("expected a single sheet in %s, got %d", fileName(courseCode, srcSuffix), f.SheetCount)
	}
	// we expect only a single sheet; assume that is the active sheet
	sheetName := f.GetSheetName(f.GetActiveSheetIndex())
	rows := f.GetRows(sheetName)
	as, err := newApproveSheet(sheetName, rows)
	if err != nil {
		return nil, fmt.Errorf("parse error in %s: %w", fileName(courseCode, srcSuffix), err)
	}
	return as, nil
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
