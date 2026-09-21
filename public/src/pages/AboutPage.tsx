import React from "react"
import BackToTop from "../components/BackToTop"
import FeatureBlock, { MiniFeatureBlock } from "../components/FeatureBlock"
import BuildLogPreview from "../components/about/BuildLogPreview"
import LabResultPreview from "../components/about/LabResultPreview"
import OrganizationPreview from "../components/about/OrganizationPreview"
import PreviewFrame from "../components/about/PreviewFrame"
import ResultsPreview from "../components/about/ResultsPreview"
import WorkflowPreview from "../components/about/WorkflowPreview"

/* AboutPage displays information about QuickFeed. Mainly displayed to non-logged in users on the LoginPage.tsx. */
const AboutPage = () => {
    return (
        <div className="w-full">
            <div className="container mx-auto px-4 max-w-7xl pb-12">
                <section className="mt-12">
                    <p className="text-sm font-semibold uppercase tracking-wider text-primary mb-3">
                        About QuickFeed
                    </p>
                    <h2 className="text-4xl md:text-5xl font-bold text-base-content mb-6 max-w-3xl">
                        Automated feedback on programming assignments
                    </h2>
                    <p className="text-xl leading-relaxed text-base-content/80 max-w-3xl">
                        QuickFeed gives students fast feedback on their lab assignments,
                        and gives the teaching staff the test results to grade from.
                    </p>
                    <p className="mt-4 text-lg leading-relaxed text-base-content/70 max-w-3xl">
                        It builds on version control and continuous integration.
                        When students push code to their repositories, QuickFeed builds it and reports
                        the result against tests supplied by the teaching staff.
                    </p>
                    <WorkflowPreview />
                </section>

                <div className="divider my-12" />

                <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-12">
                    <MiniFeatureBlock
                        title="GitHub Integration"
                        content="Manage all students and courses on GitHub. Each student gets their own repository. Teachers get separate repositories for publishing assignments and information to students. All taken care of automatically."
                        media={iconMedia("fa-brands fa-github", "text-base-content")}
                    />
                    <MiniFeatureBlock
                        title="Continuous Integration"
                        content="Instantaneous feedback to students on how well their code performs. Students can quickly identify what they need to focus on to improve. All customizable for the teaching staff."
                        media={iconMedia("fas fa-arrows-rotate", "text-primary")}
                    />
                    <MiniFeatureBlock
                        title="Fair Grading"
                        content="On due date of an assignment, the most recent version of each student's code is available through GitHub. Easily accessible for the teachers. Together with latest build log, this makes grading easier and more fair."
                        media={iconMedia("fas fa-scale-balanced", "text-success")}
                    />
                </div>

                <div className="divider my-16" />

                <FeatureBlock
                    heading="QuickFeed"
                    subheading="Automated student feedback"
                    content="QuickFeed gives students fast feedback on their lab assignments, and teaches them the tools used in the industry along the way."
                    points={[
                        "Built on version control and continuous integration",
                        "Feedback comes from tests supplied by the teaching staff",
                        "The same test results are there for the staff when grading",
                    ]}
                    media={
                        <PreviewFrame label="A QuickFeed build log listing the tests that ran on a student push and their results.">
                            <BuildLogPreview />
                        </PreviewFrame>
                    }
                />

                <div className="divider my-16" />

                <FeatureBlock
                    heading="GitHub Integration"
                    subheading="Managing courses and students"
                    content="A course is an organization on GitHub, and QuickFeed creates and manages the repositories inside it."
                    points={[
                        "Every student gets a private repository of their own",
                        "Assignments and course information are published from staff repositories",
                        "Students learn to use git for version control while they submit",
                    ]}
                    media={
                        <PreviewFrame label="The repositories QuickFeed creates in a course organization on GitHub: info, assignments, tests, one repository per student and one per group.">
                            <OrganizationPreview />
                        </PreviewFrame>
                    }
                    reverse
                />

                <div className="divider my-16" />

                <FeatureBlock
                    heading="Continuous Integration"
                    subheading="Builds and tests student code"
                    content="A push to GitHub starts a build process defined by the teacher, which generates the feedback students see."
                    points={[
                        "Feedback appears on the student's course page as soon as the build finishes",
                        "Tests defined by either teachers or students are processed",
                        "Every run shows how far the student has come on the assignment",
                    ]}
                    media={
                        <PreviewFrame
                            label="A student's lab result: a progress bar at 92 percent, a table of lab information, and a table of test scores."
                            url="quickfeed.example.org/course/1/lab/1"
                        >
                            <LabResultPreview />
                        </PreviewFrame>
                    }
                />

                <div className="divider my-16" />

                <FeatureBlock
                    heading="Grading"
                    subheading="Easy and Fair"
                    content="On the due date, teachers grade from the recorded test results rather than from a fresh reading of every submission."
                    points={[
                        "See at a glance which tests passed for each student",
                        "The build log behind every score is one click away",
                        "The most recent version of the code is on GitHub, ready to inspect",
                    ]}
                    media={
                        <PreviewFrame
                            label="A teacher's course results table, with each student's lab scores colored by approval status."
                            url="quickfeed.example.org/course/1/results"
                        >
                            <ResultsPreview />
                        </PreviewFrame>
                    }
                    reverse
                />

                <BackToTop />
            </div>
        </div>
    )
}

function iconMedia(icon: string, color: string): React.JSX.Element {
    return <i className={`${icon} text-8xl ${color}`} aria-hidden="true" />
}

export default AboutPage
