# Improving Workflows in QuickFeed

In QuickFeed, users can be enrolled in multiple courses.
Currently, courses and enrollments are entirely separate from FS, requiring manual work to create and maintain this data.

We aim to improve the current workflows for the following tasks:

- **Course creation**
- **Student enrollment**
- **Registration of approved coursework requirements**

---

## Current Workflows (Manual)

### 1. Course Creation

- An instructor or admin manually creates a course in QuickFeed.
- This requires manually entering certain course details.

(This is low effort, but can still be automated.)

### 2. Student Enrollment

- Enrollments are added manually by users signing up to their courses on QuickFeed.
- Users are then enrolled in the course by the teaching staff.
  - The teaching staff need to check a student signup against the official list of enrolled students via Fagpersonweb or Canvas.

(This is a high-effort task that should be automated.)

### 3. Registration of Approved Coursework Requirements

- Instructors manually register coursework approvals in Fagpersonweb for each student.
- Alternatively, we can download an Excel template from FS, populate it with approval data, and re-upload it.

(This is a high-effort task that should be automated.)

---

## Desired Workflows (Automated)

These are the envisioned (automated) workflows for QuickFeed after integrating with FS.

### 1. Course Creation

- A teacher authenticate via LDAP/FEIDE to grant QuickFeed access to the teacher's courses.
- QuickFeed queries FS to fetch course details and class roster.
- QuickFeed creates the course, and is ready for student signup.

### 2. Student Enrollment

#### Option 1: User-Based Enrollment

- When a user logs in via LDAP/FEIDE, QuickFeed queries FS to fetch the user's enrolled courses.
- If the user is enrolled in a course already created in QuickFeed, they are automatically added to the course.

#### Option 2: Course-Based Enrollment

- When a course is created in QuickFeed, course details (e.g., names, student numbers, roles) is fetched from FS.
- QuickFeed stores this data and automatically grants access when users log in.

### 3. Registration of Approved Coursework Requirements

- QuickFeed uses an API to register coursework approval data directly in FS.
- This can be done continuously during the semester or triggered at the end of the term with a single action.

---
