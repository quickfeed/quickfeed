import * as React from "react";
import { ViewPage } from "./ViewPage";

export class HomePage extends ViewPage {
    constructor() {
        super();
        this.template = "frontpage";
    }

    public async renderContent(page: string): Promise<JSX.Element> {
        return <div className="container">

            <div className="row marketing">
                <div className="col-lg-4">
                    <img
                        className="img-circle"
                        src="/img/GitHub-Mark-120px-plus.png"
                        alt="GitHub logo" style={{ width: "140px", height: "140px" }} />
                    <h2>GitHub Integration</h2>
                    <p>
                        Manage all students and courses on GitHub.
                        Each student get their own repository.
                        Teachers get separate repositories for publishing assignments and information to students.
                        All taken care of automatically.
                    </p>
                    <p>
                        <a className="btn btn-default" href="#versioncontrol" role="button">View details »</a>
                    </p>
                </div>
                <div className="col-lg-4">

                    <img
                        className="img-circle"
                        src="/img/overlapping-arrows.png"
                        alt=""
                        style={{ width: "140px", height: "140px" }}
                    />
                    <h2>Continuous Integration</h2>
                    <p>
                        Instantaneous feedback to students on how well their code preforms.
                        Students can quickly identify what they need to focus on to improve.
                        All customizable for the teaching staff.
                    </p>
                    <p>
                        <a className="btn btn-default" href="#ci" role="button">
                            View details »
                        </a>
                    </p>
                </div>
                <div className="col-lg-4">
                    <img
                        className="img-circle"
                        src="/img/Aplus2.png"
                        alt="A+ image" style={{ width: "140px", height: "140px" }} />
                    <h2>Fair Grading</h2>
                    <p>
                        On due date of an assignment, the most recent version
                        of each student's code is available through GitHub.
                        Easily accessible for the teachers.
                        Together with latest build log, this makes grading easier and more fair.
                    </p>
                    <p>
                        <a className="btn btn-default" href="#grading" role="button">View details »</a>
                    </p>
                </div>
            </div>
            <section id="autograder">
                <hr className="featurette-divider" />
                <div className="row featurette">
                    <div className="col-md-7">
                        <h2 className="featurette-heading">
                            Autograder: <span className="text-muted">Automated student feedback</span>
                        </h2>
                        <p className="lead">
                            Autograder aims to provide students with fast
                            feedback on their lab assignments, and is designed
                            to help students learn about state-of-the-art tools
                            used in the industry.
                            Autograder builds upon version control systems and
                            continuous integration.
                            When students upload code to their repositories,
                            Autograder automatically builds their code and
                            provides feedback based on tests supplied by the
                            teaching staff.
                            When grading assignments, teaching staff can access
                            the results of test execution and have a valuable
                            tool in the grading process.
                        </p>
                    </div>
                    <div className="col-md-5">
                        <img
                            className="featurette-image img-responsive"
                            src="/img/intro1.png"
                            alt="Generic placeholder image" />
                    </div>
                </div>
            </section>

            <section id="versioncontrol">

                <hr className="featurette-divider" />
                <div className="row featurette">
                    <div className="col-md-5">
                        <img
                            className="featurette-image img-responsive"
                            src="/img/intro3.png"
                            alt="Generic placeholder image" />
                    </div>
                    <div className="col-md-7">
                        <h2 className="featurette-heading">
                            GitHub Integration: <span className="text-muted">Managing courses and students</span>
                        </h2>
                        <p className="lead">
                            A course is an organization on GitHub.
                            Students get access to their own private GitHub repository.
                            Using this repository students can learn to use git for
                            version control and uploading their code for review or grading.
                        </p>
                    </div>
                </div>
            </section>

            <section id="ci">

                <hr className="featurette-divider" />
                <div className="row featurette">
                    <div className="col-md-7">
                        <h2 className="featurette-heading">
                            Continuous Integration: <span className="text-muted">builds and tests student code. </span>
                        </h2>
                        <p className="lead">
                            As code gets pushed up to GitHub, an automatic build process
                            defined by the teacher, generates feedback to students.
                            When the build process is completed, students immediately
                            gets access to this feedback on their personal course page.
                            Tests defined by either teachers or students will be processed
                            and tell students about their progress on the assignments.
                        </p>
                    </div>
                    <div className="col-md-5">
                        <img
                            className="featurette-image img-responsive"
                            src="/img/intro4.png"
                            alt="Generic placeholder image" />
                    </div>
                </div>
            </section>

            <section id="grading">

                <hr className="featurette-divider" />
                <div className="row featurette">
                    <div className="col-md-5">
                        <img
                            className="featurette-image img-responsive"
                            src="/img/intro2.png"
                            alt="Generic placeholder image" />
                    </div>
                    <div className="col-md-7">
                        <h2 className="featurette-heading">
                            Grading: <span className="text-muted">Easy and Fair</span>
                        </h2>
                        <p className="lead">
                            On the due date, teachers can access the test results and
                            use this as a tool in the grading process.
                            The teaching staff will immediately know which of their
                            tests passed, and how much of the code is covered by the tests.
                        </p>
                    </div>
                </div>
            </section>

            <footer>
                <hr />
                <p className="pull-right"><a href="#">Back to top</a></p>
            </footer>
        </div>;
    }

    public async renderMenu(index: number): Promise<JSX.Element[]> {
        if (index === 1) {
            return [<div id="0" className="jumbotron">
                <div className="centerblock container">
                    <h1>Automated student feedback</h1>
                    <p>
                        <strong>Autograder </strong>
                        provides instantaneous feedback to students on
                        their programming assignments. It is also a
                        valuable tool for teachers when grading lab
                        assignments.
                    </p>
                    <p>
                        <a className="btn btn-primary btn-lg" href="#autograder" role="button">Learn more »</a>
                    </p>
                </div>
            </div>];
        }
        return [];
    }
}
