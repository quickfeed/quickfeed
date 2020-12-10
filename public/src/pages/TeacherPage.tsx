import * as React from "react";

import { CourseGroup, GroupForm, Results } from "../components";
import { CourseManager, ILink, ILinkCollection, NavigationManager, UserManager } from "../managers";
import { View, ViewPage } from "./ViewPage";

import { INavInfo } from "../NavigationHelper";
import { Assignment, Course, Enrollment, Group, Repository, GradingBenchmark, GradingCriterion, SubmissionsForCourseRequest, Review } from "../../proto/ag_pb";
import { CollapsableNavMenu } from "../components/navigation/CollapsableNavMenu";
import { GroupResults } from "../components/teacher/GroupResults";
import { MemberView } from "./views/MemberView";
import { showLoader } from "../loader";
import { sortCoursesByVisibility, sortAssignmentsByOrder, submissionStatusToString } from "../componentHelper";
import { AssignmentView } from "./views/AssignmentView";
import { ISubmission } from "../models";
import { FeedbackView } from "./views/FeedbackView";
import { ReleaseView } from "./views/ReleaseView";

export class TeacherPage extends ViewPage {

    private navMan: NavigationManager;
    private userMan: UserManager;
    private courseMan: CourseManager;
    private courses: Course[] = [];
    private repositories: Map<number, Map<Repository.Type, string>>;

    private refreshState = 0;

    constructor(userMan: UserManager, navMan: NavigationManager, courseMan: CourseManager) {
        super();

        this.navMan = navMan;
        this.userMan = userMan;
        this.courseMan = courseMan;
        this.navHelper.defaultPage = "course";
        this.navHelper.checkAuthentication = () => this.checkAuthentication();

        this.navHelper.registerFunction("courses/{course}", this.course);
        this.navHelper.registerFunction("courses/{course}/members", this.courseUsers);
        this.navHelper.registerFunction("courses/{course}/results", this.results);
        this.navHelper.registerFunction("courses/{course}/groupresults", this.groupresults);
        this.navHelper.registerFunction("courses/{course}/review", this.manualReview);
        this.navHelper.registerFunction("courses/{course}/release", this.releaseReview);
        this.navHelper.registerFunction("courses/{course}/groups", this.groups);
        this.navHelper.registerFunction("courses/{cid}/new_group", this.newGroup);
        this.navHelper.registerFunction("courses/{cid}/groups/{gid}/edit", this.editGroup);

    }

    public checkAuthentication(): boolean {
        const curUser = this.userMan.getCurrentUser();
        if (curUser?.getIsadmin() || this.userMan.isTeacher()) {
            this.userMan.isAuthorizedTeacher().then((answer) => {
                if (!answer) {
                    window.location.href = "https://" + window.location.hostname + "/auth/github-teacher";
                }
            });
            return true;
        }
        return false;
    }

    public async init(): Promise<void> {
        this.courses = await this.getCourses([]);
        this.repositories = this.setupRepos();
        this.navHelper.defaultPage = "courses/";
    }

    public async course(info: INavInfo<{ course: string, page?: string }>): View {
        return this.courseFunc(info.params.course, async (course) => {
            if (info.params.page) {
                return <h3>You are now on page {info.params.page.toUpperCase()} in course {info.params.course}</h3>;
            }
            let button;
            switch (this.refreshState) {
                case 0:
                    button = <div
                        className="btn btn-primary a-button"
                        onClick={(e) => {
                            this.refreshState = 1;
                            this.courseMan.updateAssignments(course.getId())
                                .then(() => {
                                    this.refreshState = 2;
                                    this.navMan.refresh();
                                });
                            this.navMan.refresh();
                        }}>
                        Update Course Assignments
                    </div>;
                    break;
                case 1:
                    button = <div
                        className="btn btn-default a-button disabled">
                        Updating Course Assignments
                    </div>;
                    break;
                case 2:
                    button = <div
                        className="btn btn-success a-button"
                        onClick={(e) => {
                            this.refreshState = 1;
                            this.courseMan.updateAssignments(course.getId())
                                .then(() => {
                                    this.refreshState = 2;
                                    this.navMan.refresh();
                                });
                            this.navMan.refresh();
                        }}>
                        Course Assignments Updated
                    </div>;
                    break;
            }
            return <div key="head" className="col-md-12">
                <div className="row"><h1>Assignments for {course.getName()}{button}</h1></div>
                {await this.generateAssignmentList(course)}
            </div>;
        });
    }

    public async results(info: INavInfo<{ course: string }>): View {
        return this.courseFunc(info.params.course, async (course) => {
            const assignments: Assignment[] = await this.courseMan.getAssignments(course.getId());
            const results = await this.courseMan.getSubmissionsByCourse(course.getId(), SubmissionsForCourseRequest.Type.ALL);
            const labResults = await this.courseMan.fillLabLinks(course, results, assignments);
            const curUser = this.userMan.getCurrentUser();
            return <Results
                course={course}
                courseURL={await this.getCourseURL(course.getId())}
                assignments={sortAssignmentsByOrder(assignments)}
                allCourseSubmissions={labResults}
                courseCreatorView={course.getCoursecreatorid() === curUser?.getId()}
                onSubmissionRebuild={async (assignmentID: number, submissionID: number) => {
                    const ans = await this.courseMan.rebuildSubmission(assignmentID, submissionID);
                    this.navMan.refresh();
                    return ans;
                }}
                onSubmissionStatusUpdate={async (submission: ISubmission): Promise<boolean> => {
                    return this.approveFunc(submission, course.getId());
                }}>
            </Results>;
        });
    }

    public async groupresults(info: INavInfo<{ course: string }>): View {
        return this.courseFunc(info.params.course, async (course) => {
            const results = await this.courseMan.getSubmissionsByCourse(course.getId(), SubmissionsForCourseRequest.Type.GROUP);
            const labs = await this.courseMan.getAssignments(course.getId());
            const labResults = await this.courseMan.fillLabLinks(course, results, labs);
            return <GroupResults
                course={course}
                courseURL={await this.getCourseURL(course.getId())}
                assignments={sortAssignmentsByOrder(labs)}
                allGroupSubmissions={labResults}
                onSubmissionRebuild={async (assignmentID: number, submissionID: number) => {
                    const ans = await this.courseMan.rebuildSubmission(assignmentID, submissionID);
                    this.navMan.refresh();
                    return ans;
                }}
                onSubmissionStatusUpdate={async (submission: ISubmission): Promise<boolean> => {
                    return this.approveFunc(submission, course.getId());
                }}>
            </GroupResults>;
        });
    }

    public async manualReview(info: INavInfo<{ course: string }>): View {
        return this.courseFunc(info.params.course, async (course) => {
            const assignments = await this.courseMan.getAssignments(course.getId());
            const students = await this.courseMan.getSubmissionsByCourse(course.getId(), SubmissionsForCourseRequest.Type.INDIVIDUAL);
            const groups = await this.courseMan.getSubmissionsByCourse(course.getId(), SubmissionsForCourseRequest.Type.GROUP);
            const curUser = this.userMan.getCurrentUser();
            if (curUser) {
                return <FeedbackView
                    course={course}
                    courseURL={await this.getCourseURL(course.getId())}
                    assignments={assignments}
                    students={students}
                    groups={groups}
                    curUser={curUser}
                    addReview={(r: Review) => {
                        return this.courseMan.addReview(r, course.getId());
                    }}
                    updateReview={async (r: Review) => {
                        return this.courseMan.editReview(r, course.getId());
                    }}
                />;
            }
            return <div>Please log in.</div>;
        })
    }

    public async releaseReview(info: INavInfo<{ course: string }>): View {
        return this.courseFunc(info.params.course, async (course) => {
            const assignments = await this.courseMan.getAssignments(course.getId());
            const students = await this.courseMan.getSubmissionsByCourse(course.getId(), SubmissionsForCourseRequest.Type.INDIVIDUAL);
            const groups = await this.courseMan.getSubmissionsByCourse(course.getId(), SubmissionsForCourseRequest.Type.GROUP);
            const curUser = this.userMan.getCurrentUser();
            if (curUser) {
                return <ReleaseView
                    course={course}
                    courseURL={await this.getCourseURL(course.getId())}
                    assignments={assignments}
                    students={students}
                    groups={groups}
                    curUser={curUser}
                    onUpdate={(submission: ISubmission) => {
                        return this.courseMan.updateSubmission(course.getId(), submission);
                    }}
                    getReviewers={(submissionID: number) => {
                        return this.courseMan.getReviewers(submissionID, course.getId());
                    }}
                    updateAll={async (assignmentID: number, score: number, release: boolean, approve: boolean) => {
                        return this.courseMan.updateSubmissions(assignmentID, course.getId(), score, release, approve);
                    }}
                />;
            }
            return <div>Please log in.</div>;
        })
    }

    public async groups(info: INavInfo<{ course: string }>): View {
        return this.courseFunc(info.params.course, async (course) => {
            const groups = await this.courseMan.getGroupsForCourse(course.getId());
            const approvedGroups: Group[] = [];
            const pendingGroups: Group[] = [];
            for (const grp of groups) {
                switch (grp.getStatus()) {
                    case Group.GroupStatus.APPROVED:
                        approvedGroups.push(grp);
                        break;
                    case Group.GroupStatus.PENDING:
                        pendingGroups.push(grp);
                        break;
                }
            }
            return <CourseGroup
                approvedGroups={approvedGroups}
                pendingGroups={pendingGroups}
                course={course}
                courseURL={await this.getCourseURL(course.getId())}
                navMan={this.navMan}
                courseMan={this.courseMan}
                pagePath={this.pagePath}
            />;
        });
    }

    public async newGroup(info: INavInfo<{ cid: number }>): View {
        const courseId = info.params.cid;
        const course = await this.courseMan.getCourse(courseId);
        const curUser = this.userMan.getCurrentUser();

        if (course && curUser) {
            // get full list of students and teachers
            const students = await this.courseMan.getUsersForCourse(
                course, false, false, [Enrollment.UserStatus.STUDENT, Enrollment.UserStatus.TEACHER]);
            // get list of users who are not in group
            const freeStudents = await this.courseMan.getUsersForCourse(
                course, true, false, [Enrollment.UserStatus.STUDENT, Enrollment.UserStatus.TEACHER]);
            return <GroupForm
                className="form-horizontal"
                students={students}
                freeStudents={freeStudents}
                course={course}
                curUser={curUser}
                courseMan={this.courseMan}
                userMan={this.userMan}
                navMan={this.navMan}
                pagePath={this.pagePath}
            />;
        }
        return showLoader();
    }

    public async editGroup(info: INavInfo<{ cid: string, gid: string }>): View {
        const courseId = parseInt(info.params.cid, 10);
        const groupId = parseInt(info.params.gid, 10);

        const course = await this.courseMan.getCourse(courseId);
        const curUser = this.userMan.getCurrentUser();
        const group: Group | null = await this.courseMan.getGroup(groupId);
        if (course && curUser && group) {
            // get full list of students and teachers
            const students = await this.courseMan.getUsersForCourse(
                course, false, false, [Enrollment.UserStatus.STUDENT, Enrollment.UserStatus.TEACHER]);
            // get list of users who are not in group
            const freeStudents = await this.courseMan.getUsersForCourse(
                course, true, false, [Enrollment.UserStatus.STUDENT, Enrollment.UserStatus.TEACHER]);
            return <GroupForm
                className="form-horizontal"
                students={students}
                freeStudents={freeStudents}
                course={course}
                curUser={curUser}
                courseMan={this.courseMan}
                userMan={this.userMan}
                navMan={this.navMan}
                pagePath={this.pagePath}
                groupData={group}
            />;
        }
        return showLoader();
    }

    public async courseUsers(info: INavInfo<{ course: string }>): View {
        return this.courseFunc(info.params.course, async (course) => {
            const all = await this.courseMan.getUsersForCourse(course, false, true);
            const assignments = await this.courseMan.getAssignments(course.getId())
            const acceptedUsers: Enrollment[] = [];
            const pendingUsers: Enrollment[] = [];
            // TODO: Maybe move this to the Members view
            all.forEach((user) => {
                switch (user.getStatus()) {
                    case Enrollment.UserStatus.TEACHER:
                    case Enrollment.UserStatus.STUDENT:
                        acceptedUsers.push(user);
                        break;
                    case Enrollment.UserStatus.PENDING:
                        pendingUsers.push(user);
                        break;
                }
            });

            return <MemberView
                course={course}
                assignments={assignments}
                courseURL={await this.getCourseURL(course.getId())}
                navMan={this.navMan}
                pendingUsers={pendingUsers}
                acceptedUsers={acceptedUsers}
                courseMan={this.courseMan}
            >
            </MemberView>;
        });
    }

    public generateCollectionFor(link: ILink, courseID: number): ILinkCollection {
        const repoMap = this.repositories.get(courseID);
        return {
            item: link,
            children: [
                { name: "Results", uri: link.uri + "/results" },
                { name: "Group Results", uri: link.uri + "/groupresults", extra: "disabled" },
                { name: "Review", uri: link.uri + "/review" },
                { name: "Release", uri: link.uri + "/release" },
                { name: "Groups", uri: link.uri + "/groups" },
                { name: "Members", uri: link.uri + "/members" },
                { name: "New Group", uri: link.uri + "/new_group" },
                { name: "Repositories" },
                { name: "Course Info", uri: repoMap?.get(Repository.Type.COURSEINFO) ?? "", absolute: true },
                { name: "Assignments", uri: repoMap?.get(Repository.Type.ASSIGNMENTS) ?? "", absolute: true },
                { name: "Tests", uri: repoMap?.get(Repository.Type.TESTS) ?? "", absolute: true },
            ],
        };
    }

    public async approveFunc(submission: ISubmission, courseID: number): Promise<boolean> {
        if (confirm(
            `Do you want to set ${submissionStatusToString(submission.status)} status for this lab?`,
        )) {
            const ans = await this.courseMan.updateSubmission(courseID, submission);
            this.navMan.refresh();
            return ans;
        }
        return false;
    }

    public async renderMenu(menu: number): Promise<JSX.Element[]> {
        const curUser = this.userMan.getCurrentUser();
        const confirmedTeacher = await this.userMan.isTeacher();
        if (curUser) {
            if (menu === 0) {
                const status = [Enrollment.UserStatus.TEACHER];
                if (curUser.getIsadmin() || confirmedTeacher) {
                    status.push(Enrollment.UserStatus.PENDING);
                    status.push(Enrollment.UserStatus.STUDENT);
                }
                const courses = await this.getCourses(status);
                const labLinks: ILinkCollection[] = [];
                courses.forEach((e) => {
                    this.fetchCourseRepos(e.getId());
                    labLinks.push(this.generateCollectionFor({
                        name: e.getCode(),
                        uri: this.pagePath + "/courses/" + e.getId(),
                    }, e.getId()));
                });

                this.navMan.checkLinkCollection(labLinks, this);

                return [
                    <h4 key={0}>Courses</h4>,
                    <CollapsableNavMenu
                        key={1}
                        links={labLinks} onClick={(link) => this.handleClick(link)}>
                    </CollapsableNavMenu>,
                ];
            }
        }
        return [];
    }

    private handleClick(link: ILink) {
        if (link.uri) {
            this.navMan.navigateTo(link.uri);
        }
    }

    private async getCourses(statuses: Enrollment.UserStatus[]): Promise<Course[]> {
        const curUsr = this.userMan.getCurrentUser();
        if (curUsr) {
            const enrols = await this.courseMan.getEnrollmentsForUser(curUsr.getId(), statuses);
            return sortCoursesByVisibility(enrols);
        }
        return [];
    }

    private async getCourseURL(courseID: number): Promise<string> {
        const repoMap = this.repositories.get(courseID);
        return repoMap?.get(Repository.Type.COURSEINFO)?.split("course-info")[0] ?? "";
    }

    private async courseFunc(courseParam: string, fn: (course: Course) => View): View {
        const courseID = parseInt(courseParam, 10);
        const course = this.courses.find(c => c.getId() === courseID) ?? await this.courseMan.getCourse(courseID);
        if (course) {
            return fn(course);
        }
        return showLoader();
    }

    private async fetchCourseRepos(courseID: number): Promise<Map<Repository.Type, string>> {
        return this.courseMan.getRepositories(courseID,
            [Repository.Type.COURSEINFO,
            Repository.Type.ASSIGNMENTS,
            Repository.Type.TESTS]);
    }

    private setupRepos(): Map<number, Map<Repository.Type, string>> {
        const allRepoMap = new Map<number, Map<Repository.Type, string>>();
        this.courses.forEach(async (crs) => {
            const repoMap = await this.fetchCourseRepos(crs.getId());
            allRepoMap.set(crs.getId(), repoMap);
        });
        return allRepoMap;
    }

    private async generateAssignmentList(course: Course): Promise<JSX.Element> {
        const assignments: Assignment[] = await this.courseMan.getAssignments(course.getId());

        return <div className="row">{
            sortAssignmentsByOrder(assignments).map((a, i) => <AssignmentView
                key={i}
                assignment={a}
                updateBenchmark={(bm: GradingBenchmark) => {
                    return this.courseMan.updateBenchmark(bm);
                }}
                addBenchmark={(bm: GradingBenchmark) => {
                    return this.courseMan.addNewBenchmark(bm);
                }}
                removeBenchmark={(bm: GradingBenchmark) => {
                    return this.courseMan.deleteBenchmark(bm);
                }}
                updateCriterion={(c: GradingCriterion) => {
                    return this.courseMan.updateCriterion(c);
                }}
                addCriterion={(c: GradingCriterion) => {
                    return this.courseMan.addNewCriterion(c);
                }}
                removeCriterion={(c: GradingCriterion) => {
                    return this.courseMan.deleteCriterion(c);
                }}
                loadBenchmarks={async () => {
                    const ans = await this.courseMan.loadCriteria(a.getId(), course.getId());
                    if (ans.length > 0) {
                        a.setGradingbenchmarksList(ans);
                    }
                    return ans;
                }}
                loadBenchmarks={() => {
                    return this.courseMan.loadCriteria(a.getId(), course.getId())
                }}
            ></AssignmentView>)
        }</div>
    }
}
