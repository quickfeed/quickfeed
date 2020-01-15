package score

import (
	"testing"
)

func TestPassFailGradingScheme(t *testing.T) {
	var passFailTests = []struct {
		points        uint8
		expectedGrade string
	}{
		{0, "Fail"},
		{50, "Fail"},
		{59, "Fail"},
		{60, "Pass"},
		{99, "Pass"},
		{100, "Pass"},
		{101, "bad points"},
	}
	g := GradingScheme{
		Name:        "Pass/Fail (60 % pass level)",
		GradePoints: []uint8{60, 0},
		GradeNames:  []string{"Pass", "Fail"},
	}
	if g.ID != 0 {
		t.Errorf("expected ID=0 got %d", g.ID)
	}
	for _, test := range passFailTests {
		if grade := g.Grade(test.points); grade != test.expectedGrade {
			t.Errorf("got %s, want %s", grade, test.expectedGrade)
		}
	}
}

func TestCBiasGradingScheme(t *testing.T) {
	var cBiasTests = []struct {
		points        uint8
		expectedGrade string
	}{
		{0, "F"},
		{39, "F"},
		{40, "E"},
		{49, "E"},
		{50, "D"},
		{59, "D"},
		{60, "C"},
		{69, "C"},
		{70, "C"},
		{79, "C"},
		{80, "B"},
		{89, "B"},
		{90, "A"},
		{99, "A"},
		{100, "A"},
		{101, "bad points"},
	}
	g := GradingScheme{
		Name:        "C Bias (UiS Scheme)",
		GradePoints: []uint8{90, 80, 60, 50, 40, 0},
		GradeNames:  []string{"A", "B", "C", "D", "E", "F"},
	}
	if g.ID != 0 {
		t.Errorf("expected ID=0 got %d", g.ID)
	}
	for _, test := range cBiasTests {
		if grade := g.Grade(test.points); grade != test.expectedGrade {
			t.Errorf("got %s, want %s", grade, test.expectedGrade)
		}
	}
}

func TestNoBiasGradingScheme(t *testing.T) {
	var noBiasTests = []struct {
		points        uint8
		expectedGrade string
	}{
		{0, "F"},
		{39, "F"},
		{40, "E"},
		{51, "E"},
		{52, "D"},
		{63, "D"},
		{64, "C"},
		{75, "C"},
		{76, "B"},
		{87, "B"},
		{88, "A"},
		{99, "A"},
		{100, "A"},
		{101, "bad points"},
	}
	g := GradingScheme{
		Name:        "No Bias (NTNU Scheme)",
		GradePoints: []uint8{88, 76, 64, 52, 40, 0},
		GradeNames:  []string{"A", "B", "C", "D", "E", "F"},
	}
	if g.ID != 0 {
		t.Errorf("expected ID=0 got %d", g.ID)
	}
	for _, test := range noBiasTests {
		if grade := g.Grade(test.points); grade != test.expectedGrade {
			t.Errorf("got %s, want %s", grade, test.expectedGrade)
		}
	}
}
