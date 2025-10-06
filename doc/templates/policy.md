# Grading and Collaboration Policy

See the [lab plan](lab-plan.md) for grading and approval requirements.

## Lab Approval Process

Some of the assignments _may_ be automatically approved by QuickFeed when sufficiently many tests pass; these does not require any manual approval.

For assignments that require TA approval, _you must_ present your solution to a member of the teaching staff.
For group assignments, each group member must **individually** present and explain their work to a TA for approval.
This should be done during lab hours.
Approval can take place in-person or remotely via Discord.
This lets you present the thought process behind your solution, and we may also provide feedback on your solution.

When you are ready to show your solution, reach out to a member of the teaching staff.
You can also use the Discord helpbot to request lab approval, by typing `/approve`.
A TA will be notified and reach out to you to make an appointment for lab approval.
If approval is done on Discord, you will be asked to share your screen, run and explain your solution.

Please be mindful of the TA's time and be prepared to run and explain your solution in a concise manner.
Do not send messages directly to the teaching staff on Discord; use the approval queue instead.

It is expected that you can explain your code and show how it works.
The results from QuickFeed will also be taken into consideration when approving a lab.
Typically, 90% of the QuickFeed tests should pass for the lab to be approved.
Some labs may have a lower threshold, which will be communicated in the lab instructions.

Note that, while we prefer that you get approval for each lab on time, it is ok to make arrangements to get approval at a later date, as long as the handin is submitted to GitHub on time.
Please contact the teaching staff to make arrangements.

For labs requiring approval, you may not be approved.
In such cases, you will be granted **one additional attempt** per lab, limited to **max three** additional attempts overall.

### Slip Days

We have frequent lab handins in this course and to add some flexibility to your schedule, we have adopted a scheme with _slip days_.
This means that if you cannot make a handin deadline, you can use up to a total of **15 slip days** throughout the semester without failing the course.
These slip days can be used for things like illness, resit exams, offshore work, military service or other conflicting deadlines.
Weekends and holidays **are included** in your slip day budget.

_Be advised that it is the date on your lab's submission as viewed on GitHub that counts towards the slip days._

**No special extensions will be given if you have exhausted all your slip days, but we will show some flexibility towards minor overruns.**

**IMPORTANT: Slip days cannot be used for the final submission deadline. All submissions must be approved before the last deadline.**

## Assignments

### Individual Assignments

_We do not accept joint handins_ for individual assignments.
However, we do encourage collaboration on learning the material, e.g. pointing each other in the right direction and giving hints and tips.
You may discuss the assignments with others, but you may not copy answers or code from another student or make your code available to others.
This obviously includes copying the code from former students or from the Internet.

### Group Assignments

For group assignments we expect students to form groups and work together.
Group members must contribute equally to code and the lab work.

Each group member must **individually** present and explain their work to a TA for approval.

Groups cannot be composed of members whose commitment to contribute to the group work is disproportionately different.
Similarly, group members must not be in an exploitable relationship with each other that may lead to a violation of these rules.
For example, group members cannot be in a romantic or marital relationship.

_This rule implies that group members should commit a similar amount of code on GitHub._

### Commit Messages

When you commit your code to GitHub, you **must** include a commit message (in English) that describes the changes you have made.
Commit messages must follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) format, as shown below.

```log
<type>(<lab>): <description>

<longer description of change>
```

Here is an example:

```log
feat(lab1): added a function to calculate the sum of two integers

The function takes two integers as input and returns the sum of the two integers.
```

### When to Commit

Commits should represent a logical unit of work.
A unit of work can be a bug fix, a new feature, a refactoring, etc.
It is up to you to decide what constitutes a unit of work, but it should not be too large.
For instance, it is **unacceptable to commit an entire lab assignment in one commit**.

Please refrain from committing code that does not compile or is otherwise broken.
Before committing a bug fix, please make sure that the bug is fixed and that it passes all local tests.
Consider adding your own tests to verify that the bug is fixed.

_Each group member must be able to explain their own contributions to the code as indicated by the commit log:_

```sh
git log --pretty=format:"%h%x09%an%x09%ad%x09%s"
```

Group members are also expected to be able to explain the code written by other group members.
That is, you must familiarize yourself with the code written by your group members.
As such it could be a good idea to let group members do code reviews of each other's code via GitHub pull requests.

If the group is using pair programming, then each member should take turns to write code and committing it with their own GitHub account.
It is expected that group members have contributed equally to the code independent of the group's choice of development methodology.

## Plagiarism Warning

Any form of cheating, plagiarism, i.e. copying of another student’s text or source code, will result in a non-passing grade, and may be reported to the university for administrative processing.
Committing acts that violate Student Conduct policies that result in course disruption are cause for suspension or dismissal from UiS.

_Don’t cheat. It’s not worth it!_

## Generative Models

You may use generative models such as ChatGPT or GitHub Copilot to generate code.
However, you must be able to explain the code that is generated as if you had written it yourself.
You are expected to adjust any generated code so that it fits the assignment, is logically correct and efficient.
The code must obviously solve the assignment and pass sufficient number of tests on QuickFeed.
