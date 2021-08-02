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

- **Assignments** are organized into folders in a git repository, as shown in [The Assignments Repository](#The-Assignments-Repository) section below.
  An assignment may be solved individually or by a group of students.
  For individual assignments, each student is given a separate git repository.
  For group assignments, the group of students is given a separate git repository for the group assignments.

- **Submissions** are made by a student submitting code to a supported git service provider.
  Currently, only GitHub is supported.

## GitHub

To use QuickFeed, both teachers and students must have a [GitHub](https://github.com/) account.
Each course in QuickFeed is based on a separate GitHub organization.

### A course organization has several requirements

- Third-party access must not be restricted.
  This is necessary so that QuickFeed can access the organization on your behalf.
  To enable third-party access, go to your organization's main page and select **Settings > Third-party access**, and remove restrictions or go to

  <https://github.com/organizations/{organization_name}/settings/oauth_application_policy>.

- You must be able to create private repositories in your organization.
  If you are associated with University of Stavanger, you can create such organizations under the [UiS Campus Enterprise account](https://github.com/enterprises/university-of-stavanger).

- You can also [apply for an Educator discount](https://education.github.com/discount_requests/new) on GitHub.

For teachers, GitHub is happy to upgrade your organization to serve private repositories

- There should not be any course repositories in your organization before the course creation, as QuickFeed will create repositories with GitHub webhook events automatically.
  Course repositories are repositories with names `assignments`, `tests` or `course-info`.
  If you already have such repositories in your organization, you will have to remove (or temporarily rename) them in order to be able to create a new course.

### Notifications

Teachers may receive lots of email notifications when students use GitHub.
To turn off such notifications for all your organizations, you can follow these steps:

1. Click [Notifications](https://github.com/settings/notifications) (make sure you are logged in first)
2. Uncheck Automatically watch repositories

You can find more details about alternative ways to turn off notifications [here](https://stackoverflow.com/questions/25108169/how-do-i-turn-off-automatic-notification-subscription-for-new-repositories-in-a).
However, it appears there is no per-organization approach to turn off notifications, in case you do want to receive notification for some of your other organizations.

## Course

### Course repositories structure

QuickFeed uses the following repository structure.
These will be created automatically when a course is created.

| Repository name | Description                                                                    | Access                         |
|-----------------|--------------------------------------------------------------------------------|--------------------------------|
| course-info     | Holds information about the course.                                            | Public                         |
| assignments     | Contains a separate folder for each assignment.                                | Students, Teachers, QuickFeed  |
| username-labs   | Created for each student username in QuickFeed                                 | Student, Teachers, QuickFeed   |
| tests           | Contains a separate folder for each assignment with tests for that assignment. | Teachers, QuickFeed            |

*In QuickFeed, Teacher means any teaching staff, including teaching assistants and professors alike.*

The `assignments` folder has a separate folder for each assignment. The short name for each assignment can be provided in the folder name, for example `single-paxos` or `state-machine-replication`. Typically, the assignment id gleaned from the `assignment.yml` file will determine the ordering of the assignments as they appear in lists on QuickFeed. Some courses may simply use short names, such as `lab1`, `lab2`, and so on. These will be sorted by the frontend as expected.

The `username` is actually the github user name. This repository will initially be empty, and the student will need to set up a remote label called `assignments` pointing to the `assignments` repository, and pull from it to get any template code provided by the teaching staff.

The `tests` folder is used by QuickFeed to run the tests for each of the assignments.
The folder structure inside `tests` must correspond to the structure in the `assignments` repo.
Each `assignment` folder in the tests repository contains one or more test file and an `assignment.yml` configuration file that will be picked up by QuickFeed test runner.
The format of this file will describe various aspects of an assignment, such as submission deadline, approve: manual or automatic, which script file to run to test the assignment, etc.
See below for an example.

We recommend that course-info and source templates for assignments, tests and solution code are kept in a separate organization and repository that can be maintained over multiple years.
Then a member of the teaching staff can copy course-info, the code template for assignments and tests to the corresponding repositories in the present-year's organization, either manually or with scripts.

That is, these repositories should not be cloned or forked from an old version of the course.
This approach prevents accidentally revealing commit history from old course instances.

## Teaching assistants

### To give your teaching assistants access to your course you have to

- Accept their enrollments into your course
- Promote them to your course's teacher on course members page

Assistants will automatically be given organization `owner` role to be able to accept student enrollments, approve student groups and access all course repositories.
They will also be added to the `allteachers` team.

## Student enrollments

Students enroll into your course by logging in into QuickFeed with their GitHub accounts, following `Join course` link and choosing to enroll into your course. You can access the full list of students (both already enrolled into your course or waiting for enrollment approval) on the `Members` tab of your course page, and accept their enrollments.

After a student's enrollment has been accepted, the student will receive three invitations to their registered GitHub email (corresponding with the account they have used to log in to QuickFeed). One to join the course organization, and another two to access the course's `assignments` repository and the student's personal repository.

**Note: it can take GitHub some time to issue the invitation.**

Students can also navigate to

- <https://github.com/{organization_name}/assignments> and
- <https://github.com/{organization_name}/{student_git_username}-labs>

manually and accept the invitations from there. These links are also available from QuickFeed's frontend interface, in the course menu, under the User Repository heading.

All students in a course will be added to the `allstudents` team in the course's GitHub organization.

## Student groups

Students can create groups with other students on QuickFeed, which later can be approved, rejected or edited by teacher or teacher assistants.
When approved, the group will have a corresponding GitHub team created on your course organization, along with a repository for group assignments. After that the group name cannot be changed.

Group names cannot be reused: as long as a group team/repository with a certain name exists on your course organization, a new group with that name cannot be created.

## Assignments and Tests

### The Assignments Repository

A course's `assignments` repository is used to provide course assignments to students.
The `assignments` repository would typically be organized as shown below.
The `assignments` repository is the basis for each student's individual assignment repository and each group's shared repository.
Whether or not a specific assignment is an individual assignment or a group assignment is specified in an [assignment information file](#assignment-information).

A single assignment is represented as a folder containing all assignment files, e.g. `lab1` below.
Students will pull the provided code from the `assignments` repository, and push their solution attempts to their own repositories.

```text
assignments┐
           ├── lab1
           ├── lab2
           ├── lab3
           ├── lab4
           ├── lab5
           └── lab6
```

### The Tests Repository

To facilitate automated testing and scoring of student submitted solutions, a teacher must provide tests and assignment information.
This is the purpose of the `tests` repository.

The file system layout of the `tests` repository must match that of the `assignments` repository, as shown below.
The `assignment.yml` files contains the [assignment information](#assignment-information).
In addition, each folder should also contain test code for the corresponding assignment.

```text
tests┐
     ├── lab1
     │   └── assignment.yml
     ├── lab2
     │   └── assignment.yml
     ├── lab3
     │   └── assignment.yml
     ├── lab4
     │   └── assignment.yml
     ├── lab5
     │   └── assignment.yml
     └── lab6
         └── assignment.yml
```

### Assignment Information

As mentioned above, the `tests` repository must contain one `assignment.yml` file for each assignment.
This file provide assignment information used by QuickFeed.
An example is shown below for `lab1`.

```yml
assignmentid: 1
name: "lab1"
title: "Introduction to Unix"
scriptfile: "go.sh"
deadline: "2020-08-30T23:59:00"
autoapprove: true
scorelimit: 90
isgrouplab: false
hoursmin: 6
hoursmax: 7
reviewers: 2
containertimeout: 10
```

| Field              | Description                                                                                           |
|--------------------|-------------------------------------------------------------------------------------------------------|
| `assignmentid`     | TBD                                                                                                   |
| `name`             | Name of assignment folder                                                                             |
| `scriptfile`       | Script to use for running tests.                                                                      |
| `deadline`         | Submission deadline for the assignment.                                                               |
| `autoapprove`      | Automatically approve the assignment when `scorelimit` is achieved.                                   |
| `scorelimit`       | Minimal score needed for approval. Default is 80 %.                                                   |
| `isgrouplab`       | Assignment is considered a group assignment if true; otherwise it is an individual assignment.        |
| `reviewers`        | Number of teachers that must review a student submission for approval.                                |
| `containertimeout` | Timeout for CI container to finish building and testing student submitted code. Default is 10 minutes.|

## Reviewing student submissions

Assignment can be reviewed manually if the number of reviewers in the assignment's yaml file is above zero. Grading criteria can be added in groups for a selected assignment on the course's main page. Criteria descriptions and group headers can be edited at any time by simply clicking on the criterion one wishes to edit.

**Review** page gives access to creation of a manual review and feedback to a student solutions submitted for the course assignments. Only teaching staff can create reviews, and only one review per teaching staff member can be added for the same student submission for the same assignment.

Initially, a new review has *in progress* status. *Ready* status can be only set after all the grading criteria checkpoints are marked as either passed or failed. Reviews will not be shown on the **Release** page unless it is *ready*.

Comments can be left to every criterion checkpoint or to the whole group of grading criteria. A feedback to the whole submission can be added as well. Both comments and feedbacks can be edited by the reviewer.

**Release** page gives access to the overview of the results of manual reviews for all course students and assignments. There the user can see submission score for each review, the mean score for all ready reviews, set a final grade/status for a student submission (**Approved/Rejected/Revision**), look at all available reviews for each submission, and *release* the results to reveal them to students or student groups.

It is also possible to mass approve submissions or mass release reviews for an assignment by choosing a minimal score and then pressing `Approve all` or `Release all` correspondingly. Every submission with a score equal or above the set minimal score will be approved or reviews to such submissions will be released.

Grading criteria can be loaded from a file `criteria.json` in a corresponding assignment folder inside the `Tests` repository.

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
