package assignments

import "testing"

const testsFolder = "testdata/tests"

func TestWalkTestsRepository(t *testing.T) {
	wantFiles := map[string]struct{}{
		"testdata/tests/lab3/task-go-questions.md": {},
		"testdata/tests/lab3/task-learn-go.md":     {},
		"testdata/tests/lab3/task-tour-of-go.md":   {},
		"testdata/tests/scripts/Dockerfile":        {},
		"testdata/tests/scripts/run.sh":            {},
		"testdata/tests/lab1/assignment.yml":       {},
		"testdata/tests/lab2/assignment.yml":       {},
		"testdata/tests/lab3/assignment.yml":       {},
	}
	files, err := walkTestsRepository(testsFolder)
	if err != nil {
		t.Fatal(err)
	}
	for filename := range files {
		if _, ok := wantFiles[filename]; !ok {
			t.Errorf("unexpected file %q in %s", filename, testsFolder)
		}
	}
}

func TestReadTestsRepositoryContent(t *testing.T) {
	wantScriptTemplate := map[string]string{
		"lab1": `#image/quickfeed:go

printf "Custom lab1 script\n"
`,
		"lab2": `#image/quickfeed:go

printf "Default script\n"
`,
		"lab3": `#image/quickfeed:go

printf "Default script\n"
`,
	}

	assignments, _, err := readTestsRepositoryContent(testsFolder, 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, assignment := range assignments {
		if scriptTemplate, ok := wantScriptTemplate[assignment.Name]; ok {
			if scriptTemplate != assignment.ScriptFile {
				t.Errorf("assignment %q script template is %q, want %q", assignment.Name, assignment.ScriptFile, scriptTemplate)
			}
		}
		t.Logf("%+v", assignment.Name)
		for _, task := range assignment.GetTasks() {
			t.Logf("%s", task.GetTitle())
		}
	}
	// t.Logf("%s", dockerfile)
}
