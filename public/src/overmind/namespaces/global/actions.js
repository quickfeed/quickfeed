import { Code } from "@connectrpc/connect";
import { RepositoryRequestSchema, SubmissionRequest_SubmissionType, } from "../../../../proto/qf/requests_pb";
import { Enrollment_DisplayState, Enrollment_UserStatus, EnrollmentSchema, GroupSchema, SubmissionSchema, UserSchema } from "../../../../proto/qf/types_pb";
import { Color, ConnStatus, getStatusByUser, hasAllStatus, hasStudent, hasTeacher, isPending, isStudent, isTeacher, isVisible, newID, setStatusAll, setStatusByUser, SubmissionStatus, validateGroup } from "../../../Helpers";
import { isEmptyRepo } from "./internalActions";
import { clone, create, isMessage } from "@bufbuild/protobuf";
export const internal = { isEmptyRepo };
export const onInitializeOvermind = async ({ actions, effects }) => {
    effects.global.api.init(actions.global.errorHandler);
    await actions.global.fetchUserData();
    const alert = localStorage.getItem("alert");
    if (alert) {
        actions.global.alert({ text: alert, color: Color.RED });
        localStorage.removeItem("alert");
    }
};
export const handleStreamError = (context, error) => {
    context.state.connectionStatus = ConnStatus.DISCONNECTED;
    context.actions.global.alert({ text: error.message, color: Color.RED, delay: 10000 });
};
export const receiveSubmission = ({ state }, submission) => {
    state.submissions.update(submission);
};
export const getSelf = async ({ state, effects }) => {
    const response = await effects.global.api.client.getUser({});
    if (response.error) {
        return false;
    }
    state.self = response.message;
    return true;
};
export const getEnrollmentsByUser = async ({ state, effects }) => {
    const response = await effects.global.api.client.getEnrollments({
        FetchMode: {
            case: "userID",
            value: state.self.ID,
        }
    });
    if (response.error) {
        return;
    }
    state.enrollments = response.message.enrollments;
    for (const enrollment of state.enrollments) {
        state.status[enrollment.courseID.toString()] = enrollment.status;
    }
};
export const getUsers = async ({ state, effects }) => {
    const response = await effects.global.api.client.getUsers({});
    if (response.error) {
        return;
    }
    for (const user of response.message.users) {
        state.users[user.ID.toString()] = user;
    }
    state.allUsers = response.message.users.sort((a, b) => {
        if (a.IsAdmin > b.IsAdmin) {
            return -1;
        }
        if (a.IsAdmin < b.IsAdmin) {
            return 1;
        }
        return 0;
    });
};
export const updateUser = async ({ actions, effects }, user) => {
    const response = await effects.global.api.client.updateUser(user);
    if (response.error) {
        return;
    }
    await actions.global.getSelf();
};
export const getCourses = async ({ state, effects }) => {
    state.courses = [];
    const response = await effects.global.api.client.getCourses({});
    if (response.error) {
        return;
    }
    state.courses = response.message.courses;
};
export const updateAdmin = async ({ state, effects }, user) => {
    if (confirm(`Are you sure you want to ${user.IsAdmin ? "demote" : "promote"} ${user.Name}?`)) {
        const req = { ...user };
        req.IsAdmin = !user.IsAdmin;
        const response = await effects.global.api.client.updateUser(req);
        if (response.error) {
            return;
        }
        const found = state.allUsers.findIndex(s => s.ID === user.ID);
        if (found > -1) {
            state.allUsers[found].IsAdmin = req.IsAdmin;
        }
    }
};
export const getCourseData = async ({ state, effects }, { courseID }) => {
    const response = await effects.global.api.client.getCourse({
        courseID,
    });
    if (response.error) {
        return;
    }
    state.courseEnrollments[courseID.toString()] = response.message.enrollments;
    state.groups[courseID.toString()] = response.message.groups;
};
export const setEnrollmentState = async ({ effects }, enrollment) => {
    enrollment.state = isVisible(enrollment)
        ? Enrollment_DisplayState.HIDDEN
        : Enrollment_DisplayState.VISIBLE;
    await effects.global.api.client.updateCourseVisibility(enrollment);
};
export const updateSubmission = async ({ state, effects }, { owner, submission, status }) => {
    if (!submission) {
        return;
    }
    switch (owner.type) {
        case "ENROLLMENT":
            if (getStatusByUser(submission, submission.userID) === status) {
                return;
            }
            break;
        case "GROUP":
            if (hasAllStatus(submission, status)) {
                return;
            }
            break;
    }
    if (!confirm(`Are you sure you want to set status ${SubmissionStatus[status]} on this submission?`)) {
        return;
    }
    let clonedSubmission = clone(SubmissionSchema, submission);
    switch (owner.type) {
        case "ENROLLMENT":
            clonedSubmission = setStatusByUser(clonedSubmission, submission.userID, status);
            break;
        case "GROUP":
            clonedSubmission = setStatusAll(clonedSubmission, status);
            break;
    }
    const response = await effects.global.api.client.updateSubmission({
        courseID: state.activeCourse,
        submissionID: submission.ID,
        grades: clonedSubmission.Grades,
        released: submission.released,
        score: submission.score,
    });
    if (response.error) {
        return;
    }
    submission.Grades = clonedSubmission.Grades;
    state.submissionsForCourse.update(owner, submission);
};
export const updateGrade = async ({ state, effects }, { grade, status }) => {
    if (grade.Status === status || !state.selectedSubmission) {
        return;
    }
    if (!confirm(`Are you sure you want to set status ${SubmissionStatus[status]} on this grade?`)) {
        return;
    }
    const clonedSubmission = clone(SubmissionSchema, state.selectedSubmission);
    clonedSubmission.Grades = clonedSubmission.Grades.map(g => {
        if (g.UserID === grade.UserID) {
            g.Status = status;
        }
        return g;
    });
    const response = await effects.global.api.client.updateSubmission({
        courseID: state.activeCourse,
        submissionID: state.selectedSubmission.ID,
        grades: clonedSubmission.Grades,
        released: state.selectedSubmission.released,
        score: state.selectedSubmission.score,
    });
    if (response.error) {
        return;
    }
    state.selectedSubmission.Grades = clonedSubmission.Grades;
    const type = clonedSubmission.userID ? "ENROLLMENT" : "GROUP";
    switch (type) {
        case "ENROLLMENT":
            state.submissionsForCourse.update({ type, id: clonedSubmission.userID }, clonedSubmission);
            break;
        case "GROUP":
            state.submissionsForCourse.update({ type, id: clonedSubmission.groupID }, clonedSubmission);
            break;
    }
};
export const updateEnrollment = async ({ state, actions, effects }, { enrollment, status }) => {
    if (!enrollment.user) {
        return;
    }
    if (status === Enrollment_UserStatus.NONE) {
        const proceed = await actions.global.internal.isEmptyRepo(create(RepositoryRequestSchema, { userID: enrollment.userID, courseID: enrollment.courseID }));
        if (!proceed) {
            return;
        }
    }
    let confirmed = false;
    switch (status) {
        case Enrollment_UserStatus.NONE:
            confirmed = confirm("WARNING! Rejecting a student is irreversible. Are you sure?");
            break;
        case Enrollment_UserStatus.STUDENT:
            confirmed = isPending(enrollment) || confirm(`Warning! ${enrollment.user.Name} is a teacher. Are sure you want to demote?`);
            break;
        case Enrollment_UserStatus.TEACHER:
            confirmed = confirm(`Are you sure you want to promote ${enrollment.user.Name} to teacher status?`);
            break;
        case Enrollment_UserStatus.PENDING:
            return;
    }
    if (!confirmed) {
        return;
    }
    const enrollments = state.courseEnrollments[state.activeCourse.toString()] ?? [];
    const found = enrollments.findIndex(e => e.ID === enrollment.ID);
    if (found === -1) {
        return;
    }
    const clonedEnrollment = clone(EnrollmentSchema, enrollment);
    clonedEnrollment.status = status;
    const response = await effects.global.api.client.updateEnrollments({ enrollments: [clonedEnrollment] });
    if (response.error) {
        return;
    }
    if (status === Enrollment_UserStatus.NONE) {
        enrollments.splice(found, 1);
    }
    else {
        enrollments[found].status = status;
    }
};
export const approvePendingEnrollments = async ({ state, actions, effects }) => {
    if (!confirm("Please confirm that you want to approve all students")) {
        return;
    }
    const enrollments = state.pendingEnrollments.map(enrollment => {
        const temp = clone(EnrollmentSchema, enrollment);
        temp.status = Enrollment_UserStatus.STUDENT;
        return temp;
    });
    const response = await effects.global.api.client.updateEnrollments({ enrollments });
    if (response.error) {
        await actions.global.getCourseData({ courseID: state.activeCourse });
        return;
    }
    for (const enrollment of state.pendingEnrollments) {
        enrollment.status = Enrollment_UserStatus.STUDENT;
    }
};
export const getAssignments = async ({ state, actions }) => {
    await Promise.all(state.enrollments.map(async (enrollment) => {
        if (isPending(enrollment)) {
            return;
        }
        await actions.global.getAssignmentsByCourse(enrollment.courseID);
    }));
};
export const getAssignmentsByCourse = async ({ state, effects }, courseID) => {
    const response = await effects.global.api.client.getAssignments({ courseID });
    if (response.error) {
        return;
    }
    state.assignments[courseID.toString()] = response.message.assignments;
};
export const getRepositories = async ({ state, effects }) => {
    await Promise.all(state.enrollments.map(async (enrollment) => {
        if (isPending(enrollment)) {
            return;
        }
        const courseID = enrollment.courseID;
        state.repositories[courseID.toString()] = {};
        const response = await effects.global.api.client.getRepositories({ courseID });
        if (response.error) {
            return;
        }
        state.repositories[courseID.toString()] = response.message.URLs;
    }));
};
export const createGroup = async ({ state, actions, effects }, group) => {
    const check = validateGroup(group);
    if (!check.valid) {
        actions.global.alert({ text: check.message, color: Color.RED, delay: 10000 });
        return;
    }
    const response = await effects.global.api.client.createGroup({
        courseID: group.courseID,
        name: group.name,
        users: group.users.map(ID => create(UserSchema, { ID })),
    });
    if (response.error) {
        return;
    }
    state.userGroup[group.courseID.toString()] = response.message;
    state.activeGroup = null;
};
export const editCourse = async ({ actions, effects }, { course }) => {
    const response = await effects.global.api.client.updateCourse(course);
    if (response.error) {
        return;
    }
    await actions.global.getCourses();
};
export const loadCourseSubmissions = async ({ state, actions }, courseID) => {
    state.isLoading = true;
    await actions.global.refreshCourseSubmissions(courseID);
    state.loadedCourse[courseID.toString()] = true;
    state.isLoading = false;
};
export const refreshCourseSubmissions = async ({ state, effects }, courseID) => {
    const userResponse = await effects.global.api.client.getSubmissionsByCourse({
        CourseID: courseID,
        FetchMode: {
            case: "Type",
            value: SubmissionRequest_SubmissionType.ALL
        }
    });
    const groupResponse = await effects.global.api.client.getSubmissionsByCourse({
        CourseID: courseID,
        FetchMode: {
            case: "Type",
            value: SubmissionRequest_SubmissionType.GROUP
        }
    });
    if (userResponse.error || groupResponse.error) {
        return;
    }
    state.submissionsForCourse.setSubmissions("USER", userResponse.message);
    state.submissionsForCourse.setSubmissions("GROUP", groupResponse.message);
    for (const submissions of Object.values(userResponse.message.submissions)) {
        for (const submission of submissions.submissions) {
            state.review.reviews.set(submission.ID, submission.reviews);
        }
    }
};
export const getGroupsByCourse = async ({ state, effects }, courseID) => {
    state.groups[courseID.toString()] = [];
    const response = await effects.global.api.client.getGroupsByCourse({ courseID });
    if (response.error) {
        return;
    }
    state.groups[courseID.toString()] = response.message.groups;
};
export const getUserSubmissions = async ({ state, effects }, courseID) => {
    const response = await effects.global.api.client.getSubmissions({
        CourseID: courseID,
        FetchMode: {
            case: "UserID",
            value: state.self.ID,
        },
    });
    if (response.error) {
        return;
    }
    state.submissions.setSubmissions(courseID, "USER", response.message.submissions);
};
export const getGroupSubmissions = async ({ state, effects }, courseID) => {
    const enrollment = state.enrollmentsByCourseID[courseID.toString()];
    if (!enrollment?.group) {
        return;
    }
    const response = await effects.global.api.client.getSubmissions({
        CourseID: courseID,
        FetchMode: {
            case: "GroupID",
            value: enrollment.groupID,
        },
    });
    if (response.error) {
        return;
    }
    state.submissions.setSubmissions(courseID, "GROUP", response.message.submissions);
};
export const setActiveCourse = ({ state }, courseID) => {
    state.activeCourse = courseID;
};
export const toggleFavorites = ({ state }) => {
    state.showFavorites = !state.showFavorites;
};
export const setSelectedAssignmentID = ({ state }, assignmentID) => {
    state.selectedAssignmentID = assignmentID;
};
export const setSelectedSubmission = ({ state }, { submission }) => {
    if (!submission) {
        state.selectedSubmission = null;
        return;
    }
    state.selectedSubmission = clone(SubmissionSchema, submission);
};
export const getSubmission = async ({ state, effects }, { courseID, owner, submission }) => {
    const response = await effects.global.api.client.getSubmission({
        CourseID: courseID,
        FetchMode: {
            case: "SubmissionID",
            value: submission.ID,
        },
    });
    if (response.error) {
        return;
    }
    state.submissionsForCourse.update(owner, response.message);
    if (state.selectedSubmission && state.selectedSubmission.ID === submission.ID) {
        state.selectedSubmission = response.message;
    }
};
export const rebuildSubmission = async ({ state, actions, effects }, { owner, submission }) => {
    if (!(submission && state.selectedAssignment && state.activeCourse)) {
        return;
    }
    const response = await effects.global.api.client.rebuildSubmissions({
        courseID: state.activeCourse,
        assignmentID: state.selectedAssignment.ID,
        submissionID: submission.ID,
    });
    if (response.error) {
        return;
    }
    await actions.global.getSubmission({ courseID: state.activeCourse, submission, owner });
    actions.global.alert({ color: Color.GREEN, text: 'Submission rebuilt successfully' });
};
export const rebuildAllSubmissions = async ({ effects }, { courseID, assignmentID }) => {
    const response = await effects.global.api.client.rebuildSubmissions({
        courseID,
        assignmentID,
    });
    return !response.error;
};
export const enroll = async ({ state, effects }, courseID) => {
    const response = await effects.global.api.client.createEnrollment({
        courseID,
        userID: state.self.ID,
    });
    if (response.error) {
        return;
    }
    const enrolsResponse = await effects.global.api.client.getEnrollments({
        FetchMode: {
            case: "userID",
            value: state.self.ID,
        }
    });
    if (enrolsResponse.error) {
        return;
    }
    state.enrollments = enrolsResponse.message.enrollments;
};
export const updateGroupStatus = async ({ effects }, { group, status }) => {
    const oldStatus = group.status;
    group.status = status;
    const response = await effects.global.api.client.updateGroup(group);
    if (response.error) {
        group.status = oldStatus;
    }
};
export const deleteGroup = async ({ state, actions, effects }, group) => {
    if (!confirm("Deleting a group is an irreversible action. Are you sure?")) {
        return;
    }
    const proceed = await actions.global.internal.isEmptyRepo(create(RepositoryRequestSchema, { courseID: group.courseID, groupID: group.ID }));
    if (!proceed) {
        return;
    }
    const deleteResponse = await effects.global.api.client.deleteGroup({
        courseID: group.courseID,
        groupID: group.ID,
    });
    if (deleteResponse.error) {
        return;
    }
    state.groups[group.courseID.toString()] = state.groups[group.courseID.toString()].filter(g => g.ID !== group.ID);
};
export const updateGroup = async ({ state, actions, effects }, group) => {
    const response = await effects.global.api.client.updateGroup(group);
    if (response.error) {
        return;
    }
    const found = state.groups[group.courseID.toString()].find(g => g.ID === group.ID);
    if (found && response.message) {
        Object.assign(found, response.message);
        actions.global.setActiveGroup(null);
    }
};
export const createOrUpdateCriterion = async ({ effects }, { criterion, assignment }) => {
    const benchmark = assignment.gradingBenchmarks.find(bm => bm.ID === criterion.BenchmarkID);
    if (!benchmark) {
        return false;
    }
    if (criterion.ID) {
        const response = await effects.global.api.client.updateCriterion(criterion);
        if (response.error) {
            return false;
        }
        const index = benchmark.criteria.findIndex(c => c.ID === criterion.ID);
        if (index > -1) {
            benchmark.criteria[index] = criterion;
        }
    }
    else {
        criterion.CourseID = assignment.CourseID;
        const response = await effects.global.api.client.createCriterion(criterion);
        if (response.error) {
            return false;
        }
        benchmark.criteria.push(response.message);
    }
    return true;
};
export const createOrUpdateBenchmark = async ({ effects }, { benchmark, assignment }) => {
    if (benchmark.ID) {
        const response = await effects.global.api.client.updateBenchmark(benchmark);
        if (response.error) {
            return false;
        }
        const index = assignment.gradingBenchmarks.findIndex(b => b.ID === benchmark.ID);
        if (index > -1) {
            assignment.gradingBenchmarks[index] = benchmark;
        }
    }
    else {
        benchmark.CourseID = assignment.CourseID;
        const response = await effects.global.api.client.createBenchmark(benchmark);
        if (response.error) {
            return false;
        }
        assignment.gradingBenchmarks.push(response.message);
    }
    return true;
};
export const createBenchmark = async ({ effects }, { benchmark, assignment }) => {
    benchmark.AssignmentID = assignment.ID;
    const response = await effects.global.api.client.createBenchmark(benchmark);
    if (response.error) {
        return;
    }
    assignment.gradingBenchmarks.push(benchmark);
};
export const deleteCriterion = async ({ effects }, { criterion, assignment }) => {
    if (!criterion) {
        return;
    }
    const benchmarks = assignment.gradingBenchmarks;
    const benchmark = benchmarks.find(bm => bm.ID === criterion?.BenchmarkID);
    if (!benchmark) {
        return;
    }
    if (!confirm("Do you really want to delete this criterion?")) {
        return;
    }
    const response = await effects.global.api.client.deleteCriterion(criterion);
    if (response.error) {
        return;
    }
    const index = benchmarks.indexOf(benchmark);
    if (index > -1) {
        benchmarks[index].criteria = benchmarks[index].criteria.filter(c => c.ID !== criterion.ID);
    }
};
export const deleteBenchmark = async ({ effects }, { benchmark, assignment }) => {
    if (benchmark && confirm("Do you really want to delete this benchmark?")) {
        const response = await effects.global.api.client.deleteBenchmark(benchmark);
        if (response.error) {
            return;
        }
        const index = assignment.gradingBenchmarks.indexOf(benchmark);
        if (index > -1) {
            assignment.gradingBenchmarks.splice(index, 1);
        }
    }
};
export const setActiveEnrollment = ({ state }, enrollment) => {
    state.selectedEnrollment = enrollment;
};
export const startSubmissionStream = ({ actions, effects }) => {
    effects.global.streamService.submissionStream({
        onStatusChange: actions.global.setConnectionStatus,
        onMessage: actions.global.receiveSubmission,
        onError: actions.global.handleStreamError,
    });
};
export const updateAssignments = async ({ actions, effects }, courseID) => {
    const response = await effects.global.api.client.updateAssignments({ courseID });
    if (response.error) {
        return;
    }
    actions.global.alert({ text: "Assignments updated", color: Color.GREEN });
};
export const fetchUserData = async ({ state, actions }) => {
    const successful = await actions.global.getSelf();
    if (!successful) {
        state.isLoading = false;
        return false;
    }
    await actions.global.getEnrollmentsByUser();
    await actions.global.getAssignments();
    await actions.global.getCourses();
    const results = [];
    for (const enrollment of state.enrollments) {
        const courseID = enrollment.courseID;
        if (isStudent(enrollment) || isTeacher(enrollment)) {
            results.push(actions.global.getUserSubmissions(courseID));
            results.push(actions.global.getGroupSubmissions(courseID));
        }
    }
    await Promise.all(results);
    if (state.self.IsAdmin) {
        await actions.global.getUsers();
    }
    await actions.global.getRepositories();
    actions.global.startSubmissionStream();
    state.isLoading = false;
    return true;
};
export const changeView = async ({ state, effects }) => {
    const enrollment = state.enrollments.find(enrol => enrol.courseID === state.activeCourse);
    if (!enrollment) {
        return;
    }
    if (hasStudent(enrollment.status)) {
        const response = await effects.global.api.client.getEnrollments({
            FetchMode: {
                case: "userID",
                value: state.self.ID,
            },
            statuses: [Enrollment_UserStatus.TEACHER],
        });
        if (response.error) {
            return;
        }
        if (response.message.enrollments.find(enrol => enrol.courseID === state.activeCourse && hasTeacher(enrol.status))) {
            enrollment.status = Enrollment_UserStatus.TEACHER;
        }
    }
    else if (hasTeacher(enrollment.status)) {
        enrollment.status = Enrollment_UserStatus.STUDENT;
    }
};
export const loading = ({ state }) => {
    state.isLoading = !state.isLoading;
};
export const setQuery = ({ state }, query) => {
    state.query = query;
};
export const errorHandler = (context, { method, error }) => {
    if (!error) {
        return;
    }
    if (error.code === Code.Unauthenticated) {
        if (method === "GetUser") {
            return;
        }
        context.actions.global.alert({
            text: "Your session has expired. Please log in again.",
            color: Color.RED
        });
        localStorage.setItem("alert", "Your session has expired. Please log in again.");
    }
    else {
        const message = context.state.self.IsAdmin ? `${method}: ${error.message}` : error.rawMessage;
        context.actions.global.alert({
            text: message,
            color: Color.RED
        });
    }
};
export const alert = ({ state }, a) => {
    state.alerts.push({ id: newID(), ...a });
};
export const popAlert = ({ state }, alert) => {
    state.alerts = state.alerts.filter(a => a.id !== alert.id);
};
export const logout = ({ state }) => {
    state.self = create(UserSchema);
};
export const setAscending = ({ state }, ascending) => {
    state.sortAscending = ascending;
};
export const setSubmissionSort = ({ state }, sort) => {
    if (state.sortSubmissionsBy !== sort) {
        state.sortSubmissionsBy = sort;
    }
    else {
        state.sortAscending = !state.sortAscending;
    }
};
export const clearSubmissionFilter = ({ state }) => {
    state.submissionFilters = [];
};
export const setSubmissionFilter = ({ state }, filter) => {
    if (state.submissionFilters.includes(filter)) {
        state.submissionFilters = state.submissionFilters.filter(f => f !== filter);
    }
    else {
        state.submissionFilters.push(filter);
    }
};
export const setIndividualSubmissionsView = ({ state }, view) => {
    state.individualSubmissionView = view;
};
export const setGroupView = ({ state }, groupView) => {
    state.groupView = groupView;
};
export const setActiveGroup = ({ state }, group) => {
    if (group) {
        state.activeGroup = clone(GroupSchema, group);
    }
    else {
        state.activeGroup = null;
    }
};
export const updateGroupUsers = ({ state }, user) => {
    if (!state.activeGroup) {
        return;
    }
    const group = state.activeGroup;
    const index = group.users.findIndex(u => u.ID === user.ID);
    if (index >= 0) {
        group.users.splice(index, 1);
    }
    else {
        group.users.push(user);
    }
};
export const updateGroupName = ({ state }, name) => {
    if (!state.activeGroup) {
        return;
    }
    state.activeGroup.name = name;
};
export const setConnectionStatus = ({ state }, status) => {
    state.connectionStatus = status;
};
export const setSubmissionOwner = ({ state }, owner) => {
    if (isMessage(owner, GroupSchema)) {
        state.submissionOwner = { type: "GROUP", id: owner.ID };
    }
    else {
        const groupID = state.selectedSubmission?.groupID ?? 0n;
        if (groupID > 0) {
            state.submissionOwner = { type: "GROUP", id: groupID };
            return;
        }
        state.submissionOwner = { type: "ENROLLMENT", id: owner.ID };
    }
};
export const updateSubmissionOwner = ({ state }, owner) => {
    state.submissionOwner = owner;
};
