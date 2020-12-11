# Autograder user manual for teachers

## Autograder: Roles and Concepts

The system has three **user** roles.

- **Administrator** role is system-wide. Only administrators can create new courses and promote or demote other administrators.

- **Teachers** are associated with one or more courses. An administrator who creates a new course becomes **Course creator** teacher for that course. The teacher status of a course creator can never be revoked. A course creator teacher can promote other users to teachers and demote them back to students.

Teachers can view and edit all the course related data: student enrollments, student groups, lab assignment submissions.

- **Students** are associated with one or more courses. A student can view his own results and progress on individual and group assignments.

[//] #(This is copied directly from the old MD and has to be updated, or even removed? )
The following concepts are important to understand.

- **Assignments** are organized into folders in a git repository.
  - **Individual assignments** are solved by one student. There is one repository for individual assignments.
  - **Group assignments** are solved by a group of students. There is one repository for group assignments.
- **Submissions** are made by a student submitting code to a supported git service provider (e.g. github or gitlab).

*Note: Currently, GitLab support is lagging behind, so is not usable.*

## GitHub

To use Autograder, both teachers and students must have an active [GitHub](https://github.com/) account.

Each course in Autograder is based on a GitHub organization.

### A course organization has several requirements

- Third-party access must not be restricted.
  This is necessary so that Autograder can access the organization on your behalf.
  To enable third-party access, go to your organization's main page and select **Settings > Third-party access**, and remove restrictions or go to

  https://github.com/organizations/{organization_name}/settings/oauth_application_policy.

- You must be able to create private repositories in your organization.
  If you are associated with University of Stavanger, you can create such organizations under the [UiS Campus Enterprise account](https://github.com/enterprises/university-of-stavanger).

- You can also [apply for an Educator discount](https://education.github.com/discount_requests/new) on GitHub.

For teachers, GitHub is happy to upgrade your organization to serve private repositories

- There should not be any course repositories in your organization before the course creation, as Autograder will create repositories with GitHub webhook events automatically.
  Course repositories are repositories with names `assignments`, `tests` or `course-info`.
  If you already have such repositories in your organization, you will have to remove (or temporarily rename) them in order to be able to create a new course.

## Course

### Course repositories structure

Autograder uses the following repository structure. These will be created automatically when a course is created.

| Repository name | Description                                      | Access   |
|-----------------|--------------------------------------------------|----------|
| course-info   | Holds information about the course.              | Public   |
| assignments   | Contains a separate folder for each assignment.  |Students, Teachers,<br>Autograder |
| username-labs   | Created for each student username in Autograder |Student, Teachers,<br> Autograder |
| tests           | Contains a separate folder for each assignment<br> with tests for that assignment. |Teachers, Autograder|
| FIXME(meling) solutions      | Typically contains assignments, tests, and<br> solutions that pass the tests. | Teachers |

*In Autograder, Teacher means any teaching staff, including teaching assistants and professors alike.*

The `assignments` folder has a separate folder for each assignment. The short name for each assignment can be provided in the folder name, for example `single-paxos` or `state-machine-replication`. Typically, the assignment id gleaned from the `assignment.yml` file will determine the ordering of the assignments as they appear in lists on Autograder. Some courses may simply use short names, such as `lab1`, `lab2`, and so on. These will be sorted by the frontend as expected.

The `username` is actually the github user name. This repository will initially be empty, and the student will need to set up a remote label called `assignments` pointing to the `assignments` repository, and pull from it to get any template code provided by the teaching staff.

The `tests` folder is used by Autograder to run the tests for each of the assignments.
The folder structure inside `tests` must correspond to the structure in the `assignments` repo.
Each `assignment` folder in the tests repository contains one or more test file and an `assignment.yml` configuration file that will be picked up by Autograder test runner.
The format of this file will describe various aspects of an assignment, such as submission deadline, approve: manual or automatic, which script file to run to test the assignment, etc.
See below for an example.

The `solutions` folder should never be shared with anyone except teachers. This folder is not used by Autograder, but is created as a placeholder for the teaching staff to prepare and test the assignments locally. This folder will typically be used as the source for creating the `assignments` folder and `tests` folder.

Currently, teaching staff needs to populate these repositories manually for the course. This is important so as to prevent revealing commit history from old instances. That is, these repositories should not be cloned or forked from an old version of the course.

## Teaching assistants

### To give your teaching assistants access to your course you have to

- Accept their enrollments into your course
- Promote them to your course's teacher on course members page

Assistants will automatically be given organization `owner` role to be able to accept student enrollments, approve student groups and access all course repositories.
They will also be added to the `allteachers` team.

## Student enrollments

Students enroll into your course by logging in into Autograder with their GitHub accounts, following `Join course` link and choosing to enroll into your course. You can access the full list of students (both already enrolled into your course or waiting for enrollment approval) on the `Members` tab of your course page, and accept their enrollments.

After a student's enrollment has been accepted, the student will receive three invitations to their registered GitHub email (corresponding with the account they have used to log in to Autograder). One to join the course organization, and another two to access the course's `assignments` repository and the student's personal repository.

**Note: it can take GitHub some time to issue the invitation.**

Students can also navigate to

- https://github.com/{organization_name}/assignments and
- https://github.com/{organization_name}/{student_git_username}-labs

manually and accept the invitations from there. These links are also available from Autograder's frontend interface, in the course menu, under the User Repository heading.

All students in a course will be added to the `allstudents` team in the course's GitHub organization.

## Student groups

Students can create groups with other students on Autograder, which later can be approved, rejected or edited by teacher or teacher assistants.
When approved, the group will have a corresponding GitHub team created on your course organization, along with a repository for group assignments. After that the group name cannot be changed.

Group names cannot be reused: as long as a group team/repository with a certain name exists on your course organization, a new group with that name cannot be created.

## Assignments and tests

### `assignments` repository

Course repository `assignments` is used to provide course assignments for students. A single assignment is represented as a folder containing all assignment files. Students will pull the provided code from that repository and then push their solution attempt to their own student repositories.

### `tests` repository

To allow the automated build and testing of student solutions, you have to provide tests and an `assignment.yaml` template with assignment information in the `tests` repository. File structure must reflect the structure in the `assignments` repository. That is, if you have `lab1`, `lab2` folders in `assignments` repository, place the test files and the `assignment.yaml` file for lab1 in the `lab1` folder in the `tests` repository and so on for lab2.

### Example `assignment.yaml` file

The `tests` repository must contain one `assignment.yaml` file for each lab assignment, stored in the corresponding assignment's folder, e.g. for `lab1/assignment.yaml` we may have something like this:

```yml
assignmentid: 1
scriptfile: "go.sh"
deadline: "2019-10-25T23:00:00"
autoapprove: false
scorelimit: 80
isgrouplab: false
reviewers: 2
containertimeout: 10
skiptests: false
```

`scriptfile` is the name of the script used to run assignment tests. If there are no tests, set `skiptests` field to `true`. If `skiptests` field is not set to `true`,
`scriptfile` field is required.
`autoapprove` indicates whether or not Autograder will automatically approve the assignment when a sufficient score has been reached.
`reviewers` indicate the number of reviews to be created for a student submission to this assignment.
`scorelimit` defines the minimal percentage score on a student submission for the corresponding lab to be auto approved.
If `scorelimit` is not set, only submissions with 80% or higher will be approved automatically.
`containertimeout` sets a timeout (in minutes) for CI containers building and testing the code submitted by students. After the timeout for a container has been reached, the container will be stopped and removed, and a message about the timeout reached returned to user. This field is optional, the default timeout is 10 minutes.

## Reviewing student submissions

Assignment can be reviewed manually if the number of reviewers in the assignment's yaml file is above zero. Grading criteria can be added in groups for a selected assignment on the course's main page. Criteria descriptions and group headers can be edited at any time by simply clicking on the criterion one wishes to edit.

**Review** page gives access to creation of a manual review and feedback to a student solutions submitted for the course assignments. Only teaching staff can create reviews, and only one review per teaching staff member can be added for the same student submission for the same assignment.

Initially, a new review has *in progress* status. *Ready* status can be only set after all the grading criteria checkpoints are marked as either passed or failed. Reviews will not be shown on the **Release** page unless it is *ready*.

Comments can be left to every criterion checkpoint or to the whole group of grading criteria. A feedback to the whole submission can be added as well. Both comments and feedbacks can be edited by the reviewer.

**Release** page gives access to the overview of the results of manual reviews for all course students and assignments. There the user can see submission score for each review, the mean score for all ready reviews, set a final grade/status for a student submission (**Approved/Rejected/Revision**), look at all available reviews for each submission, and *release* the results to reveal them to students or student groups.

It is also possible to mass approve submissions or mass release reviews for an assignment by choosing a minimal score and then pressing `Approve all` or `Release all` correspondingly. Every submission with a score equal or above the set minimal score will be approved or reviews to such submissions will be released.

Grading criteria can be loaded from a file `criteria.json` in a corresponding assignment folder inside the `Tests` repository. 

JSON format: 

```
[
    {
        "heading": "First criteria group",
        "criteria": [
            {
                "description": "Has headers",
                "score": 5
            },
            {
                "description": "Has footers",
                "score": 10
            }
        ]
    },
    {
        "heading": "Second criteria group",
        "criteria": [
            {
                "description": "Has forms",
                "score": 5
            },
            {
                "description": "Has inputs",
                "score": 5
            },
            {
                "description": "Looks nice",
                "score": 10
            }
        ]
    }
]
```

`score` field is optional. If set, the max score for the assignment will be equal to the sum of all scores for each criteria. Otherwise, the max total score will be 100%.