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
        this.methodDescriptorGetUser = new grpcWeb.MethodDescriptor('/ag.AutograderService/GetUser', grpcWeb.MethodType.UNARY, ag_ag_pb.Void, ag_ag_pb.User, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.User.deserializeBinary);
        this.methodDescriptorGetUsers = new grpcWeb.MethodDescriptor('/ag.AutograderService/GetUsers', grpcWeb.MethodType.UNARY, ag_ag_pb.Void, ag_ag_pb.Users, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Users.deserializeBinary);
        this.methodDescriptorGetUserByCourse = new grpcWeb.MethodDescriptor('/ag.AutograderService/GetUserByCourse', grpcWeb.MethodType.UNARY, ag_ag_pb.CourseUserRequest, ag_ag_pb.User, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.User.deserializeBinary);
        this.methodDescriptorUpdateUser = new grpcWeb.MethodDescriptor('/ag.AutograderService/UpdateUser', grpcWeb.MethodType.UNARY, ag_ag_pb.User, ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodDescriptorIsAuthorizedTeacher = new grpcWeb.MethodDescriptor('/ag.AutograderService/IsAuthorizedTeacher', grpcWeb.MethodType.UNARY, ag_ag_pb.Void, ag_ag_pb.AuthorizationResponse, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.AuthorizationResponse.deserializeBinary);
        this.methodDescriptorGetGroup = new grpcWeb.MethodDescriptor('/ag.AutograderService/GetGroup', grpcWeb.MethodType.UNARY, ag_ag_pb.GetGroupRequest, ag_ag_pb.Group, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Group.deserializeBinary);
        this.methodDescriptorGetGroupByUserAndCourse = new grpcWeb.MethodDescriptor('/ag.AutograderService/GetGroupByUserAndCourse', grpcWeb.MethodType.UNARY, ag_ag_pb.GroupRequest, ag_ag_pb.Group, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Group.deserializeBinary);
        this.methodDescriptorGetGroupsByCourse = new grpcWeb.MethodDescriptor('/ag.AutograderService/GetGroupsByCourse', grpcWeb.MethodType.UNARY, ag_ag_pb.CourseRequest, ag_ag_pb.Groups, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Groups.deserializeBinary);
        this.methodDescriptorCreateGroup = new grpcWeb.MethodDescriptor('/ag.AutograderService/CreateGroup', grpcWeb.MethodType.UNARY, ag_ag_pb.Group, ag_ag_pb.Group, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Group.deserializeBinary);
        this.methodDescriptorUpdateGroup = new grpcWeb.MethodDescriptor('/ag.AutograderService/UpdateGroup', grpcWeb.MethodType.UNARY, ag_ag_pb.Group, ag_ag_pb.Group, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Group.deserializeBinary);
        this.methodDescriptorDeleteGroup = new grpcWeb.MethodDescriptor('/ag.AutograderService/DeleteGroup', grpcWeb.MethodType.UNARY, ag_ag_pb.GroupRequest, ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodDescriptorGetCourse = new grpcWeb.MethodDescriptor('/ag.AutograderService/GetCourse', grpcWeb.MethodType.UNARY, ag_ag_pb.CourseRequest, ag_ag_pb.Course, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Course.deserializeBinary);
        this.methodDescriptorGetCourses = new grpcWeb.MethodDescriptor('/ag.AutograderService/GetCourses', grpcWeb.MethodType.UNARY, ag_ag_pb.Void, ag_ag_pb.Courses, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Courses.deserializeBinary);
        this.methodDescriptorGetCoursesByUser = new grpcWeb.MethodDescriptor('/ag.AutograderService/GetCoursesByUser', grpcWeb.MethodType.UNARY, ag_ag_pb.EnrollmentStatusRequest, ag_ag_pb.Courses, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Courses.deserializeBinary);
        this.methodDescriptorCreateCourse = new grpcWeb.MethodDescriptor('/ag.AutograderService/CreateCourse', grpcWeb.MethodType.UNARY, ag_ag_pb.Course, ag_ag_pb.Course, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Course.deserializeBinary);
        this.methodDescriptorUpdateCourse = new grpcWeb.MethodDescriptor('/ag.AutograderService/UpdateCourse', grpcWeb.MethodType.UNARY, ag_ag_pb.Course, ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodDescriptorUpdateCourseVisibility = new grpcWeb.MethodDescriptor('/ag.AutograderService/UpdateCourseVisibility', grpcWeb.MethodType.UNARY, ag_ag_pb.Enrollment, ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodDescriptorGetAssignments = new grpcWeb.MethodDescriptor('/ag.AutograderService/GetAssignments', grpcWeb.MethodType.UNARY, ag_ag_pb.CourseRequest, ag_ag_pb.Assignments, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Assignments.deserializeBinary);
        this.methodDescriptorUpdateAssignments = new grpcWeb.MethodDescriptor('/ag.AutograderService/UpdateAssignments', grpcWeb.MethodType.UNARY, ag_ag_pb.CourseRequest, ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodDescriptorGetEnrollmentsByUser = new grpcWeb.MethodDescriptor('/ag.AutograderService/GetEnrollmentsByUser', grpcWeb.MethodType.UNARY, ag_ag_pb.EnrollmentStatusRequest, ag_ag_pb.Enrollments, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Enrollments.deserializeBinary);
        this.methodDescriptorGetEnrollmentsByCourse = new grpcWeb.MethodDescriptor('/ag.AutograderService/GetEnrollmentsByCourse', grpcWeb.MethodType.UNARY, ag_ag_pb.EnrollmentRequest, ag_ag_pb.Enrollments, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Enrollments.deserializeBinary);
        this.methodDescriptorCreateEnrollment = new grpcWeb.MethodDescriptor('/ag.AutograderService/CreateEnrollment', grpcWeb.MethodType.UNARY, ag_ag_pb.Enrollment, ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodDescriptorUpdateEnrollments = new grpcWeb.MethodDescriptor('/ag.AutograderService/UpdateEnrollments', grpcWeb.MethodType.UNARY, ag_ag_pb.Enrollments, ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodDescriptorGetSubmissions = new grpcWeb.MethodDescriptor('/ag.AutograderService/GetSubmissions', grpcWeb.MethodType.UNARY, ag_ag_pb.SubmissionRequest, ag_ag_pb.Submissions, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Submissions.deserializeBinary);
        this.methodDescriptorGetSubmissionsByCourse = new grpcWeb.MethodDescriptor('/ag.AutograderService/GetSubmissionsByCourse', grpcWeb.MethodType.UNARY, ag_ag_pb.SubmissionsForCourseRequest, ag_ag_pb.CourseSubmissions, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.CourseSubmissions.deserializeBinary);
        this.methodDescriptorUpdateSubmission = new grpcWeb.MethodDescriptor('/ag.AutograderService/UpdateSubmission', grpcWeb.MethodType.UNARY, ag_ag_pb.UpdateSubmissionRequest, ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodDescriptorUpdateSubmissions = new grpcWeb.MethodDescriptor('/ag.AutograderService/UpdateSubmissions', grpcWeb.MethodType.UNARY, ag_ag_pb.UpdateSubmissionsRequest, ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodDescriptorRebuildSubmissions = new grpcWeb.MethodDescriptor('/ag.AutograderService/RebuildSubmissions', grpcWeb.MethodType.UNARY, ag_ag_pb.RebuildRequest, ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodDescriptorCreateBenchmark = new grpcWeb.MethodDescriptor('/ag.AutograderService/CreateBenchmark', grpcWeb.MethodType.UNARY, ag_ag_pb.GradingBenchmark, ag_ag_pb.GradingBenchmark, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.GradingBenchmark.deserializeBinary);
        this.methodDescriptorUpdateBenchmark = new grpcWeb.MethodDescriptor('/ag.AutograderService/UpdateBenchmark', grpcWeb.MethodType.UNARY, ag_ag_pb.GradingBenchmark, ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodDescriptorDeleteBenchmark = new grpcWeb.MethodDescriptor('/ag.AutograderService/DeleteBenchmark', grpcWeb.MethodType.UNARY, ag_ag_pb.GradingBenchmark, ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodDescriptorCreateCriterion = new grpcWeb.MethodDescriptor('/ag.AutograderService/CreateCriterion', grpcWeb.MethodType.UNARY, ag_ag_pb.GradingCriterion, ag_ag_pb.GradingCriterion, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.GradingCriterion.deserializeBinary);
        this.methodDescriptorUpdateCriterion = new grpcWeb.MethodDescriptor('/ag.AutograderService/UpdateCriterion', grpcWeb.MethodType.UNARY, ag_ag_pb.GradingCriterion, ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodDescriptorDeleteCriterion = new grpcWeb.MethodDescriptor('/ag.AutograderService/DeleteCriterion', grpcWeb.MethodType.UNARY, ag_ag_pb.GradingCriterion, ag_ag_pb.Void, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Void.deserializeBinary);
        this.methodDescriptorCreateReview = new grpcWeb.MethodDescriptor('/ag.AutograderService/CreateReview', grpcWeb.MethodType.UNARY, ag_ag_pb.ReviewRequest, ag_ag_pb.Review, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Review.deserializeBinary);
        this.methodDescriptorUpdateReview = new grpcWeb.MethodDescriptor('/ag.AutograderService/UpdateReview', grpcWeb.MethodType.UNARY, ag_ag_pb.ReviewRequest, ag_ag_pb.Review, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Review.deserializeBinary);
        this.methodDescriptorGetReviewers = new grpcWeb.MethodDescriptor('/ag.AutograderService/GetReviewers', grpcWeb.MethodType.UNARY, ag_ag_pb.SubmissionReviewersRequest, ag_ag_pb.Reviewers, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Reviewers.deserializeBinary);
        this.methodDescriptorGetProviders = new grpcWeb.MethodDescriptor('/ag.AutograderService/GetProviders', grpcWeb.MethodType.UNARY, ag_ag_pb.Void, ag_ag_pb.Providers, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Providers.deserializeBinary);
        this.methodDescriptorGetOrganization = new grpcWeb.MethodDescriptor('/ag.AutograderService/GetOrganization', grpcWeb.MethodType.UNARY, ag_ag_pb.OrgRequest, ag_ag_pb.Organization, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Organization.deserializeBinary);
        this.methodDescriptorGetRepositories = new grpcWeb.MethodDescriptor('/ag.AutograderService/GetRepositories', grpcWeb.MethodType.UNARY, ag_ag_pb.URLRequest, ag_ag_pb.Repositories, function (request) {
            return request.serializeBinary();
        }, ag_ag_pb.Repositories.deserializeBinary);
        this.methodDescriptorIsEmptyRepo = new grpcWeb.MethodDescriptor('/ag.AutograderService/IsEmptyRepo', grpcWeb.MethodType.UNARY, ag_ag_pb.RepositoryRequest, ag_ag_pb.Void, function (request) {
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
                '/ag.AutograderService/GetUser', request, metadata || {}, this.methodDescriptorGetUser, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetUser', request, metadata || {}, this.methodDescriptorGetUser);
    };
    AutograderServiceClient.prototype.getUsers = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetUsers', request, metadata || {}, this.methodDescriptorGetUsers, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetUsers', request, metadata || {}, this.methodDescriptorGetUsers);
    };
    AutograderServiceClient.prototype.getUserByCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetUserByCourse', request, metadata || {}, this.methodDescriptorGetUserByCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetUserByCourse', request, metadata || {}, this.methodDescriptorGetUserByCourse);
    };
    AutograderServiceClient.prototype.updateUser = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateUser', request, metadata || {}, this.methodDescriptorUpdateUser, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateUser', request, metadata || {}, this.methodDescriptorUpdateUser);
    };
    AutograderServiceClient.prototype.isAuthorizedTeacher = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/IsAuthorizedTeacher', request, metadata || {}, this.methodDescriptorIsAuthorizedTeacher, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/IsAuthorizedTeacher', request, metadata || {}, this.methodDescriptorIsAuthorizedTeacher);
    };
    AutograderServiceClient.prototype.getGroup = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetGroup', request, metadata || {}, this.methodDescriptorGetGroup, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetGroup', request, metadata || {}, this.methodDescriptorGetGroup);
    };
    AutograderServiceClient.prototype.getGroupByUserAndCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetGroupByUserAndCourse', request, metadata || {}, this.methodDescriptorGetGroupByUserAndCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetGroupByUserAndCourse', request, metadata || {}, this.methodDescriptorGetGroupByUserAndCourse);
    };
    AutograderServiceClient.prototype.getGroupsByCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetGroupsByCourse', request, metadata || {}, this.methodDescriptorGetGroupsByCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetGroupsByCourse', request, metadata || {}, this.methodDescriptorGetGroupsByCourse);
    };
    AutograderServiceClient.prototype.createGroup = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/CreateGroup', request, metadata || {}, this.methodDescriptorCreateGroup, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/CreateGroup', request, metadata || {}, this.methodDescriptorCreateGroup);
    };
    AutograderServiceClient.prototype.updateGroup = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateGroup', request, metadata || {}, this.methodDescriptorUpdateGroup, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateGroup', request, metadata || {}, this.methodDescriptorUpdateGroup);
    };
    AutograderServiceClient.prototype.deleteGroup = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/DeleteGroup', request, metadata || {}, this.methodDescriptorDeleteGroup, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/DeleteGroup', request, metadata || {}, this.methodDescriptorDeleteGroup);
    };
    AutograderServiceClient.prototype.getCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetCourse', request, metadata || {}, this.methodDescriptorGetCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetCourse', request, metadata || {}, this.methodDescriptorGetCourse);
    };
    AutograderServiceClient.prototype.getCourses = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetCourses', request, metadata || {}, this.methodDescriptorGetCourses, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetCourses', request, metadata || {}, this.methodDescriptorGetCourses);
    };
    AutograderServiceClient.prototype.getCoursesByUser = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetCoursesByUser', request, metadata || {}, this.methodDescriptorGetCoursesByUser, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetCoursesByUser', request, metadata || {}, this.methodDescriptorGetCoursesByUser);
    };
    AutograderServiceClient.prototype.createCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/CreateCourse', request, metadata || {}, this.methodDescriptorCreateCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/CreateCourse', request, metadata || {}, this.methodDescriptorCreateCourse);
    };
    AutograderServiceClient.prototype.updateCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateCourse', request, metadata || {}, this.methodDescriptorUpdateCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateCourse', request, metadata || {}, this.methodDescriptorUpdateCourse);
    };
    AutograderServiceClient.prototype.updateCourseVisibility = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateCourseVisibility', request, metadata || {}, this.methodDescriptorUpdateCourseVisibility, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateCourseVisibility', request, metadata || {}, this.methodDescriptorUpdateCourseVisibility);
    };
    AutograderServiceClient.prototype.getAssignments = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetAssignments', request, metadata || {}, this.methodDescriptorGetAssignments, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetAssignments', request, metadata || {}, this.methodDescriptorGetAssignments);
    };
    AutograderServiceClient.prototype.updateAssignments = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateAssignments', request, metadata || {}, this.methodDescriptorUpdateAssignments, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateAssignments', request, metadata || {}, this.methodDescriptorUpdateAssignments);
    };
    AutograderServiceClient.prototype.getEnrollmentsByUser = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetEnrollmentsByUser', request, metadata || {}, this.methodDescriptorGetEnrollmentsByUser, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetEnrollmentsByUser', request, metadata || {}, this.methodDescriptorGetEnrollmentsByUser);
    };
    AutograderServiceClient.prototype.getEnrollmentsByCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetEnrollmentsByCourse', request, metadata || {}, this.methodDescriptorGetEnrollmentsByCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetEnrollmentsByCourse', request, metadata || {}, this.methodDescriptorGetEnrollmentsByCourse);
    };
    AutograderServiceClient.prototype.createEnrollment = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/CreateEnrollment', request, metadata || {}, this.methodDescriptorCreateEnrollment, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/CreateEnrollment', request, metadata || {}, this.methodDescriptorCreateEnrollment);
    };
    AutograderServiceClient.prototype.updateEnrollments = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateEnrollments', request, metadata || {}, this.methodDescriptorUpdateEnrollments, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateEnrollments', request, metadata || {}, this.methodDescriptorUpdateEnrollments);
    };
    AutograderServiceClient.prototype.getSubmissions = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetSubmissions', request, metadata || {}, this.methodDescriptorGetSubmissions, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetSubmissions', request, metadata || {}, this.methodDescriptorGetSubmissions);
    };
    AutograderServiceClient.prototype.getSubmissionsByCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetSubmissionsByCourse', request, metadata || {}, this.methodDescriptorGetSubmissionsByCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetSubmissionsByCourse', request, metadata || {}, this.methodDescriptorGetSubmissionsByCourse);
    };
    AutograderServiceClient.prototype.updateSubmission = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateSubmission', request, metadata || {}, this.methodDescriptorUpdateSubmission, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateSubmission', request, metadata || {}, this.methodDescriptorUpdateSubmission);
    };
    AutograderServiceClient.prototype.updateSubmissions = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateSubmissions', request, metadata || {}, this.methodDescriptorUpdateSubmissions, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateSubmissions', request, metadata || {}, this.methodDescriptorUpdateSubmissions);
    };
    AutograderServiceClient.prototype.rebuildSubmissions = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/RebuildSubmissions', request, metadata || {}, this.methodDescriptorRebuildSubmissions, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/RebuildSubmissions', request, metadata || {}, this.methodDescriptorRebuildSubmissions);
    };
    AutograderServiceClient.prototype.createBenchmark = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/CreateBenchmark', request, metadata || {}, this.methodDescriptorCreateBenchmark, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/CreateBenchmark', request, metadata || {}, this.methodDescriptorCreateBenchmark);
    };
    AutograderServiceClient.prototype.updateBenchmark = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateBenchmark', request, metadata || {}, this.methodDescriptorUpdateBenchmark, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateBenchmark', request, metadata || {}, this.methodDescriptorUpdateBenchmark);
    };
    AutograderServiceClient.prototype.deleteBenchmark = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/DeleteBenchmark', request, metadata || {}, this.methodDescriptorDeleteBenchmark, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/DeleteBenchmark', request, metadata || {}, this.methodDescriptorDeleteBenchmark);
    };
    AutograderServiceClient.prototype.createCriterion = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/CreateCriterion', request, metadata || {}, this.methodDescriptorCreateCriterion, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/CreateCriterion', request, metadata || {}, this.methodDescriptorCreateCriterion);
    };
    AutograderServiceClient.prototype.updateCriterion = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateCriterion', request, metadata || {}, this.methodDescriptorUpdateCriterion, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateCriterion', request, metadata || {}, this.methodDescriptorUpdateCriterion);
    };
    AutograderServiceClient.prototype.deleteCriterion = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/DeleteCriterion', request, metadata || {}, this.methodDescriptorDeleteCriterion, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/DeleteCriterion', request, metadata || {}, this.methodDescriptorDeleteCriterion);
    };
    AutograderServiceClient.prototype.createReview = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/CreateReview', request, metadata || {}, this.methodDescriptorCreateReview, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/CreateReview', request, metadata || {}, this.methodDescriptorCreateReview);
    };
    AutograderServiceClient.prototype.updateReview = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/UpdateReview', request, metadata || {}, this.methodDescriptorUpdateReview, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/UpdateReview', request, metadata || {}, this.methodDescriptorUpdateReview);
    };
    AutograderServiceClient.prototype.getReviewers = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetReviewers', request, metadata || {}, this.methodDescriptorGetReviewers, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetReviewers', request, metadata || {}, this.methodDescriptorGetReviewers);
    };
    AutograderServiceClient.prototype.getProviders = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetProviders', request, metadata || {}, this.methodDescriptorGetProviders, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetProviders', request, metadata || {}, this.methodDescriptorGetProviders);
    };
    AutograderServiceClient.prototype.getOrganization = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetOrganization', request, metadata || {}, this.methodDescriptorGetOrganization, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetOrganization', request, metadata || {}, this.methodDescriptorGetOrganization);
    };
    AutograderServiceClient.prototype.getRepositories = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/GetRepositories', request, metadata || {}, this.methodDescriptorGetRepositories, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/GetRepositories', request, metadata || {}, this.methodDescriptorGetRepositories);
    };
    AutograderServiceClient.prototype.isEmptyRepo = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/ag.AutograderService/IsEmptyRepo', request, metadata || {}, this.methodDescriptorIsEmptyRepo, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/ag.AutograderService/IsEmptyRepo', request, metadata || {}, this.methodDescriptorIsEmptyRepo);
    };
    return AutograderServiceClient;
}());
exports.AutograderServiceClient = AutograderServiceClient;
