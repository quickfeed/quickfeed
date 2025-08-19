import { clone, isMessage } from "@bufbuild/protobuf";
import React, { useCallback, useEffect, useMemo, useRef } from "react";
import { useSearchParams } from 'react-router-dom';
import { EnrollmentSchema } from "../../proto/qf/types_pb";
import { Color, getSubmissionCellColor, Icon } from "../Helpers";
import { useCourseID } from "../hooks/useCourseID";
import { useActions, useAppState } from "../overmind";
import Button, { ButtonType } from "./admin/Button";
import { generateAssignmentsHeader, generateSubmissionRows } from "./ComponentsHelpers";
import DynamicTable from "./DynamicTable";
import TableSort from "./forms/TableSort";
import LabResult from "./LabResult";
import ReviewForm from "./manual-grading/ReviewForm";
import Release from "./Release";
import Search from "./Search";
const Results = ({ review }) => {
    const state = useAppState();
    const actions = useActions();
    const courseID = useCourseID();
    const [searchParams, setSearchParams] = useSearchParams();
    const members = useMemo(() => { return state.courseMembers; }, [state.courseMembers]);
    const assignments = useMemo(() => {
        return state.assignments[courseID.toString()]?.filter(a => state.review.assignmentID <= 0 || a.ID === state.review.assignmentID);
    }, [state.assignments, courseID, state.review.assignmentID]);
    const loaded = state.loadedCourse[courseID.toString()];
    const latest = useRef({ state, actions, searchParams });
    useEffect(() => {
        latest.current = { state, actions, searchParams };
    }, [state, actions, searchParams]);
    useEffect(() => {
        if (!state.loadedCourse[courseID.toString()]) {
            actions.global.loadCourseSubmissions(courseID);
        }
        return () => {
            actions.global.setGroupView(false);
            actions.review.setAssignmentID(-1n);
            actions.global.setActiveEnrollment(null);
            actions.global.setSelectedSubmission({ submission: null });
        };
    }, [actions, courseID, state.loadedCourse]);
    useEffect(() => {
        const { state, actions, searchParams } = latest.current;
        if (state.selectedSubmission) {
            return;
        }
        const selectedLab = searchParams.get("id");
        if (selectedLab) {
            const submission = state.submissionsForCourse.ByID(BigInt(selectedLab));
            if (submission) {
                actions.global.setSelectedSubmission({ submission });
                actions.global.updateSubmissionOwner(state.submissionsForCourse.OwnerByID(submission.ID));
                if (submission.reviews.length > 0) {
                    actions.review.setSelectedReview(-1);
                }
                else {
                    actions.global.getSubmission({ submission, owner: state.submissionOwner, courseID: state.activeCourse });
                }
            }
        }
    }, [loaded]);
    const groupView = state.groupView;
    const handleSetGroupView = useCallback(() => {
        actions.global.setGroupView(!groupView);
        actions.review.setAssignmentID(BigInt(-1));
    }, [actions, groupView]);
    const handleLabClick = useCallback((submission, owner) => {
        actions.global.setSelectedSubmission({ submission });
        if (isMessage(owner, EnrollmentSchema)) {
            actions.global.setActiveEnrollment(clone(EnrollmentSchema, owner));
        }
        actions.global.setSubmissionOwner(owner);
        setSearchParams({ id: submission.ID.toString() });
    }, [actions, setSearchParams]);
    const handleReviewCellClick = useCallback((submission, owner) => () => {
        handleLabClick(submission, owner);
        actions.review.setSelectedReview(-1);
    }, [actions, handleLabClick]);
    const handleSubmissionCellClick = useCallback((submission, owner) => () => {
        handleLabClick(submission, owner);
        actions.global.getSubmission({ submission, owner: state.submissionOwner, courseID: state.activeCourse });
    }, [actions, handleLabClick, state.activeCourse, state.submissionOwner]);
    if (!state.loadedCourse[courseID.toString()]) {
        return React.createElement("h1", null, "Fetching Submissions...");
    }
    const generateReviewCell = (submission, owner) => {
        if (!state.isManuallyGraded(submission)) {
            return { iconTitle: "auto graded", iconClassName: Icon.DASH, value: "" };
        }
        const reviews = state.review.reviews.get(submission.ID) ?? [];
        const pending = reviews.some((r) => !r.ready && r.ReviewerID === state.self.ID) ? "pending-review" : "";
        const isSelected = state.selectedSubmission?.ID === submission.ID ? "selected" : "";
        const score = reviews.reduce((acc, theReview) => acc + theReview.score, 0) / reviews.length;
        const willBeReleased = state.review.minimumScore > 0 && score >= state.review.minimumScore ? "release" : "";
        const numReviewers = state.assignments[state.activeCourse.toString()]?.find((a) => a.ID === submission.AssignmentID)?.reviewers ?? 0;
        return ({
            iconTitle: submission.released ? "Released" : "Not released",
            iconClassName: submission.released ? "fa fa-unlock" : "fa fa-lock",
            value: `${reviews.length}/${numReviewers}`,
            className: `${getSubmissionCellColor(submission, owner)} ${isSelected} ${willBeReleased} ${pending}`,
            onClick: handleReviewCellClick(submission, owner),
        });
    };
    const getSubmissionCell = (submission, owner) => {
        const isSelected = state.selectedSubmission?.ID === submission.ID ? "selected" : "";
        return ({
            value: `${submission.score} %`,
            className: `${getSubmissionCellColor(submission, owner)} ${isSelected}`,
            onClick: handleSubmissionCellClick(submission, owner),
        });
    };
    const header = generateAssignmentsHeader(assignments, groupView, actions, state.isCourseManuallyGraded);
    const generator = review ? generateReviewCell : getSubmissionCell;
    const rows = generateSubmissionRows(members, generator, state);
    const divWidth = state.review.assignmentID >= 0 ? "col-md-4" : "col-md-6";
    const displayMode = state.groupView ? "Group" : "Student";
    const buttonColor = state.groupView ? Color.BLUE : Color.GREEN;
    return (React.createElement("div", { className: "row" },
        React.createElement("div", { className: `p-0 ${divWidth}` },
            review ? React.createElement(Release, null) : null,
            React.createElement(Search, { placeholder: "Search by name ...", className: "mb-2" },
                React.createElement(Button, { text: `View by ${displayMode}`, color: buttonColor, type: ButtonType.BUTTON, className: "ml-2", onClick: handleSetGroupView })),
            React.createElement(TableSort, { review: review }),
            React.createElement(DynamicTable, { header: header, data: rows })),
        React.createElement("div", { className: "col" }, review ? React.createElement(ReviewForm, null) : React.createElement(LabResult, null))));
};
export default Results;
