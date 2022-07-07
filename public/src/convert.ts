import { Assignment, Course, Enrollment, GradingBenchmark, GradingCriterion, Group, Review, Submission, User } from "../proto/qf/qf_pb"
import { BuildInfo, Score } from "../proto/kit/score/score_pb"


// Class with converter functions for the different proto types
export class Converter {

    /**
     * create creates a new object of the given type
     * @param type the type you wish to create as T.AsObject
     * @returns the given type as a T.AsObject
     * @example create<User.AsObject>(User) // returns a new User.AsObject
     */
    public static create<T>(type: (new () => any)): T {
        return (new type()).toObject() as T
    }

    public static clone<T>(obj: T): T {
        return JSON.parse(JSON.stringify(obj)) as T
    }

    /** The following functions will convert various kinds of objects
     *  into their respective protobuf message types */
    // TODO(jostein): Add converting functions for all types
    // TODO(jostein): It would be awesome to make this generic. Not sure how to do that though.
    public static toUser = (obj: User.AsObject): User => {
        const user = new User()
        user.setId(obj.id)
        user.setName(obj.name)
        user.setEmail(obj.email)
        user.setIsadmin(obj.isadmin)
        user.setLogin(obj.login)
        user.setAvatarurl(obj.avatarurl)
        user.setStudentid(obj.studentid)
        const enrollments = obj.enrollmentsList.map(e => this.toEnrollment(e))
        user.setEnrollmentsList(enrollments)

        return user
    }

    public static toEnrollment = (obj: Enrollment.AsObject): Enrollment => {
        const enrollment = new Enrollment()
        enrollment.setId(obj.id)
        enrollment.setStatus(obj.status)
        enrollment.setLastactivitydate(obj.lastactivitydate)
        enrollment.setSlipdaysremaining(obj.slipdaysremaining)
        enrollment.setTotalapproved(obj.totalapproved)
        enrollment.setCourseid(obj.courseid)
        enrollment.setUserid(obj.userid)
        enrollment.setGroupid(obj.groupid)
        enrollment.setHasteacherscopes(obj.hasteacherscopes)
        enrollment.setState(obj.state)

        // TODO: handle slipdays
        // for (const slipdays of obj.usedslipdaysList) {}
        if (obj.user) {
            enrollment.setUser(this.toUser(obj.user))
        }
        if (obj.course) {
            enrollment.setCourse(this.toCourse(obj.course))
        }
        if (obj.group) {
            enrollment.setGroup(this.toGroup(obj.group))
        }
        return enrollment
    }

    public static toCourse = (obj: Course.AsObject): Course => {
        const course = new Course()
        course.setId(obj.id)
        course.setCoursecreatorid(obj.coursecreatorid)
        course.setName(obj.name)
        course.setCode(obj.code)
        course.setYear(obj.year)
        course.setTag(obj.tag)
        course.setProvider(obj.provider)
        course.setOrganizationid(obj.organizationid)
        course.setOrganizationpath(obj.organizationpath)
        course.setSlipdays(obj.slipdays)
        course.setDockerfile(obj.dockerfile)
        course.setEnrolled(obj.enrolled)

        const enrollments = obj.enrollmentsList.map(e => this.toEnrollment(e))
        course.setEnrollmentsList(enrollments)

        const assignments = obj.assignmentsList.map(a => this.toAssignment(a))
        course.setAssignmentsList(assignments)

        const groups = obj.groupsList.map(g => this.toGroup(g))
        course.setGroupsList(groups)

        return course
    }

    public static toAssignment = (obj: Assignment.AsObject): Assignment => {
        const assignment = new Assignment()
        assignment.setId(obj.id)
        assignment.setCourseid(obj.courseid)
        assignment.setName(obj.name)
        assignment.setScriptfile(obj.scriptfile)
        assignment.setDeadline(obj.deadline)
        assignment.setAutoapprove(obj.autoapprove)
        assignment.setOrder(obj.order)
        assignment.setIsgrouplab(obj.isgrouplab)
        assignment.setScorelimit(obj.scorelimit)
        assignment.setReviewers(obj.reviewers)

        const submissions = obj.submissionsList.map(s => this.toSubmission(s))
        assignment.setSubmissionsList(submissions)

        const gradingBenchmarks = obj.gradingbenchmarksList.map(g => this.toGradingBenchmark(g))
        assignment.setGradingbenchmarksList(gradingBenchmarks)

        assignment.setContainertimeout(obj.containertimeout)

        return assignment
    }

    public static toSubmission = (obj: Submission.AsObject): Submission => {
        const submission = new Submission()
        submission.setId(obj.id)
        submission.setAssignmentid(obj.assignmentid)
        submission.setUserid(obj.userid)
        submission.setGroupid(obj.groupid)
        submission.setScore(obj.score)
        submission.setCommithash(obj.commithash)
        submission.setReleased(obj.released)
        submission.setStatus(obj.status)
        submission.setApproveddate(obj.approveddate)

        const reviews = obj.reviewsList.map(r => this.toReview(r))
        submission.setReviewsList(reviews)

        if (obj.buildinfo) {
            submission.setBuildinfo(this.toBuildInfo(obj.buildinfo))
        }

        const scores = obj.scoresList.map(s => this.toScore(s))
        submission.setScoresList(scores)

        return submission
    }

    public static toBuildInfo = (obj: BuildInfo.AsObject): BuildInfo => {
        const buildInfo = new BuildInfo()
        buildInfo.setId(obj.id)
        buildInfo.setSubmissionid(obj.submissionid)
        buildInfo.setBuilddate(obj.builddate)
        buildInfo.setBuildlog(obj.buildlog)
        buildInfo.setExectime(obj.exectime)

        return buildInfo
    }

    public static toScore = (obj: Score.AsObject): Score => {
        const score = new Score()
        score.setId(obj.id)
        score.setSubmissionid(obj.submissionid)
        score.setSecret(obj.secret)
        score.setTestname(obj.testname)
        score.setScore(obj.score)
        score.setMaxscore(obj.maxscore)
        score.setWeight(obj.weight)
        score.setTestdetails(obj.testdetails)

        return score
    }

    public static toReview = (obj: Review.AsObject): Review => {
        const review = new Review()
        review.setId(obj.id)
        review.setSubmissionid(obj.submissionid)
        review.setReviewerid(obj.reviewerid)
        review.setFeedback(obj.feedback)
        review.setReady(obj.ready)
        review.setScore(obj.score)

        const gradingBenchmarks = obj.gradingbenchmarksList.map(g => this.toGradingBenchmark(g))
        review.setGradingbenchmarksList(gradingBenchmarks)

        review.setEdited(obj.edited)

        return review
    }

    public static toGradingBenchmark = (obj: GradingBenchmark.AsObject): GradingBenchmark => {
        const gradingBenchmark = new GradingBenchmark()
        gradingBenchmark.setId(obj.id)
        gradingBenchmark.setAssignmentid(obj.assignmentid)
        gradingBenchmark.setReviewid(obj.reviewid)
        gradingBenchmark.setHeading(obj.heading)
        gradingBenchmark.setComment(obj.comment)

        const criteria = obj.criteriaList.map(c => this.toGradingCriterion(c))
        gradingBenchmark.setCriteriaList(criteria)

        return gradingBenchmark
    }

    public static toGradingCriterion = (obj: GradingCriterion.AsObject): GradingCriterion => {
        const criterion = new GradingCriterion()
        criterion.setId(obj.id)
        criterion.setBenchmarkid(obj.benchmarkid)
        criterion.setPoints(obj.points)
        criterion.setDescription(obj.description)
        criterion.setGrade(obj.grade)
        criterion.setComment(obj.comment)

        return criterion
    }

    public static toGroup = (obj: Group.AsObject): Group => {
        const group = new Group()
        group.setId(obj.id)
        group.setName(obj.name)
        group.setCourseid(obj.courseid)
        group.setTeamid(obj.teamid)
        group.setStatus(obj.status)

        const users = obj.usersList.map(u => this.toUser(u))
        group.setUsersList(users)

        const enrollments = obj.enrollmentsList.map(e => this.toEnrollment(e))
        group.setEnrollmentsList(enrollments)

        return group
    }

}
