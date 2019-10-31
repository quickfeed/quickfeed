import * as React from "react";
import { Row } from "../../components";

export class HelpView extends React.Component<{}, {}> {
    public render() {
        return (
            <Row className="container-fluid">
                <div className="col-md-2 col-sm-3 col-xs-12">
                    <div className="list-group">
                        <a href="#" className="list-group-item disabled">Help</a>
                        <a href="#autograder" className="list-group-item">Autograder</a>
                        <a href="#reg" className="list-group-item">Registration</a>
                        <a href="#signup" className="list-group-item">Sign up for a course</a>
                        <a href="#groups" className="list-group-item">Creating a group</a>

                    </div>
                </div>
                <div className="col-md-8 col-sm-9 col-xs-12">
                    <article>
                        <h1 id="autograder">Autograder</h1>
                        <p>
                            Autograder is a tool for students and teaching staff for
                            submitting and validating lab assignments and is developed at
                            the University of Stavanger. All lab submissions from students
                            are handled using Git, a source code management system, and
                            GitHub, a web-based hosting service for Git source repositories.
                        </p>
                        <p>
                            Students push their lab solutions to GitHub. Every
                             submission is then processed by a custom continuous
                            integration tool. This tool will run several test cases on the
                            submitted code. Autograder generates feedback that lets the
                            students verify that their submission implements the required
                            functionality. This feedback is available through a web interface.
                            The feedback from the Autograder system can be used by students
                            to improve their submissions.
                        </p>
                        <p>
                            Below is a step-by-step explanation on how to register and sign up
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
                                    Approve that the Autograder application may access the
                                    requested parts of your account. It is possible to make a
                                     separate GitHub account for Autograder if you do not
                                      want the application to access your personal one with
                                    the requested permissions.
                                    </p>
                            </li>
                            <li>
                                <p>
                                    Fill in your information and register as a new user.
                                     Please provide your real name in the "Name" field.
                                     This name will be used to approve your submissions
                                     to course assignments during the lab hours.
                                </p>
                            </li>
                        </ol>

                        <h1 id="signup">Signing up for a course</h1>

                        <ol>

                            <li>
                                <p>Click the course menu item.</p>
                            </li>
                            <li>
                                <p>
                                    In the course menu choose “Join Course”. Available courses will be listed.
                                </p>
                            </li>
                            <li>
                                <p>Find the course you are signing up for and click "Enroll".</p>
                            </li>
                            <li>
                                <p>
                                 Wait for the teaching staff to accept your enrollment request.
                                </p>
                            </li>
                            <li>
                                <p>
                                    After your enrollment request has been accepted,
                                     you will receive three GitHub invitations. Invitations will be sent
                                     to the email associated with the GitHub account you've signed up
                                      for the course with.
                                </p>
                            </li>
                            <li>
                                <p>
                                    You will get your own repository in the course's GitHub organization.
                                     You will also have access
                                    to the feedback pages for this course on Autograder.
                                </p>
                            </li>

                        </ol>

                        <h1 id="groups">Create a new group</h1>

                        <ol>

                            <li>
                                <p>
                                    Choose "Members" tab in the course menu on the left.
                                    Select yourself and other members of your group
                                    from the list of course members. Enter your group's
                                    desired name and click "Create".
                                </p>
                            </li>
                            <li>
                                <p>
                                The group is created after it has been approved
                                    by the course teacher. After that the group name
                                    cannot be changed. Please contact teaching staff if
                                    you wish to add or remove group members.
                                </p>
                            </li>
                        </ol>


                    </article>
                </div>
            </Row>
        );
    }
}
