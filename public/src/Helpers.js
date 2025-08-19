import { Enrollment_UserStatus, Group_GroupStatus, Enrollment_DisplayState, Submission_Status, GradeSchema, SubmissionSchema, SubmissionsSchema, GroupSchema } from "../proto/qf/types_pb";
import { timestampDate } from "@bufbuild/protobuf/wkt";
import { create, isMessage } from "@bufbuild/protobuf";
export var Color;
(function (Color) {
    Color["RED"] = "danger";
    Color["BLUE"] = "primary";
    Color["GREEN"] = "success";
    Color["YELLOW"] = "warning";
    Color["GRAY"] = "secondary";
    Color["WHITE"] = "light";
    Color["BLACK"] = "dark";
})(Color || (Color = {}));
export var Sort;
(function (Sort) {
    Sort[Sort["NAME"] = 0] = "NAME";
    Sort[Sort["STATUS"] = 1] = "STATUS";
    Sort[Sort["ID"] = 2] = "ID";
})(Sort || (Sort = {}));
export var ConnStatus;
(function (ConnStatus) {
    ConnStatus[ConnStatus["CONNECTED"] = 0] = "CONNECTED";
    ConnStatus[ConnStatus["DISCONNECTED"] = 1] = "DISCONNECTED";
    ConnStatus[ConnStatus["RECONNECTING"] = 2] = "RECONNECTING";
})(ConnStatus || (ConnStatus = {}));
export var Icon;
(function (Icon) {
    Icon["DASH"] = "fa fa-minus grey";
    Icon["USER"] = "fa fa-user";
    Icon["GROUP"] = "fa fa-users";
})(Icon || (Icon = {}));
const months = ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"];
export const getFormattedTime = (timestamp, offset) => {
    if (!timestamp) {
        return "N/A";
    }
    const date = timestampDate(timestamp);
    const tzOffset = offset ? date.getTimezoneOffset() * 60000 : 0;
    const deadline = new Date(date.getTime() + tzOffset);
    const minutes = deadline.getMinutes();
    const zero = minutes < 10 ? "0" : "";
    return `${deadline.getDate()} ${months[deadline.getMonth()]} ${deadline.getFullYear()} ${deadline.getHours()}:${zero}${minutes}`;
};
export const isExpired = (deadline) => {
    const date = timestampDate(deadline);
    const now = new Date();
    return (date.getFullYear() !== now.getFullYear() ||
        date.getMonth() > now.getMonth() + 1);
};
export var TableColor;
(function (TableColor) {
    TableColor["BLUE"] = "table-primary";
    TableColor["GREEN"] = "table-success";
    TableColor["ORANGE"] = "table-warning";
    TableColor["RED"] = "table-danger";
})(TableColor || (TableColor = {}));
const getDaysHoursAndMinutes = (deadline) => {
    const timeToDeadline = timestampDate(deadline).getTime() - Date.now();
    const days = Math.floor(timeToDeadline / (1000 * 3600 * 24));
    const hours = Math.floor(timeToDeadline / (1000 * 3600));
    const minutes = Math.floor((timeToDeadline % (1000 * 3600)) / (1000 * 60));
    return { days, hours, minutes, timeToDeadline };
};
export const deadlineFormatter = (deadline, scoreLimit, submissionScore) => {
    const { days, hours, minutes, timeToDeadline } = getDaysHoursAndMinutes(deadline);
    const daysText = Math.abs(days) === 1 ? "day" : "days";
    let className = TableColor.BLUE;
    let message = `${days} ${daysText} to deadline`;
    if (timeToDeadline < 0) {
        className = TableColor.RED;
        message = days < 0
            ? `Expired ${-days} ${daysText} ago`
            : `Expired ${-hours} hours ago`;
    }
    else if (days === 0) {
        className = TableColor.RED;
        message = `${hours === 0 ? "" : `${hours} hours and `}${minutes} minutes to deadline!`;
    }
    else if (days < 3) {
        className = TableColor.ORANGE;
        message = `${days} ${daysText} to deadline!`;
    }
    if (submissionScore >= scoreLimit) {
        className = TableColor.GREEN;
    }
    return {
        className,
        message,
        time: getFormattedTime(deadline, true),
    };
};
export const EnrollmentStatus = {
    0: "None",
    1: "Pending",
    2: "Student",
    3: "Teacher",
};
export const getPassedTestsCount = (score) => {
    let totalTests = 0;
    let passedTests = 0;
    score.forEach(s => {
        if (s.Score === s.MaxScore) {
            passedTests++;
        }
        totalTests++;
    });
    if (totalTests === 0) {
        return "";
    }
    return `${passedTests}/${totalTests}`;
};
export const hasEnrollment = (enrollments) => {
    return enrollments.some(enrollment => enrollment.status > Enrollment_UserStatus.PENDING);
};
export const isStudent = (enrollment) => { return hasStudent(enrollment.status); };
export const isTeacher = (enrollment) => { return hasTeacher(enrollment.status); };
export const isPending = (enrollment) => { return hasPending(enrollment.status); };
export const isPendingGroup = (group) => { return group.status === Group_GroupStatus.PENDING; };
export const isApprovedGroup = (group) => { return group.status === Group_GroupStatus.APPROVED; };
export const isEnrolled = (enrollment) => { return enrollment.status >= Enrollment_UserStatus.STUDENT; };
export const hasNone = (status) => { return status === Enrollment_UserStatus.NONE; };
export const hasPending = (status) => { return status === Enrollment_UserStatus.PENDING; };
export const hasStudent = (status) => { return status === Enrollment_UserStatus.STUDENT; };
export const hasTeacher = (status) => { return status === Enrollment_UserStatus.TEACHER; };
export const hasEnrolled = (status) => { return status >= Enrollment_UserStatus.PENDING; };
export const isVisible = (enrollment) => { return enrollment.state === Enrollment_DisplayState.VISIBLE; };
export const isFavorite = (enrollment) => { return enrollment.state === Enrollment_DisplayState.FAVORITE; };
export const isAuthor = (user, review) => { return user.ID === review.ReviewerID; };
export const isValidSubmissionForAssignment = (submission, assignment) => {
    return assignment.isGroupLab || submission.groupID === 0n;
};
export const isGroupSubmission = (submission) => { return submission.groupID > 0n; };
export const isManuallyGraded = (reviewers) => {
    return reviewers > 0;
};
export const isAllApproved = (submission) => { return submission.Grades.every(grade => grade.Status === Submission_Status.APPROVED); };
export const isAllRevision = (submission) => { return submission.Grades.every(grade => grade.Status === Submission_Status.REVISION); };
export const isAllRejected = (submission) => { return submission.Grades.every(grade => grade.Status === Submission_Status.REJECTED); };
export const isApproved = (status) => { return status === Submission_Status.APPROVED; };
export const isRevision = (status) => { return status === Submission_Status.REVISION; };
export const isRejected = (status) => { return status === Submission_Status.REJECTED; };
export const hasAllStatus = (submission, status) => {
    return submission.Grades.every(grade => grade.Status === status);
};
export const userHasStatus = (submission, userID, status) => {
    return submission.Grades.some(grade => grade.UserID === userID && grade.Status === status);
};
export const hasReviews = (submission) => { return submission.reviews.length > 0; };
export const hasBenchmarks = (obj) => { return obj.gradingBenchmarks.length > 0; };
export const hasCriteria = (benchmark) => { return benchmark.criteria.length > 0; };
export const hasEnrollments = (obj) => { return obj.enrollments.length > 0; };
export const hasUsers = (obj) => { return obj.users.length > 0; };
export const getStatusByUser = (submission, userID) => {
    const grade = submission.Grades.find(grade => grade.UserID === userID);
    if (!grade) {
        return Submission_Status.NONE;
    }
    return grade.Status;
};
export const setStatusByUser = (submission, userID, status) => {
    const grades = submission.Grades.map(grade => {
        if (grade.UserID === userID) {
            return create(GradeSchema, { ...grade, Status: status });
        }
        return grade;
    });
    return create(SubmissionSchema, { ...submission, Grades: grades });
};
export const setStatusAll = (submission, status) => {
    const grades = submission.Grades.map(grade => {
        return create(GradeSchema, { ...grade, Status: status });
    });
    return create(SubmissionSchema, { ...submission, Grades: grades });
};
export const isHidden = (value, query) => {
    return !value.toLowerCase().includes(query) && query.length > 0;
};
export const getSubmissionsScore = (submissions) => {
    let score = 0;
    submissions.forEach(submission => {
        score += submission.score;
    });
    return score;
};
export const getNumApproved = (submissions) => {
    let num = 0;
    submissions.forEach(submission => {
        if (isAllApproved(submission)) {
            num++;
        }
    });
    return num;
};
export const EnrollmentStatusBadge = {
    0: "",
    1: "badge badge-info",
    2: "badge badge-primary",
    3: "badge badge-danger",
};
export const SubmissionStatus = {
    0: "None",
    1: "Approved",
    2: "Rejected",
    3: "Revision",
};
export const assignmentStatusText = (assignment, submission, status) => {
    if (status === Submission_Status.NONE) {
        if (!assignment.autoApprove && submission.score >= assignment.scoreLimit) {
            return "Awaiting approval";
        }
        if (submission.score < assignment.scoreLimit) {
            return `Need ${assignment.scoreLimit}% score for approval`;
        }
    }
    return SubmissionStatus[status];
};
export const defaultTag = (date) => {
    return date.getMonth() >= 10 || date.getMonth() < 4 ? "Spring" : "Fall";
};
export const defaultYear = (date) => {
    return date.getMonth() >= 10
        ? date.getFullYear() + 1
        : date.getFullYear();
};
export const userLink = (user) => {
    return `https://github.com/${user.Login}`;
};
export const userRepoLink = (user, course) => {
    if (!course) {
        return userLink(user);
    }
    return `https://github.com/${course.ScmOrganizationName}/${user.Login}-labs`;
};
export const groupRepoLink = (group, course) => {
    if (!course) {
        return "";
    }
    return `https://github.com/${course.ScmOrganizationName}/${group.name}`;
};
export const getSubmissionCellColor = (submission, owner) => {
    if (isMessage(owner, GroupSchema)) {
        if (isAllApproved(submission)) {
            return "result-approved";
        }
        if (isAllRevision(submission)) {
            return "result-revision";
        }
        if (isAllRejected(submission)) {
            return "result-rejected";
        }
        if (submission.Grades.some(grade => grade.Status !== Submission_Status.NONE)) {
            return "result-mixed";
        }
    }
    else {
        if (userHasStatus(submission, owner.userID, Submission_Status.APPROVED)) {
            return "result-approved";
        }
        if (userHasStatus(submission, owner.userID, Submission_Status.REVISION)) {
            return "result-revision";
        }
        if (userHasStatus(submission, owner.userID, Submission_Status.REJECTED)) {
            return "result-rejected";
        }
    }
    return "clickable";
};
const pattern = /^[a-zA-Z0-9_-]+$/;
export const validateGroup = (group) => {
    if (group.name.length === 0) {
        return { valid: false, message: "Group name cannot be empty" };
    }
    if (group.name.length > 20) {
        return { valid: false, message: "Group name cannot be longer than 20 characters" };
    }
    if (group.name.includes(" ")) {
        return { valid: false, message: "Group name cannot contain spaces" };
    }
    if (!pattern.test(group.name)) {
        return { valid: false, message: "Group name can only contain letters (a-z, A-Z), numbers, underscores and dashes" };
    }
    if (group.users.length === 0) {
        return { valid: false, message: "Group must have at least one user" };
    }
    return { valid: true, message: "" };
};
export const newID = (() => {
    let id = 0;
    return () => {
        return id++;
    };
})();
export const delay = (ms) => {
    return new Promise(resolve => setTimeout(resolve, ms));
};
export var EnrollmentSort;
(function (EnrollmentSort) {
    EnrollmentSort[EnrollmentSort["Name"] = 0] = "Name";
    EnrollmentSort[EnrollmentSort["Status"] = 1] = "Status";
    EnrollmentSort[EnrollmentSort["Email"] = 2] = "Email";
    EnrollmentSort[EnrollmentSort["Activity"] = 3] = "Activity";
    EnrollmentSort[EnrollmentSort["Slipdays"] = 4] = "Slipdays";
    EnrollmentSort[EnrollmentSort["Approved"] = 5] = "Approved";
    EnrollmentSort[EnrollmentSort["StudentID"] = 6] = "StudentID";
})(EnrollmentSort || (EnrollmentSort = {}));
export var SubmissionSort;
(function (SubmissionSort) {
    SubmissionSort[SubmissionSort["ID"] = 0] = "ID";
    SubmissionSort[SubmissionSort["Name"] = 1] = "Name";
    SubmissionSort[SubmissionSort["Status"] = 2] = "Status";
    SubmissionSort[SubmissionSort["Score"] = 3] = "Score";
    SubmissionSort[SubmissionSort["Approved"] = 4] = "Approved";
})(SubmissionSort || (SubmissionSort = {}));
const enrollmentCompare = (a, b, sortBy, descending) => {
    const sortOrder = descending ? -1 : 1;
    switch (sortBy) {
        case EnrollmentSort.Name: {
            const nameA = a.user?.Name ?? "";
            const nameB = b.user?.Name ?? "";
            return sortOrder * (nameA.localeCompare(nameB));
        }
        case EnrollmentSort.Status:
            return sortOrder * (a.status - b.status);
        case EnrollmentSort.Email: {
            const emailA = a.user?.Email ?? "";
            const emailB = b.user?.Email ?? "";
            return sortOrder * (emailA.localeCompare(emailB));
        }
        case EnrollmentSort.Activity:
            if (a.lastActivityDate && b.lastActivityDate) {
                return sortOrder * (timestampDate(a.lastActivityDate).getTime() - timestampDate(b.lastActivityDate).getTime());
            }
            return 0;
        case EnrollmentSort.Slipdays:
            return sortOrder * (a.slipDaysRemaining - b.slipDaysRemaining);
        case EnrollmentSort.Approved:
            return sortOrder * Number(a.totalApproved - b.totalApproved);
        case EnrollmentSort.StudentID: {
            const aID = a.user?.ID ?? BigInt(0);
            const bID = b.user?.ID ?? BigInt(0);
            return sortOrder * Number(aID - bID);
        }
        default:
            return 0;
    }
};
export const sortEnrollments = (enrollments, sortBy, descending) => {
    return enrollments.sort((a, b) => {
        return enrollmentCompare(a, b, sortBy, descending);
    });
};
export class SubmissionsForCourse {
    userSubmissions = new Map();
    groupSubmissions = new Map();
    ForUser(enrollment) {
        return this.userSubmissions.get(enrollment.ID)?.submissions ?? [];
    }
    ForGroup(group) {
        if (isMessage(group, GroupSchema)) {
            return this.groupSubmissions.get(group.ID)?.submissions ?? [];
        }
        return this.groupSubmissions.get(group.groupID)?.submissions ?? [];
    }
    ForOwner(owner) {
        if (owner.type === "GROUP") {
            return this.groupSubmissions.get(owner.id)?.submissions ?? [];
        }
        return this.userSubmissions.get(owner.id)?.submissions ?? [];
    }
    ByID(id) {
        for (const submissions of this.userSubmissions.values()) {
            const submission = submissions.submissions.find(s => s.ID === id);
            if (submission) {
                return submission;
            }
        }
        for (const submissions of this.groupSubmissions.values()) {
            const submission = submissions.submissions.find(s => s.ID === id);
            if (submission) {
                return submission;
            }
        }
        return undefined;
    }
    OwnerByID(id) {
        for (const [key, submissions] of this.userSubmissions.entries()) {
            const submission = submissions.submissions.find(s => s.ID === id);
            if (submission) {
                if (submission.groupID > 0) {
                    return { type: "GROUP", id: submission.groupID };
                }
                return { type: "ENROLLMENT", id: key };
            }
        }
        for (const [key, submissions] of this.groupSubmissions.entries()) {
            const submission = submissions.submissions.find(s => s.ID === id);
            if (submission) {
                return { type: "GROUP", id: key };
            }
        }
        return undefined;
    }
    update(owner, submission) {
        const submissions = this.ForOwner(owner);
        const index = submissions.findIndex(s => s.AssignmentID === submission.AssignmentID);
        if (index === -1) {
            return;
        }
        else {
            submissions[index] = submission;
        }
        if (owner.type === "GROUP") {
            const clone = new Map(this.groupSubmissions);
            this.groupSubmissions = clone.set(owner.id, create(SubmissionsSchema, { submissions }));
        }
        else {
            const clone = new Map(this.userSubmissions);
            this.userSubmissions = clone.set(owner.id, create(SubmissionsSchema, { submissions }));
        }
    }
    setSubmissions(type, submissions) {
        const map = new Map();
        for (const [key, value] of Object.entries(submissions.submissions)) {
            map.set(BigInt(key), value);
        }
        switch (type) {
            case "USER":
                this.userSubmissions = map;
                break;
            case "GROUP":
                this.groupSubmissions = map;
                break;
        }
    }
}
export class SubmissionsForUser {
    submissions = new Map();
    groupSubmissions = new Map();
    ForGroup(courseID) {
        return this.groupSubmissions.get(courseID) ?? [];
    }
    ForAssignment(assignment) {
        const submissions = [];
        const groupSubs = this.groupSubmissions.get(assignment.CourseID) ?? [];
        const userSubs = this.submissions.get(assignment.CourseID) ?? [];
        for (const sub of groupSubs) {
            if (sub.AssignmentID === assignment.ID) {
                submissions.push(sub);
            }
        }
        for (const sub of userSubs) {
            if (sub.AssignmentID === assignment.ID) {
                submissions.push(sub);
            }
        }
        return submissions;
    }
    ByID(submissionID) {
        for (const submissions of this.submissions.values()) {
            const submission = submissions.find(s => s.ID === submissionID);
            if (submission) {
                return submission;
            }
        }
        for (const submissions of this.groupSubmissions.values()) {
            const submission = submissions.find(s => s.ID === submissionID);
            if (submission) {
                return submission;
            }
        }
        return undefined;
    }
    update(submission) {
        for (const [courseID, submissions] of this.submissions) {
            const index = submissions.findIndex(s => s.ID === submission.ID);
            if (index !== -1) {
                submissions[index] = submission;
                const clone = new Map(this.submissions);
                this.submissions = clone.set(courseID, submissions);
                return;
            }
        }
        for (const [courseID, submissions] of this.groupSubmissions) {
            const index = submissions.findIndex(s => s.ID === submission.ID);
            if (index !== -1) {
                submissions[index] = submission;
                const clone = new Map(this.groupSubmissions);
                this.groupSubmissions = clone.set(courseID, submissions);
                return;
            }
        }
    }
    setSubmissions(courseID, type, submissions) {
        if (type === "USER") {
            const clone = new Map(this.submissions);
            this.submissions = clone.set(courseID, submissions);
        }
        if (type === "GROUP") {
            const clone = new Map(this.groupSubmissions);
            this.groupSubmissions = clone.set(courseID, submissions);
        }
    }
}
