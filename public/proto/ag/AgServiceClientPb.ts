/**
 * @fileoverview gRPC-Web generated client stub for ag
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck


import * as grpcWeb from 'grpc-web';

import * as ag_ag_pb from '../ag/ag_pb';


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

  methodDescriptorGetUser = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetUser',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.Void,
    ag_ag_pb.User,
    (request: ag_ag_pb.Void) => {
      return request.serializeBinary();
    },
    ag_ag_pb.User.deserializeBinary
  );

  getUser(
    request: ag_ag_pb.Void,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.User>;

  getUser(
    request: ag_ag_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.User) => void): grpcWeb.ClientReadableStream<ag_ag_pb.User>;

  getUser(
    request: ag_ag_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.User) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetUser',
        request,
        metadata || {},
        this.methodDescriptorGetUser,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetUser',
    request,
    metadata || {},
    this.methodDescriptorGetUser);
  }

  methodDescriptorGetUsers = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetUsers',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.Void,
    ag_ag_pb.Users,
    (request: ag_ag_pb.Void) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Users.deserializeBinary
  );

  getUsers(
    request: ag_ag_pb.Void,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Users>;

  getUsers(
    request: ag_ag_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Users) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Users>;

  getUsers(
    request: ag_ag_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Users) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetUsers',
        request,
        metadata || {},
        this.methodDescriptorGetUsers,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetUsers',
    request,
    metadata || {},
    this.methodDescriptorGetUsers);
  }

  methodDescriptorGetUserByCourse = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetUserByCourse',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.CourseUserRequest,
    ag_ag_pb.User,
    (request: ag_ag_pb.CourseUserRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.User.deserializeBinary
  );

  getUserByCourse(
    request: ag_ag_pb.CourseUserRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.User>;

  getUserByCourse(
    request: ag_ag_pb.CourseUserRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.User) => void): grpcWeb.ClientReadableStream<ag_ag_pb.User>;

  getUserByCourse(
    request: ag_ag_pb.CourseUserRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.User) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetUserByCourse',
        request,
        metadata || {},
        this.methodDescriptorGetUserByCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetUserByCourse',
    request,
    metadata || {},
    this.methodDescriptorGetUserByCourse);
  }

  methodDescriptorUpdateUser = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateUser',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.User,
    ag_ag_pb.Void,
    (request: ag_ag_pb.User) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Void.deserializeBinary
  );

  updateUser(
    request: ag_ag_pb.User,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Void>;

  updateUser(
    request: ag_ag_pb.User,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Void>;

  updateUser(
    request: ag_ag_pb.User,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateUser',
        request,
        metadata || {},
        this.methodDescriptorUpdateUser,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateUser',
    request,
    metadata || {},
    this.methodDescriptorUpdateUser);
  }

  methodDescriptorIsAuthorizedTeacher = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/IsAuthorizedTeacher',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.Void,
    ag_ag_pb.AuthorizationResponse,
    (request: ag_ag_pb.Void) => {
      return request.serializeBinary();
    },
    ag_ag_pb.AuthorizationResponse.deserializeBinary
  );

  isAuthorizedTeacher(
    request: ag_ag_pb.Void,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.AuthorizationResponse>;

  isAuthorizedTeacher(
    request: ag_ag_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.AuthorizationResponse) => void): grpcWeb.ClientReadableStream<ag_ag_pb.AuthorizationResponse>;

  isAuthorizedTeacher(
    request: ag_ag_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.AuthorizationResponse) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/IsAuthorizedTeacher',
        request,
        metadata || {},
        this.methodDescriptorIsAuthorizedTeacher,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/IsAuthorizedTeacher',
    request,
    metadata || {},
    this.methodDescriptorIsAuthorizedTeacher);
  }

  methodDescriptorGetGroup = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetGroup',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.GetGroupRequest,
    ag_ag_pb.Group,
    (request: ag_ag_pb.GetGroupRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Group.deserializeBinary
  );

  getGroup(
    request: ag_ag_pb.GetGroupRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Group>;

  getGroup(
    request: ag_ag_pb.GetGroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Group) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Group>;

  getGroup(
    request: ag_ag_pb.GetGroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Group) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetGroup',
        request,
        metadata || {},
        this.methodDescriptorGetGroup,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetGroup',
    request,
    metadata || {},
    this.methodDescriptorGetGroup);
  }

  methodDescriptorGetGroupByUserAndCourse = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetGroupByUserAndCourse',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.GroupRequest,
    ag_ag_pb.Group,
    (request: ag_ag_pb.GroupRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Group.deserializeBinary
  );

  getGroupByUserAndCourse(
    request: ag_ag_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Group>;

  getGroupByUserAndCourse(
    request: ag_ag_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Group) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Group>;

  getGroupByUserAndCourse(
    request: ag_ag_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Group) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetGroupByUserAndCourse',
        request,
        metadata || {},
        this.methodDescriptorGetGroupByUserAndCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetGroupByUserAndCourse',
    request,
    metadata || {},
    this.methodDescriptorGetGroupByUserAndCourse);
  }

  methodDescriptorGetGroupsByCourse = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetGroupsByCourse',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.CourseRequest,
    ag_ag_pb.Groups,
    (request: ag_ag_pb.CourseRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Groups.deserializeBinary
  );

  getGroupsByCourse(
    request: ag_ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Groups>;

  getGroupsByCourse(
    request: ag_ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Groups) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Groups>;

  getGroupsByCourse(
    request: ag_ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Groups) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetGroupsByCourse',
        request,
        metadata || {},
        this.methodDescriptorGetGroupsByCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetGroupsByCourse',
    request,
    metadata || {},
    this.methodDescriptorGetGroupsByCourse);
  }

  methodDescriptorCreateGroup = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/CreateGroup',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.Group,
    ag_ag_pb.Group,
    (request: ag_ag_pb.Group) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Group.deserializeBinary
  );

  createGroup(
    request: ag_ag_pb.Group,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Group>;

  createGroup(
    request: ag_ag_pb.Group,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Group) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Group>;

  createGroup(
    request: ag_ag_pb.Group,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Group) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/CreateGroup',
        request,
        metadata || {},
        this.methodDescriptorCreateGroup,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/CreateGroup',
    request,
    metadata || {},
    this.methodDescriptorCreateGroup);
  }

  methodDescriptorUpdateGroup = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateGroup',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.Group,
    ag_ag_pb.Void,
    (request: ag_ag_pb.Group) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Void.deserializeBinary
  );

  updateGroup(
    request: ag_ag_pb.Group,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Void>;

  updateGroup(
    request: ag_ag_pb.Group,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Void>;

  updateGroup(
    request: ag_ag_pb.Group,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateGroup',
        request,
        metadata || {},
        this.methodDescriptorUpdateGroup,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateGroup',
    request,
    metadata || {},
    this.methodDescriptorUpdateGroup);
  }

  methodDescriptorDeleteGroup = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/DeleteGroup',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.GroupRequest,
    ag_ag_pb.Void,
    (request: ag_ag_pb.GroupRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Void.deserializeBinary
  );

  deleteGroup(
    request: ag_ag_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Void>;

  deleteGroup(
    request: ag_ag_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Void>;

  deleteGroup(
    request: ag_ag_pb.GroupRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/DeleteGroup',
        request,
        metadata || {},
        this.methodDescriptorDeleteGroup,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/DeleteGroup',
    request,
    metadata || {},
    this.methodDescriptorDeleteGroup);
  }

  methodDescriptorGetCourse = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetCourse',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.CourseRequest,
    ag_ag_pb.Course,
    (request: ag_ag_pb.CourseRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Course.deserializeBinary
  );

  getCourse(
    request: ag_ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Course>;

  getCourse(
    request: ag_ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Course) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Course>;

  getCourse(
    request: ag_ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Course) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetCourse',
        request,
        metadata || {},
        this.methodDescriptorGetCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetCourse',
    request,
    metadata || {},
    this.methodDescriptorGetCourse);
  }

  methodDescriptorGetCourses = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetCourses',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.Void,
    ag_ag_pb.Courses,
    (request: ag_ag_pb.Void) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Courses.deserializeBinary
  );

  getCourses(
    request: ag_ag_pb.Void,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Courses>;

  getCourses(
    request: ag_ag_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Courses) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Courses>;

  getCourses(
    request: ag_ag_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Courses) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetCourses',
        request,
        metadata || {},
        this.methodDescriptorGetCourses,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetCourses',
    request,
    metadata || {},
    this.methodDescriptorGetCourses);
  }

  methodDescriptorGetCoursesByUser = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetCoursesByUser',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.EnrollmentStatusRequest,
    ag_ag_pb.Courses,
    (request: ag_ag_pb.EnrollmentStatusRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Courses.deserializeBinary
  );

  getCoursesByUser(
    request: ag_ag_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Courses>;

  getCoursesByUser(
    request: ag_ag_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Courses) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Courses>;

  getCoursesByUser(
    request: ag_ag_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Courses) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetCoursesByUser',
        request,
        metadata || {},
        this.methodDescriptorGetCoursesByUser,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetCoursesByUser',
    request,
    metadata || {},
    this.methodDescriptorGetCoursesByUser);
  }

  methodDescriptorCreateCourse = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/CreateCourse',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.Course,
    ag_ag_pb.Course,
    (request: ag_ag_pb.Course) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Course.deserializeBinary
  );

  createCourse(
    request: ag_ag_pb.Course,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Course>;

  createCourse(
    request: ag_ag_pb.Course,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Course) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Course>;

  createCourse(
    request: ag_ag_pb.Course,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Course) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/CreateCourse',
        request,
        metadata || {},
        this.methodDescriptorCreateCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/CreateCourse',
    request,
    metadata || {},
    this.methodDescriptorCreateCourse);
  }

  methodDescriptorUpdateCourse = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateCourse',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.Course,
    ag_ag_pb.Void,
    (request: ag_ag_pb.Course) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Void.deserializeBinary
  );

  updateCourse(
    request: ag_ag_pb.Course,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Void>;

  updateCourse(
    request: ag_ag_pb.Course,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Void>;

  updateCourse(
    request: ag_ag_pb.Course,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateCourse',
        request,
        metadata || {},
        this.methodDescriptorUpdateCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateCourse',
    request,
    metadata || {},
    this.methodDescriptorUpdateCourse);
  }

  methodDescriptorUpdateCourseVisibility = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateCourseVisibility',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.Enrollment,
    ag_ag_pb.Void,
    (request: ag_ag_pb.Enrollment) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Void.deserializeBinary
  );

  updateCourseVisibility(
    request: ag_ag_pb.Enrollment,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Void>;

  updateCourseVisibility(
    request: ag_ag_pb.Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Void>;

  updateCourseVisibility(
    request: ag_ag_pb.Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateCourseVisibility',
        request,
        metadata || {},
        this.methodDescriptorUpdateCourseVisibility,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateCourseVisibility',
    request,
    metadata || {},
    this.methodDescriptorUpdateCourseVisibility);
  }

  methodDescriptorGetAssignments = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetAssignments',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.CourseRequest,
    ag_ag_pb.Assignments,
    (request: ag_ag_pb.CourseRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Assignments.deserializeBinary
  );

  getAssignments(
    request: ag_ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Assignments>;

  getAssignments(
    request: ag_ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Assignments) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Assignments>;

  getAssignments(
    request: ag_ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Assignments) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetAssignments',
        request,
        metadata || {},
        this.methodDescriptorGetAssignments,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetAssignments',
    request,
    metadata || {},
    this.methodDescriptorGetAssignments);
  }

  methodDescriptorUpdateAssignments = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateAssignments',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.CourseRequest,
    ag_ag_pb.Void,
    (request: ag_ag_pb.CourseRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Void.deserializeBinary
  );

  updateAssignments(
    request: ag_ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Void>;

  updateAssignments(
    request: ag_ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Void>;

  updateAssignments(
    request: ag_ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateAssignments',
        request,
        metadata || {},
        this.methodDescriptorUpdateAssignments,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateAssignments',
    request,
    metadata || {},
    this.methodDescriptorUpdateAssignments);
  }

  methodDescriptorGetEnrollmentsByUser = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetEnrollmentsByUser',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.EnrollmentStatusRequest,
    ag_ag_pb.Enrollments,
    (request: ag_ag_pb.EnrollmentStatusRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Enrollments.deserializeBinary
  );

  getEnrollmentsByUser(
    request: ag_ag_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Enrollments>;

  getEnrollmentsByUser(
    request: ag_ag_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Enrollments) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Enrollments>;

  getEnrollmentsByUser(
    request: ag_ag_pb.EnrollmentStatusRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Enrollments) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetEnrollmentsByUser',
        request,
        metadata || {},
        this.methodDescriptorGetEnrollmentsByUser,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetEnrollmentsByUser',
    request,
    metadata || {},
    this.methodDescriptorGetEnrollmentsByUser);
  }

  methodDescriptorGetEnrollmentsByCourse = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetEnrollmentsByCourse',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.EnrollmentRequest,
    ag_ag_pb.Enrollments,
    (request: ag_ag_pb.EnrollmentRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Enrollments.deserializeBinary
  );

  getEnrollmentsByCourse(
    request: ag_ag_pb.EnrollmentRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Enrollments>;

  getEnrollmentsByCourse(
    request: ag_ag_pb.EnrollmentRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Enrollments) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Enrollments>;

  getEnrollmentsByCourse(
    request: ag_ag_pb.EnrollmentRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Enrollments) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetEnrollmentsByCourse',
        request,
        metadata || {},
        this.methodDescriptorGetEnrollmentsByCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetEnrollmentsByCourse',
    request,
    metadata || {},
    this.methodDescriptorGetEnrollmentsByCourse);
  }

  methodDescriptorCreateEnrollment = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/CreateEnrollment',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.Enrollment,
    ag_ag_pb.Void,
    (request: ag_ag_pb.Enrollment) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Void.deserializeBinary
  );

  createEnrollment(
    request: ag_ag_pb.Enrollment,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Void>;

  createEnrollment(
    request: ag_ag_pb.Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Void>;

  createEnrollment(
    request: ag_ag_pb.Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/CreateEnrollment',
        request,
        metadata || {},
        this.methodDescriptorCreateEnrollment,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/CreateEnrollment',
    request,
    metadata || {},
    this.methodDescriptorCreateEnrollment);
  }

  methodDescriptorUpdateEnrollment = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateEnrollment',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.Enrollment,
    ag_ag_pb.Void,
    (request: ag_ag_pb.Enrollment) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Void.deserializeBinary
  );

  updateEnrollment(
    request: ag_ag_pb.Enrollment,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Void>;

  updateEnrollment(
    request: ag_ag_pb.Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Void>;

  updateEnrollment(
    request: ag_ag_pb.Enrollment,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateEnrollment',
        request,
        metadata || {},
        this.methodDescriptorUpdateEnrollment,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateEnrollment',
    request,
    metadata || {},
    this.methodDescriptorUpdateEnrollment);
  }

  methodDescriptorUpdateEnrollments = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateEnrollments',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.CourseRequest,
    ag_ag_pb.Void,
    (request: ag_ag_pb.CourseRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Void.deserializeBinary
  );

  updateEnrollments(
    request: ag_ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Void>;

  updateEnrollments(
    request: ag_ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Void>;

  updateEnrollments(
    request: ag_ag_pb.CourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateEnrollments',
        request,
        metadata || {},
        this.methodDescriptorUpdateEnrollments,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateEnrollments',
    request,
    metadata || {},
    this.methodDescriptorUpdateEnrollments);
  }

  methodDescriptorGetSubmissions = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetSubmissions',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.SubmissionRequest,
    ag_ag_pb.Submissions,
    (request: ag_ag_pb.SubmissionRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Submissions.deserializeBinary
  );

  getSubmissions(
    request: ag_ag_pb.SubmissionRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Submissions>;

  getSubmissions(
    request: ag_ag_pb.SubmissionRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Submissions) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Submissions>;

  getSubmissions(
    request: ag_ag_pb.SubmissionRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Submissions) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetSubmissions',
        request,
        metadata || {},
        this.methodDescriptorGetSubmissions,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetSubmissions',
    request,
    metadata || {},
    this.methodDescriptorGetSubmissions);
  }

  methodDescriptorGetSubmissionsByCourse = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetSubmissionsByCourse',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.SubmissionsForCourseRequest,
    ag_ag_pb.CourseSubmissions,
    (request: ag_ag_pb.SubmissionsForCourseRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.CourseSubmissions.deserializeBinary
  );

  getSubmissionsByCourse(
    request: ag_ag_pb.SubmissionsForCourseRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.CourseSubmissions>;

  getSubmissionsByCourse(
    request: ag_ag_pb.SubmissionsForCourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.CourseSubmissions) => void): grpcWeb.ClientReadableStream<ag_ag_pb.CourseSubmissions>;

  getSubmissionsByCourse(
    request: ag_ag_pb.SubmissionsForCourseRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.CourseSubmissions) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetSubmissionsByCourse',
        request,
        metadata || {},
        this.methodDescriptorGetSubmissionsByCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetSubmissionsByCourse',
    request,
    metadata || {},
    this.methodDescriptorGetSubmissionsByCourse);
  }

  methodDescriptorUpdateSubmission = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateSubmission',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.UpdateSubmissionRequest,
    ag_ag_pb.Void,
    (request: ag_ag_pb.UpdateSubmissionRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Void.deserializeBinary
  );

  updateSubmission(
    request: ag_ag_pb.UpdateSubmissionRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Void>;

  updateSubmission(
    request: ag_ag_pb.UpdateSubmissionRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Void>;

  updateSubmission(
    request: ag_ag_pb.UpdateSubmissionRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateSubmission',
        request,
        metadata || {},
        this.methodDescriptorUpdateSubmission,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateSubmission',
    request,
    metadata || {},
    this.methodDescriptorUpdateSubmission);
  }

  methodDescriptorUpdateSubmissions = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateSubmissions',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.UpdateSubmissionsRequest,
    ag_ag_pb.Void,
    (request: ag_ag_pb.UpdateSubmissionsRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Void.deserializeBinary
  );

  updateSubmissions(
    request: ag_ag_pb.UpdateSubmissionsRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Void>;

  updateSubmissions(
    request: ag_ag_pb.UpdateSubmissionsRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Void>;

  updateSubmissions(
    request: ag_ag_pb.UpdateSubmissionsRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateSubmissions',
        request,
        metadata || {},
        this.methodDescriptorUpdateSubmissions,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateSubmissions',
    request,
    metadata || {},
    this.methodDescriptorUpdateSubmissions);
  }

  methodDescriptorRebuildSubmissions = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/RebuildSubmissions',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.RebuildRequest,
    ag_ag_pb.Void,
    (request: ag_ag_pb.RebuildRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Void.deserializeBinary
  );

  rebuildSubmissions(
    request: ag_ag_pb.RebuildRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Void>;

  rebuildSubmissions(
    request: ag_ag_pb.RebuildRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Void>;

  rebuildSubmissions(
    request: ag_ag_pb.RebuildRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/RebuildSubmissions',
        request,
        metadata || {},
        this.methodDescriptorRebuildSubmissions,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/RebuildSubmissions',
    request,
    metadata || {},
    this.methodDescriptorRebuildSubmissions);
  }

  methodDescriptorCreateBenchmark = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/CreateBenchmark',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.GradingBenchmark,
    ag_ag_pb.GradingBenchmark,
    (request: ag_ag_pb.GradingBenchmark) => {
      return request.serializeBinary();
    },
    ag_ag_pb.GradingBenchmark.deserializeBinary
  );

  createBenchmark(
    request: ag_ag_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.GradingBenchmark>;

  createBenchmark(
    request: ag_ag_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.GradingBenchmark) => void): grpcWeb.ClientReadableStream<ag_ag_pb.GradingBenchmark>;

  createBenchmark(
    request: ag_ag_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.GradingBenchmark) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/CreateBenchmark',
        request,
        metadata || {},
        this.methodDescriptorCreateBenchmark,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/CreateBenchmark',
    request,
    metadata || {},
    this.methodDescriptorCreateBenchmark);
  }

  methodDescriptorUpdateBenchmark = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateBenchmark',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.GradingBenchmark,
    ag_ag_pb.Void,
    (request: ag_ag_pb.GradingBenchmark) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Void.deserializeBinary
  );

  updateBenchmark(
    request: ag_ag_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Void>;

  updateBenchmark(
    request: ag_ag_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Void>;

  updateBenchmark(
    request: ag_ag_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateBenchmark',
        request,
        metadata || {},
        this.methodDescriptorUpdateBenchmark,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateBenchmark',
    request,
    metadata || {},
    this.methodDescriptorUpdateBenchmark);
  }

  methodDescriptorDeleteBenchmark = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/DeleteBenchmark',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.GradingBenchmark,
    ag_ag_pb.Void,
    (request: ag_ag_pb.GradingBenchmark) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Void.deserializeBinary
  );

  deleteBenchmark(
    request: ag_ag_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Void>;

  deleteBenchmark(
    request: ag_ag_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Void>;

  deleteBenchmark(
    request: ag_ag_pb.GradingBenchmark,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/DeleteBenchmark',
        request,
        metadata || {},
        this.methodDescriptorDeleteBenchmark,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/DeleteBenchmark',
    request,
    metadata || {},
    this.methodDescriptorDeleteBenchmark);
  }

  methodDescriptorCreateCriterion = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/CreateCriterion',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.GradingCriterion,
    ag_ag_pb.GradingCriterion,
    (request: ag_ag_pb.GradingCriterion) => {
      return request.serializeBinary();
    },
    ag_ag_pb.GradingCriterion.deserializeBinary
  );

  createCriterion(
    request: ag_ag_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.GradingCriterion>;

  createCriterion(
    request: ag_ag_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.GradingCriterion) => void): grpcWeb.ClientReadableStream<ag_ag_pb.GradingCriterion>;

  createCriterion(
    request: ag_ag_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.GradingCriterion) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/CreateCriterion',
        request,
        metadata || {},
        this.methodDescriptorCreateCriterion,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/CreateCriterion',
    request,
    metadata || {},
    this.methodDescriptorCreateCriterion);
  }

  methodDescriptorUpdateCriterion = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateCriterion',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.GradingCriterion,
    ag_ag_pb.Void,
    (request: ag_ag_pb.GradingCriterion) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Void.deserializeBinary
  );

  updateCriterion(
    request: ag_ag_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Void>;

  updateCriterion(
    request: ag_ag_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Void>;

  updateCriterion(
    request: ag_ag_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateCriterion',
        request,
        metadata || {},
        this.methodDescriptorUpdateCriterion,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateCriterion',
    request,
    metadata || {},
    this.methodDescriptorUpdateCriterion);
  }

  methodDescriptorDeleteCriterion = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/DeleteCriterion',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.GradingCriterion,
    ag_ag_pb.Void,
    (request: ag_ag_pb.GradingCriterion) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Void.deserializeBinary
  );

  deleteCriterion(
    request: ag_ag_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Void>;

  deleteCriterion(
    request: ag_ag_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Void>;

  deleteCriterion(
    request: ag_ag_pb.GradingCriterion,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/DeleteCriterion',
        request,
        metadata || {},
        this.methodDescriptorDeleteCriterion,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/DeleteCriterion',
    request,
    metadata || {},
    this.methodDescriptorDeleteCriterion);
  }

  methodDescriptorCreateReview = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/CreateReview',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.ReviewRequest,
    ag_ag_pb.Review,
    (request: ag_ag_pb.ReviewRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Review.deserializeBinary
  );

  createReview(
    request: ag_ag_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Review>;

  createReview(
    request: ag_ag_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Review) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Review>;

  createReview(
    request: ag_ag_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Review) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/CreateReview',
        request,
        metadata || {},
        this.methodDescriptorCreateReview,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/CreateReview',
    request,
    metadata || {},
    this.methodDescriptorCreateReview);
  }

  methodDescriptorUpdateReview = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/UpdateReview',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.ReviewRequest,
    ag_ag_pb.Review,
    (request: ag_ag_pb.ReviewRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Review.deserializeBinary
  );

  updateReview(
    request: ag_ag_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Review>;

  updateReview(
    request: ag_ag_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Review) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Review>;

  updateReview(
    request: ag_ag_pb.ReviewRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Review) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/UpdateReview',
        request,
        metadata || {},
        this.methodDescriptorUpdateReview,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/UpdateReview',
    request,
    metadata || {},
    this.methodDescriptorUpdateReview);
  }

  methodDescriptorGetReviewers = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetReviewers',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.SubmissionReviewersRequest,
    ag_ag_pb.Reviewers,
    (request: ag_ag_pb.SubmissionReviewersRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Reviewers.deserializeBinary
  );

  getReviewers(
    request: ag_ag_pb.SubmissionReviewersRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Reviewers>;

  getReviewers(
    request: ag_ag_pb.SubmissionReviewersRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Reviewers) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Reviewers>;

  getReviewers(
    request: ag_ag_pb.SubmissionReviewersRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Reviewers) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetReviewers',
        request,
        metadata || {},
        this.methodDescriptorGetReviewers,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetReviewers',
    request,
    metadata || {},
    this.methodDescriptorGetReviewers);
  }

  methodDescriptorGetProviders = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetProviders',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.Void,
    ag_ag_pb.Providers,
    (request: ag_ag_pb.Void) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Providers.deserializeBinary
  );

  getProviders(
    request: ag_ag_pb.Void,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Providers>;

  getProviders(
    request: ag_ag_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Providers) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Providers>;

  getProviders(
    request: ag_ag_pb.Void,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Providers) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetProviders',
        request,
        metadata || {},
        this.methodDescriptorGetProviders,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetProviders',
    request,
    metadata || {},
    this.methodDescriptorGetProviders);
  }

  methodDescriptorGetOrganization = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetOrganization',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.OrgRequest,
    ag_ag_pb.Organization,
    (request: ag_ag_pb.OrgRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Organization.deserializeBinary
  );

  getOrganization(
    request: ag_ag_pb.OrgRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Organization>;

  getOrganization(
    request: ag_ag_pb.OrgRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Organization) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Organization>;

  getOrganization(
    request: ag_ag_pb.OrgRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Organization) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetOrganization',
        request,
        metadata || {},
        this.methodDescriptorGetOrganization,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetOrganization',
    request,
    metadata || {},
    this.methodDescriptorGetOrganization);
  }

  methodDescriptorGetRepositories = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/GetRepositories',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.URLRequest,
    ag_ag_pb.Repositories,
    (request: ag_ag_pb.URLRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Repositories.deserializeBinary
  );

  getRepositories(
    request: ag_ag_pb.URLRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Repositories>;

  getRepositories(
    request: ag_ag_pb.URLRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Repositories) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Repositories>;

  getRepositories(
    request: ag_ag_pb.URLRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Repositories) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/GetRepositories',
        request,
        metadata || {},
        this.methodDescriptorGetRepositories,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/GetRepositories',
    request,
    metadata || {},
    this.methodDescriptorGetRepositories);
  }

  methodDescriptorIsEmptyRepo = new grpcWeb.MethodDescriptor(
    '/ag.AutograderService/IsEmptyRepo',
    grpcWeb.MethodType.UNARY,
    ag_ag_pb.RepositoryRequest,
    ag_ag_pb.Void,
    (request: ag_ag_pb.RepositoryRequest) => {
      return request.serializeBinary();
    },
    ag_ag_pb.Void.deserializeBinary
  );

  isEmptyRepo(
    request: ag_ag_pb.RepositoryRequest,
    metadata: grpcWeb.Metadata | null): Promise<ag_ag_pb.Void>;

  isEmptyRepo(
    request: ag_ag_pb.RepositoryRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void): grpcWeb.ClientReadableStream<ag_ag_pb.Void>;

  isEmptyRepo(
    request: ag_ag_pb.RepositoryRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: ag_ag_pb.Void) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/ag.AutograderService/IsEmptyRepo',
        request,
        metadata || {},
        this.methodDescriptorIsEmptyRepo,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/ag.AutograderService/IsEmptyRepo',
    request,
    metadata || {},
    this.methodDescriptorIsEmptyRepo);
  }

}

