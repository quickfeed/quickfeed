import React, { useEffect, useState } from 'react';
import { useLocation, Link } from 'react-router-dom';
import { useAppState } from '../../overmind';


const Breadcrumbs = () => {
    const state = useAppState()
    const location = useLocation();
    const [courseName, setCourseName] = useState<string | null>(null);
    const [assignmentName, setAssignmentName] = useState<string | null>(null);
    const pathnames = location.pathname.split('/').filter(x => x);

    const getCourseNameById = async (id: string): Promise<string | null> => {
        const course = state.courses.find(course => course.ID.toString() === id);
        return course ? course.name : null
    };

    const getAssignmentNameById = async (id: string): Promise<string | null> => {
        if (pathnames[0] === 'course' && pathnames[1]) {
            const assignment = state.assignments[pathnames[1]].find(assignment => assignment.ID.toString() === id);
            console.log(assignment)
            return assignment ? assignment.name : null
        }
        return null
    }

    useEffect(() => {
        if (pathnames[0] === 'course' && pathnames[1]) {
            getCourseNameById(pathnames[1]).then(setCourseName);
        }
        if (pathnames[2] === 'lab' && pathnames[3]) {
            getAssignmentNameById(pathnames[3]).then(setAssignmentName);
        }
        console.log(pathnames)
    }, [pathnames]);

    return (
        <nav aria-label="breadcrumb">
            <ol className="breadcrumb m-0 bg-transparent">
                <li className="breadcrumb-item">
                    <Link to="/">Dashboard</Link>
                </li>
                {pathnames.map((value, index) => {
                    const last = index === pathnames.length - 1;
                    const to = `/${pathnames.slice(0, index + 1).join('/')}`;
                    // title case the path segment.
                    let breadcrumbName = decodeURIComponent(value.charAt(0).toUpperCase() + value.slice(1));

                    // skip the first path segment (e.g., 'course/ID').
                    if (index === 0 && value === 'course') {
                        return null;
                    }

                    // skip the second path segment (e.g., 'course/ID/lab/ID').
                    if (index === 2 && value === 'lab') {
                        return null;
                    }

                    // Replace 'course/ID' with 'course/Course Name' in the breadcrumb.
                    if (index === 1 && courseName && pathnames[0] === 'course') {
                        breadcrumbName = courseName;
                    }

                    // Replace 'lab/ID' with 'lab/Assignment Name' in the breadcrumb.
                    if (index === 3 && assignmentName && pathnames[2] === 'lab') {
                        breadcrumbName = assignmentName;
                    }

                    return last ? (
                        <li key={to} className="breadcrumb-item active" aria-current="page">
                            {breadcrumbName}
                        </li>
                    ) : (
                        <li key={to} className="breadcrumb-item">
                            <Link to={to}>{breadcrumbName}</Link>
                        </li>
                    );
                })}
            </ol>
        </nav>
    );
};

export default Breadcrumbs;
