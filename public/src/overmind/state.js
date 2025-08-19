import { create, isMessage } from "@bufbuild/protobuf";
import { derived } from "overmind";
import { Enrollment_UserStatus, EnrollmentSchema, GroupSchema, UserSchema } from "../../proto/qf/types_pb";
import { ConnStatus, getNumApproved, getSubmissionsScore, isAllApproved, isManuallyGraded, isPending, isPendingGroup, isTeacher, SubmissionsForCourse, SubmissionsForUser, SubmissionSort } from "../Helpers";
export const state = {
    self: create(UserSchema),
    isLoggedIn: derived(({ self }) => {
        return Number(self.ID) !== 0;
    }),
    isValid: derived(({ self }) => {
        return self.Name.length > 0 && self.StudentID.length > 0 && self.Email.length > 0;
    }),
    enrollments: [],
    enrollmentsByCourseID: derived(({ enrollments }) => {
        const enrollmentsByCourseID = {};
        for (const enrollment of enrollments) {
            enrollmentsByCourseID[enrollment.courseID.toString()] = enrollment;
        }
        return enrollmentsByCourseID;
    }),
    submissions: new SubmissionsForUser(),
    userGroup: derived(({ enrollments }) => {
        const userGroup = {};
        for (const enrollment of enrollments) {
            if (enrollment.group) {
                userGroup[enrollment.courseID.toString()] = enrollment.group;
            }
        }
        return userGroup;
    }),
    isTeacher: derived(({ enrollmentsByCourseID, activeCourse }) => {
        if (activeCourse > 0 && enrollmentsByCourseID[activeCourse.toString()]) {
            return isTeacher(enrollmentsByCourseID[activeCourse.toString()]);
        }
        return false;
    }),
    isCourseCreator: derived(({ courses, activeCourse, self }) => {
        const course = courses.find(c => c.ID === activeCourse);
        return course !== undefined && course.courseCreatorID === self.ID;
    }),
    status: {},
    users: {},
    allUsers: [],
    courses: [],
    courseTeachers: derived(({ courseEnrollments, activeCourse }) => {
        if (!activeCourse || !courseEnrollments[activeCourse.toString()]) {
            return {};
        }
        const teachersMap = {};
        courseEnrollments[activeCourse.toString()].forEach(enrollment => {
            if (isTeacher(enrollment) && enrollment.user) {
                teachersMap[enrollment.userID.toString()] = enrollment.user;
            }
        });
        return teachersMap;
    }),
    courseMembers: derived(({ activeCourse, groupView, submissionsForCourse, assignments, groups, courseEnrollments, submissionFilters, sortAscending, sortSubmissionsBy }, { review: { assignmentID } }) => {
        if (!activeCourse) {
            return [];
        }
        const submissions = groupView
            ? submissionsForCourse.groupSubmissions
            : submissionsForCourse.userSubmissions;
        if (submissions.size === 0) {
            return [];
        }
        let numAssignments = 0;
        if (assignmentID > 0) {
            numAssignments = 1;
        }
        else if (groupView) {
            numAssignments = assignments[activeCourse.toString()]?.filter(a => a.isGroupLab).length || 0;
        }
        else {
            numAssignments = assignments[activeCourse.toString()]?.length ?? 0;
        }
        let filtered = groupView ? groups[activeCourse.toString()] : courseEnrollments[activeCourse.toString()] ?? [];
        for (const filter of submissionFilters) {
            switch (filter) {
                case "teachers":
                    filtered = filtered.filter(el => {
                        return el.status !== Enrollment_UserStatus.TEACHER;
                    });
                    break;
                case "approved":
                    filtered = filtered.filter(el => {
                        if (assignmentID > 0) {
                            const sub = submissions.get(el.ID)?.submissions?.find(s => s.AssignmentID === assignmentID);
                            return sub !== undefined && !isAllApproved(sub);
                        }
                        const numApproved = submissions.get(el.ID)?.submissions?.reduce((acc, cur) => {
                            return acc + ((cur &&
                                isAllApproved(cur)) ? 1 : 0);
                        }, 0) ?? 0;
                        return numApproved < numAssignments;
                    });
                    break;
                case "released":
                    filtered = filtered.filter(el => {
                        if (assignmentID > 0) {
                            const sub = submissions.get(el.ID)?.submissions?.find(s => s.AssignmentID === assignmentID);
                            return sub !== undefined && !sub.released;
                        }
                        const hasReleased = submissions.get(el.ID)?.submissions.some(sub => sub.released);
                        return !hasReleased;
                    });
                    break;
                default:
                    break;
            }
        }
        const sortOrder = sortAscending ? -1 : 1;
        const sortedSubmissions = Object.values(filtered).sort((a, b) => {
            let subA;
            let subB;
            if (assignmentID > 0) {
                subA = submissions.get(a.ID)?.submissions.find(sub => sub.AssignmentID === assignmentID);
                subB = submissions.get(b.ID)?.submissions.find(sub => sub.AssignmentID === assignmentID);
            }
            const subsA = submissions.get(a.ID)?.submissions;
            const subsB = submissions.get(b.ID)?.submissions;
            switch (sortSubmissionsBy) {
                case SubmissionSort.ID: {
                    if (isMessage(a, EnrollmentSchema) && isMessage(b, EnrollmentSchema)) {
                        return sortOrder * (Number(a.userID) - Number(b.userID));
                    }
                    else {
                        return sortOrder * (Number(a.ID) - Number(b.ID));
                    }
                }
                case SubmissionSort.Score: {
                    if (assignmentID > 0) {
                        const sA = subA?.score;
                        const sB = subB?.score;
                        if (sA !== undefined && sB !== undefined) {
                            return sortOrder * (sB - sA);
                        }
                        else if (sA !== undefined) {
                            return -sortOrder;
                        }
                        return sortOrder;
                    }
                    const aSubs = subsA ? getSubmissionsScore(subsA) : 0;
                    const bSubs = subsB ? getSubmissionsScore(subsB) : 0;
                    return sortOrder * (aSubs - bSubs);
                }
                case SubmissionSort.Approved: {
                    if (assignmentID > 0) {
                        const sA = subA && isAllApproved(subA) ? 1 : 0;
                        const sB = subB && isAllApproved(subB) ? 1 : 0;
                        return sortOrder * (sA - sB);
                    }
                    const aApproved = subsA ? getNumApproved(subsA) : 0;
                    const bApproved = subsB ? getNumApproved(subsB) : 0;
                    return sortOrder * (aApproved - bApproved);
                }
                case SubmissionSort.Name: {
                    let nameA = "";
                    let nameB = "";
                    if (!groupView && isMessage(a, EnrollmentSchema) && isMessage(b, EnrollmentSchema)) {
                        nameA = a.user?.Name ?? "";
                        nameB = b.user?.Name ?? "";
                    }
                    else if (groupView && isMessage(a, GroupSchema) && isMessage(b, GroupSchema)) {
                        nameA = a.name ?? "";
                        nameB = b.name ?? "";
                    }
                    return sortOrder * (nameA.localeCompare(nameB));
                }
                default:
                    return 0;
            }
        });
        return sortedSubmissions;
    }),
    selectedEnrollment: null,
    selectedSubmission: null,
    selectedAssignment: derived(({ activeCourse, selectedSubmission, assignments }) => {
        return assignments[activeCourse.toString()]?.find(a => a.ID === selectedSubmission?.AssignmentID) ?? null;
    }),
    assignments: {},
    repositories: {},
    courseGroup: { courseID: 0n, users: [], name: "" },
    alerts: [],
    isLoading: true,
    activeCourse: BigInt(-1),
    selectedAssignmentID: -1,
    courseEnrollments: {},
    groups: {},
    pendingGroups: derived(({ activeCourse, groups }) => {
        if (activeCourse > 0 && groups[activeCourse.toString()]) {
            return groups[activeCourse.toString()]?.filter(group => isPendingGroup(group));
        }
        return [];
    }),
    pendingEnrollments: derived(({ activeCourse, courseEnrollments }) => {
        if (activeCourse > 0 && courseEnrollments[activeCourse.toString()]) {
            return courseEnrollments[activeCourse.toString()].filter(enrollment => isPending(enrollment));
        }
        return [];
    }),
    numGroups: derived(({ groups, activeCourse }) => {
        if (activeCourse > 0 && groups[activeCourse.toString()]) {
            return groups[activeCourse.toString()]?.filter(group => !isPendingGroup(group)).length;
        }
        return 0;
    }),
    numEnrolled: derived(({ activeCourse, courseEnrollments }) => {
        if (activeCourse > 0 && courseEnrollments[activeCourse.toString()]) {
            return courseEnrollments[activeCourse.toString()]?.filter(enrollment => !isPending(enrollment)).length;
        }
        return 0;
    }),
    isCourseManuallyGraded: derived(({ activeCourse, assignments }) => {
        if (activeCourse > 0 && assignments[activeCourse.toString()]) {
            return assignments[activeCourse.toString()].some(a => isManuallyGraded(a.reviewers));
        }
        return false;
    }),
    query: "",
    sortSubmissionsBy: SubmissionSort.Approved,
    sortAscending: true,
    submissionFilters: [],
    individualSubmissionView: false,
    groupView: false,
    activeGroup: null,
    hasGroup: derived(({ userGroup }) => courseID => {
        return userGroup[courseID] !== undefined;
    }),
    showFavorites: false,
    connectionStatus: ConnStatus.DISCONNECTED,
    isManuallyGraded: derived(({ activeCourse, assignments }) => submission => {
        const assignment = assignments[activeCourse.toString()]?.find(a => a.ID === submission.AssignmentID);
        return assignment ? assignment.reviewers > 0 : false;
    }),
    getAssignmentsMap: derived(({ assignments }, { review: { assignmentID } }) => courseID => {
        const asgmts = assignments[courseID.toString()]?.filter(assignment => (assignmentID < 0) || assignment.ID === assignmentID) ?? [];
        const assignmentsMap = {};
        asgmts.forEach(assignment => {
            assignmentsMap[assignment.ID.toString()] = assignment.isGroupLab;
        });
        return assignmentsMap;
    }),
    submissionOwner: { type: "ENROLLMENT", id: 0n },
    loadedCourse: {},
    submissionsForCourse: new SubmissionsForCourse()
};
