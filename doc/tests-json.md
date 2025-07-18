# Pre-fetching Test Dependencies and Initializing Assignment Scores

This document describes the new functionality for pre-fetching test dependencies and initializing assignment scores to improve test execution performance.

## Overview

The system now supports pre-initializing assignment scores by storing test information (TestName, MaxScore, Weight) in the database when the tests repository is updated. This avoids the need to run tests on the assignment repository every time to initialize scores.

## Using tests.json Files

The simplest way to provide test information is to create a `tests.json` file in each assignment directory alongside the `assignment.yml` file.

### Format

The `tests.json` file should contain an array of test objects with the following structure:

```json
[
    {
        "TestName": "TestExample1",
        "MaxScore": 100,
        "Weight": 10
    },
    {
        "TestName": "TestExample2", 
        "MaxScore": 50,
        "Weight": 5
    }
]
```

### Required Fields

- `TestName`: The name of the test function
- `MaxScore`: The maximum score possible for this test
- `Weight`: The weight of this test in the overall assignment score

### Example Directory Structure

```
tests/
├── lab1/
│   ├── assignment.yml
│   ├── tests.json
│   └── ...
├── lab2/
│   ├── assignment.yml
│   ├── tests.json
│   └── ...
```

## Automatic Generation for Go Assignments

For Go assignments, you can automatically generate the `tests.json` file using the provided script:

```bash
./scripts/gen-tests-json.sh /path/to/tests/lab1
```

This script:
1. Checks if the assignment is a Go assignment (has go.mod)
2. Sets up Go dependencies
3. Runs tests with `SCORE_INIT=1` to extract test information
4. Generates the `tests.json` file

### Requirements for Go Tests

For the automatic generation to work, your Go tests must use the score package properly:

```go
func TestExample(t *testing.T) {
    score.MaxScore(100, 10)  // MaxScore: 100, Weight: 10
    // your test code here
}
```

## How It Works

1. When the tests repository is updated, the system scans for `tests.json` files
2. If found, the test information is parsed and stored in the database
3. The test information is stored as a dummy submission with ID 0
4. During test execution, if tests fail early, the stored information provides complete score details
5. This prevents incomplete responses in the web interface when tests panic or fail

## Benefits

- **Faster Test Execution**: No need to run tests on assignment repo for score initialization
- **Better User Experience**: Complete score information even when tests fail
- **Reliable Scoring**: Prevents issues with incomplete test results
- **Pre-cached Dependencies**: Can pre-fetch test dependencies in Docker images

## Implementation Details

The functionality is implemented in `assignments/walk_tests_repo.go` and follows the same pattern as `criteria.json` handling. The test information is stored in the assignment's submissions array with a dummy submission (ID 0) containing the score information.

This approach is minimal and follows existing patterns in the codebase without requiring changes to the protobuf schema or database structure.