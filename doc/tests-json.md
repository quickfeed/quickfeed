# Pre-fetching Test Dependencies and Initializing Assignment Scores

This document describes the new functionality for pre-fetching test dependencies and initializing assignment scores to improve test execution performance.

## Overview

The system now supports pre-initializing assignment scores by storing test information (TestName, MaxScore, Weight) in the database when the tests repository is updated. This avoids the need to run tests on the assignment repository every time to initialize scores.

## Using tests.json Files

The simplest way to provide test information is to create a `tests.json` file in each assignment directory alongside the `assignment.json` file.

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

```text
tests/
├── lab1/
│   ├── assignment.json
│   ├── tests.json
│   └── ...
├── lab2/
│   ├── assignment.json
│   ├── tests.json
│   └── ...
```

## Automatic Generation for Go Assignments

For Go assignments, you can automatically generate the `tests.json` file using `cm` tool:

```sh
cm gen-tests-json -view -labs lab1
```

This command will generate the `tests.json` file in the specified lab directory based on the Go tests defined in the assignment.

Usually, this command is run as part of the `tests` target in the `Justfile`:

```sh
just tests lab1
```
