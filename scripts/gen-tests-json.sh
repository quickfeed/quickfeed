#!/bin/bash

# gen-tests-json.sh - Generate tests.json for an assignment
# Usage: ./gen-tests-json.sh <assignment-directory>

if [ $# -ne 1 ]; then
    echo "Usage: $0 <assignment-directory>"
    echo "Example: $0 /path/to/tests/lab1"
    exit 1
fi

ASSIGNMENT_DIR="$1"

if [ ! -d "$ASSIGNMENT_DIR" ]; then
    echo "Error: Assignment directory '$ASSIGNMENT_DIR' does not exist"
    exit 1
fi

# Check if this is a Go assignment
if [ ! -f "$ASSIGNMENT_DIR/go.mod" ] && [ ! -f "$ASSIGNMENT_DIR/go.sum" ]; then
    echo "Error: This appears to be a non-Go assignment (no go.mod found)"
    echo "This script only works with Go assignments"
    exit 1
fi

echo "Generating tests.json for assignment in: $ASSIGNMENT_DIR"

# Change to the assignment directory
cd "$ASSIGNMENT_DIR"

# Ensure dependencies are available
echo "Setting up Go dependencies..."
go get -t github.com/quickfeed/quickfeed/kit/score
go mod tidy

# Run tests with SCORE_INIT to extract test information
echo "Extracting test information..."
SCORE_INIT=1 go test -v ./... 2>&1 | grep -E '^\s*\{.*"TestName"' > tests.json

# Check if tests.json was created and has content
if [ -f "tests.json" ] && [ -s "tests.json" ]; then
    echo "Successfully generated tests.json with the following test information:"
    cat tests.json
    echo
    echo "Tests.json file created at: $ASSIGNMENT_DIR/tests.json"
else
    echo "Warning: No test information found. This could mean:"
    echo "1. No tests use the score.MaxScore() or score.MaxScoreWithTask() functions"
    echo "2. Tests failed to compile or run"
    echo "3. Tests don't follow the expected pattern"
    echo
    echo "Please check that your tests use the score package properly."
    echo "Example:"
    echo "  func TestExample(t *testing.T) {"
    echo "      score.MaxScore(100, 10)"
    echo "      // your test code here"
    echo "  }"
fi