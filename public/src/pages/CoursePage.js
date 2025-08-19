import React, { useLayoutEffect } from "react";
import { Navigate } from "react-router";
import { isEnrolled, isTeacher } from "../Helpers";
import { useActions, useAppState } from "../overmind";
import StudentPage from "./StudentPage";
import TeacherPage from "./TeacherPage";
import { useCourseID } from "../hooks/useCourseID";
const CoursePage = () => {
    const state = useAppState();
    const actions = useActions().global;
    const courseID = useCourseID();
    const enrollment = state.enrollmentsByCourseID[courseID.toString()];
    useLayoutEffect(() => {
        if (!state.showFavorites) {
            actions.toggleFavorites();
        }
        actions.setActiveCourse(courseID);
        actions.getCourseData({ courseID });
    }, [actions, courseID]);
    if (state.enrollmentsByCourseID[courseID.toString()] && isEnrolled(enrollment)) {
        if (isTeacher(enrollment)) {
            return React.createElement(TeacherPage, null);
        }
        return React.createElement(StudentPage, null);
    }
    else {
        return React.createElement(Navigate, { to: "/", replace: true });
    }
};
export default CoursePage;
