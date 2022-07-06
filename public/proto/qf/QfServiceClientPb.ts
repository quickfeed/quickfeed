/**
 * @fileoverview gRPC-Web generated client stub for qf
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck


import * as grpcWeb from 'grpc-web';

import * as qf_qf_pb from '../qf/qf_pb';


export class QuickFeedServiceClient {
  client_: grpcWeb.AbstractClientBase;
  hostname_: string;
  credentials_: null | { [index: string]: string; };
  options_: null | { [index: string]: any; };

  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; }) {
    if (!options) options = {};
    if (!credentials) credentials = {};
    options['format'] = 'text';

    this.client_ = new grpcWeb.GrpcWebClientBase(options);
    this.hostname_ = hostname;
    this.credentials_ = credentials;
    this.options_ = options;
  }

  methodInfoGetUser = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/GetUser',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.Void,
    qf_qf_pb.User,
    (request: qf_qf_pb.Void) => {
      return request.serializeBinary();
    },
    qf_qf_pb.User.deserializeBinary
  );

  getUser(
    request: qf_qf_pb.Void,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.User>;

  getUser(
    request: qf_qf_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.User) => void): grpcWeb.ClientReadableStream<qf_qf_pb.User>;

  getUser(
    request: qf_qf_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.User) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/GetUser',
        request,
        metadata || {},
        this.methodInfoGetUser,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/GetUser',
    request,
    metadata || {},
    this.methodInfoGetUser);
  }

  methodInfoGetUsers = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/GetUsers',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.Void,
    qf_qf_pb.Users,
    (request: qf_qf_pb.Void) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Users.deserializeBinary
  );

  getUsers(
    request: qf_qf_pb.Void,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Users>;

  getUsers(
    request: qf_qf_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Users) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Users>;

  getUsers(
    request: qf_qf_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Users) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/GetUsers',
        request,
        metadata || {},
        this.methodInfoGetUsers,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/GetUsers',
    request,
    metadata || {},
    this.methodInfoGetUsers);
  }

  methodInfoGetUserByCourse = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/GetUserByCourse',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.CourseUserRequest,
    qf_qf_pb.User,
    (request: qf_qf_pb.CourseUserRequest) => {
      return request.serializeBinary();
    },
    qf_qf_pb.User.deserializeBinary
  );

  getUserByCourse(
    request: qf_qf_pb.CourseUserRequest,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.User>;

  getUserByCourse(
    request: qf_qf_pb.CourseUserRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.User) => void): grpcWeb.ClientReadableStream<qf_qf_pb.User>;

  getUserByCourse(
    request: qf_qf_pb.CourseUserRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.User) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/GetUserByCourse',
        request,
        metadata || {},
        this.methodInfoGetUserByCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/GetUserByCourse',
    request,
    metadata || {},
    this.methodInfoGetUserByCourse);
  }

  methodInfoUpdateUser = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/UpdateUser',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.User,
    qf_qf_pb.Void,
    (request: qf_qf_pb.User) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Void.deserializeBinary
  );

  updateUser(
    request: qf_qf_pb.User,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Void>;

  updateUser(
    request: qf_qf_pb.User,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Void>;

  updateUser(
    request: qf_qf_pb.User,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/UpdateUser',
        request,
        metadata || {},
        this.methodInfoUpdateUser,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/UpdateUser',
    request,
    metadata || {},
    this.methodInfoUpdateUser);
  }

  methodInfoIsAuthorizedTeacher = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/IsAuthorizedTeacher',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.Void,
    qf_qf_pb.AuthorizationResponse,
    (request: qf_qf_pb.Void) => {
      return request.serializeBinary();
    },
    qf_qf_pb.AuthorizationResponse.deserializeBinary
  );

  isAuthorizedTeacher(
    request: qf_qf_pb.Void,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.AuthorizationResponse>;

  isAuthorizedTeacher(
    request: qf_qf_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.AuthorizationResponse) => void): grpcWeb.ClientReadableStream<qf_qf_pb.AuthorizationResponse>;

  isAuthorizedTeacher(
    request: qf_qf_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.AuthorizationResponse) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/IsAuthorizedTeacher',
        request,
        metadata || {},
        this.methodInfoIsAuthorizedTeacher,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/IsAuthorizedTeacher',
    request,
    metadata || {},
    this.methodInfoIsAuthorizedTeacher);
  }

  methodInfoGetGroup = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/GetGroup',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.GetGroupRequest,
    qf_qf_pb.Group,
    (request: qf_qf_pb.GetGroupRequest) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Group.deserializeBinary
  );

  getGroup(
    request: qf_qf_pb.GetGroupRequest,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Group>;

  getGroup(
    request: qf_qf_pb.GetGroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Group) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Group>;

  getGroup(
    request: qf_qf_pb.GetGroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Group) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/GetGroup',
        request,
        metadata || {},
        this.methodInfoGetGroup,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/GetGroup',
    request,
    metadata || {},
    this.methodInfoGetGroup);
  }

  methodInfoGetGroupByUserAndCourse = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/GetGroupByUserAndCourse',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.GroupRequest,
    qf_qf_pb.Group,
    (request: qf_qf_pb.GroupRequest) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Group.deserializeBinary
  );

  getGroupByUserAndCourse(
    request: qf_qf_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Group>;

  getGroupByUserAndCourse(
    request: qf_qf_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Group) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Group>;

  getGroupByUserAndCourse(
    request: qf_qf_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Group) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/GetGroupByUserAndCourse',
        request,
        metadata || {},
        this.methodInfoGetGroupByUserAndCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/GetGroupByUserAndCourse',
    request,
    metadata || {},
    this.methodInfoGetGroupByUserAndCourse);
  }

  methodInfoGetGroupsByCourse = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/GetGroupsByCourse',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.CourseRequest,
    qf_qf_pb.Groups,
    (request: qf_qf_pb.CourseRequest) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Groups.deserializeBinary
  );

  getGroupsByCourse(
    request: qf_qf_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Groups>;

  getGroupsByCourse(
    request: qf_qf_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Groups) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Groups>;

  getGroupsByCourse(
    request: qf_qf_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Groups) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/GetGroupsByCourse',
        request,
        metadata || {},
        this.methodInfoGetGroupsByCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/GetGroupsByCourse',
    request,
    metadata || {},
    this.methodInfoGetGroupsByCourse);
  }

  methodInfoCreateGroup = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/CreateGroup',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.Group,
    qf_qf_pb.Group,
    (request: qf_qf_pb.Group) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Group.deserializeBinary
  );

  createGroup(
    request: qf_qf_pb.Group,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Group>;

  createGroup(
    request: qf_qf_pb.Group,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Group) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Group>;

  createGroup(
    request: qf_qf_pb.Group,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Group) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/CreateGroup',
        request,
        metadata || {},
        this.methodInfoCreateGroup,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/CreateGroup',
    request,
    metadata || {},
    this.methodInfoCreateGroup);
  }

  methodInfoUpdateGroup = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/UpdateGroup',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.Group,
    qf_qf_pb.Group,
    (request: qf_qf_pb.Group) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Group.deserializeBinary
  );

  updateGroup(
    request: qf_qf_pb.Group,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Group>;

  updateGroup(
    request: qf_qf_pb.Group,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Group) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Group>;

  updateGroup(
    request: qf_qf_pb.Group,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Group) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/UpdateGroup',
        request,
        metadata || {},
        this.methodInfoUpdateGroup,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/UpdateGroup',
    request,
    metadata || {},
    this.methodInfoUpdateGroup);
  }

  methodInfoDeleteGroup = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/DeleteGroup',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.GroupRequest,
    qf_qf_pb.Void,
    (request: qf_qf_pb.GroupRequest) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Void.deserializeBinary
  );

  deleteGroup(
    request: qf_qf_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Void>;

  deleteGroup(
    request: qf_qf_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Void>;

  deleteGroup(
    request: qf_qf_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/DeleteGroup',
        request,
        metadata || {},
        this.methodInfoDeleteGroup,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/DeleteGroup',
    request,
    metadata || {},
    this.methodInfoDeleteGroup);
  }

  methodInfoGetCourse = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/GetCourse',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.CourseRequest,
    qf_qf_pb.Course,
    (request: qf_qf_pb.CourseRequest) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Course.deserializeBinary
  );

  getCourse(
    request: qf_qf_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Course>;

  getCourse(
    request: qf_qf_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Course) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Course>;

  getCourse(
    request: qf_qf_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Course) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/GetCourse',
        request,
        metadata || {},
        this.methodInfoGetCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/GetCourse',
    request,
    metadata || {},
    this.methodInfoGetCourse);
  }

  methodInfoGetCourses = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/GetCourses',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.Void,
    qf_qf_pb.Courses,
    (request: qf_qf_pb.Void) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Courses.deserializeBinary
  );

  getCourses(
    request: qf_qf_pb.Void,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Courses>;

  getCourses(
    request: qf_qf_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Courses) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Courses>;

  getCourses(
    request: qf_qf_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Courses) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/GetCourses',
        request,
        metadata || {},
        this.methodInfoGetCourses,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/GetCourses',
    request,
    metadata || {},
    this.methodInfoGetCourses);
  }

  methodInfoGetCoursesByUser = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/GetCoursesByUser',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.EnrollmentStatusRequest,
    qf_qf_pb.Courses,
    (request: qf_qf_pb.EnrollmentStatusRequest) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Courses.deserializeBinary
  );

  getCoursesByUser(
    request: qf_qf_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Courses>;

  getCoursesByUser(
    request: qf_qf_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Courses) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Courses>;

  getCoursesByUser(
    request: qf_qf_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Courses) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/GetCoursesByUser',
        request,
        metadata || {},
        this.methodInfoGetCoursesByUser,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/GetCoursesByUser',
    request,
    metadata || {},
    this.methodInfoGetCoursesByUser);
  }

  methodInfoCreateCourse = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/CreateCourse',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.Course,
    qf_qf_pb.Course,
    (request: qf_qf_pb.Course) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Course.deserializeBinary
  );

  createCourse(
    request: qf_qf_pb.Course,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Course>;

  createCourse(
    request: qf_qf_pb.Course,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Course) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Course>;

  createCourse(
    request: qf_qf_pb.Course,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Course) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/CreateCourse',
        request,
        metadata || {},
        this.methodInfoCreateCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/CreateCourse',
    request,
    metadata || {},
    this.methodInfoCreateCourse);
  }

  methodInfoUpdateCourse = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/UpdateCourse',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.Course,
    qf_qf_pb.Void,
    (request: qf_qf_pb.Course) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Void.deserializeBinary
  );

  updateCourse(
    request: qf_qf_pb.Course,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Void>;

  updateCourse(
    request: qf_qf_pb.Course,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Void>;

  updateCourse(
    request: qf_qf_pb.Course,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/UpdateCourse',
        request,
        metadata || {},
        this.methodInfoUpdateCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/UpdateCourse',
    request,
    metadata || {},
    this.methodInfoUpdateCourse);
  }

  methodInfoUpdateCourseVisibility = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/UpdateCourseVisibility',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.Enrollment,
    qf_qf_pb.Void,
    (request: qf_qf_pb.Enrollment) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Void.deserializeBinary
  );

  updateCourseVisibility(
    request: qf_qf_pb.Enrollment,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Void>;

  updateCourseVisibility(
    request: qf_qf_pb.Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Void>;

  updateCourseVisibility(
    request: qf_qf_pb.Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/UpdateCourseVisibility',
        request,
        metadata || {},
        this.methodInfoUpdateCourseVisibility,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/UpdateCourseVisibility',
    request,
    metadata || {},
    this.methodInfoUpdateCourseVisibility);
  }

  methodInfoGetAssignments = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/GetAssignments',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.CourseRequest,
    qf_qf_pb.Assignments,
    (request: qf_qf_pb.CourseRequest) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Assignments.deserializeBinary
  );

  getAssignments(
    request: qf_qf_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Assignments>;

  getAssignments(
    request: qf_qf_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Assignments) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Assignments>;

  getAssignments(
    request: qf_qf_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Assignments) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/GetAssignments',
        request,
        metadata || {},
        this.methodInfoGetAssignments,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/GetAssignments',
    request,
    metadata || {},
    this.methodInfoGetAssignments);
  }

  methodInfoUpdateAssignments = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/UpdateAssignments',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.CourseRequest,
    qf_qf_pb.Void,
    (request: qf_qf_pb.CourseRequest) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Void.deserializeBinary
  );

  updateAssignments(
    request: qf_qf_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Void>;

  updateAssignments(
    request: qf_qf_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Void>;

  updateAssignments(
    request: qf_qf_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/UpdateAssignments',
        request,
        metadata || {},
        this.methodInfoUpdateAssignments,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/UpdateAssignments',
    request,
    metadata || {},
    this.methodInfoUpdateAssignments);
  }

  methodInfoGetEnrollmentsByUser = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/GetEnrollmentsByUser',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.EnrollmentStatusRequest,
    qf_qf_pb.Enrollments,
    (request: qf_qf_pb.EnrollmentStatusRequest) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Enrollments.deserializeBinary
  );

  getEnrollmentsByUser(
    request: qf_qf_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Enrollments>;

  getEnrollmentsByUser(
    request: qf_qf_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Enrollments) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Enrollments>;

  getEnrollmentsByUser(
    request: qf_qf_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Enrollments) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/GetEnrollmentsByUser',
        request,
        metadata || {},
        this.methodInfoGetEnrollmentsByUser,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/GetEnrollmentsByUser',
    request,
    metadata || {},
    this.methodInfoGetEnrollmentsByUser);
  }

  methodInfoGetEnrollmentsByCourse = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/GetEnrollmentsByCourse',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.EnrollmentRequest,
    qf_qf_pb.Enrollments,
    (request: qf_qf_pb.EnrollmentRequest) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Enrollments.deserializeBinary
  );

  getEnrollmentsByCourse(
    request: qf_qf_pb.EnrollmentRequest,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Enrollments>;

  getEnrollmentsByCourse(
    request: qf_qf_pb.EnrollmentRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Enrollments) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Enrollments>;

  getEnrollmentsByCourse(
    request: qf_qf_pb.EnrollmentRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Enrollments) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/GetEnrollmentsByCourse',
        request,
        metadata || {},
        this.methodInfoGetEnrollmentsByCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/GetEnrollmentsByCourse',
    request,
    metadata || {},
    this.methodInfoGetEnrollmentsByCourse);
  }

  methodInfoCreateEnrollment = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/CreateEnrollment',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.Enrollment,
    qf_qf_pb.Void,
    (request: qf_qf_pb.Enrollment) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Void.deserializeBinary
  );

  createEnrollment(
    request: qf_qf_pb.Enrollment,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Void>;

  createEnrollment(
    request: qf_qf_pb.Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Void>;

  createEnrollment(
    request: qf_qf_pb.Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/CreateEnrollment',
        request,
        metadata || {},
        this.methodInfoCreateEnrollment,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/CreateEnrollment',
    request,
    metadata || {},
    this.methodInfoCreateEnrollment);
  }

  methodInfoUpdateEnrollments = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/UpdateEnrollments',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.Enrollments,
    qf_qf_pb.Void,
    (request: qf_qf_pb.Enrollments) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Void.deserializeBinary
  );

  updateEnrollments(
    request: qf_qf_pb.Enrollments,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Void>;

  updateEnrollments(
    request: qf_qf_pb.Enrollments,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Void>;

  updateEnrollments(
    request: qf_qf_pb.Enrollments,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/UpdateEnrollments',
        request,
        metadata || {},
        this.methodInfoUpdateEnrollments,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/UpdateEnrollments',
    request,
    metadata || {},
    this.methodInfoUpdateEnrollments);
  }

  methodInfoGetSubmissions = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/GetSubmissions',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.SubmissionRequest,
    qf_qf_pb.Submissions,
    (request: qf_qf_pb.SubmissionRequest) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Submissions.deserializeBinary
  );

  getSubmissions(
    request: qf_qf_pb.SubmissionRequest,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Submissions>;

  getSubmissions(
    request: qf_qf_pb.SubmissionRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Submissions) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Submissions>;

  getSubmissions(
    request: qf_qf_pb.SubmissionRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Submissions) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/GetSubmissions',
        request,
        metadata || {},
        this.methodInfoGetSubmissions,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/GetSubmissions',
    request,
    metadata || {},
    this.methodInfoGetSubmissions);
  }

  methodInfoGetSubmissionsByCourse = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/GetSubmissionsByCourse',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.SubmissionsForCourseRequest,
    qf_qf_pb.CourseSubmissions,
    (request: qf_qf_pb.SubmissionsForCourseRequest) => {
      return request.serializeBinary();
    },
    qf_qf_pb.CourseSubmissions.deserializeBinary
  );

  getSubmissionsByCourse(
    request: qf_qf_pb.SubmissionsForCourseRequest,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.CourseSubmissions>;

  getSubmissionsByCourse(
    request: qf_qf_pb.SubmissionsForCourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.CourseSubmissions) => void): grpcWeb.ClientReadableStream<qf_qf_pb.CourseSubmissions>;

  getSubmissionsByCourse(
    request: qf_qf_pb.SubmissionsForCourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.CourseSubmissions) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/GetSubmissionsByCourse',
        request,
        metadata || {},
        this.methodInfoGetSubmissionsByCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/GetSubmissionsByCourse',
    request,
    metadata || {},
    this.methodInfoGetSubmissionsByCourse);
  }

  methodInfoUpdateSubmission = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/UpdateSubmission',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.UpdateSubmissionRequest,
    qf_qf_pb.Void,
    (request: qf_qf_pb.UpdateSubmissionRequest) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Void.deserializeBinary
  );

  updateSubmission(
    request: qf_qf_pb.UpdateSubmissionRequest,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Void>;

  updateSubmission(
    request: qf_qf_pb.UpdateSubmissionRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Void>;

  updateSubmission(
    request: qf_qf_pb.UpdateSubmissionRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/UpdateSubmission',
        request,
        metadata || {},
        this.methodInfoUpdateSubmission,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/UpdateSubmission',
    request,
    metadata || {},
    this.methodInfoUpdateSubmission);
  }

  methodInfoUpdateSubmissions = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/UpdateSubmissions',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.UpdateSubmissionsRequest,
    qf_qf_pb.Void,
    (request: qf_qf_pb.UpdateSubmissionsRequest) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Void.deserializeBinary
  );

  updateSubmissions(
    request: qf_qf_pb.UpdateSubmissionsRequest,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Void>;

  updateSubmissions(
    request: qf_qf_pb.UpdateSubmissionsRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Void>;

  updateSubmissions(
    request: qf_qf_pb.UpdateSubmissionsRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/UpdateSubmissions',
        request,
        metadata || {},
        this.methodInfoUpdateSubmissions,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/UpdateSubmissions',
    request,
    metadata || {},
    this.methodInfoUpdateSubmissions);
  }

  methodInfoRebuildSubmissions = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/RebuildSubmissions',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.RebuildRequest,
    qf_qf_pb.Void,
    (request: qf_qf_pb.RebuildRequest) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Void.deserializeBinary
  );

  rebuildSubmissions(
    request: qf_qf_pb.RebuildRequest,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Void>;

  rebuildSubmissions(
    request: qf_qf_pb.RebuildRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Void>;

  rebuildSubmissions(
    request: qf_qf_pb.RebuildRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/RebuildSubmissions',
        request,
        metadata || {},
        this.methodInfoRebuildSubmissions,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/RebuildSubmissions',
    request,
    metadata || {},
    this.methodInfoRebuildSubmissions);
  }

  methodInfoCreateBenchmark = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/CreateBenchmark',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.GradingBenchmark,
    qf_qf_pb.GradingBenchmark,
    (request: qf_qf_pb.GradingBenchmark) => {
      return request.serializeBinary();
    },
    qf_qf_pb.GradingBenchmark.deserializeBinary
  );

  createBenchmark(
    request: qf_qf_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.GradingBenchmark>;

  createBenchmark(
    request: qf_qf_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.GradingBenchmark) => void): grpcWeb.ClientReadableStream<qf_qf_pb.GradingBenchmark>;

  createBenchmark(
    request: qf_qf_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.GradingBenchmark) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/CreateBenchmark',
        request,
        metadata || {},
        this.methodInfoCreateBenchmark,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/CreateBenchmark',
    request,
    metadata || {},
    this.methodInfoCreateBenchmark);
  }

  methodInfoUpdateBenchmark = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/UpdateBenchmark',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.GradingBenchmark,
    qf_qf_pb.Void,
    (request: qf_qf_pb.GradingBenchmark) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Void.deserializeBinary
  );

  updateBenchmark(
    request: qf_qf_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Void>;

  updateBenchmark(
    request: qf_qf_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Void>;

  updateBenchmark(
    request: qf_qf_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/UpdateBenchmark',
        request,
        metadata || {},
        this.methodInfoUpdateBenchmark,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/UpdateBenchmark',
    request,
    metadata || {},
    this.methodInfoUpdateBenchmark);
  }

  methodInfoDeleteBenchmark = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/DeleteBenchmark',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.GradingBenchmark,
    qf_qf_pb.Void,
    (request: qf_qf_pb.GradingBenchmark) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Void.deserializeBinary
  );

  deleteBenchmark(
    request: qf_qf_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Void>;

  deleteBenchmark(
    request: qf_qf_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Void>;

  deleteBenchmark(
    request: qf_qf_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/DeleteBenchmark',
        request,
        metadata || {},
        this.methodInfoDeleteBenchmark,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/DeleteBenchmark',
    request,
    metadata || {},
    this.methodInfoDeleteBenchmark);
  }

  methodInfoCreateCriterion = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/CreateCriterion',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.GradingCriterion,
    qf_qf_pb.GradingCriterion,
    (request: qf_qf_pb.GradingCriterion) => {
      return request.serializeBinary();
    },
    qf_qf_pb.GradingCriterion.deserializeBinary
  );

  createCriterion(
    request: qf_qf_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.GradingCriterion>;

  createCriterion(
    request: qf_qf_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.GradingCriterion) => void): grpcWeb.ClientReadableStream<qf_qf_pb.GradingCriterion>;

  createCriterion(
    request: qf_qf_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.GradingCriterion) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/CreateCriterion',
        request,
        metadata || {},
        this.methodInfoCreateCriterion,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/CreateCriterion',
    request,
    metadata || {},
    this.methodInfoCreateCriterion);
  }

  methodInfoUpdateCriterion = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/UpdateCriterion',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.GradingCriterion,
    qf_qf_pb.Void,
    (request: qf_qf_pb.GradingCriterion) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Void.deserializeBinary
  );

  updateCriterion(
    request: qf_qf_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Void>;

  updateCriterion(
    request: qf_qf_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Void>;

  updateCriterion(
    request: qf_qf_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/UpdateCriterion',
        request,
        metadata || {},
        this.methodInfoUpdateCriterion,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/UpdateCriterion',
    request,
    metadata || {},
    this.methodInfoUpdateCriterion);
  }

  methodInfoDeleteCriterion = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/DeleteCriterion',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.GradingCriterion,
    qf_qf_pb.Void,
    (request: qf_qf_pb.GradingCriterion) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Void.deserializeBinary
  );

  deleteCriterion(
    request: qf_qf_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Void>;

  deleteCriterion(
    request: qf_qf_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Void>;

  deleteCriterion(
    request: qf_qf_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/DeleteCriterion',
        request,
        metadata || {},
        this.methodInfoDeleteCriterion,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/DeleteCriterion',
    request,
    metadata || {},
    this.methodInfoDeleteCriterion);
  }

  methodInfoCreateReview = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/CreateReview',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.ReviewRequest,
    qf_qf_pb.Review,
    (request: qf_qf_pb.ReviewRequest) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Review.deserializeBinary
  );

  createReview(
    request: qf_qf_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Review>;

  createReview(
    request: qf_qf_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Review) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Review>;

  createReview(
    request: qf_qf_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Review) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/CreateReview',
        request,
        metadata || {},
        this.methodInfoCreateReview,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/CreateReview',
    request,
    metadata || {},
    this.methodInfoCreateReview);
  }

  methodInfoUpdateReview = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/UpdateReview',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.ReviewRequest,
    qf_qf_pb.Review,
    (request: qf_qf_pb.ReviewRequest) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Review.deserializeBinary
  );

  updateReview(
    request: qf_qf_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Review>;

  updateReview(
    request: qf_qf_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Review) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Review>;

  updateReview(
    request: qf_qf_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Review) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/UpdateReview',
        request,
        metadata || {},
        this.methodInfoUpdateReview,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/UpdateReview',
    request,
    metadata || {},
    this.methodInfoUpdateReview);
  }

  methodInfoGetReviewers = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/GetReviewers',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.SubmissionReviewersRequest,
    qf_qf_pb.Reviewers,
    (request: qf_qf_pb.SubmissionReviewersRequest) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Reviewers.deserializeBinary
  );

  getReviewers(
    request: qf_qf_pb.SubmissionReviewersRequest,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Reviewers>;

  getReviewers(
    request: qf_qf_pb.SubmissionReviewersRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Reviewers) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Reviewers>;

  getReviewers(
    request: qf_qf_pb.SubmissionReviewersRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Reviewers) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/GetReviewers',
        request,
        metadata || {},
        this.methodInfoGetReviewers,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/GetReviewers',
    request,
    metadata || {},
    this.methodInfoGetReviewers);
  }

  methodInfoGetProviders = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/GetProviders',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.Void,
    qf_qf_pb.Providers,
    (request: qf_qf_pb.Void) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Providers.deserializeBinary
  );

  getProviders(
    request: qf_qf_pb.Void,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Providers>;

  getProviders(
    request: qf_qf_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Providers) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Providers>;

  getProviders(
    request: qf_qf_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Providers) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/GetProviders',
        request,
        metadata || {},
        this.methodInfoGetProviders,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/GetProviders',
    request,
    metadata || {},
    this.methodInfoGetProviders);
  }

  methodInfoGetOrganization = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/GetOrganization',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.OrgRequest,
    qf_qf_pb.Organization,
    (request: qf_qf_pb.OrgRequest) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Organization.deserializeBinary
  );

  getOrganization(
    request: qf_qf_pb.OrgRequest,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Organization>;

  getOrganization(
    request: qf_qf_pb.OrgRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Organization) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Organization>;

  getOrganization(
    request: qf_qf_pb.OrgRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Organization) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/GetOrganization',
        request,
        metadata || {},
        this.methodInfoGetOrganization,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/GetOrganization',
    request,
    metadata || {},
    this.methodInfoGetOrganization);
  }

  methodInfoGetRepositories = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/GetRepositories',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.URLRequest,
    qf_qf_pb.Repositories,
    (request: qf_qf_pb.URLRequest) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Repositories.deserializeBinary
  );

  getRepositories(
    request: qf_qf_pb.URLRequest,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Repositories>;

  getRepositories(
    request: qf_qf_pb.URLRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Repositories) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Repositories>;

  getRepositories(
    request: qf_qf_pb.URLRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Repositories) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/GetRepositories',
        request,
        metadata || {},
        this.methodInfoGetRepositories,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/GetRepositories',
    request,
    metadata || {},
    this.methodInfoGetRepositories);
  }

  methodInfoIsEmptyRepo = new grpcWeb.MethodDescriptor(
    '/qf.QuickFeedService/IsEmptyRepo',
    grpcWeb.MethodType.UNARY,
    qf_qf_pb.RepositoryRequest,
    qf_qf_pb.Void,
    (request: qf_qf_pb.RepositoryRequest) => {
      return request.serializeBinary();
    },
    qf_qf_pb.Void.deserializeBinary
  );

  isEmptyRepo(
    request: qf_qf_pb.RepositoryRequest,
    metadata: grpcWeb.Metadata | null): Promise<qf_qf_pb.Void>;

  isEmptyRepo(
    request: qf_qf_pb.RepositoryRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void): grpcWeb.ClientReadableStream<qf_qf_pb.Void>;

  isEmptyRepo(
    request: qf_qf_pb.RepositoryRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: qf_qf_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/qf.QuickFeedService/IsEmptyRepo',
        request,
        metadata || {},
        this.methodInfoIsEmptyRepo,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/qf.QuickFeedService/IsEmptyRepo',
    request,
    metadata || {},
    this.methodInfoIsEmptyRepo);
  }

}

