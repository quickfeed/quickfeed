# QuickFeed User Manual for Teachers

## Roles and Concepts

The system has three **user** roles.

- **Administrators** can create new courses and promote other users to become administrator.
  The administrator role is system-wide.

  It is recommended that teachers responsible for one or more courses be promoted to administrator.

- **Teachers** are associated with one or more courses.
  A course can have many teachers.
  A teacher is anyone associated with the course that are not students, such as professors and teaching assistants.

  The administrator that creates a new course becomes teacher for that course.
  The teacher status of a **course creator** can never be revoked.
  The teacher that created the course can promote users to teacher status and demote them back to the student role.

  Teachers can view all course related data, such as student enrollments, student groups, lab submissions, and results.
  A teacher can also accept, reject and update student enrollments and groups.

- **Students** are associated with one or more courses.
  A student can view his own results and progress on individual assignments.

The following concepts are important to understand.

- **Assignments** are organized into folders in a git repository, as shown in [The Assignments Repository](#the-assignments-repository) section below.
  An assignment may be solved individually or by a group of students.
  For individual assignments, each student is given a separate git repository.
  For group assignments, the group of students is given a separate git repository for the group assignments.

- **Submissions** are made by a student submitting code to a supported git service provider.
  Currently, only GitHub is supported.

## GitHub

To use QuickFeed, both teachers and students must have a [GitHub](https://github.com/) account.
Each course in QuickFeed is based on a separate GitHub organization.

### A Course Organization Has Several Requirements

- You must be able to create private repositories in your organization.
  If you are associated with University of Stavanger, you can create such organizations under the [UiS Campus Enterprise account](https://github.com/enterprises/university-of-stavanger).

- You can also [apply for an Educator discount](https://education.github.com/discount_requests/new) on GitHub.

For teachers, GitHub is happy to upgrade your organization to serve private repositories.

- There should not be any repositories in your organization before the course creation, as QuickFeed will create repositories with GitHub webhook events automatically.
  Course repositories are named `assignments`, `tests`, and `info`.
  If you already have such repositories in your organization, you will have to remove (or temporarily rename) them in order to be able to create a new course.

### Notifications

Teachers may receive lots of email notifications when students use GitHub.
To turn off such notifications for all your organizations, you can follow these steps:

1. Click [Notifications](https://github.com/settings/notifications) (make sure you are logged in first)
2. Uncheck Automatically watch repositories

You can find more details about alternative ways to turn off notifications [here](https://stackoverflow.com/questions/25108169/how-do-i-turn-off-automatic-notification-subscription-for-new-repositories-in-a).
However, it appears there is no per-organization approach to turn off notifications, in case you do want to receive notification for some of your other organizations.

## Course

### Course Repositories Structure

QuickFeed uses the following repository structure.
These will be created automatically when a course is created.

| Repository name | Description                                                                    | Access                        |
|-----------------|--------------------------------------------------------------------------------|-------------------------------|
| info            | Holds information about the course.                                            | Public                        |
| assignments     | Contains a separate folder for each assignment.                                | Students, Teachers, QuickFeed |
| username-labs   | Created for each student username in QuickFeed                                 | Student, Teachers, QuickFeed  |
| groupname       | Created by a group of students; `groupname` is decided by the students.        | Students, Teachers, QuickFeed |
| tests           | Contains a separate folder for each assignment with tests for that assignment. | Teachers, QuickFeed           |

*In QuickFeed, Teacher means any teaching staff, including teaching assistants and professors alike.*

The `assignments` folder has a separate folder for each assignment.
See section [The Assignments Repository](#the-assignments-repository) for more details.

The `username` is actually the github user name. This repository will initially be empty, and the student will need to set up a remote label called `assignments` pointing to the `assignments` repository, and pull from it to get any template code provided by the teaching staff.

The `tests` folder is used by QuickFeed to run the tests for each of the assignments.
The folder structure inside `tests` must correspond to the structure in the `assignments` repository.
Each `assignment` folder in the tests repository contains one or more test file and an `assignment.yml` configuration file that will be picked up by QuickFeed test runner.
The format of this file will describe various aspects of an assignment, such as submission deadline, approve: manual or automatic, which script file to run to test the assignment, etc.
See below for an example.

We recommend that course information in the `info` repository and source templates for assignments, tests and solution code are kept in a separate organization/repository that can be maintained over multiple years.
A member of the teaching staff can then copy course `info`, the code template for assignments and tests to the corresponding repositories in the present-year's organization, either manually or with scripts.

That is, these repositories should not be cloned or forked from an old version of the course.
This approach prevents accidentally revealing commit history from old course instances.

## Teaching Assistants

### To Give Your Teaching Assistants Access To Your Course You Have To

- Accept their enrollments into your course
- Promote them to your course's teacher on course members page

Assistants will automatically be given organization `owner` role to be able to accept student enrollments, approve student groups and access all course repositories.
They will also be added to the `allteachers` team.

## Student Enrollments

Students enroll in your course by logging in on QuickFeed with their GitHub accounts, find the course among the course cards and click the `Enroll` link.
You can access the full list of students (both already enrolled and those waiting for approval) in the `Members` menu of your course page, and accept their enrollments.

After a student's enrollment has been accepted into a course, the student will have access to the course's `assignment` repository and the student's personal repository, e.g., `student-labs`.
All students in a course will be added to the `allstudents` team in the course's GitHub organization.

Note that the student may receive three invitation emails from `quickfeed-uis[bot]`.
These emails can be ignored.

## Student Groups

Students can create groups with other students on QuickFeed, which later can be approved, rejected or edited by teacher or teacher assistants.
When approved, the group will have a corresponding GitHub team created on your course organization, along with a repository for group assignments. After that the group name cannot be changed.

Group names cannot be reused: as long as a group team/repository with a certain name exists on your course organization, a new group with that name cannot be created.

## Assignments and Tests

### The Assignments Repository

A course's `assignments` repository provides course assignments to students and is typically organized as shown below.
The `assignments` repository is the basis for each student's individual assignment repository and each group's shared repository.
A single assignment is represented by a folder containing all assignment files, e.g., `lab1` below.
Students will need to pull the provided code from the `assignments` repository, and push their solution attempts to their own repositories.

While each assignment folder can be named anything you want, we recommend using the naming convention below to deliver a pleasant user experience in the web frontend.

```text
assignments┐
           ├── lab1
           ├── lab2
           ├── lab3
           └── lab4
```

### The Tests Repository

To facilitate automated testing and scoring of student submitted solutions, a teacher must provide tests and assignment information.
This is the purpose of the `tests` repository.

The file system layout of the `tests` repository must match that of the `assignments` repository, as shown below.
The `assignment.yml` files contains the [assignment information](#assignment-information).
In addition, each assignment folder should also contain test code for the corresponding assignment.

The `scripts` folder may contain a course-specific [test runner](#test-runners), named `run.sh`, for running the tests.
If an assignment requires a different test runner, you can supply a custom `run.sh` script for that assignment.

The `scripts` folder may also contain a custom Dockerfile for the course.
Otherwise, the [test runner](#test-runners) for each assignment specifies which Docker image to use.

**(Beta feature: Issues and Pull Requests)**
In addition, an assignment folder may contain one or more `task-*.md` files with exercise task descriptions.
These task files must contain markdown content with a title specified on the first line.
That is, the first line must start with `#` followed by the title.
The title must then be followed by a blank line before the task description body text.

Tasks will be used to create issues on the repositories of students and groups.
Tasks are sorted within an assignment grouping by their title.
Henceforth, if a particular ordering is desired, the teacher may prefix the title with `Task 1:` and so on.

```text
tests┐
     ├── lab1
     │   ├── assignment.yml
     │   └── run.sh
     ├── lab2
     │   └── assignment.yml
     ├── lab3
     │   ├── assignment.yml
     │   ├── task-go-questions.md
     │   ├── task-learn-go.md
     │   └── task-tour-of-go.md
     ├── lab4
     │   ├── assignment.yml
     │   └── criteria.json
     └── scripts
         ├── Dockerfile
         └── run.sh
```

### Assignment Information

As mentioned above, the `tests` repository must contain one `assignment.yml` file for each assignment.
This file provide assignment information used by QuickFeed.
An example is shown below.

```yml
order: 1
title: "Introduction to Unix"
deadline: "2020-08-30T23:59:00"
effort: "8-10 hours"
isgrouplab: false
autoapprove: true
scorelimit: 90
reviewers: 2
containertimeout: 10
```

QuickFeed only use the fields in the table below.
The `title` and `effort` are used by other tooling to create a README.md file for an assignment.

| Field              | Description                                                                                    |
|--------------------|------------------------------------------------------------------------------------------------|
| `order`            | Assignment's sequence number; used to order the assignments in the frontend.                   |
| `deadline`         | Submission deadline for the assignment.                                                        |
| `isgrouplab`       | Assignment is considered a group assignment if true; otherwise it is an individual assignment. |
| `autoapprove`      | Automatically approve the assignment when `scorelimit` is achieved.                            |
| `scorelimit`       | Minimal score needed for approval. Default is 80 %.                                            |
| `reviewers`        | Number of teachers that must review a student submission for manual approval. Default is 1.    |
| `containertimeout` | Timeout for CI container to finish building and testing submitted code. Default is 10 minutes. |

### Test Runners

A course may specify a test runner that runs the tests for all assignments.
The course-specific test runner is located in `scripts/run.sh`.
Assignment-specific test runners are located in the individual assignment folders.

The test runner is a bash script; an example is shown below.

The first line of the script specifies which Docker image to use for the tests.
For example, the test runner can specify a publicly available Docker image, such as `#image/mcr.microsoft.com/dotnet/sdk:5.0`.
However, it is also possible to use a custom Docker image, which is built from the course's `scripts/Dockerfile`.
In this case, the test runner should specify the course code as the image to use, i.e., `#image/{course_code}`.
The example below is for our QF101 test course.
Note that the image will only be built/downloaded once, and will be cached for subsequent test runs.

QuickFeed will clone a student's repository or a group repository, and make them available via the `/quickfeed/assignments` folder inside the docker image.
Similarly, QuickFeed will also clone the `tests` repository and make it available via the `/quickfeed/tests` folder.

To simplify the test runner script QuickFeed makes the following environment variables available:

- `$TESTS`: Path to the root of the course's `tests` repository.
- `$ASSIGNMENTS`: Path to the root of the course's `assignments` repository.
- `$SUBMITTED`: Path to the root of the student's or group's clone  of the `assignments` repository, where submissions are received.
- `$CURRENT`: The current assignment folder; this folder should exist in all three repositories.

The first three environment variables are always set to the following paths:

```bash
TESTS=/quickfeed/tests
ASSIGNMENTS=/quickfeed/assignments
SUBMITTED=/quickfeed/submitted
```

Whereas the `$CURRENT` variable is set to the current assignment folder, e.g.,:

```bash
CURRENT=lab1
```

Thus, the test runner script can use these variables to manipulate the filesystem as needed.

To prepare a custom test runner, it is recommended to use the `docker run` command to ensure that the code is accessible at the appropriate locations.
You may use the `ls` command to list the contents of the various `/quickfeed` folders.

```sh
% docker run -it -v/my/local/path:/quickfeed image bash
```

Where `/my/local/path` contains the `assignments` and `tests` folders side-by-side.

Note that QuickFeed performs a lightweight sanity check of the cloned student repository before running the tests.

```shell
#image/qf101

start=$SECONDS
printf "*** Initializing Tests for %s ***\n" "$CURRENT"

# Move to folder with assignment handout code for the current assignment to test.
cd "$ASSIGNMENTS/$CURRENT"
# Remove assignment handout tests to avoid interference
find . -name '*_test.go' -exec rm -rf {} \;

# Copy tests into the base assignments folder for initializing test scores
cp -r "$TESTS"/* "$ASSIGNMENTS"/

# $TESTS does not contain go.mod and go.sum: make sure to get the kit/score package
go get -t github.com/quickfeed/quickfeed/kit/score
go mod tidy
# Initialize test scores
SCORE_INIT=1 go test -v ./... 2>&1 | grep TestName

printf "*** Preparing Test Execution for %s ***\n" "$CURRENT"

# Move to folder with submitted code for the current assignment to test.
cd "$SUBMITTED/$CURRENT"
# Remove student written tests to avoid interference
find . -name '*_test.go' -exec rm -rf {} \;

# Copy tests into student assignments folder for running tests
cp -r "$TESTS"/* "$SUBMITTED"/

# $TESTS does not contain go.mod and go.sum: make sure to get the kit/score package
go get -t github.com/quickfeed/quickfeed/kit/score
go mod tidy

printf "\n*** Finished Test Setup in %s seconds ***\n" "$(( SECONDS - start ))"
start=$SECONDS
printf "\n*** Running Tests ***\n\n"
go test -v -timeout 30s ./... 2>&1
printf "\n*** Finished Running Tests in %s seconds ***\n" "$(( SECONDS - start ))"
```

## Writing Tests

The test runner script will run the tests for the current assignment.
However, the teacher must write the tests so that each test emits a `Score` JSON object.
This `Score` object must be written to `stdout` and must contain the following fields:

```json
{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"Gradle","Score":100,"MaxScore":100,"Weight":1}
```

The session secret is generated by QuickFeed and is used to identify the test run.
A test execution can read the session secret from the `$QUICKFEED_SESSION_SECRET` environment variable.
However, once the test code has read the session secret into memory, it should set the environment variable to the empty string `""`.

For additional information about writing tests, please see the Go-based `score` package in the `kit` module.

## Tasks and Pull Requests (Experimental feature)

As mentioned above, an assignment folder may contain one or more `task-*.md` files with exercise task descriptions.
These task files must contain markdown content with a title specified on the first line.

```md
# Task 1: Go Questions

Here are some questions about Go.
```

These tasks will be used to create issues on group repositories.

The idea is that the students in a group solve the tasks/issues and then submit a pull request.
The students of the group can then review each other's code and make suggestions for improvements.
Once all the tests pass for a particular issue, the pull request can be reviewed by one or more teachers.
Once the pull request is approved, the students of the group can then merge the pull request.

Note: We don't support creating issues on student repositories since we don't have a good way to prevent cheating if we were to give access between student repositories.

## Reviewing student submissions

Assignment can be reviewed manually if the number of reviewers in the assignment's yaml file is above zero. Grading criteria can be added in groups for a selected assignment on the course's main page. Criteria descriptions and group headers can be edited at any time by simply clicking on the criterion one wishes to edit.

**Review** page gives access to creation of a manual review and feedback to a student solutions submitted for the course assignments. Only teaching staff can create reviews, and only one review per teaching staff member can be added for the same student submission for the same assignment.

Initially, a new review has *in progress* status. *Ready* status can be only set after all the grading criteria checkpoints are marked as either passed or failed. Reviews will not be shown on the **Release** page unless it is *ready*.

Comments can be left to every criterion checkpoint or to the whole group of grading criteria. A feedback to the whole submission can be added as well. Both comments and feedbacks can be edited by the reviewer.

**Release** page gives access to the overview of the results of manual reviews for all course students and assignments. There the user can see submission score for each review, the mean score for all ready reviews, set a final grade/status for a student submission (**Approved/Rejected/Revision**), look at all available reviews for each submission, and *release* the results to reveal them to students or student groups.

It is also possible to mass approve submissions or mass release reviews for an assignment by choosing a minimal score and then pressing `Approve all` or `Release all` correspondingly. Every submission with a score equal or above the set minimal score will be approved or reviews to such submissions will be released.

Grading criteria will be loaded from a `criteria.json` file if it is added to the corresponding assignment folder inside the `tests` repository.

JSON format:

```json
[
    {
        "heading": "First criteria group",
        "criteria": [
            {
                "description": "Has headers",
                "points": 5
            },
            {
                "description": "Has footers",
                "points": 10
            }
        ]
    },
    {
        "heading": "Second criteria group",
        "criteria": [
            {
                "description": "Has forms",
                "points": 5
            },
            {
                "description": "Has inputs",
                "points": 5
            },
            {
                "description": "Looks nice",
                "points": 10
            }
        ]
    }
]
```

`points` field is optional. If set, the total score for the assignment will be equal to the sum of all points for all criteria. Otherwise, each criterion counts equally towards the total score of 100%.
