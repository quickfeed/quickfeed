import * as React from "react";
import { Row } from "../../components";

class HelpView extends React.Component<any, undefined> {
    public render() {
        return (

            <Row className="container-fluid">
                <div className="col-md-2 col-sm-3 col-xs-12">
                    <div className="list-group">
                        <a href="#" className="list-group-item disabled">Help</a>
                        <a href="#autograder" className="list-group-item">Autograder</a>
                        <a href="#reg" className="list-group-item">Registration</a>
                        <a href="#signup" className="list-group-item">Sign up for a course</a>

                    </div>
                </div>
                <div className="col-md-8 col-sm-9 col-xs-12">
                    <article>
                        <h1 id="autograder">Autograder</h1>
                        <p>
                            Autograder is a new tool for students and teaching staff for
                            submitting and validating lab assignments and is developed at
                            the University of Stavanger. All lab submissions from students
                            are handled using Git, a source code management system, and
                            GitHub, a web-based hosting service for Git source repositories.
                        </p>
                        <p>
                            Students push their updated lab submissions to GitHub. Every
                            lab submission is then processed by a custom continuous
                            integration tool. This tool will run several test cases on the
                            submitted code. Autograder generates feedback that let the
                            students verify if their submission implements the required
                            functionality. This feedback is available through a web interface.
                            The feedback from the Autograder system can be used by students
                            to improve their submissions.
                        </p>
                        <p>
                            Below is a step-by-step explanation of how to register and sign up
                            for the lab project in Autograder.
                        </p>

                        <h1 id="reg">Registration</h1>

                        <ol>
                            <li>
                                <p>
                                    Go to <a href="http://github.com">GitHub</a> and register.
                                    A GitHub account is required to sign in to Autograder.
                                    You can skip this step if you already have an account.</p>
                            </li>
                            <li>
                                <p>
                                    Click the "Sign in with GitHub" button to register.
                                    You will then be taken to GitHub's website.
                                </p>
                            </li>
                            <li>
                                <p>
                                    Approve that our Autograder application may have permission
                                    to access to the requested parts of your account. It is
                                    possible to make a separate GitHub account for system if
                                    you do not want Autograder to access your personal one with
                                    the requested permissions.</p>
                            </li>
                        </ol>

                        <h1 id="signup">Signing up for a course</h1>

                        <ol>

                            <li>
                                <p>Click the course menu item.</p>
                            </li>
                            <li>
                                <p>
                                    In the course menu click on “New Course”. Available courses will be listed.
                                </p>
                            </li>
                            <li>
                                <p>Find the course you are signing up for and click sign up.</p>
                            </li>
                            <li>
                                <p>
                                    Read through and accept the terms. You will then be invited
                                    to the course organization on GitHub.
                                </p>
                            </li>
                            <li>
                                <p>
                                    An invitation will be sent to your email address registered with
                                    GitHub account. Accept the invitation using the received email.
                                </p>
                            </li>
                            <li>
                                <p>Wait for the teaching staff to verify your Autograder-registration.</p>
                            </li>
                            <li>
                                <p>
                                    You will get your own repository in the organization "uis-dat520" on
                                    GitHub after your registration is verified. You will also have access
                                    to the feedback pages for this course on Autograder.
                                </p>
                            </li>

                        </ol>
                    </article>

                </div>
            </Row>
        );
    }
}

export { HelpView };
