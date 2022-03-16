# Quickfeed Support for Feedback via Pull Requests and Issues
The following is a proposed solution for implementing support for feedback via pull requests and issues in quickfeed.

## Goals
- Teachers should be able to create "task*.md" files in an assignment in the tests-repository. In doing so quickfeed will create one issue per task on the students repository.
- Students will create a pull request on relating issue when finished with a task, so that they can receive feedback from various reviewers. These reviewers can come in the form of teachers or other students.
- Students should also receive automatic feedback on their pull requests, based on results from running tests on the code or markdown content. 
- When a task has been completed, the pull request should be closed, and said part of the assignment should be set as approved.

## Challenges
Multiple challenges become apparent when trying to implement these features.

### New branch to create PR
In order to create a pull request, the student or group will first need to create a new branch. This can lead to problems and confusion on its own, especially for newer students, but will also give rise to a number of complications that might occur.
- Quickfeed currently only supports running tests on master branches.
- What if a student merges their branch into `main` before having the pull request reviewed and approved?

#### Possible solution
Making quickfeed able to run tests on other branches than `main` would be necessary to support pull requests. As for the other issue, a simple solution would be to simply manually re-open the pull request.

### Co-student access to student repository on review
If a PR is to be reviewed by one or more co-students, how do we handle the fact that those students now will have access to everything on that students repository.

#### Possible solution
One solution could be to create a clone repository of the student repository that only contains the given assignment in question. This way, any co-student assigned to review a students code would only have access to that specific assignment.
This solution does seem to have its own range of problems and complications however. For example, if an assignment has multiple tasks, a PR would be created for each of them (maybe each one with its own set of reviewers), how do we in this
situation handle the cloning and assigning reviewers to the new repository.
One approach could be to always create a clone repository for each assignment. This repository would handle everything that had to do with the reviewing of said assignment, and would have to be kept up to date with the student repository in question.
This solution though seems like it creates more problems than it solves, given that it would give rise to a myriad of complexity and potential problems.
Another solution would be to give co-students access to the repositories they are reviewing, and then removing it after the assignment is completed. 

### Correctly creating issues on late enrolling student
If an assignment has been pushed, then quickfeed will automatically create issues based on the tasks in that assignment. If a student enrolls to the courser after this has happened, they will be left without these issues created.

#### Possible solution
If we save tasks in the database, we can create issues on enrolling students as part of the function handling their enrollment.

## Open questions

### How reviewers are handled
When a student creates a pull request for a certain task, reviewers will have to be assigned to that pull request. Questions arise as to who these reviewers should be, and how they are to be assigned.
Should a teacher always be assigned, should co-students? And if so, how are these selected? Would it be automatically by quickfeed, or more manually by the teachers themselves?

### When should a pull request be created?
The purpose of using pull requests is to give the student a more accessible hub to track their process (via actions). It would therefore make sense for the student to have the pull request accessible as they work, and not just when they
feel they are ready for approval. This question is also somewhat related to the previous one, in the sense of when is the `review` and the `approval` process supposed to begin, and is there any difference between them?
If there is a difference between them, when are reviewers assigned, and when do we go from the `review` to the `approval` phase? When the automatic tests give the student a passing score? 