# Autograder user manual for teachers

## Autograder: Roles and Concepts

The system has three **user** roles.

- **Administrators** can create new courses and promote other users to become administrator. It is common that all teachers that are responsible for one or more courses be an administrator. The administrator role is system-wide.

- **Teachers** are associated with one or more courses. A teacher can view results for all students in his course(s). A course can have many teachers. A teacher is anyone associated with the course that are not students, such as professors and teaching assistants.

- **Students** are associated with one or more courses. A student can view his own results and progress on individual and group assignments.

The following concepts are important to understand.

- **Assignments** are organized into folders in a git repository.
  - **Individual assignments** are solved by one student. There is one repository for individual assignments.
  - **Group assignments** are solved by a group of students. There is one repository for group assignments.
- **Submissions** are made by a student submitting code to a supported git service provider (e.g. github or gitlab).

*Note: Currently, GitLab support is lagging behind, so is not usable.*

## GitHub

 To use Autograder, both teachers and students must have an active [GitHub](https://github.com/) account. 

 Each course in Autograder is based on a GitHub organization. 
 
 ### A course organization has several requirements:

 - Third-party access must not be restricted. This is necessary so that Autograder can access the organization on your behalf. To enable third-party access, go to your organization's main page and select **Settings > Third-party access**, and remove restrictions or go to 
 https://github.com/organizations/{organization_name}/settings/oauth_application_policy.
 
 - You must be able to create private repositories in your organization. If you are associated with University of Stavanger, you can create such organizations under the [UiS Campus Enterprise account](https://github.com/enterprises/university-of-stavanger).

 - You can also [apply for an Educator discount](https://education.github.com/discount_requests/new) on GitHub.

For teachers, GitHub is happy to upgrade your organization to serve private repositories

 - There should not be any course repositories in your organization before the course creation, as Autograder will create repositories with GitHub webhook events automatically. Course repositories are repositories with names `assignments`, `tests`, `solutions` or `course-info`. If you already have such repositories in your organization, you will have to remove (or temporarily rename) them in order to be able to create a new course.

## Course

### Course repositories structure

Autograder uses the following repository structure. These will be created automatically when a course is created.

| Repository name |	Description                                      | Access   |
|-----------------|--------------------------------------------------|----------|
| course-info	  | Holds information about the course.              | Public   |
| assignments	  | Contains a separate folder for each assignment.  |Students, Teachers,<br>Autograder |
| username-labs   |	Created for each student username in autograder	 |Student, Teachers,<br> Autograder |
| tests	          | Contains a separate folder for each assignment<br> with tests for that assignment. |Teachers, Autograder|
| solutions	      | Typically contains assignments, tests, and<br> solutions that pass the tests. |	Teachers |

*In Autograder, Teacher means any teaching staff, including teaching assistants and professors alike.*

The `assignments` folder has a separate folder for each assignment. The short name for each assignment can be provided in the folder name, for example `single-paxos` or `state-machine-replication`. Typically, the deadline gleaned from the `assignment.yml` file will determine the ordering of the assignments as they appear in lists on Autograder. Some courses may simply use short names, such as `lab1`, `lab2`, and so on. These will be sorted by the frontend as expected.

The `username` is actually the github user name. This repository will initially be empty, and the student will need to set up a remote label called `assignments` pointing to the `assignments` repository, and pull from it to get any template code provided by the teaching staff.

The `tests` folder is used by autograder to run the tests for each of the assignments. The folder structure inside `tests` must correspond to the structure in the `assignments` repo. Each `assignment` folder in the tests repository contains one or more test file and an `assignment.yml` configuration file that will be picked up by autograder test runner. The format of this file will describe various aspects of an assignment, such as submission deadline, approve: manual or automatic, programming language, test commands, etc.

The `solutions` folder should never be shared with anyone except teachers. This folder is not used by autograder, but is created as a placeholder for the teaching staff to prepare and test the assignments locally. This folder will typically be used as the source for creating the `assignments` folder and `tests` folder.

Currently, teaching staff needs to populate these repositories manually for the course. This is important so as to prevent revealing commit history from old instances. That is, these repositories should not be cloned or forked from an old version of the course.

## Teaching assistants

### To give your teaching assistants access to your course you have to:

- Accept their enrollments into your course
- Promote them to your course's teacher on course members page

Assistants will automatically be given organization `owner` role to be able to accept student enrollments, approve student groups and access all course repositories. They will also be added to the `allteachers` team.

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

Students can create groups with other students on Autograder, which later can be approved, rejected or edited by teacher or teacher assistants. When approved, the group will have a corresponding GitHub team created on your course organization, along with a repository for group assignments. After that the group name cannot be changed. 

Group names cannot be reused: as long as a group team/repository with a certain name exists on your course organization, a new group with that name cannot be created.

## Assignments and tests

### `assignments` repository

Course repository `assignments` is used to provide course assignments for students. A single assignment is represented as a folder containing all assignment files. Students will pull the provided code from that repository and then push their solution attempt to their own student repositories.

### `tests` repository

To allow the automated build and testing of student solutions, you have to provide tests and an `assignment.yaml` template with assignment information in the `tests` repository. File structure must reflect the structure in the `assignments` repository. That is, if you have `lab1`, `lab2` folders in `assignments` repository, place the test files and the `assignment.yaml` file for lab1 in the `lab1` folder in the `tests` repository and so on for lab2.

### Example `assignment.yaml` file

The `tests` repository must contain one `assignment.yaml` file for each lab assignment, stored in the corresponding assignment's folder, e.g. for `lab1/assignment.yaml` we may have something like this:
```
assignmentid: 1
name: "lab1"
language: "go"
deadline: "2019-09-23 12:00:00"
autoapprove: false
scorelimit: 80
isgrouplab: false
```

The `autoapprove` indicates whether or not Autograder will automatically approve the assignment when sufficient number of tests pass.
The `scorelimit` field defines minimal score for a student submission to be reached before the lab will be auto approved.
If `scorelimit` is not set, every student submission that scores 80% or higher will be approved automatically.

## Student labs

Student solutions to a new assignment will not be built and tested if the previous assignment for that student is not approved.
