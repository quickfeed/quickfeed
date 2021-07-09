/**
 * @fileoverview gRPC-Web generated client stub for ag
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


import * as grpcWeb from 'grpc-web';


import {
  Assignments,
  AuthorizationResponse,
  Benchmarks,
  CommitHashRequest,
  CommitHashResponse,
  Course,
  CourseRequest,
  CourseSubmissions,
  CourseUserRequest,
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
  LoadCriteriaRequest,
  OrgRequest,
  Organization,
  Providers,
  RebuildRequest,
  Repositories,
  RepositoryRequest,
  Review,
  ReviewRequest,
  Reviewers,
  Submission,
  SubmissionRequest,
  SubmissionReviewersRequest,
  Submissions,
  SubmissionsForCourseRequest,
  URLRequest,
  UpdateSubmissionRequest,
  UpdateSubmissionsRequest,
  User,
  Users,
  Void} from './ag_pb';

export class AutograderServiceClient {
  client_: grpcWeb.AbstractClientBase;
  hostname_: string;
  credentials_: null | { [index: string]: string; };
  options_: null | { [index: string]: string; };

  constructor (hostname: string,
               credentials: null | { [index: string]: string; },
               options: null | { [index: string]: string; }) {
    if (!options) options = {};
    if (!credentials) credentials = {};
    options['format'] = 'text';

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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/GetUser', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoGetUser,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetUser',
    request,
    metadata || {},
    this.methodInfoGetUser);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/GetUsers', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoGetUsers,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetUsers',
    request,
    metadata || {},
    this.methodInfoGetUsers);
  }

  methodInfoGetUserByCourse = new grpcWeb.AbstractClientBase.MethodInfo(
    User,
    (request: CourseUserRequest) => {
      return request.serializeBinary();
    },
    User.deserializeBinary
  );

  getUserByCourse(
    request: CourseUserRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: User) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/GetUserByCourse', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoGetUserByCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetUserByCourse',
    request,
    metadata || {},
    this.methodInfoGetUserByCourse);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/UpdateUser', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoUpdateUser,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateUser',
    request,
    metadata || {},
    this.methodInfoUpdateUser);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/IsAuthorizedTeacher', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoIsAuthorizedTeacher,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/IsAuthorizedTeacher',
    request,
    metadata || {},
    this.methodInfoIsAuthorizedTeacher);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/GetGroup', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoGetGroup,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetGroup',
    request,
    metadata || {},
    this.methodInfoGetGroup);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/GetGroupByUserAndCourse', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoGetGroupByUserAndCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetGroupByUserAndCourse',
    request,
    metadata || {},
    this.methodInfoGetGroupByUserAndCourse);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/GetGroupsByCourse', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoGetGroupsByCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetGroupsByCourse',
    request,
    metadata || {},
    this.methodInfoGetGroupsByCourse);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/CreateGroup', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoCreateGroup,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/CreateGroup',
    request,
    metadata || {},
    this.methodInfoCreateGroup);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/UpdateGroup', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoUpdateGroup,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateGroup',
    request,
    metadata || {},
    this.methodInfoUpdateGroup);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/DeleteGroup', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoDeleteGroup,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/DeleteGroup',
    request,
    metadata || {},
    this.methodInfoDeleteGroup);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/GetCourse', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoGetCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetCourse',
    request,
    metadata || {},
    this.methodInfoGetCourse);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/GetCourses', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoGetCourses,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetCourses',
    request,
    metadata || {},
    this.methodInfoGetCourses);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/GetCoursesByUser', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoGetCoursesByUser,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetCoursesByUser',
    request,
    metadata || {},
    this.methodInfoGetCoursesByUser);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/CreateCourse', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoCreateCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/CreateCourse',
    request,
    metadata || {},
    this.methodInfoCreateCourse);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/UpdateCourse', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoUpdateCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateCourse',
    request,
    metadata || {},
    this.methodInfoUpdateCourse);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/UpdateCourseVisibility', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoUpdateCourseVisibility,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateCourseVisibility',
    request,
    metadata || {},
    this.methodInfoUpdateCourseVisibility);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/GetAssignments', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoGetAssignments,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetAssignments',
    request,
    metadata || {},
    this.methodInfoGetAssignments);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/UpdateAssignments', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoUpdateAssignments,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateAssignments',
    request,
    metadata || {},
    this.methodInfoUpdateAssignments);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/GetEnrollmentsByUser', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoGetEnrollmentsByUser,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetEnrollmentsByUser',
    request,
    metadata || {},
    this.methodInfoGetEnrollmentsByUser);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/GetEnrollmentsByCourse', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoGetEnrollmentsByCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetEnrollmentsByCourse',
    request,
    metadata || {},
    this.methodInfoGetEnrollmentsByCourse);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/CreateEnrollment', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoCreateEnrollment,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/CreateEnrollment',
    request,
    metadata || {},
    this.methodInfoCreateEnrollment);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/UpdateEnrollment', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoUpdateEnrollment,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateEnrollment',
    request,
    metadata || {},
    this.methodInfoUpdateEnrollment);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/UpdateEnrollments', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoUpdateEnrollments,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateEnrollments',
    request,
    metadata || {},
    this.methodInfoUpdateEnrollments);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/GetSubmissions', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoGetSubmissions,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetSubmissions',
    request,
    metadata || {},
    this.methodInfoGetSubmissions);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/GetSubmissionsByCourse', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoGetSubmissionsByCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetSubmissionsByCourse',
    request,
    metadata || {},
    this.methodInfoGetSubmissionsByCourse);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/UpdateSubmission', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoUpdateSubmission,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateSubmission',
    request,
    metadata || {},
    this.methodInfoUpdateSubmission);
  }

  methodInfoUpdateSubmissions = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: UpdateSubmissionsRequest) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  updateSubmissions(
    request: UpdateSubmissionsRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/UpdateSubmissions', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoUpdateSubmissions,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateSubmissions',
    request,
    metadata || {},
    this.methodInfoUpdateSubmissions);
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

  methodInfoGetSubmissionCommitHash = new grpcWeb.AbstractClientBase.MethodInfo(
    CommitHashResponse,
    (request: CommitHashRequest) => {
      return request.serializeBinary();
    },
    CommitHashResponse.deserializeBinary
  );

  getSubmissionCommitHash(
    request: CommitHashRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: Submission) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/RebuildSubmission', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoRebuildSubmission,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/RebuildSubmission',
    request,
    metadata || {},
    this.methodInfoRebuildSubmission);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/CreateBenchmark', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoCreateBenchmark,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/CreateBenchmark',
    request,
    metadata || {},
    this.methodInfoCreateBenchmark);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/UpdateBenchmark', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoUpdateBenchmark,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateBenchmark',
    request,
    metadata || {},
    this.methodInfoUpdateBenchmark);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/DeleteBenchmark', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoDeleteBenchmark,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/DeleteBenchmark',
    request,
    metadata || {},
    this.methodInfoDeleteBenchmark);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/CreateCriterion', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoCreateCriterion,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/CreateCriterion',
    request,
    metadata || {},
    this.methodInfoCreateCriterion);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/UpdateCriterion', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoUpdateCriterion,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateCriterion',
    request,
    metadata || {},
    this.methodInfoUpdateCriterion);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/DeleteCriterion', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoDeleteCriterion,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/DeleteCriterion',
    request,
    metadata || {},
    this.methodInfoDeleteCriterion);
  }

  methodInfoCreateReview = new grpcWeb.AbstractClientBase.MethodInfo(
    Review,
    (request: ReviewRequest) => {
      return request.serializeBinary();
    },
    Review.deserializeBinary
  );

  createReview(
    request: ReviewRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Review) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/CreateReview', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoCreateReview,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/CreateReview',
    request,
    metadata || {},
    this.methodInfoCreateReview);
  }

  methodInfoUpdateReview = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: ReviewRequest) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  updateReview(
    request: ReviewRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/UpdateReview', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoUpdateReview,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateReview',
    request,
    metadata || {},
    this.methodInfoUpdateReview);
  }

  methodInfoGetReviewers = new grpcWeb.AbstractClientBase.MethodInfo(
    Reviewers,
    (request: SubmissionReviewersRequest) => {
      return request.serializeBinary();
    },
    Reviewers.deserializeBinary
  );

  getReviewers(
    request: SubmissionReviewersRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Reviewers) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/GetReviewers', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoGetReviewers,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetReviewers',
    request,
    metadata || {},
    this.methodInfoGetReviewers);
  }

  methodInfoLoadCriteria = new grpcWeb.AbstractClientBase.MethodInfo(
    Benchmarks,
    (request: LoadCriteriaRequest) => {
      return request.serializeBinary();
    },
    Benchmarks.deserializeBinary
  );

  loadCriteria(
    request: LoadCriteriaRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Benchmarks) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/LoadCriteria', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoLoadCriteria,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/LoadCriteria',
    request,
    metadata || {},
    this.methodInfoLoadCriteria);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/GetProviders', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoGetProviders,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetProviders',
    request,
    metadata || {},
    this.methodInfoGetProviders);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/GetOrganization', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoGetOrganization,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetOrganization',
    request,
    metadata || {},
    this.methodInfoGetOrganization);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/GetRepositories', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoGetRepositories,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetRepositories',
    request,
    metadata || {},
    this.methodInfoGetRepositories);
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
    if (callback !== undefined) {
      return this.client_.rpcCall(
        new URL('/ag.AutograderService/IsEmptyRepo', this.hostname_).toString(),
        request,
        metadata || {},
        this.methodInfoIsEmptyRepo,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/IsEmptyRepo',
    request,
    metadata || {},
    this.methodInfoIsEmptyRepo);
  }

}

