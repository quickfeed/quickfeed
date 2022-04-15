/**
 * @fileoverview gRPC-Web generated client stub for admin
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck


import * as grpcWeb from 'grpc-web';

import * as ag_ag_pb from '../ag/ag_pb';


export class AdminServiceClient {
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

  methodInfoGetUsers = new grpcWeb.MethodDescriptor(
    '/admin.AdminService/GetUsers',
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
          '/admin.AdminService/GetUsers',
        request,
        metadata || {},
        this.methodInfoGetUsers,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/admin.AdminService/GetUsers',
    request,
    metadata || {},
    this.methodInfoGetUsers);
  }

  methodInfoCreateCourse = new grpcWeb.MethodDescriptor(
    '/admin.AdminService/CreateCourse',
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
          '/admin.AdminService/CreateCourse',
        request,
        metadata || {},
        this.methodInfoCreateCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/admin.AdminService/CreateCourse',
    request,
    metadata || {},
    this.methodInfoCreateCourse);
  }

  methodInfoUpdateCourse = new grpcWeb.MethodDescriptor(
    '/admin.AdminService/UpdateCourse',
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
          '/admin.AdminService/UpdateCourse',
        request,
        metadata || {},
        this.methodInfoUpdateCourse,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/admin.AdminService/UpdateCourse',
    request,
    metadata || {},
    this.methodInfoUpdateCourse);
  }

  methodInfoGetOrganization = new grpcWeb.MethodDescriptor(
    '/admin.AdminService/GetOrganization',
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
          '/admin.AdminService/GetOrganization',
        request,
        metadata || {},
        this.methodInfoGetOrganization,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/admin.AdminService/GetOrganization',
    request,
    metadata || {},
    this.methodInfoGetOrganization);
  }

}

