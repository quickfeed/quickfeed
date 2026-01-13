import React, { JSX } from "react"
import FeatureBlock, { MiniFeatureBlock } from "../components/FeatureBlock"
import BackToTop from "../components/BackToTop"

/* AboutPage displays information about QuickFeed. Mainly displayed to non-logged in users on the LoginPage.tsx. */
const AboutPage = () => {
    return (
        <div>
            <div className="container">
                <h2 className="featurette-heading mt-5">About QuickFeed</h2>
                <p className="lead">
                    QuickFeed is a tool for providing automated feedback to students on their lab assignments.
                    QuickFeed builds upon version control systems and continuous integration.
                    When students upload code to their repositories, QuickFeed automatically builds their code and provides feedback based on tests supplied by the teaching staff.
                    When grading assignments, teaching staff can access the results of test execution and have a valuable tool in the grading process.
                </p>
                <hr className="loginDivider" />
                <div key="rowheader" className="row marketing">
                    <MiniFeatureBlock
                        title="GitHub Integration"
                        content="Manage all students and courses on GitHub. Each student gets their own repository. Teachers get separate repositories for publishing assignments and information to students. All taken care of automatically."
                        media={<i className="fa fa-github" style={{ fontSize: "140px" }} />}
                    />
                    <MiniFeatureBlock
                        title="Continuous Integration"
                        content="Instantaneous feedback to students on how well their code performs. Students can quickly identify what they need to focus on to improve. All customizable for the teaching staff."
                        media={imageMedia("/assets/img/overlapping-arrows-no-background.webp", "Continuous Integration")}
                    />
                    <MiniFeatureBlock
                        title="Fair Grading"
                        content="On due date of an assignment, the most recent version of each student&apos;s code is available through GitHub. Easily accessible for the teachers. Together with latest build log, this makes grading easier and more fair."
                        media={imageMedia("/assets/img/Aplus2-no-background.webp", "Fair Grading")}
                    />
                </div>
                <hr className="loginDivider" />
                <FeatureBlock
                    heading="QuickFeed"
                    subheading="Automated student feedback"
                    content="QuickFeed aims to provide students with fast feedback on their lab assignments, and is designed to help students learn about state-of-the-art tools used in the industry. QuickFeed builds upon version control systems and continuous integration. When students upload code to their repositories, QuickFeed automatically builds their code and provides feedback based on tests supplied by the teaching staff. When grading assignments, teaching staff can access the results of test execution and have a valuable tool in the grading process."
                    imageSrc="/assets/img/intro1.png"
                />
                <hr className="loginDivider" />
                <FeatureBlock
                    heading="GitHub Integration"
                    subheading="Managing courses and students"
                    content="A course is an organization on GitHub. Students get access to their own private GitHub repository. Uploading their code for review or grading, students can learn to use git for version control."
                    imageSrc="/assets/img/intro3.png"
                    reverse
                />
                <hr className="loginDivider" />
                <FeatureBlock
                    heading="Continuous Integration"
                    subheading="Builds and tests student code"
                    content="As code gets pushed up to GitHub, an automatic build process defined by the teacher, generates feedback to students. When the build process is completed, student gets immediate access to this feedback on their personal course page. Tests defined by either teachers or students will be processed and tell students about their progress on the assignments."
                    imageSrc="/assets/img/intro2.png"
                />
                <hr className="loginDivider" />
                <FeatureBlock
                    heading="Grading"
                    subheading="Easy and Fair"
                    content="On the due date, teachers can access the test results and use this as a tool in the grading process. The teaching staff will immediately know which of their tests passed, and how much of the code is covered by the tests."
                    imageSrc="/assets/img/intro4.png"
                    reverse
                />
                <BackToTop />
            </div>
        </div>

    )
}

function imageMedia(src: string, alt: string): JSX.Element {
    return (
        <img
            src={src}
            alt={alt}
            style={{ width: "100%", height: "100%", objectFit: "contain" }}
        />
    )
}
export default AboutPage
