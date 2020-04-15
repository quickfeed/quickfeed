# Autograder

This course uses Autograder, a tool developed at the University of Stavanger
for students and teaching staff to manage the submission and validation of
lab assignments. All lab submissions from students are handled using Git,
a source code management system, and GitHub, a web-based hosting service for
Git source repositories. Thus, basic knowledge of these tools are necessary.
The procedure used to submit your lab assignments is explained below.

Students push their updated lab submissions to GitHub. Every lab submission is
then processed by a custom continuous integration tool. This tool will run
several test cases on the submitted code. Autograder generates feedback that
let the students verify if their submission implements the required functionality.
This feedback is available through a web interface. The feedback from the
Autograder system can be used by students to improve their submissions.

## Git and GitHub

Git is a distributed revision control and source code management system. Basic
knowledge of Git is required for handing in the lab assignments. There are many
resources available online for learning Git. A good book is Pro Git by Scott
Chacon. The book is available for free [here](https://git-scm.com/book).
Chapter 2.1 and 2.2 should contain the necessary information for delivering the
lab assignments.

GitHub is a web-based hosting service for software development projects that
use the Git revision control system. An introduction to Git and GitHub is
available in [this video](http://youtu.be/U8GBXvdmHT4). Students need to sign
up for a GitHub account to get access to the needed course material.

## Autograder Registration

Follow the steps below to register and sign up for the course on Autograder.

1. Go to [GitHub](http://github.com) and register. A GitHub account is required
   to sign in to Autograder. You can skip this step if you already have an
   account.

2. Click the "Sign in with GitHub" button in
   [Autograder](http://ag.itest.run) to register. You will then be
   taken to GitHub's website.

3. Approve that our Autograder application may have permission to access to the
   requested parts of your account. It is possible to make a separate GitHub
   account for only this (and other) courses if you do not want Autograder to
   access your personal one with the requested permissions.

## Signing up for the course

1. Click the Plus (+) menu and select “Join course”. Available courses will be listed.

2. Find the course and click Enroll.

3. Wait for the teaching staff to confirm your Autograder registration.

4. You will then be invited to the course organization on GitHub.
   An invitation will be sent to the email address associated with your GitHub account.
   Accept the invitation using the received email.
   Note that emails from GitHub can sometimes take a while to arrive.
   You can avoid this delay by navigating to the [assignments](https://github.com/dat320-2020/assignments)
   repository and accept the invitation.

5. Once you have accepted the invitation, you will get your own repository under the course organization on GitHub.

## Instructions for submitting a lab assignment

This section give step-by-step instructions on how to submit assignments.

1. Initially, you will get access to two repositories when you have signed up on Autograder.
   The first is the [`assignments`](https://github.com/dat320-2020/assignments)
   repository, which is where we publish all lab assignments, skeleton code
   and additional information.
   You only have read access to this repository, and its content may change
   throughout the semester, as we add new assignments or fix problems.
   The second is your own private repository named `username-labs`.
   You will have write access to this repository.
   Your solution to the assignments should be pushed here.

2. To get started, decide on a suitable location for your workspace for the course.
   In this guide we will use `$HOME/dat320-2020` as the workspace. Do the following
   making sure to replace `username` with your GitHub user name:

   ```console
   mkdir $HOME/dat320-2020
   cd $HOME/dat320-2020
   git clone https://github.com/dat320-2020/username-labs assignments
   cd assignments
   git remote add course-assignments https://github.com/dat320-2020/assignments
   git pull course-assignments master
   ```

   (you may be asked for username and password on GitHub above.)

3. To avoid having to type your password every time, follow these
   [steps](https://github.com/dat320-2020/course-info/blob/master/github-ssh.md)
   to set up SSH for GitHub authentication.

4. One of the most useful git commands is: `git status`. This will most often
   be able to tell you what you should be doing with your working copy.

5. When working with git you typically iterate between the following steps:

    1. Edit files
    2. `git status` (check to see which files have changed)
    3. `git add <edited files>`
    4. `git status` (check that all intended files have been added to the staging area.)
    5. `git commit`
    6. `git status` (check that changes have been committed.)

6. You may iterate over the steps in Step 5 many times. But eventually,
   you will want to push your changes to GitHub with the following command:

   ```console
   git push
   ```

   Note that this will only push your committed changes!

7. If you make changes to your own `username-labs` repository using the GitHub
   web interface, and want to pull or fetch those changes down to your own
   computer's (working copy), you can run the following commands:

   ```console
   git pull
   ```

   Or (depending on which merge strategy you prefer)

   ```console
   git fetch
   git rebase
   ```

   If there are conflicting changes, you will need to edit the files
   with conflicts and remove the lines that should not be there, along with the
   `>>>>`, `====`, and `<<<<<` lines.

   In later assignments you will work in groups. This approach is also the way that
   you can download (pull) your group's code changes from GitHub, assuming that
   another group member has previously pushed it to GitHub.

8. As time goes by the teaching staff may publish updates to the
   original [`assignments`](https://github.com/dat320-2020/assignments) repo,
   e.g. new or updated lab assignments. To fetch and integrate these
   updates into your own working copy, you will need to run the following commands:

   ```console
   cd $HOME/dat320-2020/assignments
   git pull course-assignments master
   ```

   Once again, if there are conflicting changes, you will need to edit the files
   with conflicts and remove the lines that should not be there, along with the
   `>>>>`, `====`, and `<<<<<` lines.

9. In summary, these are the typical steps that are necessary to make changes to
   files, add those changes to staging, commit changes and push changes to your
   own private repository on GitHub:

   ```console
   cd $HOME/dat320-2020/assignments/lab1
   vim shell-answers.md
   # make your edits and save
   git add shell-answers.md
   git commit
   # This will open an editor for you to write a commit message
   # Use for example "Implemented Assignment 1"
   git push
   ```

10. When you have pushed a change to GitHub, Autograder's built-in Continuous Integration
    system will pick up your code and run a custom tests suite on your code.

    Note that Autograder will only run tests against your master branch.
    If you do not want Autograder to test your code, you can create a separate branch,
    e.g. featureX, and work on that branch until you are finished. When you are ready
    to submit, simply merge the featureX branch into master and commit and push.
    Autograder will then pick up your code and run our tests on your code.

11. You can check the output by going to the [Autograder web
    interface](http://ag.itest.run/). The results (scores and build log) should be
    available under assignment's menu item.

12. If some of the autograder tests fail, you may make changes to your code/answers
    as many times as you like up until the deadline. Changes after the deadline
    will count against the slip days.

## Lab Approval

To have your lab assignment approved, you must come to the lab during lab hours
and present your solution. This lets you present the thought process behind your
solution, and gives us a more information for grading purposes. When you are
ready to show your solution, reach out to a member of the teaching staff.
It is expected that you can explain your code and show how it works.
You may show your solution on a lab workstation or your own
computer. The results from Autograder will also be taken into consideration
when approving a lab. At least 80% of the Autograder tests should pass for the
lab to be approved. A lab needs to be approved before Autograder will provide
feedback on the next lab assignment.

Also see the [Grading and Collaboration
Policy](https://github.com/dat320-2020/course-info/blob/master/policy.md)
document for additional information.
