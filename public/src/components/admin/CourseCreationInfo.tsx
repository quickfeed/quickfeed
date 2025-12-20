import React from "react"


// CourseCreationInfo is a component that displays information about the course creation process.
const CourseCreationInfo = () => {
    return (
        <div className="jumbotron">
            <h1 className="display-4">Create Course</h1>
            <p className="lead">
                For each new semester of a course, QuickFeed requires a new GitHub organization.
                This is to keep the student roster for the different runs of the course separate.
            </p>
            <p>
                <a className="badge-pill badge-success" href="https://github.com/organizations/plan" target="_blank" rel="noopener noreferrer">
                    Create an organization
                </a> for your course.
                The course organization must allow private repositories.
            </p>
            <p>
                Add the <a className="badge-pill badge-info" href={process.env.QUICKFEED_APP_URL} target="_blank" rel="noopener noreferrer">QuickFeed application</a> to your GitHub organization to create a course.
            </p>
            <p>
                QuickFeed will create the following repositories for you:
            </p>
            <ul>
                <li>info</li>
                <li>assignments</li>
                <li>tests</li>
            </ul>
            <p>
                <span>Please refer to the </span>
                <a className="badge-pill badge-primary" href="https://github.com/quickfeed/quickfeed/blob/master/doc/teacher.md" target="_blank" rel="noopener noreferrer">
                    documentation
                </a>
                <span> for further instructions on how to work with the various repositories.</span>
            </p>
            <p>
                <span>After you have installed the QuickFeed application, enter the name of the organization in the field below to find the created course.</span>
            </p>
        </div>
    )
}

export default CourseCreationInfo
