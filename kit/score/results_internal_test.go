package score

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestScoreNil(t *testing.T) {
	const want = 0
	results := &Results{Scores: nil}
	got := results.Sum()
	if got != want {
		t.Errorf("Sum() = %d, want %d", got, want)
	}
}

func TestScoresSum(t *testing.T) {
	// scoreObjects is obtained using this query (dat320-2020/lab4):
	// select score_objects from submissions where user_id='19' and assignment_id='8';
	scoreObjects := `[{"Secret":"hidden","TestName":"TestLintAG","Score":3,"MaxScore":3,"Weight":5},{"Secret":"hidden","TestName":"TestSchedulersAG/FIFO/No_jobs","Score":0,"MaxScore":10,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/FIFO/Two_jobs","Score":2,"MaxScore":2,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/FIFO/Three_jobs","Score":3,"MaxScore":3,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/FIFO/Five_jobs","Score":5,"MaxScore":5,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/FIFO/Six_jobs","Score":6,"MaxScore":6,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/FIFO/Six_jobs_unordered","Score":6,"MaxScore":6,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(2)/No_jobs","Score":0,"MaxScore":10,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(2)/Two_jobs","Score":10,"MaxScore":10,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(2)/Three_jobs","Score":15,"MaxScore":15,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(2)/Five_jobs","Score":25,"MaxScore":25,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(2)/Six_jobs","Score":28,"MaxScore":28,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(2)/Six_jobs_unordered","Score":28,"MaxScore":28,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(5)/No_jobs","Score":0,"MaxScore":10,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(5)/Two_jobs","Score":4,"MaxScore":4,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(5)/Three_jobs","Score":6,"MaxScore":6,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(5)/Five_jobs","Score":10,"MaxScore":10,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(5)/Six_jobs","Score":12,"MaxScore":12,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(5)/Six_jobs_unordered","Score":12,"MaxScore":12,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(10)/No_jobs","Score":0,"MaxScore":10,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(10)/Two_jobs","Score":2,"MaxScore":2,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(10)/Three_jobs","Score":3,"MaxScore":3,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(10)/Five_jobs","Score":5,"MaxScore":5,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(10)/Six_jobs","Score":8,"MaxScore":8,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(10)/Six_jobs_unordered","Score":8,"MaxScore":8,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SJF/No_jobs","Score":0,"MaxScore":10,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SJF/Two_jobs","Score":2,"MaxScore":2,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SJF/Three_jobs","Score":3,"MaxScore":3,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SJF/Five_jobs","Score":5,"MaxScore":5,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SJF/Six_jobs","Score":6,"MaxScore":6,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SJF/Six_jobs_unordered","Score":6,"MaxScore":6,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SJF/Six_jobs_different_unordered","Score":6,"MaxScore":6,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SS(5)/No_jobs","Score":0,"MaxScore":10,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SS(5)/ABC_jobs","Score":12,"MaxScore":12,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SS(5)/ABC_jobs_long","Score":60,"MaxScore":60,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SS(5)/Varying_length_ABC_jobs","Score":32,"MaxScore":32,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SS(5)/ABCDE_jobs","Score":84,"MaxScore":84,"Weight":2}]`
	scores := make([]*Score, 0)
	dec := json.NewDecoder(strings.NewReader(scoreObjects))
	for {
		if err := dec.Decode(&scores); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
	}
	results := newResults(scores...)
	const secret = "hidden"
	if err := results.validate(secret); err != nil {
		t.Errorf("Validate() = %v, expected <nil>", err)
	}
	got := results.Sum()
	const want = 84
	if got != want {
		t.Errorf("Sum() = %d, want %d", got, want)
	}
}

func TestScore100(t *testing.T) {
	// RegExp patterns to use to extract from JSON output.
	//
	//	 Search: \{\W+"Secret": "hidden",\W+"(\w+)"(:.*)\W+"(\w+)"(:.*)\W+"(\w+)"(:.*)\W+"(\w+)"(:\W+\d+)\n(.*)
	//	Replace: {$1$2$3$4$5$6$7$8$9
	//
	// To use, copy the JSON string start on the line after: "Scores": [
	// And stop on the line before the corresponding ].
	// You will need to add the final comma for the last element.
	const want = 100
	score100 := []*Score{
		{TestName: "TestVetCheckAG", Score: 1, MaxScore: 1, Weight: 5},
		{TestName: "TestFormattingAG", Score: 1, MaxScore: 1, Weight: 5},
		{TestName: "TestTODOItemsAG", Score: 1, MaxScore: 1, Weight: 5},
		{TestName: "TestLintAG", Score: 1, MaxScore: 1, Weight: 5},
		{TestName: "TestAverageMetrics/fifo/book_schedule1", Score: 4, MaxScore: 4, Weight: 4},
		{TestName: "TestAverageMetrics/fifo/book_schedule2", Score: 4, MaxScore: 4, Weight: 4},
		{TestName: "TestAverageMetrics/fifo/book_schedule3", Score: 4, MaxScore: 4, Weight: 4},
		{TestName: "TestAverageMetrics/rr/book_schedule1/q=1ms", Score: 4, MaxScore: 4, Weight: 4},
		{TestName: "TestRoundRobin", Score: 169, MaxScore: 169, Weight: 30},
		{TestName: "TestSingleJobMetrics/rr/book_schedule3/q=1ms", Score: 2, MaxScore: 2, Weight: 2},
		{TestName: "TestAverageMetrics/rr/book_schedule2/q=1ms", Score: 4, MaxScore: 4, Weight: 4},
		{TestName: "TestAverageMetrics/rr/book_schedule3/q=1ms", Score: 4, MaxScore: 4, Weight: 4},
		{TestName: "TestShortestJobFirst", Score: 163, MaxScore: 163, Weight: 20},
		{TestName: "TestStride", Score: 248, MaxScore: 248, Weight: 30},
		{TestName: "TestMinPass", Score: 5, MaxScore: 5, Weight: 5},
		{TestName: "TestStrideNewJob", Score: 2, MaxScore: 2, Weight: 2},
		{TestName: "TestSingleJobMetrics/fifo/book_schedule1", Score: 2, MaxScore: 2, Weight: 2},
		{TestName: "TestSingleJobMetrics/fifo/book_schedule2", Score: 2, MaxScore: 2, Weight: 2},
		{TestName: "TestSingleJobMetrics/fifo/book_schedule3", Score: 2, MaxScore: 2, Weight: 2},
		{TestName: "TestSingleJobMetrics/rr/book_schedule1/q=1ms", Score: 2, MaxScore: 2, Weight: 2},
		{TestName: "TestSingleJobMetrics/rr/book_schedule2/q=1ms", Score: 2, MaxScore: 2, Weight: 2},
	}
	score100v2 := []*Score{
		{TestName: "TestTODOItemsAG", Score: 1, MaxScore: 1, Weight: 5},
		{TestName: "TestAllocAG", Score: 14, MaxScore: 14, Weight: 20},
		{TestName: "TestAllocMultipleAG", Score: 63, MaxScore: 63, Weight: 10},
		{TestName: "TestFreeAG", Score: 40, MaxScore: 40, Weight: 20},
		{TestName: "TestPTLookupAG", Score: 12, MaxScore: 12, Weight: 10},
		{TestName: "TestNewMMUAG", Score: 12, MaxScore: 12, Weight: 10},
		{TestName: "TestReadAG", Score: 13, MaxScore: 13, Weight: 30},
		{TestName: "TestPTAppendAG", Score: 4, MaxScore: 4, Weight: 10},
		{TestName: "TestFormattingAG", Score: 1, MaxScore: 1, Weight: 5},
		{TestName: "TestLintAG", Score: 1, MaxScore: 1, Weight: 5},
		{TestName: "TestVetCheckAG", Score: 1, MaxScore: 1, Weight: 5},
		{TestName: "TestExtractAG", Score: 20, MaxScore: 20, Weight: 10},
		{TestName: "TestWriteAG", Score: 48, MaxScore: 48, Weight: 10},
		{TestName: "TestSequencesAG", Score: 16, MaxScore: 16, Weight: 40},
		{TestName: "TestMemoryManagementMultipleChoiceAG", Score: 3, MaxScore: 3, Weight: 5},
		{TestName: "TestPTFreeAG", Score: 18, MaxScore: 18, Weight: 10},
	}

	for i, sc100 := range [][]*Score{score100, score100v2} {
		t.Run(fmt.Sprintf("Sample%d", i), func(t *testing.T) {
			for _, sc := range sc100 {
				if sc.Score != sc.MaxScore {
					// sanity check; all scores must be max
					t.Errorf("%s Score=%d, expected %d", sc.TestName, sc.Score, sc.MaxScore)
				}
			}
			results := newResults(sc100...)
			if err := results.validate(""); err != nil {
				t.Error(err)
			}
			got := results.Sum()
			if got != want {
				t.Errorf("Sum() = %d, want %d", got, want)
			}
		})
	}
}

func TestAddScore(t *testing.T) {
	scoreTests := []struct {
		name string
		desc string
		in   []*Score
		want *Results
	}{
		{
			name: "Record the score of the second emitted score object",
			desc: "First score is registration of the test, second score is the actual score.",
			in: []*Score{
				{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 0},
				{TestName: "B", Secret: theSecret, Weight: 20, MaxScore: 100, Score: 0},
				{TestName: "C", Secret: theSecret, Weight: 30, MaxScore: 100, Score: 0},
				{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 50},
				{TestName: "B", Secret: theSecret, Weight: 20, MaxScore: 100, Score: 60},
				{TestName: "C", Secret: theSecret, Weight: 30, MaxScore: 100, Score: 70},
			},
			want: &Results{
				Scores: []*Score{
					{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 50},
					{TestName: "B", Secret: theSecret, Weight: 20, MaxScore: 100, Score: 60},
					{TestName: "C", Secret: theSecret, Weight: 30, MaxScore: 100, Score: 70},
				},
			},
		},
		{
			name: "TestName D is missing score",
			desc: "Can be due to test D panicking or some other reason for not emitting a score object",
			in: []*Score{
				{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 0},
				{TestName: "B", Secret: theSecret, Weight: 20, MaxScore: 100, Score: 0},
				{TestName: "C", Secret: theSecret, Weight: 30, MaxScore: 100, Score: 0},
				{TestName: "D", Secret: theSecret, Weight: 30, MaxScore: 100, Score: 0},
				{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 50},
				{TestName: "B", Secret: theSecret, Weight: 20, MaxScore: 100, Score: 60},
				{TestName: "C", Secret: theSecret, Weight: 30, MaxScore: 100, Score: 70},
			},
			want: &Results{
				Scores: []*Score{
					{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 50},
					{TestName: "B", Secret: theSecret, Weight: 20, MaxScore: 100, Score: 60},
					{TestName: "C", Secret: theSecret, Weight: 30, MaxScore: 100, Score: 70},
					{TestName: "D", Secret: theSecret, Weight: 30, MaxScore: 100, Score: 0},
				},
			},
		},
		{
			name: "Test A recorded 3 times",
			desc: "We only allow the same test to be recorded two times",
			in: []*Score{
				{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 0},
				{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 50},
				{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 100},
			},
			want: &Results{
				Scores: []*Score{
					{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: -1},
				},
			},
		},
		{
			name: "Test A with non-zero score recorded 3 times",
			desc: "We only allow the same test to be recorded two times",
			in: []*Score{
				{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 40},
				{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 50},
				{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 100},
			},
			want: &Results{
				Scores: []*Score{
					{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: -1},
				},
			},
		},
		{
			name: "Test A with non-zero score recorded 5 times",
			desc: "We only allow the same test to be recorded two times",
			in: []*Score{
				{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 40},
				{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 50},
				{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 100},
				{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 100},
				{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 100},
			},
			want: &Results{
				Scores: []*Score{
					{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: -1},
				},
			},
		},
		{
			name: "nil scores",
			desc: "nil score slice",
			in:   nil,
			want: &Results{
				Scores: []*Score{},
			},
		},
	}
	for _, test := range scoreTests {
		t.Run(test.name, func(t *testing.T) {
			results := newResults(test.in...)
			// results may contain negative scores so we do not validate here
			if diff := cmp.Diff(test.want, results, cmpopts.IgnoreUnexported(Results{})); diff != "" {
				t.Errorf("\nDescription: %s\nScores are different (-want +got):\n%s", test.desc, diff)
			}
		})
	}
}

func TestSumGrade(t *testing.T) {
	g := GradingScheme{
		Name:        "C Bias (UiS Scheme)",
		GradePoints: []uint32{90, 80, 60, 50, 40, 0},
		GradeNames:  []string{"A", "B", "C", "D", "E", "F"},
	}

	scoreGrades := []struct {
		in        []*Score
		out       uint32
		wantGrade string
	}{
		{
			in: []*Score{
				{TestName: "A", Score: 10, MaxScore: 10, Weight: 1},
				{TestName: "B", Score: 5, MaxScore: 5, Weight: 1},
				{TestName: "C", Score: 15, MaxScore: 15, Weight: 1},
			},
			out:       100,
			wantGrade: "A",
		},
		{
			in: []*Score{
				{TestName: "A", Score: 5, MaxScore: 10, Weight: 1},
				{TestName: "B", Score: 5, MaxScore: 5, Weight: 1},
				{TestName: "C", Score: 20, MaxScore: 40, Weight: 1},
			},
			out:       67,
			wantGrade: "C",
		},
		{
			in: []*Score{
				{TestName: "A", Score: 5, MaxScore: 10, Weight: 1},
				{TestName: "B", Score: 5, MaxScore: 10, Weight: 1},
				{TestName: "C", Score: 20, MaxScore: 40, Weight: 1},
			},
			out:       50,
			wantGrade: "D",
		},
		{
			in: []*Score{
				{TestName: "A", Score: 10, MaxScore: 10, Weight: 2},
				{TestName: "B", Score: 5, MaxScore: 10, Weight: 1},
				{TestName: "C", Score: 20, MaxScore: 40, Weight: 1},
			},
			out:       75,
			wantGrade: "C",
		},
		{
			in: []*Score{
				{TestName: "A", Score: 0, MaxScore: 10, Weight: 2},
				{TestName: "B", Score: 0, MaxScore: 10, Weight: 1},
				{TestName: "C", Score: 0, MaxScore: 40, Weight: 1},
			},
			out:       0,
			wantGrade: "F",
		},
	}
	for _, s := range scoreGrades {
		results := newResults(s.in...)
		if err := results.validate(""); err != nil {
			t.Error(err)
		}
		tot := results.Sum()
		grade := g.Grade(tot)
		if grade != s.wantGrade {
			t.Errorf("Grade(%d) = %s, expected %s", tot, grade, s.wantGrade)
		}
		if tot != s.out {
			t.Errorf("Sum() = %d, expected %d", tot, s.out)
		}
	}
}

func TestTaskSum(t *testing.T) {
	tests := []struct {
		scores   []*Score
		wantSums map[string]uint32
	}{
		{
			scores: []*Score{
				{TestName: "A", TaskName: "task-1", Score: 12, MaxScore: 12, Weight: 1},
				{TestName: "B", TaskName: "task-1", Score: 12, MaxScore: 12, Weight: 1},
				{TestName: "C", TaskName: "task-1", Score: 6, MaxScore: 12, Weight: 1},
				{TestName: "D", TaskName: "task-1", Score: 6, MaxScore: 12, Weight: 1},
				{TestName: "E", TaskName: "task-2", Score: 10, MaxScore: 10, Weight: 1},
				{TestName: "F", TaskName: "task-2", Score: 3, MaxScore: 12, Weight: 1},
				{TestName: "G", TaskName: "", Score: 10, MaxScore: 10, Weight: 1},
				{TestName: "H", TaskName: "", Score: 0, MaxScore: 10, Weight: 1},
				{TestName: "I", TaskName: "", Score: 0, MaxScore: 10, Weight: 1},
				{TestName: "J", TaskName: "", Score: 0, MaxScore: 10, Weight: 1},
			},
			wantSums: map[string]uint32{
				"task-1": 75,
				"task-2": 63,
				"":       53,
			},
		},
		{
			scores: []*Score{
				{TestName: "A", TaskName: "task-1", Score: 3, MaxScore: 12, Weight: 1},
				{TestName: "B", TaskName: "task-2", Score: 4, MaxScore: 12, Weight: 1},
				{TestName: "C", TaskName: "task-3", Score: 9, MaxScore: 12, Weight: 1},
				{TestName: "D", TaskName: "task-4", Score: 6, MaxScore: 12, Weight: 7},
			},
			wantSums: map[string]uint32{
				"task-1": 25,
				"task-2": 33,
				"task-3": 75,
				"task-4": 50,
				"":       48,
			},
		},
		{
			scores: []*Score{
				{TestName: "A", TaskName: "task-1", Score: 6, MaxScore: 12, Weight: 1},
				{TestName: "A", TaskName: "task-1", Score: 6, MaxScore: 12, Weight: 1},
				{TestName: "B", TaskName: "task-2", Score: 0, MaxScore: 12, Weight: 1},
				{TestName: "C", TaskName: "task-3", Score: 0, MaxScore: 12, Weight: 1},
				{TestName: "D", TaskName: "task-4", Score: 0, MaxScore: 12, Weight: 7},
			},
			wantSums: map[string]uint32{
				"task-1": 0, // duplicate test, should be ignored
				"task-2": 0,
				"task-3": 0,
				"task-4": 0,
				"":       0,
			},
		},
	}
	for _, tt := range tests {
		results := newResults(tt.scores...)
		// results may contain negative scores so we do not validate here
		for taskName, wantSum := range tt.wantSums {
			taskSum := results.TaskSum(taskName)
			if taskSum != wantSum {
				t.Errorf("TaskSum(%s) = %d, expected %d", taskName, taskSum, wantSum)
			}
		}
	}
}

func TestValidate(t *testing.T) {
	validateScores := []struct {
		desc    string
		in      []*Score
		wantErr error
	}{
		{
			in:      nil,
			wantErr: nil,
		},
		{
			in:      []*Score{},
			wantErr: nil,
		},
		{
			in: []*Score{
				{TestName: "A", Score: 10, MaxScore: 10, Weight: 1},
				{TestName: "B", Score: 5, MaxScore: 5, Weight: 1},
				{TestName: "C", Score: 15, MaxScore: 15, Weight: 1},
			},
			wantErr: nil,
		},
		{
			in: []*Score{
				{TestName: "A", Score: 5, MaxScore: 10, Weight: 1},
				{TestName: "B", Score: 5, MaxScore: 5, Weight: 1},
				{TestName: "C", Score: 20, MaxScore: 40, Weight: 1},
			},
			wantErr: nil,
		},
		{
			in: []*Score{
				{TestName: "A", Score: 5, MaxScore: 10, Weight: 1},
				{TestName: "B", Score: 5, MaxScore: 10, Weight: 1},
				{TestName: "C", Score: 20, MaxScore: 40, Weight: 1},
			},
			wantErr: nil,
		},
		{
			in: []*Score{
				{TestName: "A", Score: 10, MaxScore: 10, Weight: 2},
				{TestName: "B", Score: 5, MaxScore: 10, Weight: 1},
				{TestName: "C", Score: 20, MaxScore: 40, Weight: 1},
			},
			wantErr: nil,
		},
		{
			in: []*Score{
				{TestName: "A", Score: 0, MaxScore: 10, Weight: 2},
				{TestName: "B", Score: 0, MaxScore: 10, Weight: 1},
				{TestName: "C", Score: 0, MaxScore: 40, Weight: 1},
			},
			wantErr: nil,
		},
		{
			in: []*Score{
				{TestName: "A", Score: -10, MaxScore: 10, Weight: 1},
				{TestName: "B", Score: 5, MaxScore: 5, Weight: 1},
				{TestName: "C", Score: 15, MaxScore: 15, Weight: 1},
			},
			wantErr: ErrScoreInterval,
		},
		{
			in: []*Score{
				{TestName: "A", Score: 10, MaxScore: 10, Weight: 1},
				{TestName: "B", Score: 5, MaxScore: 5, Weight: 1},
				{TestName: "C", Score: -1, MaxScore: 15, Weight: 1},
			},
			wantErr: ErrScoreInterval,
		},
		{
			desc: "score = 0",
			in: []*Score{
				{TestName: "A", Score: 0, MaxScore: 10, Weight: 1},
			},
			wantErr: nil,
		},
		{
			desc: "score = maxScore",
			in: []*Score{
				{TestName: "A", Score: 10, MaxScore: 10, Weight: 1},
			},
			wantErr: nil,
		},
		{
			desc: "large maxScore",
			in: []*Score{
				{TestName: "A", Score: 10, MaxScore: 1000, Weight: 10},
			},
			wantErr: nil,
		},
		{
			desc: "large weight",
			in: []*Score{
				{TestName: "A", Score: 10, MaxScore: 10, Weight: 1000},
			},
			wantErr: nil,
		},
		{
			desc: "score > maxScore",
			in: []*Score{
				{TestName: "A", Score: 10, MaxScore: 1, Weight: 1},
			},
			wantErr: ErrScoreInterval,
		},
		{
			desc: "score < 0",
			in: []*Score{
				{TestName: "A", Score: -1, MaxScore: 1, Weight: 1},
			},
			wantErr: ErrScoreInterval,
		},
		{
			desc: "maxScore = 0 (would normally panic during Add)",
			in: []*Score{
				{TestName: "A", Score: 10, MaxScore: 0, Weight: 1},
			},
			wantErr: ErrMaxScore,
		},
		{
			desc: "maxScore < 0 (would normally panic during Add)",
			in: []*Score{
				{TestName: "A", Score: 10, MaxScore: -1, Weight: 1},
			},
			wantErr: ErrMaxScore,
		},
		{
			desc: "weight = 0 (would normally panic during Add)",
			in: []*Score{
				{TestName: "A", Score: 10, MaxScore: 1, Weight: 0},
			},
			wantErr: ErrWeight,
		},
		{
			desc: "weight < 0 (would normally panic during Add)",
			in: []*Score{
				{TestName: "A", Score: 10, MaxScore: 1, Weight: -1},
			},
			wantErr: ErrWeight,
		},
	}
	for _, s := range validateScores {
		results := newResults(s.in...)
		if err := results.validate(""); err != s.wantErr {
			var e, se string
			if err != nil {
				e = err.Error()
			}
			if s.wantErr != nil {
				se = s.wantErr.Error()
			}
			if !(len(se) > 0 && strings.Contains(e, se)) {
				t.Errorf("Validate() = %q, expected %v", err, s.wantErr)
			}
		}
	}
}
