"use strict";
/**
 * @fileoverview gRPC-Web generated client stub for
 * @enhanceable
 * @public
 */
exports.__esModule = true;
// GENERATED CODE -- DO NOT EDIT!
var grpcWeb = require("grpc-web");
var ag_pb_1 = require("./ag_pb");
var AutograderServiceClient = /** @class */ (function () {
    function AutograderServiceClient(hostname, credentials, options) {
        this.methodInfoGetUser = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.User, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.User.deserializeBinary);
        this.methodInfoGetUsers = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Users, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Users.deserializeBinary);
        this.methodInfoUpdateUser = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.User, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.User.deserializeBinary);
        this.methodInfoIsAuthorizedTeacher = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.AuthorizationResponse, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.AuthorizationResponse.deserializeBinary);
        this.methodInfoGetGroup = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Group, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Group.deserializeBinary);
        this.methodInfoGetGroupByUserAndCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Group, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Group.deserializeBinary);
        this.methodInfoGetGroups = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Groups, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Groups.deserializeBinary);
        this.methodInfoCreateGroup = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Group, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Group.deserializeBinary);
        this.methodInfoUpdateGroup = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoDeleteGroup = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoGetCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Course, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Course.deserializeBinary);
        this.methodInfoGetCourses = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Courses, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Courses.deserializeBinary);
        this.methodInfoGetCoursesWithEnrollment = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Courses, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Courses.deserializeBinary);
        this.methodInfoCreateCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Course, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Course.deserializeBinary);
        this.methodInfoUpdateCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoGetAssignments = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Assignments, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Assignments.deserializeBinary);
        this.methodInfoUpdateAssignments = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoGetEnrollmentsByCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Enrollments, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Enrollments.deserializeBinary);
        this.methodInfoCreateEnrollment = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoUpdateEnrollment = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoGetSubmissions = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Submissions, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Submissions.deserializeBinary);
        this.methodInfoApproveSubmission = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoRefreshSubmission = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoGetProviders = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Providers, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Providers.deserializeBinary);
        this.methodInfoGetOrganizations = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Organizations, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Organizations.deserializeBinary);
        this.methodInfoGetRepositories = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Repositories, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Repositories.deserializeBinary);
        if (!options)
            options = {};
        options['format'] = 'binary';
        this.client_ = new grpcWeb.GrpcWebClientBase(options);
        this.hostname_ = hostname;
        this.credentials_ = credentials;
        this.options_ = options;
    }
    AutograderServiceClient.prototype.getUser = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetUser', request, metadata || {}, this.methodInfoGetUser, callback);
    };
    AutograderServiceClient.prototype.getUsers = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetUsers', request, metadata || {}, this.methodInfoGetUsers, callback);
    };
    AutograderServiceClient.prototype.updateUser = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/UpdateUser', request, metadata || {}, this.methodInfoUpdateUser, callback);
    };
    AutograderServiceClient.prototype.isAuthorizedTeacher = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/IsAuthorizedTeacher', request, metadata || {}, this.methodInfoIsAuthorizedTeacher, callback);
    };
    AutograderServiceClient.prototype.getGroup = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetGroup', request, metadata || {}, this.methodInfoGetGroup, callback);
    };
    AutograderServiceClient.prototype.getGroupByUserAndCourse = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetGroupByUserAndCourse', request, metadata || {}, this.methodInfoGetGroupByUserAndCourse, callback);
    };
    AutograderServiceClient.prototype.getGroups = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetGroups', request, metadata || {}, this.methodInfoGetGroups, callback);
    };
    AutograderServiceClient.prototype.createGroup = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/CreateGroup', request, metadata || {}, this.methodInfoCreateGroup, callback);
    };
    AutograderServiceClient.prototype.updateGroup = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/UpdateGroup', request, metadata || {}, this.methodInfoUpdateGroup, callback);
    };
    AutograderServiceClient.prototype.deleteGroup = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/DeleteGroup', request, metadata || {}, this.methodInfoDeleteGroup, callback);
    };
    AutograderServiceClient.prototype.getCourse = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetCourse', request, metadata || {}, this.methodInfoGetCourse, callback);
    };
    AutograderServiceClient.prototype.getCourses = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetCourses', request, metadata || {}, this.methodInfoGetCourses, callback);
    };
    AutograderServiceClient.prototype.getCoursesWithEnrollment = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetCoursesWithEnrollment', request, metadata || {}, this.methodInfoGetCoursesWithEnrollment, callback);
    };
    AutograderServiceClient.prototype.createCourse = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/CreateCourse', request, metadata || {}, this.methodInfoCreateCourse, callback);
    };
    AutograderServiceClient.prototype.updateCourse = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/UpdateCourse', request, metadata || {}, this.methodInfoUpdateCourse, callback);
    };
    AutograderServiceClient.prototype.getAssignments = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetAssignments', request, metadata || {}, this.methodInfoGetAssignments, callback);
    };
    AutograderServiceClient.prototype.updateAssignments = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/UpdateAssignments', request, metadata || {}, this.methodInfoUpdateAssignments, callback);
    };
    AutograderServiceClient.prototype.getEnrollmentsByCourse = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetEnrollmentsByCourse', request, metadata || {}, this.methodInfoGetEnrollmentsByCourse, callback);
    };
    AutograderServiceClient.prototype.createEnrollment = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/CreateEnrollment', request, metadata || {}, this.methodInfoCreateEnrollment, callback);
    };
    AutograderServiceClient.prototype.updateEnrollment = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/UpdateEnrollment', request, metadata || {}, this.methodInfoUpdateEnrollment, callback);
    };
    AutograderServiceClient.prototype.getSubmissions = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetSubmissions', request, metadata || {}, this.methodInfoGetSubmissions, callback);
    };
    AutograderServiceClient.prototype.approveSubmission = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/ApproveSubmission', request, metadata || {}, this.methodInfoApproveSubmission, callback);
    };
    AutograderServiceClient.prototype.refreshSubmission = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/RefreshSubmission', request, metadata || {}, this.methodInfoRefreshSubmission, callback);
    };
    AutograderServiceClient.prototype.getProviders = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetProviders', request, metadata || {}, this.methodInfoGetProviders, callback);
    };
    AutograderServiceClient.prototype.getOrganizations = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetOrganizations', request, metadata || {}, this.methodInfoGetOrganizations, callback);
    };
    AutograderServiceClient.prototype.getRepositories = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetRepositories', request, metadata || {}, this.methodInfoGetRepositories, callback);
    };
    return AutograderServiceClient;
}());
exports.AutograderServiceClient = AutograderServiceClient;
