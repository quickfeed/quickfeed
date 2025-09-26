# Instructions for Submitting a Lab Assignment to QuickFeed

This section give step-by-step instructions on how to submit assignments.
In the following, you are expected to run commands from a terminal environment.

Here are two videos describing these steps: [Part 1 (~10 minutes)](https://youtu.be/Ewoax8goysg) and [Part 2 (~19 minutes)](https://youtu.be/rWrrrmgur4g).
Recommend watching videos in 1.25x or 1.5x speed.
Note that the videos were recorded in a previous semester, and some details may have changed.

- On macOS, Terminal can be started via Spotlight, by typing the first few letters of `terminal`.
- On Ubuntu Linux, you can click on the Activities item at the top left of the screen, then type the first few letters of `terminal`.
- On Windows, follow these [instructions](setup-wsl.md) to install the Windows Subsystem for Linux, if you haven't done so already.

1. Initially, you will get access to two repositories when you have signed up on QuickFeed.

   The first is the [`assignments`](https://github.com/COURSE_ORG/assignments) repository, which is where we publish all lab assignments, skeleton code and additional information.
   You only have read access to this repository, and its content may change throughout the semester, as we add new assignments or fix problems.

   The second is your own private repository named `username-labs`.
   You will have write access to this repository.
   This repository is automatically populated with the initial content from the `assignments` repository when created, so you can start working immediately.
   Your solution to the assignments should be pushed here.

2. First, you will need to set up GitHub authentication.

   Alternative 1:
   Install the [GitHub command-line tool](https://cli.github.com).

   On macOS, you can install it using [Homebrew](https://brew.sh/):

   ```console
   brew install gh
   ```

   On Ubuntu Linux, you can install it using [Snap](https://snapcraft.io/):

   ```console
   sudo snap install gh
   ```

   On Windows, you can install it using [Chocolatey](https://chocolatey.org/):

   ```console
   choco install gh
   ```

   Then, you can authenticate using:

   ```console
   gh auth login
   ```

   Follow the instructions to authenticate using your GitHub account.

   Alternative 2 (old method, but should still work):
   We made a video [tutorial (19:34)](https://youtu.be/qik3HHZW6C0) describing the steps to set up SSH keys for GitHub authentication.

3. To get started, decide on a suitable location for your workspace for the course.
   In this guide we will use `$HOME/COURSE_ORG` as the workspace.
   Do the following making sure to replace `username` with your GitHub user name:

   ```console
   mkdir $HOME/COURSE_ORG
   cd $HOME/COURSE_ORG
   ```

   Alternative 1 (using the `gh` tool):

   ```console
   gh repo clone COURSE_ORG/username-labs assignments
   ```

   Alternative 2 (using `git` directly; assuming you have set up your GitHub user with SSH keys):

   ```console
   git clone git@github.com:COURSE_ORG/username-labs assignments
   ```

4. Next, you will need to add the course's `assignments` repository as a remote repository for future updates.

   ```console
   cd assignments
   git remote add course-assignments git@github.com:COURSE_ORG/assignments
   ```

   This will allow you to pull in updates from the `assignments` repository, as we publish them.
   Since your repository was automatically populated with the initial assignments content, you typically won't need to pull immediately unless there are updates.
   See the [Update Local Working Copy from Course Assignments](#update-local-working-copy-from-course-assignments) section for more details.

5. One of the most useful git commands is: `git status`.
   This will most often be able to tell you what you should be doing with your working copy.

6. When working with `git` you typically iterate between the following steps:

    1. Edit files
    2. `git status` (check to see which files have changed)
    3. `git add <edited files>` (only add source files, not binaries)
    4. `git status` (check that all intended files have been added to the staging area)
    5. `git commit`
    6. `git status` (check that changes have been committed)

7. You may iterate over the steps above many times.
   But eventually, you will want to push your changes to GitHub with the following command:

   ```console
   git push
   ```

   Note that this will only push your committed changes!

8. In summary, these are the typical steps that are necessary to make changes to files, add those changes to staging, commit changes and push changes to your own private repository on GitHub:

   ```console
   cd $HOME/COURSE_ORG/assignments/lab1
   vim shell_questions.md
   # make your edits and save
   git add shell_questions.md
   git commit
   # This will open an editor for you to write a commit message
   # See policy.md for instructions on how to construct commit messages
   git push
   ```

9. When you have pushed a change to GitHub, QuickFeed's built-in Continuous Integration system will pick up your code and run a set of tests against your code.

   Note that QuickFeed will only run tests against your `main` branch.
   If you do not want QuickFeed to test your code, you can create a separate branch, e.g. `featureX`, and work on that branch until you are finished.

   When you are ready to submit, simply merge the `featureX` branch into `main` and commit and push.
   QuickFeed will then pick up your code and run our tests on your code.

10. You can check the output by going to the [QuickFeed web interface](http://uis.itest.run/).
    The results (scores and build log) should be available under the assignment's menu item.

11. If some of the QuickFeed tests fail, you may make changes to your code/answers as many times as you like up until the deadline.
    Changes after the deadline will count against the slip days.

## Final Submission and Resubmission of LabX

1. When you are finished with all the tasks for some `labX`, and you wish to submit, you may issue a commit message as follows to indicate that you are done:

   ```console
   git commit -m "qf(labX): final submission"
   ```

   The above text should be on the first line of the commit message, where `X` with the lab number.
   If you have no changes to commit, then you can use:

   ```console
   git commit --allow-empty -m "qf(labX): final submission"
   ```

   If there are any issues you want us to pay attention to, please add those comments after an empty line in the commit message.

   If you later find a mistake and want to resubmit, please use `labX resubmission` as the commit message.

   ```console
   git commit -m "qf(labX): resubmission"
   ```

   Notes:

   - These commit messages are not used by QuickFeed, they are only used to identify your lab submission commit when we do manual review.
   - Your slip days usage is calculated based on the deadline of `labX` and the time when you pushed the last commit to GitHub, that touched any of the files in the `labX` folder.

2. Push your changes to GitHub using:

   ```console
   git push
   ```

   After a while, you should be able to view your results in the QuickFeed web interface as described earlier.

## Update Local Working Copy from Course Assignments

1. The teaching staff may publish updates to the course's [assignments](https://github.com/COURSE_ORG/assignments) repository, e.g. new or updated lab assignments.

   First, check that your local working copy is clean using `git status`, which should instruct you to either commit your local changes or to restore any files whose changes you want to discard.

   Once your working copy is clean, you can fetch and integrate any updates from the course's `assignments` repository into your working copy, with the following command:

   ```console
   git pull course-assignments main
   ```

2. If there are conflicting changes, you will need to edit the files with conflicts.
   Normally, the conflicts are relatively straight forward to fix by picking one of the two changes:
   (i) your local change, or (2) the course assignment change.
   Sometimes you need to merge the two changes, if both are relevant for your code.
   Remember to remove the lines that start with `>>>>`, `====`, and `<<<<<`.

   Tools like VSCode has built-in support for helping to resolve merge conflicts.

3. Next, you will need to commit the merge:

   ```console
   git commit
   ```

   This will open an editor for you to write a commit message.
   However, usually you can just use the default commit message that is provided since it is a merge commit.

## Updating Local Working Copy with Changes from Web Interface

1. If you make changes to your own `username-labs` repository using the GitHub web interface, and you want to pull or fetch those changes to your local computer's working copy, you can run the following command:

   ```console
   git pull
   ```

   Or (depending on which merge strategy you prefer)

   ```console
   git fetch
   git rebase
   ```

2. If there are conflicting changes, you will need to edit the files with conflicts.
   Normally, the conflicts are relatively straight forward to fix by picking one of the two changes:
   (i) your local change, or (2) the course assignment change.
   Sometimes you need to merge the two changes, if both are relevant for your code.
   Remember to remove the lines that start with `>>>>`, `====`, and `<<<<<`.

   Tools like VSCode has built-in support for helping to resolve merge conflicts.

3. Next, you will need to commit the merge:

   ```console
   git commit
   ```

   This will open an editor for you to write a commit message.
   However, usually you can just use the default commit message that is provided since it is a merge commit.

## Working with Group Assignments

To work on group assignments, you need to clone your group's repository to your own machine, and set up the `assignments` repository as a remote for future updates.
In the instructions below, replace `groupname` with your group's repository name.
We assume you have already created the `COURSE_ORG` directory on your machine.
Note that group repositories are automatically populated with the initial assignments content when created.

```console
cd $HOME/COURSE_ORG
git clone git@github.com:COURSE_ORG/groupname.git
cd groupname
git remote add course-assignments git@github.com:COURSE_ORG/assignments
```

All group members will have write access to the `groupname` repository, and it is this repository that your solutions should be pushed to.
QuickFeed will run our tests against your `groupname` repository.

Remember that you should run:

```console
git pull course-assignments main
```

Every once in a while, to check if we have posted updates to the assignments, including new assignments.

Read the next section, for instructions on pulling in changes from your group partners.

## Updating Local Working Copy with Changes from Other Group Members

1. If another group member has made changes that has been pushed to GitHub, and you want to pull or fetch those changes to your local computer's working copy, you can run the following commands:

   ```console
   git pull
   ```

   Or (depending on which merge strategy you prefer)

   ```console
   git fetch
   git rebase
   ```

2. If there are conflicting changes, you will need to edit the files with conflicts.
   Normally, the conflicts are relatively straight forward to fix by picking one of the two changes:
   (i) your local change, or (2) the course assignment change.
   Sometimes you need to merge the two changes, if both are relevant for your code.
   Remember to remove the lines that start with `>>>>`, `====`, and `<<<<<`.

   Tools like VSCode has built-in support for helping to resolve merge conflicts.

3. Next, you will need to commit the merge:

   ```console
   git commit
   ```

   This will open an editor for you to write a commit message.
   However, usually you can just use the default commit message that is provided since it is a merge commit.
