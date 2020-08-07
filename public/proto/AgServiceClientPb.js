"use strict";
/**
 * @fileoverview gRPC-Web generated client stub for
 * @enhanceable
 * @public
 */
exports.__esModule = true;
exports.AutograderServiceClient = void 0;
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
        this.methodInfoUpdateUser = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoIsAuthorizedTeacher = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.AuthorizationResponse, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.AuthorizationResponse.deserializeBinary);
        this.methodInfoGetGroup = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Group, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Group.deserializeBinary);
        this.methodInfoGetGroupByUserAndCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Group, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Group.deserializeBinary);
        this.methodInfoGetGroupsByCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Groups, function (request) {
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
        this.methodInfoGetCoursesByUser = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Courses, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Courses.deserializeBinary);
        this.methodInfoCreateCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Course, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Course.deserializeBinary);
        this.methodInfoUpdateCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoUpdateCourseVisibility = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoGetAssignments = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Assignments, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Assignments.deserializeBinary);
        this.methodInfoUpdateAssignments = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoGetEnrollmentsByUser = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Enrollments, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Enrollments.deserializeBinary);
        this.methodInfoGetEnrollmentsByCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Enrollments, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Enrollments.deserializeBinary);
        this.methodInfoCreateEnrollment = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoUpdateEnrollment = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoUpdateEnrollments = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoGetSubmissions = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Submissions, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Submissions.deserializeBinary);
        this.methodInfoGetSubmissionsByCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.CourseSubmissions, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.CourseSubmissions.deserializeBinary);
        this.methodInfoUpdateSubmission = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoUpdateSubmissions = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoRebuildSubmission = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Submission, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Submission.deserializeBinary);
        this.methodInfoCreateBenchmark = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.GradingBenchmark, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.GradingBenchmark.deserializeBinary);
        this.methodInfoUpdateBenchmark = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoDeleteBenchmark = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoCreateCriterion = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.GradingCriterion, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.GradingCriterion.deserializeBinary);
        this.methodInfoUpdateCriterion = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoDeleteCriterion = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoCreateReview = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Review, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Review.deserializeBinary);
        this.methodInfoUpdateReview = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoGetReviewers = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Reviewers, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Reviewers.deserializeBinary);
        this.methodInfoGetProviders = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Providers, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Providers.deserializeBinary);
        this.methodInfoGetOrganization = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Organization, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Organization.deserializeBinary);
        this.methodInfoGetRepositories = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Repositories, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Repositories.deserializeBinary);
        this.methodInfoIsEmptyRepo = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.Void, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.Void.deserializeBinary);
        this.methodInfoGetStudentForDiscord = new grpcWeb.AbstractClientBase.MethodInfo(ag_pb_1.DiscordResponse, function (request) {
            return request.serializeBinary();
        }, ag_pb_1.DiscordResponse.deserializeBinary);
        if (!options)
            options = {};
        if (!credentials)
            credentials = {};
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
    AutograderServiceClient.prototype.getGroupsByCourse = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetGroupsByCourse', request, metadata || {}, this.methodInfoGetGroupsByCourse, callback);
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
    AutograderServiceClient.prototype.getCoursesByUser = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetCoursesByUser', request, metadata || {}, this.methodInfoGetCoursesByUser, callback);
    };
    AutograderServiceClient.prototype.createCourse = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/CreateCourse', request, metadata || {}, this.methodInfoCreateCourse, callback);
    };
    AutograderServiceClient.prototype.updateCourse = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/UpdateCourse', request, metadata || {}, this.methodInfoUpdateCourse, callback);
    };
    AutograderServiceClient.prototype.updateCourseVisibility = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/UpdateCourseVisibility', request, metadata || {}, this.methodInfoUpdateCourseVisibility, callback);
    };
    AutograderServiceClient.prototype.getAssignments = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetAssignments', request, metadata || {}, this.methodInfoGetAssignments, callback);
    };
    AutograderServiceClient.prototype.updateAssignments = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/UpdateAssignments', request, metadata || {}, this.methodInfoUpdateAssignments, callback);
    };
    AutograderServiceClient.prototype.getEnrollmentsByUser = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetEnrollmentsByUser', request, metadata || {}, this.methodInfoGetEnrollmentsByUser, callback);
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
    AutograderServiceClient.prototype.updateEnrollments = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/UpdateEnrollments', request, metadata || {}, this.methodInfoUpdateEnrollments, callback);
    };
    AutograderServiceClient.prototype.getSubmissions = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetSubmissions', request, metadata || {}, this.methodInfoGetSubmissions, callback);
    };
    AutograderServiceClient.prototype.getSubmissionsByCourse = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetSubmissionsByCourse', request, metadata || {}, this.methodInfoGetSubmissionsByCourse, callback);
    };
    AutograderServiceClient.prototype.updateSubmission = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/UpdateSubmission', request, metadata || {}, this.methodInfoUpdateSubmission, callback);
    };
    AutograderServiceClient.prototype.updateSubmissions = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/UpdateSubmissions', request, metadata || {}, this.methodInfoUpdateSubmissions, callback);
    };
    AutograderServiceClient.prototype.rebuildSubmission = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/RebuildSubmission', request, metadata || {}, this.methodInfoRebuildSubmission, callback);
    };
    AutograderServiceClient.prototype.createBenchmark = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/CreateBenchmark', request, metadata || {}, this.methodInfoCreateBenchmark, callback);
    };
    AutograderServiceClient.prototype.updateBenchmark = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/UpdateBenchmark', request, metadata || {}, this.methodInfoUpdateBenchmark, callback);
    };
    AutograderServiceClient.prototype.deleteBenchmark = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/DeleteBenchmark', request, metadata || {}, this.methodInfoDeleteBenchmark, callback);
    };
    AutograderServiceClient.prototype.createCriterion = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/CreateCriterion', request, metadata || {}, this.methodInfoCreateCriterion, callback);
    };
    AutograderServiceClient.prototype.updateCriterion = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/UpdateCriterion', request, metadata || {}, this.methodInfoUpdateCriterion, callback);
    };
    AutograderServiceClient.prototype.deleteCriterion = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/DeleteCriterion', request, metadata || {}, this.methodInfoDeleteCriterion, callback);
    };
    AutograderServiceClient.prototype.createReview = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/CreateReview', request, metadata || {}, this.methodInfoCreateReview, callback);
    };
    AutograderServiceClient.prototype.updateReview = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/UpdateReview', request, metadata || {}, this.methodInfoUpdateReview, callback);
    };
    AutograderServiceClient.prototype.getReviewers = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetReviewers', request, metadata || {}, this.methodInfoGetReviewers, callback);
    };
    AutograderServiceClient.prototype.getProviders = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetProviders', request, metadata || {}, this.methodInfoGetProviders, callback);
    };
    AutograderServiceClient.prototype.getOrganization = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetOrganization', request, metadata || {}, this.methodInfoGetOrganization, callback);
    };
    AutograderServiceClient.prototype.getRepositories = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetRepositories', request, metadata || {}, this.methodInfoGetRepositories, callback);
    };
    AutograderServiceClient.prototype.isEmptyRepo = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/IsEmptyRepo', request, metadata || {}, this.methodInfoIsEmptyRepo, callback);
    };
    AutograderServiceClient.prototype.getStudentForDiscord = function (request, metadata, callback) {
        return this.client_.rpcCall(this.hostname_ +
            '/AutograderService/GetStudentForDiscord', request, metadata || {}, this.methodInfoGetStudentForDiscord, callback);
    };
    return AutograderServiceClient;
}());
exports.AutograderServiceClient = AutograderServiceClient;
