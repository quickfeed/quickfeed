/**
 * @fileoverview gRPC-Web generated client stub for ag
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck


import * as grpcWeb from 'grpc-web';

import * as ag_types_types_pb from '../ag/types/types_pb';
import * as ag_types_requests_pb from '../ag/types/requests_pb';


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
    options['format'] = 'text';

    this.client_ = new grpcWeb.GrpcWebClientBase(options);
    this.hostname_ = hostname;
    this.credentials_ = credentials;
    this.options_ = options;
  }

  methodInfoGetUser = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetUser',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.Void,
    ag_types_types_pb.User,
    (request: ag_types_requests_pb.Void) => {
      return request.serializeBinary();
    },
    ag_types_types_pb.User.deserializeBinary
  );

  getUser(
    request: ag_types_requests_pb.Void,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_types_pb.User>;

  getUser(
    request: ag_types_requests_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.User) => void): grpcWeb.ClientReadableStream<ag_types_types_pb.User>;

  getUser(
    request: ag_types_requests_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.User) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetUser',
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

  methodInfoGetUsers = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetUsers',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.Void,
    ag_types_types_pb.Users,
    (request: ag_types_requests_pb.Void) => {
      return request.serializeBinary();
    },
    ag_types_types_pb.Users.deserializeBinary
  );

  getUsers(
    request: ag_types_requests_pb.Void,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_types_pb.Users>;

  getUsers(
    request: ag_types_requests_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Users) => void): grpcWeb.ClientReadableStream<ag_types_types_pb.Users>;

  getUsers(
    request: ag_types_requests_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Users) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetUsers',
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

  methodInfoGetUserByCourse = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetUserByCourse',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.CourseUserRequest,
    ag_types_types_pb.User,
    (request: ag_types_requests_pb.CourseUserRequest) => {
      return request.serializeBinary();
    },
    ag_types_types_pb.User.deserializeBinary
  );

  getUserByCourse(
    request: ag_types_requests_pb.CourseUserRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_types_pb.User>;

  getUserByCourse(
    request: ag_types_requests_pb.CourseUserRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.User) => void): grpcWeb.ClientReadableStream<ag_types_types_pb.User>;

  getUserByCourse(
    request: ag_types_requests_pb.CourseUserRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.User) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetUserByCourse',
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

  methodInfoUpdateUser = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateUser',
    grpcWeb.MethodType.UNARY,
    ag_types_types_pb.User,
    ag_types_requests_pb.Void,
    (request: ag_types_types_pb.User) => {
      return request.serializeBinary();
    },
    ag_types_requests_pb.Void.deserializeBinary
  );

  updateUser(
    request: ag_types_types_pb.User,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_requests_pb.Void>;

  updateUser(
    request: ag_types_types_pb.User,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void): grpcWeb.ClientReadableStream<ag_types_requests_pb.Void>;

  updateUser(
    request: ag_types_types_pb.User,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateUser',
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

  methodInfoGetGroup = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetGroup',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.GetGroupRequest,
    ag_types_types_pb.Group,
    (request: ag_types_requests_pb.GetGroupRequest) => {
      return request.serializeBinary();
    },
    ag_types_types_pb.Group.deserializeBinary
  );

  getGroup(
    request: ag_types_requests_pb.GetGroupRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_types_pb.Group>;

  getGroup(
    request: ag_types_requests_pb.GetGroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Group) => void): grpcWeb.ClientReadableStream<ag_types_types_pb.Group>;

  getGroup(
    request: ag_types_requests_pb.GetGroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Group) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetGroup',
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

  methodInfoGetGroupByUserAndCourse = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetGroupByUserAndCourse',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.GroupRequest,
    ag_types_types_pb.Group,
    (request: ag_types_requests_pb.GroupRequest) => {
      return request.serializeBinary();
    },
    ag_types_types_pb.Group.deserializeBinary
  );

  getGroupByUserAndCourse(
    request: ag_types_requests_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_types_pb.Group>;

  getGroupByUserAndCourse(
    request: ag_types_requests_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Group) => void): grpcWeb.ClientReadableStream<ag_types_types_pb.Group>;

  getGroupByUserAndCourse(
    request: ag_types_requests_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Group) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetGroupByUserAndCourse',
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

  methodInfoGetGroupsByCourse = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetGroupsByCourse',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.CourseRequest,
    ag_types_types_pb.Groups,
    (request: ag_types_requests_pb.CourseRequest) => {
      return request.serializeBinary();
    },
    ag_types_types_pb.Groups.deserializeBinary
  );

  getGroupsByCourse(
    request: ag_types_requests_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_types_pb.Groups>;

  getGroupsByCourse(
    request: ag_types_requests_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Groups) => void): grpcWeb.ClientReadableStream<ag_types_types_pb.Groups>;

  getGroupsByCourse(
    request: ag_types_requests_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Groups) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetGroupsByCourse',
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

  methodInfoCreateGroup = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/CreateGroup',
    grpcWeb.MethodType.UNARY,
    ag_types_types_pb.Group,
    ag_types_types_pb.Group,
    (request: ag_types_types_pb.Group) => {
      return request.serializeBinary();
    },
    ag_types_types_pb.Group.deserializeBinary
  );

  createGroup(
    request: ag_types_types_pb.Group,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_types_pb.Group>;

  createGroup(
    request: ag_types_types_pb.Group,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Group) => void): grpcWeb.ClientReadableStream<ag_types_types_pb.Group>;

  createGroup(
    request: ag_types_types_pb.Group,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Group) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/CreateGroup',
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

  methodInfoUpdateGroup = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateGroup',
    grpcWeb.MethodType.UNARY,
    ag_types_types_pb.Group,
    ag_types_requests_pb.Void,
    (request: ag_types_types_pb.Group) => {
      return request.serializeBinary();
    },
    ag_types_requests_pb.Void.deserializeBinary
  );

  updateGroup(
    request: ag_types_types_pb.Group,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_requests_pb.Void>;

  updateGroup(
    request: ag_types_types_pb.Group,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void): grpcWeb.ClientReadableStream<ag_types_requests_pb.Void>;

  updateGroup(
    request: ag_types_types_pb.Group,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateGroup',
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

  methodInfoDeleteGroup = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/DeleteGroup',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.GroupRequest,
    ag_types_requests_pb.Void,
    (request: ag_types_requests_pb.GroupRequest) => {
      return request.serializeBinary();
    },
    ag_types_requests_pb.Void.deserializeBinary
  );

  deleteGroup(
    request: ag_types_requests_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_requests_pb.Void>;

  deleteGroup(
    request: ag_types_requests_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void): grpcWeb.ClientReadableStream<ag_types_requests_pb.Void>;

  deleteGroup(
    request: ag_types_requests_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/DeleteGroup',
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

  methodInfoGetCourse = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetCourse',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.CourseRequest,
    ag_types_types_pb.Course,
    (request: ag_types_requests_pb.CourseRequest) => {
      return request.serializeBinary();
    },
    ag_types_types_pb.Course.deserializeBinary
  );

  getCourse(
    request: ag_types_requests_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_types_pb.Course>;

  getCourse(
    request: ag_types_requests_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Course) => void): grpcWeb.ClientReadableStream<ag_types_types_pb.Course>;

  getCourse(
    request: ag_types_requests_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Course) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetCourse',
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

  methodInfoGetCourses = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetCourses',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.Void,
    ag_types_types_pb.Courses,
    (request: ag_types_requests_pb.Void) => {
      return request.serializeBinary();
    },
    ag_types_types_pb.Courses.deserializeBinary
  );

  getCourses(
    request: ag_types_requests_pb.Void,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_types_pb.Courses>;

  getCourses(
    request: ag_types_requests_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Courses) => void): grpcWeb.ClientReadableStream<ag_types_types_pb.Courses>;

  getCourses(
    request: ag_types_requests_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Courses) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetCourses',
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

  methodInfoGetCoursesByUser = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetCoursesByUser',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.EnrollmentStatusRequest,
    ag_types_types_pb.Courses,
    (request: ag_types_requests_pb.EnrollmentStatusRequest) => {
      return request.serializeBinary();
    },
    ag_types_types_pb.Courses.deserializeBinary
  );

  getCoursesByUser(
    request: ag_types_requests_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_types_pb.Courses>;

  getCoursesByUser(
    request: ag_types_requests_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Courses) => void): grpcWeb.ClientReadableStream<ag_types_types_pb.Courses>;

  getCoursesByUser(
    request: ag_types_requests_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Courses) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetCoursesByUser',
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

  methodInfoCreateCourse = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/CreateCourse',
    grpcWeb.MethodType.UNARY,
    ag_types_types_pb.Course,
    ag_types_types_pb.Course,
    (request: ag_types_types_pb.Course) => {
      return request.serializeBinary();
    },
    ag_types_types_pb.Course.deserializeBinary
  );

  createCourse(
    request: ag_types_types_pb.Course,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_types_pb.Course>;

  createCourse(
    request: ag_types_types_pb.Course,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Course) => void): grpcWeb.ClientReadableStream<ag_types_types_pb.Course>;

  createCourse(
    request: ag_types_types_pb.Course,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Course) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/CreateCourse',
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

  methodInfoUpdateCourse = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateCourse',
    grpcWeb.MethodType.UNARY,
    ag_types_types_pb.Course,
    ag_types_requests_pb.Void,
    (request: ag_types_types_pb.Course) => {
      return request.serializeBinary();
    },
    ag_types_requests_pb.Void.deserializeBinary
  );

  updateCourse(
    request: ag_types_types_pb.Course,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_requests_pb.Void>;

  updateCourse(
    request: ag_types_types_pb.Course,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void): grpcWeb.ClientReadableStream<ag_types_requests_pb.Void>;

  updateCourse(
    request: ag_types_types_pb.Course,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateCourse',
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

  methodInfoUpdateCourseVisibility = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateCourseVisibility',
    grpcWeb.MethodType.UNARY,
    ag_types_types_pb.Enrollment,
    ag_types_requests_pb.Void,
    (request: ag_types_types_pb.Enrollment) => {
      return request.serializeBinary();
    },
    ag_types_requests_pb.Void.deserializeBinary
  );

  updateCourseVisibility(
    request: ag_types_types_pb.Enrollment,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_requests_pb.Void>;

  updateCourseVisibility(
    request: ag_types_types_pb.Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void): grpcWeb.ClientReadableStream<ag_types_requests_pb.Void>;

  updateCourseVisibility(
    request: ag_types_types_pb.Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateCourseVisibility',
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

  methodInfoGetAssignments = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetAssignments',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.CourseRequest,
    ag_types_types_pb.Assignments,
    (request: ag_types_requests_pb.CourseRequest) => {
      return request.serializeBinary();
    },
    ag_types_types_pb.Assignments.deserializeBinary
  );

  getAssignments(
    request: ag_types_requests_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_types_pb.Assignments>;

  getAssignments(
    request: ag_types_requests_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Assignments) => void): grpcWeb.ClientReadableStream<ag_types_types_pb.Assignments>;

  getAssignments(
    request: ag_types_requests_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Assignments) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetAssignments',
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

  methodInfoUpdateAssignments = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateAssignments',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.CourseRequest,
    ag_types_requests_pb.Void,
    (request: ag_types_requests_pb.CourseRequest) => {
      return request.serializeBinary();
    },
    ag_types_requests_pb.Void.deserializeBinary
  );

  updateAssignments(
    request: ag_types_requests_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_requests_pb.Void>;

  updateAssignments(
    request: ag_types_requests_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void): grpcWeb.ClientReadableStream<ag_types_requests_pb.Void>;

  updateAssignments(
    request: ag_types_requests_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateAssignments',
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

  methodInfoGetEnrollmentsByUser = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetEnrollmentsByUser',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.EnrollmentStatusRequest,
    ag_types_types_pb.Enrollments,
    (request: ag_types_requests_pb.EnrollmentStatusRequest) => {
      return request.serializeBinary();
    },
    ag_types_types_pb.Enrollments.deserializeBinary
  );

  getEnrollmentsByUser(
    request: ag_types_requests_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_types_pb.Enrollments>;

  getEnrollmentsByUser(
    request: ag_types_requests_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Enrollments) => void): grpcWeb.ClientReadableStream<ag_types_types_pb.Enrollments>;

  getEnrollmentsByUser(
    request: ag_types_requests_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Enrollments) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetEnrollmentsByUser',
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

  methodInfoGetEnrollmentsByCourse = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetEnrollmentsByCourse',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.EnrollmentRequest,
    ag_types_types_pb.Enrollments,
    (request: ag_types_requests_pb.EnrollmentRequest) => {
      return request.serializeBinary();
    },
    ag_types_types_pb.Enrollments.deserializeBinary
  );

  getEnrollmentsByCourse(
    request: ag_types_requests_pb.EnrollmentRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_types_pb.Enrollments>;

  getEnrollmentsByCourse(
    request: ag_types_requests_pb.EnrollmentRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Enrollments) => void): grpcWeb.ClientReadableStream<ag_types_types_pb.Enrollments>;

  getEnrollmentsByCourse(
    request: ag_types_requests_pb.EnrollmentRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Enrollments) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetEnrollmentsByCourse',
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

  methodInfoCreateEnrollment = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/CreateEnrollment',
    grpcWeb.MethodType.UNARY,
    ag_types_types_pb.Enrollment,
    ag_types_requests_pb.Void,
    (request: ag_types_types_pb.Enrollment) => {
      return request.serializeBinary();
    },
    ag_types_requests_pb.Void.deserializeBinary
  );

  createEnrollment(
    request: ag_types_types_pb.Enrollment,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_requests_pb.Void>;

  createEnrollment(
    request: ag_types_types_pb.Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void): grpcWeb.ClientReadableStream<ag_types_requests_pb.Void>;

  createEnrollment(
    request: ag_types_types_pb.Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/CreateEnrollment',
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

  methodInfoUpdateEnrollments = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateEnrollments',
    grpcWeb.MethodType.UNARY,
    ag_types_types_pb.Enrollments,
    ag_types_requests_pb.Void,
    (request: ag_types_types_pb.Enrollments) => {
      return request.serializeBinary();
    },
    ag_types_requests_pb.Void.deserializeBinary
  );

  updateEnrollments(
    request: ag_types_types_pb.Enrollments,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_requests_pb.Void>;

  updateEnrollments(
    request: ag_types_types_pb.Enrollments,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void): grpcWeb.ClientReadableStream<ag_types_requests_pb.Void>;

  updateEnrollments(
    request: ag_types_types_pb.Enrollments,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateEnrollments',
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

  methodInfoGetSubmissions = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetSubmissions',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.SubmissionRequest,
    ag_types_types_pb.Submissions,
    (request: ag_types_requests_pb.SubmissionRequest) => {
      return request.serializeBinary();
    },
    ag_types_types_pb.Submissions.deserializeBinary
  );

  getSubmissions(
    request: ag_types_requests_pb.SubmissionRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_types_pb.Submissions>;

  getSubmissions(
    request: ag_types_requests_pb.SubmissionRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Submissions) => void): grpcWeb.ClientReadableStream<ag_types_types_pb.Submissions>;

  getSubmissions(
    request: ag_types_requests_pb.SubmissionRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Submissions) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetSubmissions',
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

  methodInfoGetSubmissionsByCourse = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetSubmissionsByCourse',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.SubmissionsForCourseRequest,
    ag_types_requests_pb.CourseSubmissions,
    (request: ag_types_requests_pb.SubmissionsForCourseRequest) => {
      return request.serializeBinary();
    },
    ag_types_requests_pb.CourseSubmissions.deserializeBinary
  );

  getSubmissionsByCourse(
    request: ag_types_requests_pb.SubmissionsForCourseRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_requests_pb.CourseSubmissions>;

  getSubmissionsByCourse(
    request: ag_types_requests_pb.SubmissionsForCourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.CourseSubmissions) => void): grpcWeb.ClientReadableStream<ag_types_requests_pb.CourseSubmissions>;

  getSubmissionsByCourse(
    request: ag_types_requests_pb.SubmissionsForCourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.CourseSubmissions) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetSubmissionsByCourse',
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

  methodInfoUpdateSubmission = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateSubmission',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.UpdateSubmissionRequest,
    ag_types_requests_pb.Void,
    (request: ag_types_requests_pb.UpdateSubmissionRequest) => {
      return request.serializeBinary();
    },
    ag_types_requests_pb.Void.deserializeBinary
  );

  updateSubmission(
    request: ag_types_requests_pb.UpdateSubmissionRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_requests_pb.Void>;

  updateSubmission(
    request: ag_types_requests_pb.UpdateSubmissionRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void): grpcWeb.ClientReadableStream<ag_types_requests_pb.Void>;

  updateSubmission(
    request: ag_types_requests_pb.UpdateSubmissionRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateSubmission',
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

  methodInfoUpdateSubmissions = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateSubmissions',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.UpdateSubmissionsRequest,
    ag_types_requests_pb.Void,
    (request: ag_types_requests_pb.UpdateSubmissionsRequest) => {
      return request.serializeBinary();
    },
    ag_types_requests_pb.Void.deserializeBinary
  );

  updateSubmissions(
    request: ag_types_requests_pb.UpdateSubmissionsRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_requests_pb.Void>;

  updateSubmissions(
    request: ag_types_requests_pb.UpdateSubmissionsRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void): grpcWeb.ClientReadableStream<ag_types_requests_pb.Void>;

  updateSubmissions(
    request: ag_types_requests_pb.UpdateSubmissionsRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateSubmissions',
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

  methodInfoRebuildSubmissions = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/RebuildSubmissions',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.RebuildRequest,
    ag_types_requests_pb.Void,
    (request: ag_types_requests_pb.RebuildRequest) => {
      return request.serializeBinary();
    },
    ag_types_requests_pb.Void.deserializeBinary
  );

  rebuildSubmissions(
    request: ag_types_requests_pb.RebuildRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_requests_pb.Void>;

  rebuildSubmissions(
    request: ag_types_requests_pb.RebuildRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void): grpcWeb.ClientReadableStream<ag_types_requests_pb.Void>;

  rebuildSubmissions(
    request: ag_types_requests_pb.RebuildRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/RebuildSubmissions',
        request,
        metadata || {},
        this.methodInfoRebuildSubmissions,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/RebuildSubmissions',
    request,
    metadata || {},
    this.methodInfoRebuildSubmissions);
  }

  methodInfoCreateBenchmark = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/CreateBenchmark',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.BenchmarkRequest,
    ag_types_types_pb.GradingBenchmark,
    (request: ag_types_requests_pb.BenchmarkRequest) => {
      return request.serializeBinary();
    },
    ag_types_types_pb.GradingBenchmark.deserializeBinary
  );

  createBenchmark(
    request: ag_types_requests_pb.BenchmarkRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_types_pb.GradingBenchmark>;

  createBenchmark(
    request: ag_types_requests_pb.BenchmarkRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.GradingBenchmark) => void): grpcWeb.ClientReadableStream<ag_types_types_pb.GradingBenchmark>;

  createBenchmark(
    request: ag_types_requests_pb.BenchmarkRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.GradingBenchmark) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/CreateBenchmark',
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

  methodInfoUpdateBenchmark = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateBenchmark',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.BenchmarkRequest,
    ag_types_requests_pb.Void,
    (request: ag_types_requests_pb.BenchmarkRequest) => {
      return request.serializeBinary();
    },
    ag_types_requests_pb.Void.deserializeBinary
  );

  updateBenchmark(
    request: ag_types_requests_pb.BenchmarkRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_requests_pb.Void>;

  updateBenchmark(
    request: ag_types_requests_pb.BenchmarkRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void): grpcWeb.ClientReadableStream<ag_types_requests_pb.Void>;

  updateBenchmark(
    request: ag_types_requests_pb.BenchmarkRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateBenchmark',
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

  methodInfoDeleteBenchmark = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/DeleteBenchmark',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.BenchmarkRequest,
    ag_types_requests_pb.Void,
    (request: ag_types_requests_pb.BenchmarkRequest) => {
      return request.serializeBinary();
    },
    ag_types_requests_pb.Void.deserializeBinary
  );

  deleteBenchmark(
    request: ag_types_requests_pb.BenchmarkRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_requests_pb.Void>;

  deleteBenchmark(
    request: ag_types_requests_pb.BenchmarkRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void): grpcWeb.ClientReadableStream<ag_types_requests_pb.Void>;

  deleteBenchmark(
    request: ag_types_requests_pb.BenchmarkRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/DeleteBenchmark',
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

  methodInfoCreateCriterion = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/CreateCriterion',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.CriteriaRequest,
    ag_types_types_pb.GradingCriterion,
    (request: ag_types_requests_pb.CriteriaRequest) => {
      return request.serializeBinary();
    },
    ag_types_types_pb.GradingCriterion.deserializeBinary
  );

  createCriterion(
    request: ag_types_requests_pb.CriteriaRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_types_pb.GradingCriterion>;

  createCriterion(
    request: ag_types_requests_pb.CriteriaRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.GradingCriterion) => void): grpcWeb.ClientReadableStream<ag_types_types_pb.GradingCriterion>;

  createCriterion(
    request: ag_types_requests_pb.CriteriaRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.GradingCriterion) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/CreateCriterion',
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

  methodInfoUpdateCriterion = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateCriterion',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.CriteriaRequest,
    ag_types_requests_pb.Void,
    (request: ag_types_requests_pb.CriteriaRequest) => {
      return request.serializeBinary();
    },
    ag_types_requests_pb.Void.deserializeBinary
  );

  updateCriterion(
    request: ag_types_requests_pb.CriteriaRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_requests_pb.Void>;

  updateCriterion(
    request: ag_types_requests_pb.CriteriaRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void): grpcWeb.ClientReadableStream<ag_types_requests_pb.Void>;

  updateCriterion(
    request: ag_types_requests_pb.CriteriaRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateCriterion',
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

  methodInfoDeleteCriterion = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/DeleteCriterion',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.CriteriaRequest,
    ag_types_requests_pb.Void,
    (request: ag_types_requests_pb.CriteriaRequest) => {
      return request.serializeBinary();
    },
    ag_types_requests_pb.Void.deserializeBinary
  );

  deleteCriterion(
    request: ag_types_requests_pb.CriteriaRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_requests_pb.Void>;

  deleteCriterion(
    request: ag_types_requests_pb.CriteriaRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void): grpcWeb.ClientReadableStream<ag_types_requests_pb.Void>;

  deleteCriterion(
    request: ag_types_requests_pb.CriteriaRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/DeleteCriterion',
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

  methodInfoCreateReview = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/CreateReview',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.ReviewRequest,
    ag_types_types_pb.Review,
    (request: ag_types_requests_pb.ReviewRequest) => {
      return request.serializeBinary();
    },
    ag_types_types_pb.Review.deserializeBinary
  );

  createReview(
    request: ag_types_requests_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_types_pb.Review>;

  createReview(
    request: ag_types_requests_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Review) => void): grpcWeb.ClientReadableStream<ag_types_types_pb.Review>;

  createReview(
    request: ag_types_requests_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Review) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/CreateReview',
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

  methodInfoUpdateReview = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateReview',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.ReviewRequest,
    ag_types_types_pb.Review,
    (request: ag_types_requests_pb.ReviewRequest) => {
      return request.serializeBinary();
    },
    ag_types_types_pb.Review.deserializeBinary
  );

  updateReview(
    request: ag_types_requests_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_types_pb.Review>;

  updateReview(
    request: ag_types_requests_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Review) => void): grpcWeb.ClientReadableStream<ag_types_types_pb.Review>;

  updateReview(
    request: ag_types_requests_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_types_pb.Review) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateReview',
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

  methodInfoGetReviewers = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetReviewers',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.SubmissionReviewersRequest,
    ag_types_requests_pb.Reviewers,
    (request: ag_types_requests_pb.SubmissionReviewersRequest) => {
      return request.serializeBinary();
    },
    ag_types_requests_pb.Reviewers.deserializeBinary
  );

  getReviewers(
    request: ag_types_requests_pb.SubmissionReviewersRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_requests_pb.Reviewers>;

  getReviewers(
    request: ag_types_requests_pb.SubmissionReviewersRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Reviewers) => void): grpcWeb.ClientReadableStream<ag_types_requests_pb.Reviewers>;

  getReviewers(
    request: ag_types_requests_pb.SubmissionReviewersRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Reviewers) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetReviewers',
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

  methodInfoGetProviders = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetProviders',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.Void,
    ag_types_requests_pb.Providers,
    (request: ag_types_requests_pb.Void) => {
      return request.serializeBinary();
    },
    ag_types_requests_pb.Providers.deserializeBinary
  );

  getProviders(
    request: ag_types_requests_pb.Void,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_requests_pb.Providers>;

  getProviders(
    request: ag_types_requests_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Providers) => void): grpcWeb.ClientReadableStream<ag_types_requests_pb.Providers>;

  getProviders(
    request: ag_types_requests_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Providers) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetProviders',
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

  methodInfoGetOrganization = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetOrganization',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.OrgRequest,
    ag_types_requests_pb.Organization,
    (request: ag_types_requests_pb.OrgRequest) => {
      return request.serializeBinary();
    },
    ag_types_requests_pb.Organization.deserializeBinary
  );

  getOrganization(
    request: ag_types_requests_pb.OrgRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_requests_pb.Organization>;

  getOrganization(
    request: ag_types_requests_pb.OrgRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Organization) => void): grpcWeb.ClientReadableStream<ag_types_requests_pb.Organization>;

  getOrganization(
    request: ag_types_requests_pb.OrgRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Organization) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetOrganization',
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

  methodInfoGetRepositories = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetRepositories',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.URLRequest,
    ag_types_requests_pb.Repositories,
    (request: ag_types_requests_pb.URLRequest) => {
      return request.serializeBinary();
    },
    ag_types_requests_pb.Repositories.deserializeBinary
  );

  getRepositories(
    request: ag_types_requests_pb.URLRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_requests_pb.Repositories>;

  getRepositories(
    request: ag_types_requests_pb.URLRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Repositories) => void): grpcWeb.ClientReadableStream<ag_types_requests_pb.Repositories>;

  getRepositories(
    request: ag_types_requests_pb.URLRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Repositories) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetRepositories',
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

  methodInfoIsEmptyRepo = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/IsEmptyRepo',
    grpcWeb.MethodType.UNARY,
    ag_types_requests_pb.RepositoryRequest,
    ag_types_requests_pb.Void,
    (request: ag_types_requests_pb.RepositoryRequest) => {
      return request.serializeBinary();
    },
    ag_types_requests_pb.Void.deserializeBinary
  );

  isEmptyRepo(
    request: ag_types_requests_pb.RepositoryRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_types_requests_pb.Void>;

  isEmptyRepo(
    request: ag_types_requests_pb.RepositoryRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void): grpcWeb.ClientReadableStream<ag_types_requests_pb.Void>;

  isEmptyRepo(
    request: ag_types_requests_pb.RepositoryRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_types_requests_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/IsEmptyRepo',
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

