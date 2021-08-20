"use strict";
/**
 * @fileoverview gRPC-Web generated client stub for ag
 * @enhanceable
 * @public
 */
exports.__esModule = true;
exports.AutograderServiceClient = void 0;
// GENERATED CODE -- DO NOT EDIT!
/* eslint-disable */
// @ts-nocheck
var grpcWeb = require("grpc-web");
var ag_ag_pb = require("../ag/ag_pb");
var AutograderServiceClient = /** @class */ (function () {
    function AutograderServiceClient(hostname, credentials, options) {
        this.methodInfoGetUser = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.User, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.User.deserializeBinary);
        this.methodInfoGetUsers = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Users, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Users.deserializeBinary);
        this.methodInfoGetUserByCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.User, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.User.deserializeBinary);
        this.methodInfoUpdateUser = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodInfoIsAuthorizedTeacher = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.AuthorizationResponse, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.AuthorizationResponse.deserializeBinary);
        this.methodInfoGetGroup = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Group, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Group.deserializeBinary);
        this.methodInfoGetGroupByUserAndCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Group, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Group.deserializeBinary);
        this.methodInfoGetGroupsByCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Groups, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Groups.deserializeBinary);
        this.methodInfoCreateGroup = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Group, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Group.deserializeBinary);
        this.methodInfoUpdateGroup = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodInfoDeleteGroup = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodInfoGetCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Course, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Course.deserializeBinary);
        this.methodInfoGetCourses = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Courses, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Courses.deserializeBinary);
        this.methodInfoGetCoursesByUser = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Courses, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Courses.deserializeBinary);
        this.methodInfoCreateCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Course, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Course.deserializeBinary);
        this.methodInfoUpdateCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodInfoUpdateCourseVisibility = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodInfoGetAssignments = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Assignments, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Assignments.deserializeBinary);
        this.methodInfoUpdateAssignments = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodInfoGetEnrollmentsByUser = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Enrollments, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Enrollments.deserializeBinary);
        this.methodInfoGetEnrollmentsByCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Enrollments, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Enrollments.deserializeBinary);
        this.methodInfoCreateEnrollment = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodInfoUpdateEnrollment = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodInfoUpdateEnrollments = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodInfoGetSubmissions = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Submissions, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Submissions.deserializeBinary);
        this.methodInfoGetSubmissionsByCourse = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.CourseSubmissions, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.CourseSubmissions.deserializeBinary);
        this.methodInfoUpdateSubmission = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodInfoUpdateSubmissions = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodInfoRebuildSubmission = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Submission, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Submission.deserializeBinary);
        this.methodInfoRebuildSubmissions = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodInfoCreateBenchmark = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.GradingBenchmark, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.GradingBenchmark.deserializeBinary);
        this.methodInfoUpdateBenchmark = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodInfoDeleteBenchmark = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodInfoCreateCriterion = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.GradingCriterion, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.GradingCriterion.deserializeBinary);
        this.methodInfoUpdateCriterion = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodInfoDeleteCriterion = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodInfoCreateReview = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Review, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Review.deserializeBinary);
        this.methodInfoUpdateReview = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Review, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Review.deserializeBinary);
        this.methodInfoGetReviewers = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Reviewers, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Reviewers.deserializeBinary);
        this.methodInfoGetProviders = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Providers, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Providers.deserializeBinary);
        this.methodInfoGetOrganization = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Organization, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Organization.deserializeBinary);
        this.methodInfoGetRepositories = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Repositories, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Repositories.deserializeBinary);
        this.methodInfoIsEmptyRepo = new grpcWeb.AbstractClientBase.MethodInfo(ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        if (!options)
            options = {};
        if (!credentials)
            credentials = {};
        options['format'] = 'text';
        this.client_ = new grpcWeb.GrpcWebClientBase(options);
        this.hostname_ = hostname;
        this.credentials_ = credentials;
        this.options_ = options;
    }
    AutograderServiceClient.prototype.getUser = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetUser', request, metadata || {}, this.methodInfoGetUser, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetUser', request, metadata || {}, this.methodInfoGetUser);
    };
    AutograderServiceClient.prototype.getUsers = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetUsers', request, metadata || {}, this.methodInfoGetUsers, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetUsers', request, metadata || {}, this.methodInfoGetUsers);
    };
    AutograderServiceClient.prototype.getUserByCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetUserByCourse', request, metadata || {}, this.methodInfoGetUserByCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetUserByCourse', request, metadata || {}, this.methodInfoGetUserByCourse);
    };
    AutograderServiceClient.prototype.updateUser = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateUser', request, metadata || {}, this.methodInfoUpdateUser, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateUser', request, metadata || {}, this.methodInfoUpdateUser);
    };
    AutograderServiceClient.prototype.isAuthorizedTeacher = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/IsAuthorizedTeacher', request, metadata || {}, this.methodInfoIsAuthorizedTeacher, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/IsAuthorizedTeacher', request, metadata || {}, this.methodInfoIsAuthorizedTeacher);
    };
    AutograderServiceClient.prototype.getGroup = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetGroup', request, metadata || {}, this.methodInfoGetGroup, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetGroup', request, metadata || {}, this.methodInfoGetGroup);
    };
    AutograderServiceClient.prototype.getGroupByUserAndCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetGroupByUserAndCourse', request, metadata || {}, this.methodInfoGetGroupByUserAndCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetGroupByUserAndCourse', request, metadata || {}, this.methodInfoGetGroupByUserAndCourse);
    };
    AutograderServiceClient.prototype.getGroupsByCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetGroupsByCourse', request, metadata || {}, this.methodInfoGetGroupsByCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetGroupsByCourse', request, metadata || {}, this.methodInfoGetGroupsByCourse);
    };
    AutograderServiceClient.prototype.createGroup = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/CreateGroup', request, metadata || {}, this.methodInfoCreateGroup, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/CreateGroup', request, metadata || {}, this.methodInfoCreateGroup);
    };
    AutograderServiceClient.prototype.updateGroup = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateGroup', request, metadata || {}, this.methodInfoUpdateGroup, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateGroup', request, metadata || {}, this.methodInfoUpdateGroup);
    };
    AutograderServiceClient.prototype.deleteGroup = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/DeleteGroup', request, metadata || {}, this.methodInfoDeleteGroup, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/DeleteGroup', request, metadata || {}, this.methodInfoDeleteGroup);
    };
    AutograderServiceClient.prototype.getCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetCourse', request, metadata || {}, this.methodInfoGetCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetCourse', request, metadata || {}, this.methodInfoGetCourse);
    };
    AutograderServiceClient.prototype.getCourses = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetCourses', request, metadata || {}, this.methodInfoGetCourses, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetCourses', request, metadata || {}, this.methodInfoGetCourses);
    };
    AutograderServiceClient.prototype.getCoursesByUser = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetCoursesByUser', request, metadata || {}, this.methodInfoGetCoursesByUser, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetCoursesByUser', request, metadata || {}, this.methodInfoGetCoursesByUser);
    };
    AutograderServiceClient.prototype.createCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/CreateCourse', request, metadata || {}, this.methodInfoCreateCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/CreateCourse', request, metadata || {}, this.methodInfoCreateCourse);
    };
    AutograderServiceClient.prototype.updateCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateCourse', request, metadata || {}, this.methodInfoUpdateCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateCourse', request, metadata || {}, this.methodInfoUpdateCourse);
    };
    AutograderServiceClient.prototype.updateCourseVisibility = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateCourseVisibility', request, metadata || {}, this.methodInfoUpdateCourseVisibility, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateCourseVisibility', request, metadata || {}, this.methodInfoUpdateCourseVisibility);
    };
    AutograderServiceClient.prototype.getAssignments = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetAssignments', request, metadata || {}, this.methodInfoGetAssignments, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetAssignments', request, metadata || {}, this.methodInfoGetAssignments);
    };
    AutograderServiceClient.prototype.updateAssignments = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateAssignments', request, metadata || {}, this.methodInfoUpdateAssignments, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateAssignments', request, metadata || {}, this.methodInfoUpdateAssignments);
    };
    AutograderServiceClient.prototype.getEnrollmentsByUser = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetEnrollmentsByUser', request, metadata || {}, this.methodInfoGetEnrollmentsByUser, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetEnrollmentsByUser', request, metadata || {}, this.methodInfoGetEnrollmentsByUser);
    };
    AutograderServiceClient.prototype.getEnrollmentsByCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetEnrollmentsByCourse', request, metadata || {}, this.methodInfoGetEnrollmentsByCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetEnrollmentsByCourse', request, metadata || {}, this.methodInfoGetEnrollmentsByCourse);
    };
    AutograderServiceClient.prototype.createEnrollment = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/CreateEnrollment', request, metadata || {}, this.methodInfoCreateEnrollment, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/CreateEnrollment', request, metadata || {}, this.methodInfoCreateEnrollment);
    };
    AutograderServiceClient.prototype.updateEnrollment = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateEnrollment', request, metadata || {}, this.methodInfoUpdateEnrollment, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateEnrollment', request, metadata || {}, this.methodInfoUpdateEnrollment);
    };
    AutograderServiceClient.prototype.updateEnrollments = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateEnrollments', request, metadata || {}, this.methodInfoUpdateEnrollments, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateEnrollments', request, metadata || {}, this.methodInfoUpdateEnrollments);
    };
    AutograderServiceClient.prototype.getSubmissions = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetSubmissions', request, metadata || {}, this.methodInfoGetSubmissions, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetSubmissions', request, metadata || {}, this.methodInfoGetSubmissions);
    };
    AutograderServiceClient.prototype.getSubmissionsByCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetSubmissionsByCourse', request, metadata || {}, this.methodInfoGetSubmissionsByCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetSubmissionsByCourse', request, metadata || {}, this.methodInfoGetSubmissionsByCourse);
    };
    AutograderServiceClient.prototype.updateSubmission = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateSubmission', request, metadata || {}, this.methodInfoUpdateSubmission, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateSubmission', request, metadata || {}, this.methodInfoUpdateSubmission);
    };
    AutograderServiceClient.prototype.updateSubmissions = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateSubmissions', request, metadata || {}, this.methodInfoUpdateSubmissions, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateSubmissions', request, metadata || {}, this.methodInfoUpdateSubmissions);
    };
    AutograderServiceClient.prototype.rebuildSubmission = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/RebuildSubmission', request, metadata || {}, this.methodInfoRebuildSubmission, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/RebuildSubmission', request, metadata || {}, this.methodInfoRebuildSubmission);
    };
    AutograderServiceClient.prototype.rebuildSubmissions = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/RebuildSubmissions', request, metadata || {}, this.methodInfoRebuildSubmissions, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/RebuildSubmissions', request, metadata || {}, this.methodInfoRebuildSubmissions);
    };
    AutograderServiceClient.prototype.createBenchmark = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/CreateBenchmark', request, metadata || {}, this.methodInfoCreateBenchmark, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/CreateBenchmark', request, metadata || {}, this.methodInfoCreateBenchmark);
    };
    AutograderServiceClient.prototype.updateBenchmark = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateBenchmark', request, metadata || {}, this.methodInfoUpdateBenchmark, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateBenchmark', request, metadata || {}, this.methodInfoUpdateBenchmark);
    };
    AutograderServiceClient.prototype.deleteBenchmark = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/DeleteBenchmark', request, metadata || {}, this.methodInfoDeleteBenchmark, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/DeleteBenchmark', request, metadata || {}, this.methodInfoDeleteBenchmark);
    };
    AutograderServiceClient.prototype.createCriterion = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/CreateCriterion', request, metadata || {}, this.methodInfoCreateCriterion, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/CreateCriterion', request, metadata || {}, this.methodInfoCreateCriterion);
    };
    AutograderServiceClient.prototype.updateCriterion = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateCriterion', request, metadata || {}, this.methodInfoUpdateCriterion, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateCriterion', request, metadata || {}, this.methodInfoUpdateCriterion);
    };
    AutograderServiceClient.prototype.deleteCriterion = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/DeleteCriterion', request, metadata || {}, this.methodInfoDeleteCriterion, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/DeleteCriterion', request, metadata || {}, this.methodInfoDeleteCriterion);
    };
    AutograderServiceClient.prototype.createReview = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/CreateReview', request, metadata || {}, this.methodInfoCreateReview, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/CreateReview', request, metadata || {}, this.methodInfoCreateReview);
    };
    AutograderServiceClient.prototype.updateReview = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateReview', request, metadata || {}, this.methodInfoUpdateReview, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateReview', request, metadata || {}, this.methodInfoUpdateReview);
    };
    AutograderServiceClient.prototype.getReviewers = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetReviewers', request, metadata || {}, this.methodInfoGetReviewers, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetReviewers', request, metadata || {}, this.methodInfoGetReviewers);
    };
    AutograderServiceClient.prototype.getProviders = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetProviders', request, metadata || {}, this.methodInfoGetProviders, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetProviders', request, metadata || {}, this.methodInfoGetProviders);
    };
    AutograderServiceClient.prototype.getOrganization = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetOrganization', request, metadata || {}, this.methodInfoGetOrganization, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetOrganization', request, metadata || {}, this.methodInfoGetOrganization);
    };
    AutograderServiceClient.prototype.getRepositories = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetRepositories', request, metadata || {}, this.methodInfoGetRepositories, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetRepositories', request, metadata || {}, this.methodInfoGetRepositories);
    };
    AutograderServiceClient.prototype.isEmptyRepo = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/IsEmptyRepo', request, metadata || {}, this.methodInfoIsEmptyRepo, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/IsEmptyRepo', request, metadata || {}, this.methodInfoIsEmptyRepo);
    };
    return AutograderServiceClient;
}());
exports.AutograderServiceClient = AutograderServiceClient;
