/**
 * @fileoverview gRPC-Web generated client stub for 
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


import * as grpcWeb from 'grpc-web';


import {
  Assignments,
  AuthorizationResponse,
  Course,
  CourseRequest,
  CourseSubmissions,
  Courses,
  Enrollment,
  EnrollmentRequest,
  EnrollmentStatusRequest,
  Enrollments,
  GetGroupRequest,
  GradingBenchmark,
  GradingCriterion,
  Group,
  GroupRequest,
  Groups,
  OrgRequest,
  Organization,
  Providers,
  RebuildRequest,
  Repositories,
  RepositoryRequest,
  Submission,
  SubmissionRequest,
  Submissions,
  SubmissionsForCourseRequest,
  URLRequest,
  UpdateSubmissionRequest,
  User,
  Users,
  Void} from './ag_pb';

export class AutograderServiceClient {
  client_: grpcWeb.AbstractClientBase;
  hostname_: string;
  credentials_: null | { [index: string]: string; };
  options_: null | { [index: string]: string; };

  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: string; }) {
    if (!options) options = {};
    if (!credentials) credentials = {};
    options['format'] = 'binary';

    this.client_ = new grpcWeb.GrpcWebClientBase(options);
    this.hostname_ = hostname;
    this.credentials_ = credentials;
    this.options_ = options;
  }

  methodInfoGetUser = new grpcWeb.AbstractClientBase.MethodInfo(
    User,
    (request: Void) => {
      return request.serializeBinary();
    },
    User.deserializeBinary
  );

  getUser(
    request: Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: User) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetUser',
      request,
      metadata || {},
      this.methodInfoGetUser,
      callback);
  }

  methodInfoGetUsers = new grpcWeb.AbstractClientBase.MethodInfo(
    Users,
    (request: Void) => {
      return request.serializeBinary();
    },
    Users.deserializeBinary
  );

  getUsers(
    request: Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Users) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetUsers',
      request,
      metadata || {},
      this.methodInfoGetUsers,
      callback);
  }

  methodInfoUpdateUser = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: User) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  updateUser(
    request: User,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/UpdateUser',
      request,
      metadata || {},
      this.methodInfoUpdateUser,
      callback);
  }

  methodInfoIsAuthorizedTeacher = new grpcWeb.AbstractClientBase.MethodInfo(
    AuthorizationResponse,
    (request: Void) => {
      return request.serializeBinary();
    },
    AuthorizationResponse.deserializeBinary
  );

  isAuthorizedTeacher(
    request: Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: AuthorizationResponse) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/IsAuthorizedTeacher',
      request,
      metadata || {},
      this.methodInfoIsAuthorizedTeacher,
      callback);
  }

  methodInfoGetGroup = new grpcWeb.AbstractClientBase.MethodInfo(
    Group,
    (request: GetGroupRequest) => {
      return request.serializeBinary();
    },
    Group.deserializeBinary
  );

  getGroup(
    request: GetGroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Group) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetGroup',
      request,
      metadata || {},
      this.methodInfoGetGroup,
      callback);
  }

  methodInfoGetGroupByUserAndCourse = new grpcWeb.AbstractClientBase.MethodInfo(
    Group,
    (request: GroupRequest) => {
      return request.serializeBinary();
    },
    Group.deserializeBinary
  );

  getGroupByUserAndCourse(
    request: GroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Group) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetGroupByUserAndCourse',
      request,
      metadata || {},
      this.methodInfoGetGroupByUserAndCourse,
      callback);
  }

  methodInfoGetGroupsByCourse = new grpcWeb.AbstractClientBase.MethodInfo(
    Groups,
    (request: CourseRequest) => {
      return request.serializeBinary();
    },
    Groups.deserializeBinary
  );

  getGroupsByCourse(
    request: CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Groups) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetGroupsByCourse',
      request,
      metadata || {},
      this.methodInfoGetGroupsByCourse,
      callback);
  }

  methodInfoCreateGroup = new grpcWeb.AbstractClientBase.MethodInfo(
    Group,
    (request: Group) => {
      return request.serializeBinary();
    },
    Group.deserializeBinary
  );

  createGroup(
    request: Group,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Group) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/CreateGroup',
      request,
      metadata || {},
      this.methodInfoCreateGroup,
      callback);
  }

  methodInfoUpdateGroup = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: Group) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  updateGroup(
    request: Group,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/UpdateGroup',
      request,
      metadata || {},
      this.methodInfoUpdateGroup,
      callback);
  }

  methodInfoDeleteGroup = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: GroupRequest) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  deleteGroup(
    request: GroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/DeleteGroup',
      request,
      metadata || {},
      this.methodInfoDeleteGroup,
      callback);
  }

  methodInfoGetCourse = new grpcWeb.AbstractClientBase.MethodInfo(
    Course,
    (request: CourseRequest) => {
      return request.serializeBinary();
    },
    Course.deserializeBinary
  );

  getCourse(
    request: CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Course) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetCourse',
      request,
      metadata || {},
      this.methodInfoGetCourse,
      callback);
  }

  methodInfoGetCourses = new grpcWeb.AbstractClientBase.MethodInfo(
    Courses,
    (request: Void) => {
      return request.serializeBinary();
    },
    Courses.deserializeBinary
  );

  getCourses(
    request: Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Courses) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetCourses',
      request,
      metadata || {},
      this.methodInfoGetCourses,
      callback);
  }

  methodInfoGetCoursesByUser = new grpcWeb.AbstractClientBase.MethodInfo(
    Courses,
    (request: EnrollmentStatusRequest) => {
      return request.serializeBinary();
    },
    Courses.deserializeBinary
  );

  getCoursesByUser(
    request: EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Courses) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetCoursesByUser',
      request,
      metadata || {},
      this.methodInfoGetCoursesByUser,
      callback);
  }

  methodInfoCreateCourse = new grpcWeb.AbstractClientBase.MethodInfo(
    Course,
    (request: Course) => {
      return request.serializeBinary();
    },
    Course.deserializeBinary
  );

  createCourse(
    request: Course,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Course) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/CreateCourse',
      request,
      metadata || {},
      this.methodInfoCreateCourse,
      callback);
  }

  methodInfoUpdateCourse = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: Course) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  updateCourse(
    request: Course,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/UpdateCourse',
      request,
      metadata || {},
      this.methodInfoUpdateCourse,
      callback);
  }

  methodInfoUpdateCourseVisibility = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: Enrollment) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  updateCourseVisibility(
    request: Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/UpdateCourseVisibility',
      request,
      metadata || {},
      this.methodInfoUpdateCourseVisibility,
      callback);
  }

  methodInfoGetAssignments = new grpcWeb.AbstractClientBase.MethodInfo(
    Assignments,
    (request: CourseRequest) => {
      return request.serializeBinary();
    },
    Assignments.deserializeBinary
  );

  getAssignments(
    request: CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Assignments) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetAssignments',
      request,
      metadata || {},
      this.methodInfoGetAssignments,
      callback);
  }

  methodInfoUpdateAssignments = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: CourseRequest) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  updateAssignments(
    request: CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/UpdateAssignments',
      request,
      metadata || {},
      this.methodInfoUpdateAssignments,
      callback);
  }

  methodInfoGetEnrollmentsByUser = new grpcWeb.AbstractClientBase.MethodInfo(
    Enrollments,
    (request: EnrollmentStatusRequest) => {
      return request.serializeBinary();
    },
    Enrollments.deserializeBinary
  );

  getEnrollmentsByUser(
    request: EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Enrollments) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetEnrollmentsByUser',
      request,
      metadata || {},
      this.methodInfoGetEnrollmentsByUser,
      callback);
  }

  methodInfoGetEnrollmentsByCourse = new grpcWeb.AbstractClientBase.MethodInfo(
    Enrollments,
    (request: EnrollmentRequest) => {
      return request.serializeBinary();
    },
    Enrollments.deserializeBinary
  );

  getEnrollmentsByCourse(
    request: EnrollmentRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Enrollments) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetEnrollmentsByCourse',
      request,
      metadata || {},
      this.methodInfoGetEnrollmentsByCourse,
      callback);
  }

  methodInfoCreateEnrollment = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: Enrollment) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  createEnrollment(
    request: Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/CreateEnrollment',
      request,
      metadata || {},
      this.methodInfoCreateEnrollment,
      callback);
  }

  methodInfoUpdateEnrollment = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: Enrollment) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  updateEnrollment(
    request: Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/UpdateEnrollment',
      request,
      metadata || {},
      this.methodInfoUpdateEnrollment,
      callback);
  }

  methodInfoUpdateEnrollments = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: CourseRequest) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  updateEnrollments(
    request: CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/UpdateEnrollments',
      request,
      metadata || {},
      this.methodInfoUpdateEnrollments,
      callback);
  }

  methodInfoGetSubmissions = new grpcWeb.AbstractClientBase.MethodInfo(
    Submissions,
    (request: SubmissionRequest) => {
      return request.serializeBinary();
    },
    Submissions.deserializeBinary
  );

  getSubmissions(
    request: SubmissionRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Submissions) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetSubmissions',
      request,
      metadata || {},
      this.methodInfoGetSubmissions,
      callback);
  }

  methodInfoGetSubmissionsByCourse = new grpcWeb.AbstractClientBase.MethodInfo(
    CourseSubmissions,
    (request: SubmissionsForCourseRequest) => {
      return request.serializeBinary();
    },
    CourseSubmissions.deserializeBinary
  );

  getSubmissionsByCourse(
    request: SubmissionsForCourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: CourseSubmissions) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetSubmissionsByCourse',
      request,
      metadata || {},
      this.methodInfoGetSubmissionsByCourse,
      callback);
  }

  methodInfoUpdateSubmission = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: UpdateSubmissionRequest) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  updateSubmission(
    request: UpdateSubmissionRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/UpdateSubmission',
      request,
      metadata || {},
      this.methodInfoUpdateSubmission,
      callback);
  }

  methodInfoRebuildSubmission = new grpcWeb.AbstractClientBase.MethodInfo(
    Submission,
    (request: RebuildRequest) => {
      return request.serializeBinary();
    },
    Submission.deserializeBinary
  );

  rebuildSubmission(
    request: RebuildRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Submission) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/RebuildSubmission',
      request,
      metadata || {},
      this.methodInfoRebuildSubmission,
      callback);
  }

  methodInfoCreateBenchmark = new grpcWeb.AbstractClientBase.MethodInfo(
    GradingBenchmark,
    (request: GradingBenchmark) => {
      return request.serializeBinary();
    },
    GradingBenchmark.deserializeBinary
  );

  createBenchmark(
    request: GradingBenchmark,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: GradingBenchmark) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/CreateBenchmark',
      request,
      metadata || {},
      this.methodInfoCreateBenchmark,
      callback);
  }

  methodInfoUpdateBenchmark = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: GradingBenchmark) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  updateBenchmark(
    request: GradingBenchmark,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/UpdateBenchmark',
      request,
      metadata || {},
      this.methodInfoUpdateBenchmark,
      callback);
  }

  methodInfoDeleteBenchmark = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: GradingBenchmark) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  deleteBenchmark(
    request: GradingBenchmark,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/DeleteBenchmark',
      request,
      metadata || {},
      this.methodInfoDeleteBenchmark,
      callback);
  }

  methodInfoCreateCriterion = new grpcWeb.AbstractClientBase.MethodInfo(
    GradingCriterion,
    (request: GradingCriterion) => {
      return request.serializeBinary();
    },
    GradingCriterion.deserializeBinary
  );

  createCriterion(
    request: GradingCriterion,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: GradingCriterion) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/CreateCriterion',
      request,
      metadata || {},
      this.methodInfoCreateCriterion,
      callback);
  }

  methodInfoUpdateCriterion = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: GradingCriterion) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  updateCriterion(
    request: GradingCriterion,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/UpdateCriterion',
      request,
      metadata || {},
      this.methodInfoUpdateCriterion,
      callback);
  }

  methodInfoDeleteCriterion = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: GradingCriterion) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  deleteCriterion(
    request: GradingCriterion,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/DeleteCriterion',
      request,
      metadata || {},
      this.methodInfoDeleteCriterion,
      callback);
  }

  methodInfoUpdateFeedback = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: Submission) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  updateFeedback(
    request: Submission,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/UpdateFeedback',
      request,
      metadata || {},
      this.methodInfoUpdateFeedback,
      callback);
  }

  methodInfoGetProviders = new grpcWeb.AbstractClientBase.MethodInfo(
    Providers,
    (request: Void) => {
      return request.serializeBinary();
    },
    Providers.deserializeBinary
  );

  getProviders(
    request: Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Providers) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetProviders',
      request,
      metadata || {},
      this.methodInfoGetProviders,
      callback);
  }

  methodInfoGetOrganization = new grpcWeb.AbstractClientBase.MethodInfo(
    Organization,
    (request: OrgRequest) => {
      return request.serializeBinary();
    },
    Organization.deserializeBinary
  );

  getOrganization(
    request: OrgRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Organization) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetOrganization',
      request,
      metadata || {},
      this.methodInfoGetOrganization,
      callback);
  }

  methodInfoGetRepositories = new grpcWeb.AbstractClientBase.MethodInfo(
    Repositories,
    (request: URLRequest) => {
      return request.serializeBinary();
    },
    Repositories.deserializeBinary
  );

  getRepositories(
    request: URLRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Repositories) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetRepositories',
      request,
      metadata || {},
      this.methodInfoGetRepositories,
      callback);
  }

  methodInfoIsEmptyRepo = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: RepositoryRequest) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  isEmptyRepo(
    request: RepositoryRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/IsEmptyRepo',
      request,
      metadata || {},
      this.methodInfoIsEmptyRepo,
      callback);
  }

}

