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
        this.methodInfoGetSelf = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.User, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.User.deserializeBinary);
        this.methodInfoGetUser = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.User, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.User.deserializeBinary);
        this.methodInfoGetUsers = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Users, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Users.deserializeBinary);
        this.methodInfoUpdateUser = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.User, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.User.deserializeBinary);
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
        this.methodInfoUpdateGroup = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.StatusCode, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.StatusCode.deserializeBinary);
        this.methodInfoUpdateGroupStatus = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.StatusCode, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.StatusCode.deserializeBinary);
        this.methodInfoDeleteGroup = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.StatusCode, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.StatusCode.deserializeBinary);
        this.methodInfoGetCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Course, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Course.deserializeBinary);
        this.methodInfoGetCourses = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Courses, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Courses.deserializeBinary);
        this.methodInfoGetCoursesWithEnrollment = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Courses, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Courses.deserializeBinary);
        this.methodInfoGetCourseInformationURL = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.URLResponse, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.URLResponse.deserializeBinary);
        this.methodInfoCreateCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Course, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Course.deserializeBinary);
        this.methodInfoUpdateCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.StatusCode, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.StatusCode.deserializeBinary);
        this.methodInfoRefreshCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.StatusCode, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.StatusCode.deserializeBinary);
        this.methodInfoGetEnrollmentsByCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Enrollments, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Enrollments.deserializeBinary);
        this.methodInfoCreateEnrollment = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.StatusCode, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.StatusCode.deserializeBinary);
        this.methodInfoUpdateEnrollment = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.StatusCode, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.StatusCode.deserializeBinary);
        this.methodInfoGetSubmissions = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Submissions, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Submissions.deserializeBinary);
        this.methodInfoGetSubmission = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Submission, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Submission.deserializeBinary);
        this.methodInfoGetGroupSubmissions = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Submissions, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Submissions.deserializeBinary);
        this.methodInfoUpdateSubmission = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.StatusCode, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.StatusCode.deserializeBinary);
        this.methodInfoGetAssignments = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Assignments, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Assignments.deserializeBinary);
        this.methodInfoGetRepositoryURL = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.URLResponse, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.URLResponse.deserializeBinary);
        this.methodInfoGetProviders = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Providers, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Providers.deserializeBinary);
        this.methodInfoGetDirectories = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Directories, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Directories.deserializeBinary);
        this.methodInfoGetRepository = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Repository, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Repository.deserializeBinary);
        if (!options)
            options = {};
        options['format'] = 'binary';
        this.client_ = new grpcWeb.GrpcWebClientBase(options);
        this.hostname_ = hostname;
        this.credentials_ = credentials;
        this.options_ = options;
    }
    AutograderServiceClient.prototype.getSelf = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetSelf', request, metadata, this.methodInfoGetSelf, callback);
    };
    AutograderServiceClient.prototype.getUser = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetUser', request, metadata, this.methodInfoGetUser, callback);
    };
    AutograderServiceClient.prototype.getUsers = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetUsers', request, metadata, this.methodInfoGetUsers, callback);
    };
    AutograderServiceClient.prototype.updateUser = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/UpdateUser', request, metadata, this.methodInfoUpdateUser, callback);
    };
    AutograderServiceClient.prototype.getGroup = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetGroup', request, metadata, this.methodInfoGetGroup, callback);
    };
    AutograderServiceClient.prototype.getGroupByUserAndCourse = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetGroupByUserAndCourse', request, metadata, this.methodInfoGetGroupByUserAndCourse, callback);
    };
    AutograderServiceClient.prototype.getGroups = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetGroups', request, metadata, this.methodInfoGetGroups, callback);
    };
    AutograderServiceClient.prototype.createGroup = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/CreateGroup', request, metadata, this.methodInfoCreateGroup, callback);
    };
    AutograderServiceClient.prototype.updateGroup = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/UpdateGroup', request, metadata, this.methodInfoUpdateGroup, callback);
    };
    AutograderServiceClient.prototype.updateGroupStatus = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/UpdateGroupStatus', request, metadata, this.methodInfoUpdateGroupStatus, callback);
    };
    AutograderServiceClient.prototype.deleteGroup = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/DeleteGroup', request, metadata, this.methodInfoDeleteGroup, callback);
    };
    AutograderServiceClient.prototype.getCourse = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetCourse', request, metadata, this.methodInfoGetCourse, callback);
    };
    AutograderServiceClient.prototype.getCourses = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetCourses', request, metadata, this.methodInfoGetCourses, callback);
    };
    AutograderServiceClient.prototype.getCoursesWithEnrollment = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetCoursesWithEnrollment', request, metadata, this.methodInfoGetCoursesWithEnrollment, callback);
    };
    AutograderServiceClient.prototype.getCourseInformationURL = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetCourseInformationURL', request, metadata, this.methodInfoGetCourseInformationURL, callback);
    };
    AutograderServiceClient.prototype.createCourse = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/CreateCourse', request, metadata, this.methodInfoCreateCourse, callback);
    };
    AutograderServiceClient.prototype.updateCourse = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/UpdateCourse', request, metadata, this.methodInfoUpdateCourse, callback);
    };
    AutograderServiceClient.prototype.refreshCourse = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/RefreshCourse', request, metadata, this.methodInfoRefreshCourse, callback);
    };
    AutograderServiceClient.prototype.getEnrollmentsByCourse = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetEnrollmentsByCourse', request, metadata, this.methodInfoGetEnrollmentsByCourse, callback);
    };
    AutograderServiceClient.prototype.createEnrollment = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/CreateEnrollment', request, metadata, this.methodInfoCreateEnrollment, callback);
    };
    AutograderServiceClient.prototype.updateEnrollment = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/UpdateEnrollment', request, metadata, this.methodInfoUpdateEnrollment, callback);
    };
    AutograderServiceClient.prototype.getSubmissions = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetSubmissions', request, metadata, this.methodInfoGetSubmissions, callback);
    };
    AutograderServiceClient.prototype.getSubmission = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetSubmission', request, metadata, this.methodInfoGetSubmission, callback);
    };
    AutograderServiceClient.prototype.getGroupSubmissions = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetGroupSubmissions', request, metadata, this.methodInfoGetGroupSubmissions, callback);
    };
    AutograderServiceClient.prototype.updateSubmission = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/UpdateSubmission', request, metadata, this.methodInfoUpdateSubmission, callback);
    };
    AutograderServiceClient.prototype.getAssignments = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetAssignments', request, metadata, this.methodInfoGetAssignments, callback);
    };
    AutograderServiceClient.prototype.getRepositoryURL = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetRepositoryURL', request, metadata, this.methodInfoGetRepositoryURL, callback);
    };
    AutograderServiceClient.prototype.getProviders = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetProviders', request, metadata, this.methodInfoGetProviders, callback);
    };
    AutograderServiceClient.prototype.getDirectories = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetDirectories', request, metadata, this.methodInfoGetDirectories, callback);
    };
    AutograderServiceClient.prototype.getRepository = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetRepository', request, metadata, this.methodInfoGetRepository, callback);
    };
    return AutograderServiceClient;
}());
exports.AutograderServiceClient = AutograderServiceClient;
