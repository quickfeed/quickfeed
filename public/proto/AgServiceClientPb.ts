/**
 * @fileoverview gRPC-Web generated client stub for 
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck


import * as grpcWeb from 'grpc-web';

import * as ag_pb from './ag_pb';


export class AutograderServiceClient {
  client_: grpcWeb.AbstractClientBase;
  hostname_: string;
  credentials_: null | { [index: string]: string; };
  options_: null | { [index: string]: any; };

  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; }) {
    if (!options) options = {};
    if (!credentials) credentials = {};
    options['format'] = 'binary';

    this.client_ = new grpcWeb.GrpcWebClientBase(options);
    this.hostname_ = hostname;
    this.credentials_ = credentials;
    this.options_ = options;
  }

  methodInfoGetUser = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.User,
    (request: ag_pb.Void) => {
      return request.serializeBinary();
    },
    ag_pb.User.deserializeBinary
  );

  getUser(
    request: ag_pb.Void,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.User>;

  getUser(
    request: ag_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.User) => void): grpcWeb.ClientReadableStream<ag_pb.User>;

  getUser(
    request: ag_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.User) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/GetUser',
        request,
        metadata || {},
        this.methodInfoGetUser,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/GetUser',
    request,
    metadata || {},
    this.methodInfoGetUser);
  }

  methodInfoGetUsers = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Users,
    (request: ag_pb.Void) => {
      return request.serializeBinary();
    },
    ag_pb.Users.deserializeBinary
  );

  getUsers(
    request: ag_pb.Void,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Users>;

  getUsers(
    request: ag_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Users) => void): grpcWeb.ClientReadableStream<ag_pb.Users>;

  getUsers(
    request: ag_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Users) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/GetUsers',
        request,
        metadata || {},
        this.methodInfoGetUsers,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/GetUsers',
    request,
    metadata || {},
    this.methodInfoGetUsers);
  }

  methodInfoGetUserByCourse = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.User,
    (request: ag_pb.CourseUserRequest) => {
      return request.serializeBinary();
    },
    ag_pb.User.deserializeBinary
  );

  getUserByCourse(
    request: ag_pb.CourseUserRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.User>;

  getUserByCourse(
    request: ag_pb.CourseUserRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.User) => void): grpcWeb.ClientReadableStream<ag_pb.User>;

  getUserByCourse(
    request: ag_pb.CourseUserRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.User) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/GetUserByCourse',
        request,
        metadata || {},
        this.methodInfoGetUserByCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/GetUserByCourse',
    request,
    metadata || {},
    this.methodInfoGetUserByCourse);
  }

  methodInfoUpdateUser = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Void,
    (request: ag_pb.User) => {
      return request.serializeBinary();
    },
    ag_pb.Void.deserializeBinary
  );

  updateUser(
    request: ag_pb.User,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Void>;

  updateUser(
    request: ag_pb.User,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_pb.Void>;

  updateUser(
    request: ag_pb.User,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/UpdateUser',
        request,
        metadata || {},
        this.methodInfoUpdateUser,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/UpdateUser',
    request,
    metadata || {},
    this.methodInfoUpdateUser);
  }

  methodInfoIsAuthorizedTeacher = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.AuthorizationResponse,
    (request: ag_pb.Void) => {
      return request.serializeBinary();
    },
    ag_pb.AuthorizationResponse.deserializeBinary
  );

  isAuthorizedTeacher(
    request: ag_pb.Void,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.AuthorizationResponse>;

  isAuthorizedTeacher(
    request: ag_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.AuthorizationResponse) => void): grpcWeb.ClientReadableStream<ag_pb.AuthorizationResponse>;

  isAuthorizedTeacher(
    request: ag_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.AuthorizationResponse) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/IsAuthorizedTeacher',
        request,
        metadata || {},
        this.methodInfoIsAuthorizedTeacher,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/IsAuthorizedTeacher',
    request,
    metadata || {},
    this.methodInfoIsAuthorizedTeacher);
  }

  methodInfoGetGroup = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Group,
    (request: ag_pb.GetGroupRequest) => {
      return request.serializeBinary();
    },
    ag_pb.Group.deserializeBinary
  );

  getGroup(
    request: ag_pb.GetGroupRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Group>;

  getGroup(
    request: ag_pb.GetGroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Group) => void): grpcWeb.ClientReadableStream<ag_pb.Group>;

  getGroup(
    request: ag_pb.GetGroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Group) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/GetGroup',
        request,
        metadata || {},
        this.methodInfoGetGroup,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/GetGroup',
    request,
    metadata || {},
    this.methodInfoGetGroup);
  }

  methodInfoGetGroupByUserAndCourse = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Group,
    (request: ag_pb.GroupRequest) => {
      return request.serializeBinary();
    },
    ag_pb.Group.deserializeBinary
  );

  getGroupByUserAndCourse(
    request: ag_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Group>;

  getGroupByUserAndCourse(
    request: ag_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Group) => void): grpcWeb.ClientReadableStream<ag_pb.Group>;

  getGroupByUserAndCourse(
    request: ag_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Group) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/GetGroupByUserAndCourse',
        request,
        metadata || {},
        this.methodInfoGetGroupByUserAndCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/GetGroupByUserAndCourse',
    request,
    metadata || {},
    this.methodInfoGetGroupByUserAndCourse);
  }

  methodInfoGetGroupsByCourse = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Groups,
    (request: ag_pb.CourseRequest) => {
      return request.serializeBinary();
    },
    ag_pb.Groups.deserializeBinary
  );

  getGroupsByCourse(
    request: ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Groups>;

  getGroupsByCourse(
    request: ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Groups) => void): grpcWeb.ClientReadableStream<ag_pb.Groups>;

  getGroupsByCourse(
    request: ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Groups) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/GetGroupsByCourse',
        request,
        metadata || {},
        this.methodInfoGetGroupsByCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/GetGroupsByCourse',
    request,
    metadata || {},
    this.methodInfoGetGroupsByCourse);
  }

  methodInfoCreateGroup = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Group,
    (request: ag_pb.Group) => {
      return request.serializeBinary();
    },
    ag_pb.Group.deserializeBinary
  );

  createGroup(
    request: ag_pb.Group,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Group>;

  createGroup(
    request: ag_pb.Group,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Group) => void): grpcWeb.ClientReadableStream<ag_pb.Group>;

  createGroup(
    request: ag_pb.Group,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Group) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/CreateGroup',
        request,
        metadata || {},
        this.methodInfoCreateGroup,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/CreateGroup',
    request,
    metadata || {},
    this.methodInfoCreateGroup);
  }

  methodInfoUpdateGroup = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Void,
    (request: ag_pb.Group) => {
      return request.serializeBinary();
    },
    ag_pb.Void.deserializeBinary
  );

  updateGroup(
    request: ag_pb.Group,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Void>;

  updateGroup(
    request: ag_pb.Group,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_pb.Void>;

  updateGroup(
    request: ag_pb.Group,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/UpdateGroup',
        request,
        metadata || {},
        this.methodInfoUpdateGroup,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/UpdateGroup',
    request,
    metadata || {},
    this.methodInfoUpdateGroup);
  }

  methodInfoDeleteGroup = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Void,
    (request: ag_pb.GroupRequest) => {
      return request.serializeBinary();
    },
    ag_pb.Void.deserializeBinary
  );

  deleteGroup(
    request: ag_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Void>;

  deleteGroup(
    request: ag_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_pb.Void>;

  deleteGroup(
    request: ag_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/DeleteGroup',
        request,
        metadata || {},
        this.methodInfoDeleteGroup,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/DeleteGroup',
    request,
    metadata || {},
    this.methodInfoDeleteGroup);
  }

  methodInfoGetCourse = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Course,
    (request: ag_pb.CourseRequest) => {
      return request.serializeBinary();
    },
    ag_pb.Course.deserializeBinary
  );

  getCourse(
    request: ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Course>;

  getCourse(
    request: ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Course) => void): grpcWeb.ClientReadableStream<ag_pb.Course>;

  getCourse(
    request: ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Course) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/GetCourse',
        request,
        metadata || {},
        this.methodInfoGetCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/GetCourse',
    request,
    metadata || {},
    this.methodInfoGetCourse);
  }

  methodInfoGetCourses = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Courses,
    (request: ag_pb.Void) => {
      return request.serializeBinary();
    },
    ag_pb.Courses.deserializeBinary
  );

  getCourses(
    request: ag_pb.Void,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Courses>;

  getCourses(
    request: ag_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Courses) => void): grpcWeb.ClientReadableStream<ag_pb.Courses>;

  getCourses(
    request: ag_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Courses) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/GetCourses',
        request,
        metadata || {},
        this.methodInfoGetCourses,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/GetCourses',
    request,
    metadata || {},
    this.methodInfoGetCourses);
  }

  methodInfoGetCoursesByUser = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Courses,
    (request: ag_pb.EnrollmentStatusRequest) => {
      return request.serializeBinary();
    },
    ag_pb.Courses.deserializeBinary
  );

  getCoursesByUser(
    request: ag_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Courses>;

  getCoursesByUser(
    request: ag_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Courses) => void): grpcWeb.ClientReadableStream<ag_pb.Courses>;

  getCoursesByUser(
    request: ag_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Courses) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/GetCoursesByUser',
        request,
        metadata || {},
        this.methodInfoGetCoursesByUser,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/GetCoursesByUser',
    request,
    metadata || {},
    this.methodInfoGetCoursesByUser);
  }

  methodInfoCreateCourse = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Course,
    (request: ag_pb.Course) => {
      return request.serializeBinary();
    },
    ag_pb.Course.deserializeBinary
  );

  createCourse(
    request: ag_pb.Course,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Course>;

  createCourse(
    request: ag_pb.Course,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Course) => void): grpcWeb.ClientReadableStream<ag_pb.Course>;

  createCourse(
    request: ag_pb.Course,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Course) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/CreateCourse',
        request,
        metadata || {},
        this.methodInfoCreateCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/CreateCourse',
    request,
    metadata || {},
    this.methodInfoCreateCourse);
  }

  methodInfoUpdateCourse = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Void,
    (request: ag_pb.Course) => {
      return request.serializeBinary();
    },
    ag_pb.Void.deserializeBinary
  );

  updateCourse(
    request: ag_pb.Course,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Void>;

  updateCourse(
    request: ag_pb.Course,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_pb.Void>;

  updateCourse(
    request: ag_pb.Course,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/UpdateCourse',
        request,
        metadata || {},
        this.methodInfoUpdateCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/UpdateCourse',
    request,
    metadata || {},
    this.methodInfoUpdateCourse);
  }

  methodInfoUpdateCourseVisibility = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Void,
    (request: ag_pb.Enrollment) => {
      return request.serializeBinary();
    },
    ag_pb.Void.deserializeBinary
  );

  updateCourseVisibility(
    request: ag_pb.Enrollment,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Void>;

  updateCourseVisibility(
    request: ag_pb.Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_pb.Void>;

  updateCourseVisibility(
    request: ag_pb.Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/UpdateCourseVisibility',
        request,
        metadata || {},
        this.methodInfoUpdateCourseVisibility,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/UpdateCourseVisibility',
    request,
    metadata || {},
    this.methodInfoUpdateCourseVisibility);
  }

  methodInfoGetAssignments = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Assignments,
    (request: ag_pb.CourseRequest) => {
      return request.serializeBinary();
    },
    ag_pb.Assignments.deserializeBinary
  );

  getAssignments(
    request: ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Assignments>;

  getAssignments(
    request: ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Assignments) => void): grpcWeb.ClientReadableStream<ag_pb.Assignments>;

  getAssignments(
    request: ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Assignments) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/GetAssignments',
        request,
        metadata || {},
        this.methodInfoGetAssignments,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/GetAssignments',
    request,
    metadata || {},
    this.methodInfoGetAssignments);
  }

  methodInfoUpdateAssignments = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Void,
    (request: ag_pb.CourseRequest) => {
      return request.serializeBinary();
    },
    ag_pb.Void.deserializeBinary
  );

  updateAssignments(
    request: ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Void>;

  updateAssignments(
    request: ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_pb.Void>;

  updateAssignments(
    request: ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/UpdateAssignments',
        request,
        metadata || {},
        this.methodInfoUpdateAssignments,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/UpdateAssignments',
    request,
    metadata || {},
    this.methodInfoUpdateAssignments);
  }

  methodInfoGetEnrollmentsByUser = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Enrollments,
    (request: ag_pb.EnrollmentStatusRequest) => {
      return request.serializeBinary();
    },
    ag_pb.Enrollments.deserializeBinary
  );

  getEnrollmentsByUser(
    request: ag_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Enrollments>;

  getEnrollmentsByUser(
    request: ag_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Enrollments) => void): grpcWeb.ClientReadableStream<ag_pb.Enrollments>;

  getEnrollmentsByUser(
    request: ag_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Enrollments) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/GetEnrollmentsByUser',
        request,
        metadata || {},
        this.methodInfoGetEnrollmentsByUser,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/GetEnrollmentsByUser',
    request,
    metadata || {},
    this.methodInfoGetEnrollmentsByUser);
  }

  methodInfoGetEnrollmentsByCourse = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Enrollments,
    (request: ag_pb.EnrollmentRequest) => {
      return request.serializeBinary();
    },
    ag_pb.Enrollments.deserializeBinary
  );

  getEnrollmentsByCourse(
    request: ag_pb.EnrollmentRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Enrollments>;

  getEnrollmentsByCourse(
    request: ag_pb.EnrollmentRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Enrollments) => void): grpcWeb.ClientReadableStream<ag_pb.Enrollments>;

  getEnrollmentsByCourse(
    request: ag_pb.EnrollmentRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Enrollments) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/GetEnrollmentsByCourse',
        request,
        metadata || {},
        this.methodInfoGetEnrollmentsByCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/GetEnrollmentsByCourse',
    request,
    metadata || {},
    this.methodInfoGetEnrollmentsByCourse);
  }

  methodInfoCreateEnrollment = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Void,
    (request: ag_pb.Enrollment) => {
      return request.serializeBinary();
    },
    ag_pb.Void.deserializeBinary
  );

  createEnrollment(
    request: ag_pb.Enrollment,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Void>;

  createEnrollment(
    request: ag_pb.Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_pb.Void>;

  createEnrollment(
    request: ag_pb.Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/CreateEnrollment',
        request,
        metadata || {},
        this.methodInfoCreateEnrollment,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/CreateEnrollment',
    request,
    metadata || {},
    this.methodInfoCreateEnrollment);
  }

  methodInfoUpdateEnrollment = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Void,
    (request: ag_pb.Enrollment) => {
      return request.serializeBinary();
    },
    ag_pb.Void.deserializeBinary
  );

  updateEnrollment(
    request: ag_pb.Enrollment,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Void>;

  updateEnrollment(
    request: ag_pb.Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_pb.Void>;

  updateEnrollment(
    request: ag_pb.Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/UpdateEnrollment',
        request,
        metadata || {},
        this.methodInfoUpdateEnrollment,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/UpdateEnrollment',
    request,
    metadata || {},
    this.methodInfoUpdateEnrollment);
  }

  methodInfoUpdateEnrollments = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Void,
    (request: ag_pb.CourseRequest) => {
      return request.serializeBinary();
    },
    ag_pb.Void.deserializeBinary
  );

  updateEnrollments(
    request: ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Void>;

  updateEnrollments(
    request: ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_pb.Void>;

  updateEnrollments(
    request: ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/UpdateEnrollments',
        request,
        metadata || {},
        this.methodInfoUpdateEnrollments,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/UpdateEnrollments',
    request,
    metadata || {},
    this.methodInfoUpdateEnrollments);
  }

  methodInfoGetSubmissions = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Submissions,
    (request: ag_pb.SubmissionRequest) => {
      return request.serializeBinary();
    },
    ag_pb.Submissions.deserializeBinary
  );

  getSubmissions(
    request: ag_pb.SubmissionRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Submissions>;

  getSubmissions(
    request: ag_pb.SubmissionRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Submissions) => void): grpcWeb.ClientReadableStream<ag_pb.Submissions>;

  getSubmissions(
    request: ag_pb.SubmissionRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Submissions) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/GetSubmissions',
        request,
        metadata || {},
        this.methodInfoGetSubmissions,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/GetSubmissions',
    request,
    metadata || {},
    this.methodInfoGetSubmissions);
  }

  methodInfoGetSubmissionsByCourse = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.CourseSubmissions,
    (request: ag_pb.SubmissionsForCourseRequest) => {
      return request.serializeBinary();
    },
    ag_pb.CourseSubmissions.deserializeBinary
  );

  getSubmissionsByCourse(
    request: ag_pb.SubmissionsForCourseRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.CourseSubmissions>;

  getSubmissionsByCourse(
    request: ag_pb.SubmissionsForCourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.CourseSubmissions) => void): grpcWeb.ClientReadableStream<ag_pb.CourseSubmissions>;

  getSubmissionsByCourse(
    request: ag_pb.SubmissionsForCourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.CourseSubmissions) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/GetSubmissionsByCourse',
        request,
        metadata || {},
        this.methodInfoGetSubmissionsByCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/GetSubmissionsByCourse',
    request,
    metadata || {},
    this.methodInfoGetSubmissionsByCourse);
  }

  methodInfoUpdateSubmission = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Void,
    (request: ag_pb.UpdateSubmissionRequest) => {
      return request.serializeBinary();
    },
    ag_pb.Void.deserializeBinary
  );

  updateSubmission(
    request: ag_pb.UpdateSubmissionRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Void>;

  updateSubmission(
    request: ag_pb.UpdateSubmissionRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_pb.Void>;

  updateSubmission(
    request: ag_pb.UpdateSubmissionRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/UpdateSubmission',
        request,
        metadata || {},
        this.methodInfoUpdateSubmission,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/UpdateSubmission',
    request,
    metadata || {},
    this.methodInfoUpdateSubmission);
  }

  methodInfoUpdateSubmissions = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Void,
    (request: ag_pb.UpdateSubmissionsRequest) => {
      return request.serializeBinary();
    },
    ag_pb.Void.deserializeBinary
  );

  updateSubmissions(
    request: ag_pb.UpdateSubmissionsRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Void>;

  updateSubmissions(
    request: ag_pb.UpdateSubmissionsRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_pb.Void>;

  updateSubmissions(
    request: ag_pb.UpdateSubmissionsRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/UpdateSubmissions',
        request,
        metadata || {},
        this.methodInfoUpdateSubmissions,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/UpdateSubmissions',
    request,
    metadata || {},
    this.methodInfoUpdateSubmissions);
  }

  methodInfoRebuildSubmission = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Submission,
    (request: ag_pb.RebuildRequest) => {
      return request.serializeBinary();
    },
    ag_pb.Submission.deserializeBinary
  );

  rebuildSubmission(
    request: ag_pb.RebuildRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Submission>;

  rebuildSubmission(
    request: ag_pb.RebuildRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Submission) => void): grpcWeb.ClientReadableStream<ag_pb.Submission>;

  rebuildSubmission(
    request: ag_pb.RebuildRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Submission) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/RebuildSubmission',
        request,
        metadata || {},
        this.methodInfoRebuildSubmission,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/RebuildSubmission',
    request,
    metadata || {},
    this.methodInfoRebuildSubmission);
  }

  methodInfoCreateBenchmark = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.GradingBenchmark,
    (request: ag_pb.GradingBenchmark) => {
      return request.serializeBinary();
    },
    ag_pb.GradingBenchmark.deserializeBinary
  );

  createBenchmark(
    request: ag_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.GradingBenchmark>;

  createBenchmark(
    request: ag_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.GradingBenchmark) => void): grpcWeb.ClientReadableStream<ag_pb.GradingBenchmark>;

  createBenchmark(
    request: ag_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.GradingBenchmark) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/CreateBenchmark',
        request,
        metadata || {},
        this.methodInfoCreateBenchmark,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/CreateBenchmark',
    request,
    metadata || {},
    this.methodInfoCreateBenchmark);
  }

  methodInfoUpdateBenchmark = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Void,
    (request: ag_pb.GradingBenchmark) => {
      return request.serializeBinary();
    },
    ag_pb.Void.deserializeBinary
  );

  updateBenchmark(
    request: ag_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Void>;

  updateBenchmark(
    request: ag_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_pb.Void>;

  updateBenchmark(
    request: ag_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/UpdateBenchmark',
        request,
        metadata || {},
        this.methodInfoUpdateBenchmark,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/UpdateBenchmark',
    request,
    metadata || {},
    this.methodInfoUpdateBenchmark);
  }

  methodInfoDeleteBenchmark = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Void,
    (request: ag_pb.GradingBenchmark) => {
      return request.serializeBinary();
    },
    ag_pb.Void.deserializeBinary
  );

  deleteBenchmark(
    request: ag_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Void>;

  deleteBenchmark(
    request: ag_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_pb.Void>;

  deleteBenchmark(
    request: ag_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/DeleteBenchmark',
        request,
        metadata || {},
        this.methodInfoDeleteBenchmark,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/DeleteBenchmark',
    request,
    metadata || {},
    this.methodInfoDeleteBenchmark);
  }

  methodInfoCreateCriterion = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.GradingCriterion,
    (request: ag_pb.GradingCriterion) => {
      return request.serializeBinary();
    },
    ag_pb.GradingCriterion.deserializeBinary
  );

  createCriterion(
    request: ag_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.GradingCriterion>;

  createCriterion(
    request: ag_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.GradingCriterion) => void): grpcWeb.ClientReadableStream<ag_pb.GradingCriterion>;

  createCriterion(
    request: ag_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.GradingCriterion) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/CreateCriterion',
        request,
        metadata || {},
        this.methodInfoCreateCriterion,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/CreateCriterion',
    request,
    metadata || {},
    this.methodInfoCreateCriterion);
  }

  methodInfoUpdateCriterion = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Void,
    (request: ag_pb.GradingCriterion) => {
      return request.serializeBinary();
    },
    ag_pb.Void.deserializeBinary
  );

  updateCriterion(
    request: ag_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Void>;

  updateCriterion(
    request: ag_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_pb.Void>;

  updateCriterion(
    request: ag_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/UpdateCriterion',
        request,
        metadata || {},
        this.methodInfoUpdateCriterion,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/UpdateCriterion',
    request,
    metadata || {},
    this.methodInfoUpdateCriterion);
  }

  methodInfoDeleteCriterion = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Void,
    (request: ag_pb.GradingCriterion) => {
      return request.serializeBinary();
    },
    ag_pb.Void.deserializeBinary
  );

  deleteCriterion(
    request: ag_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Void>;

  deleteCriterion(
    request: ag_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_pb.Void>;

  deleteCriterion(
    request: ag_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/DeleteCriterion',
        request,
        metadata || {},
        this.methodInfoDeleteCriterion,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/DeleteCriterion',
    request,
    metadata || {},
    this.methodInfoDeleteCriterion);
  }

  methodInfoCreateReview = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Review,
    (request: ag_pb.ReviewRequest) => {
      return request.serializeBinary();
    },
    ag_pb.Review.deserializeBinary
  );

  createReview(
    request: ag_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Review>;

  createReview(
    request: ag_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Review) => void): grpcWeb.ClientReadableStream<ag_pb.Review>;

  createReview(
    request: ag_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Review) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/CreateReview',
        request,
        metadata || {},
        this.methodInfoCreateReview,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/CreateReview',
    request,
    metadata || {},
    this.methodInfoCreateReview);
  }

  methodInfoUpdateReview = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Void,
    (request: ag_pb.ReviewRequest) => {
      return request.serializeBinary();
    },
    ag_pb.Void.deserializeBinary
  );

  updateReview(
    request: ag_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Void>;

  updateReview(
    request: ag_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_pb.Void>;

  updateReview(
    request: ag_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/UpdateReview',
        request,
        metadata || {},
        this.methodInfoUpdateReview,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/UpdateReview',
    request,
    metadata || {},
    this.methodInfoUpdateReview);
  }

  methodInfoGetReviewers = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Reviewers,
    (request: ag_pb.SubmissionReviewersRequest) => {
      return request.serializeBinary();
    },
    ag_pb.Reviewers.deserializeBinary
  );

  getReviewers(
    request: ag_pb.SubmissionReviewersRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Reviewers>;

  getReviewers(
    request: ag_pb.SubmissionReviewersRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Reviewers) => void): grpcWeb.ClientReadableStream<ag_pb.Reviewers>;

  getReviewers(
    request: ag_pb.SubmissionReviewersRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Reviewers) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/GetReviewers',
        request,
        metadata || {},
        this.methodInfoGetReviewers,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/GetReviewers',
    request,
    metadata || {},
    this.methodInfoGetReviewers);
  }

  methodInfoLoadCriteria = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Benchmarks,
    (request: ag_pb.LoadCriteriaRequest) => {
      return request.serializeBinary();
    },
    ag_pb.Benchmarks.deserializeBinary
  );

  loadCriteria(
    request: ag_pb.LoadCriteriaRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Benchmarks>;

  loadCriteria(
    request: ag_pb.LoadCriteriaRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Benchmarks) => void): grpcWeb.ClientReadableStream<ag_pb.Benchmarks>;

  loadCriteria(
    request: ag_pb.LoadCriteriaRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Benchmarks) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/LoadCriteria',
        request,
        metadata || {},
        this.methodInfoLoadCriteria,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/LoadCriteria',
    request,
    metadata || {},
    this.methodInfoLoadCriteria);
  }

  methodInfoGetProviders = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Providers,
    (request: ag_pb.Void) => {
      return request.serializeBinary();
    },
    ag_pb.Providers.deserializeBinary
  );

  getProviders(
    request: ag_pb.Void,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Providers>;

  getProviders(
    request: ag_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Providers) => void): grpcWeb.ClientReadableStream<ag_pb.Providers>;

  getProviders(
    request: ag_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Providers) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/GetProviders',
        request,
        metadata || {},
        this.methodInfoGetProviders,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/GetProviders',
    request,
    metadata || {},
    this.methodInfoGetProviders);
  }

  methodInfoGetOrganization = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Organization,
    (request: ag_pb.OrgRequest) => {
      return request.serializeBinary();
    },
    ag_pb.Organization.deserializeBinary
  );

  getOrganization(
    request: ag_pb.OrgRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Organization>;

  getOrganization(
    request: ag_pb.OrgRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Organization) => void): grpcWeb.ClientReadableStream<ag_pb.Organization>;

  getOrganization(
    request: ag_pb.OrgRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Organization) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/GetOrganization',
        request,
        metadata || {},
        this.methodInfoGetOrganization,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/GetOrganization',
    request,
    metadata || {},
    this.methodInfoGetOrganization);
  }

  methodInfoGetRepositories = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Repositories,
    (request: ag_pb.URLRequest) => {
      return request.serializeBinary();
    },
    ag_pb.Repositories.deserializeBinary
  );

  getRepositories(
    request: ag_pb.URLRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Repositories>;

  getRepositories(
    request: ag_pb.URLRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Repositories) => void): grpcWeb.ClientReadableStream<ag_pb.Repositories>;

  getRepositories(
    request: ag_pb.URLRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Repositories) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/GetRepositories',
        request,
        metadata || {},
        this.methodInfoGetRepositories,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/GetRepositories',
    request,
    metadata || {},
    this.methodInfoGetRepositories);
  }

  methodInfoIsEmptyRepo = new grpcWeb.AbstractClientBase.MethodInfo(
    ag_pb.Void,
    (request: ag_pb.RepositoryRequest) => {
      return request.serializeBinary();
    },
    ag_pb.Void.deserializeBinary
  );

  isEmptyRepo(
    request: ag_pb.RepositoryRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_pb.Void>;

  isEmptyRepo(
    request: ag_pb.RepositoryRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_pb.Void>;

  isEmptyRepo(
    request: ag_pb.RepositoryRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.Error,
               response: ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/AutograderService/IsEmptyRepo',
        request,
        metadata || {},
        this.methodInfoIsEmptyRepo,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/AutograderService/IsEmptyRepo',
    request,
    metadata || {},
    this.methodInfoIsEmptyRepo);
  }

}

