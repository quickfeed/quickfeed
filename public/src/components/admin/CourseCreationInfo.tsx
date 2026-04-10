import React from "react"


// CourseCreationInfo is a component that displays information about the course creation process.
const CourseCreationInfo = () => {
    return (
        <div className="card bg-base-200 shadow-xl">
            <div className="card-body">
                <h1 className="card-title text-3xl mb-4">Create Course</h1>

                <p className="text-lg text-base-content/80 mb-6">
                    For each new semester of a course, QuickFeed requires a new GitHub organization.
                    This is to keep the student roster for the different runs of the course separate.
                </p>

                <div className="space-y-4 text-base-content/80">
                    <p>
                        <a
                            className="btn btn-success btn-sm mr-2"
                            href="https://github.com/organizations/plan"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            Create an organization
                        </a>
                        for your course. The course organization must allow private repositories.
                    </p>

                    <p>
                        Add the
                        <a
                            className="btn btn-info btn-sm mx-2"
                            href={process.env.QUICKFEED_APP_URL}
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            QuickFeed application
                        </a>
                        to your GitHub organization to create a course.
                    </p>

                    <div>
                        <p className="mb-2">QuickFeed will create the following repositories for you:</p>
                        <div className="flex gap-2 ml-4">
                            <code className="bg-base-300 px-2 py-0.5 rounded text-sm">info</code>
                            <code className="bg-base-300 px-2 py-0.5 rounded text-sm">assignments</code>
                            <code className="bg-base-300 px-2 py-0.5 rounded text-sm">tests</code>
                        </div>
                    </div>

                    <p>
                        Please refer to the
                        <a
                            className="btn btn-primary btn-sm mx-2"
                            href="https://github.com/quickfeed/quickfeed/blob/master/doc/teacher.md"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            documentation
                        </a>
                        for further instructions on how to work with the various repositories.
                    </p>

                    <div className="alert alert-info mt-6">
                        <span>
                            After you have installed the QuickFeed application, enter the name of the organization in the field below to find the created course.
                        </span>
                    </div>
                </div>
            </div>
        </div>
    )
}

export default CourseCreationInfo
