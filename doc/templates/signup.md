# Signing up to the Course

In this course we use various systems that require additional signup procedures.

## QuickFeed

This course uses [QuickFeed](https://uis.itest.run/), a tool developed at the University of Stavanger for students and teaching staff to manage the submission and validation of lab assignments.
All lab submissions from students are handled using Git, a source code management system, and GitHub, a web-based hosting service for Git source repositories.
Thus, basic knowledge of these tools are necessary.
The procedure used to submit your lab assignments is explained in the [lab submission process](lab-submission.md).

Students push their updated lab submissions to GitHub.
Every lab submission is then processed by a custom continuous integration tool.
This tool will run several test cases on the submitted code.
QuickFeed generates feedback that let the students verify if their submission implements the required functionality.
This feedback is available through a web interface.
The feedback from the QuickFeed system can be used by students to improve their submissions.

## Git and GitHub

Git is a distributed revision control and source code management system.
Basic knowledge of Git is required for handing in the lab assignments.
There are many great resources available online for learning Git.
A good (free) book is [Pro Git](https://git-scm.com/book).
Chapter 2.1 and 2.2 should contain the necessary information for delivering the lab assignments.

GitHub is a web-based hosting service for software development projects that use the Git revision control system.
An introduction to Git and GitHub is available in [this video](http://youtu.be/U8GBXvdmHT4).

You need to sign up for a GitHub account to get access to the needed course material.

## QuickFeed Registration

Follow the steps below to register and sign up for the course on QuickFeed.
Here are two short videos describing these steps: [Part 1](https://youtu.be/3KJm4ABvTAo) and [Part 2](https://youtu.be/kMyH_-8xMGc).

1. Go to [GitHub](http://github.com) and register.
   A GitHub account is required to sign in to QuickFeed.
   You can skip this step if you already have an account.

2. Click the "Sign in with GitHub" button in [QuickFeed](http://uis.itest.run) to register.
   You will then be taken to GitHub's website.

3. Approve that our QuickFeed application may have permission to access to the requested parts of your account.
   It is possible to make a separate GitHub account for only this (and other) courses if you do not want QuickFeed to access your personal one with the requested permissions.

## Signing up for the Course on QuickFeed

1. Click the Plus (+) menu and select “Join course”.
   Available courses will be listed.

2. Find the course and click Enroll.

3. Wait for the teaching staff to confirm your QuickFeed registration.

4. You will then be invited to the course organization on GitHub and two separate repositories.
   QuickFeed should now accept these invitations on your behalf.

   However, should this fail, you can accept the invitations manually.

   - Navigate to the course organization [COURSE_ORG](https://github.com/COURSE_ORG) accept the invitation.
   - Navigate to the [assignments](https://github.com/COURSE_ORG/assignments) repository and accept the invitation.
   - Navigate to your private <https://github.com/COURSE_ORG/username-labs> repository and accept the invitation.
     Remember to replace `username` in this link with your own GitHub `username`.

   Several invitation emails will also be sent to the email address associated with your GitHub account.
   However, emails from GitHub can sometimes take a while to arrive.

5. Once you have accepted the invitations, you will get your own repository under `COURSE_ORG`, which is the course's organization on GitHub.

## Group Signup on QuickFeed

**If you prefer to work alone, you do not need to sign up for a group.**

1. Read the [policy about group assignments](policy.md#group-assignments).
   Find and agree with another student to form a group.
   We prefer groups of two, but will allow groups of three.
   It is important that all group members agree to contribute equally to the group assignments.

2. Agree on a name for the group.
   The name will be used as the group's GitHub repository.
   We prefer group names that identifies the persons in the group.
   But we will allow neat and descriptive project names as well.
   Do not use profanity or other inappropriate names.
   **The group name cannot be changed later.**

3. Navigate to the course's left menu bar and select “New Group”.

4. Enter the name of the group in the textbox above the list of students.

5. In the dialog, find your own name via the “Search for students” text box.
   Click the Plus (+) symbol to add yourself to the group.

6. Repeat the above step for the other group members.

7. Click the “Create” button.

## Discord Course Server Registration

1. Go to [Discord](https://discord.com/register) and register.
   A Discord account is required to sign in to communicate with the teaching staff during lab hours.
   You can skip this step if you already have an account.

2. To join the [COURSE_ORG Discord server](DISCORD_JOIN_URL).

3. Once connected to the server, please register with our bot, `@BOT_USER`, by sending it a direct message or by sending the command in the `#general` channel:

   ```text
   /register
   ```

   Registration requires that you provide your GitHub `username` and select the `Cloud Computing` course.

   Note that to register with the bot, you must previously have registered with QuickFeed with the same GitHub `username`.
   Hence, please make sure that you have joined the [`COURSE_ORG`](https://github.com/COURSE_ORG) GitHub organization and registered with QuickFeed first.

   If you need help with registering on the server, send a message in the `#help` channel.
