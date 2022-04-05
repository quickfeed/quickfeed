# Quickfeed Support for Feedback via Pull Requests and Issues

The following document discusses how to implement quickfeed support for github pull requests.

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

#### Hein's response

It is easy to support running tests on other branches.
See `web/hooks/github.go`, Lines 56-61.
The questions I have regarding this is: Should we run tests for all branches, all PR-branches, or some other scheme that limits it to task-related PR branches?
Perhaps I would favor the latter approach, so as the limit students from generating too much load on the test runner.
On the other hand, it might be complicated for students to switch between branches or forget to switch before making commits and push.
I'm open to discuss alternative simplification strategies.
I think this feature is for more advanced courses, so perhaps we don't need to simplify things.

If a student accidentally or on purpose merges their branch into `main`, will re-opening the PR bring it back to a state where it can be reviewed easily?

### Co-student access to student repository on review

If a PR is to be reviewed by one or more co-students, how do we handle the fact that those students now will have access to everything on that students repository.

#### Possible solution

One solution could be to create a clone repository of the student repository that only contains the given assignment in question. This way, any co-student assigned to review a students code would only have access to that specific assignment.
This solution does seem to have its own range of problems and complications however. For example, if an assignment has multiple tasks, a PR would be created for each of them (maybe each one with its own set of reviewers), how do we in this
situation handle the cloning and assigning reviewers to the new repository.
One approach could be to always create a clone repository for each assignment. This repository would handle everything that had to do with the reviewing of said assignment, and would have to be kept up to date with the student repository in question.
This solution though seems like it creates more problems than it solves, given that it would give rise to a myriad of complexity and potential problems.
Another solution would be to give co-students access to the repositories they are reviewing, and then removing it after the assignment is completed.

#### Hein's response

I agree with the assessment that the first two solutions are too complex.
Of these solutions, I would go for the last solution to manage this via teams with (temporary) access to the repository.

However, another idea could be that we don't do co-student code review of the same assignments/tasks.
Instead, we could require that this feature be used as part of a group/team.
That is, the students within a group that already have access should review each other's PRs.
This requires that students divide the work in some way.
One idea is that each assignment could have _X_ tasks of similar difficulty, where _X_ is the number of group members.

### Correctly creating issues on late enrolling student

If an assignment has been pushed, then quickfeed will automatically create issues based on the tasks in that assignment. If a student enrolls to the course after this has happened, they will be left without these issues created.

#### Possible solution

If we save tasks in the database, we can create issues on enrolling students as part of the function handling their enrollment.

#### Hein's response

I think it makes sense to save tasks in the database.
Another solution could be to git clone the tests repository every time a student enrolls to get the tasks.

Since I'm leaning towards making this a group-only feature, we wouldn't need to tackle student enrollments.
Instead we would have to deal with the same problem for late-to-enroll groups.

## Open questions

### How reviewers are handled

When a student creates a pull request for a certain task, reviewers will have to be assigned to that pull request. Questions arise as to who these reviewers should be, and how they are to be assigned.
Should a teacher always be assigned, should co-students? And if so, how are these selected? Would it be automatically done by quickfeed, or more manually by the teachers themselves?

#### Hein's response

If we go with the group-only feature as described above, I see two options that we should support:

1. The group members could self-organize
2. QuickFeed could select reviewers

Obviously, a teacher/TA must be involved at some stage.
Hence, we should support these cases:

1. One teacher is selected and assigned as reviewer
2. One teacher is selected and assigned as just approver (if possible)

The selection would typically be round-robin among the teachers.
It should also be possible to select which teachers to draw from, that is, if some teachers are not involved in the lab work.

Another idea: maybe we could assign all teachers as reviewer/approver, but require that only one of them approve in the end.

### When should a pull request be created?

The purpose of using pull requests is to give the student a more accessible hub to track their process via actions and reviews. It would therefore make sense for the student to have the pull request accessible as they work, and not just when they
feel they are ready for approval. This question is also somewhat related to the previous one, in the sense of when are the `review` and the `approval` processes supposed to begin, and is there any difference between them?
If there is a difference between them, when are reviewers assigned, and when do we go from the `review` to the `approval` phase? When the automatic tests give the student a passing score?

#### Hein's response

I do think it makes sense to separate the two into a review process and approval.
I think this is sort of already part of the PR process.

I think students should review each others PRs.
Teachers may review PRs as well.

The final step is approval to be done by a teacher.

There are three stages:

1. (initial or draft stage) Tests are not passing
2. (review stage) Tests are passing; reviewers do their job and approve
3. (approval stage) Teachers check that tests are passing, that the reviewers did their job, and then approves the PR to be merged

Maybe in the initial stage, the PR is marked as draft; only when the all tests pass will the PR be moved out of draft mode to be reviewed.
Moving the PR out of draft mode could be automatic, but students can do it manually also.

### Oje's notes on responses

- How do we know that a task is completed? Currently quickfeed grades only based on the entire assignment.
- If group assignments are to be separated into individual tasks, one drawback would be that these tasks would have to be independent of each other. Otherwise one task can not be implemented before another is complete.
- How does the teacher communicate to quickfeed their desired "settings" for the assignment, e.g. that quickfeed is supposed to automatically assign reviewers. I assume via .yaml file.
- How would a student signal that their pull request is ready for approval? Would they have to be reliant on the teacher assigned as a reviewer simply checking in every once in a while, to see if they have gotten a passing grade?
  Or would they maybe signal in a comment on the pull request that they now want their assignment reviewed? It still leads to a situation where the approver would have to check in, in order to know.
- How does a teacher approve a task? By creating a comment on the pull request, saying that it is approved and ready to be merged with the main branch? If so this will have to be explicitly specified to students,
  otherwise we may end up with situations where students see that they have gotten a passing score on a task, and therefore merge it back into the main branch without getting approval from a teacher.
- How do we communicate to quickfeed that a task is approved? Currently assignments as a whole can be manually approved by teachers, but not tasks(?). In this sense we have no way of checking, when a pull request is closed,
  whether or not it has been approved by a teacher. This is a problem that needs to be solved, otherwise we have no good way of checking if a closing pull request is legitimate or not, i.e. that it has gone through
  all the checks that need to be fulfilled, in order to be closed.
- Many of the comments above highlight a reoccurring issue; what if a pull request is closed when it is not supposed to? When this happens, it is very important that quickfeed handles the event correctly, and that it does
  not corrupt the state of the pull request in question.
- If a teacher sets the assignment to automatically assign assigners, how is this handled? Internally we could have a data record of each pull request, with a list of users as assigners.
  How would this be communicated to the students in question? The most logical solution would be that quickfeed automatically sets reviewers on the pull request on github.
  This information would still have to be somehow communicated to students. Probably the easiest way of doing this would simply be to state in the assignment that users should check reviewers on their pull request.
- If we now are going for a group only implementation, what should happen with issues/tasks? Should issues now only be created on group repositories?
