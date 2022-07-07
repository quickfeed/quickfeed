"use strict";
/**
 * @fileoverview gRPC-Web generated client stub for qf
 * @enhanceable
 * @public
 */
exports.__esModule = true;
exports.QuickFeedServiceClient = void 0;
// GENERATED CODE -- DO NOT EDIT!
/* eslint-disable */
// @ts-nocheck
var grpcWeb = require("grpc-web");
var qf_types_types_pb = require("../qf/types/types_pb");
var qf_types_requests_pb = require("../qf/types/requests_pb");
var QuickFeedServiceClient = /** @class */ (function () {
    function QuickFeedServiceClient(hostname, credentials, options) {
        this.methodInfoGetUser = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetUser', grpcWeb.MethodType.UNARY, qf_types_requests_pb.Void, qf_types_types_pb.User, function (request) {
            return request.serializeBinary();
        }, qf_types_types_pb.User.deserializeBinary);
        this.methodInfoGetUsers = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetUsers', grpcWeb.MethodType.UNARY, qf_types_requests_pb.Void, qf_types_types_pb.Users, function (request) {
            return request.serializeBinary();
        }, qf_types_types_pb.Users.deserializeBinary);
        this.methodInfoGetUserByCourse = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetUserByCourse', grpcWeb.MethodType.UNARY, qf_types_requests_pb.CourseUserRequest, qf_types_types_pb.User, function (request) {
            return request.serializeBinary();
        }, qf_types_types_pb.User.deserializeBinary);
        this.methodInfoUpdateUser = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/UpdateUser', grpcWeb.MethodType.UNARY, qf_types_types_pb.User, qf_types_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_types_requests_pb.Void.deserializeBinary);
        this.methodInfoIsAuthorizedTeacher = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/IsAuthorizedTeacher', grpcWeb.MethodType.UNARY, qf_types_requests_pb.Void, qf_types_requests_pb.AuthorizationResponse, function (request) {
            return request.serializeBinary();
        }, qf_types_requests_pb.AuthorizationResponse.deserializeBinary);
        this.methodInfoGetGroup = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetGroup', grpcWeb.MethodType.UNARY, qf_types_requests_pb.GetGroupRequest, qf_types_types_pb.Group, function (request) {
            return request.serializeBinary();
        }, qf_types_types_pb.Group.deserializeBinary);
        this.methodInfoGetGroupByUserAndCourse = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetGroupByUserAndCourse', grpcWeb.MethodType.UNARY, qf_types_requests_pb.GroupRequest, qf_types_types_pb.Group, function (request) {
            return request.serializeBinary();
        }, qf_types_types_pb.Group.deserializeBinary);
        this.methodInfoGetGroupsByCourse = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetGroupsByCourse', grpcWeb.MethodType.UNARY, qf_types_requests_pb.CourseRequest, qf_types_types_pb.Groups, function (request) {
            return request.serializeBinary();
        }, qf_types_types_pb.Groups.deserializeBinary);
        this.methodInfoCreateGroup = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/CreateGroup', grpcWeb.MethodType.UNARY, qf_types_types_pb.Group, qf_types_types_pb.Group, function (request) {
            return request.serializeBinary();
        }, qf_types_types_pb.Group.deserializeBinary);
        this.methodInfoUpdateGroup = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/UpdateGroup', grpcWeb.MethodType.UNARY, qf_types_types_pb.Group, qf_types_types_pb.Group, function (request) {
            return request.serializeBinary();
        }, qf_types_types_pb.Group.deserializeBinary);
        this.methodInfoDeleteGroup = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/DeleteGroup', grpcWeb.MethodType.UNARY, qf_types_requests_pb.GroupRequest, qf_types_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_types_requests_pb.Void.deserializeBinary);
        this.methodInfoGetCourse = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetCourse', grpcWeb.MethodType.UNARY, qf_types_requests_pb.CourseRequest, qf_types_types_pb.Course, function (request) {
            return request.serializeBinary();
        }, qf_types_types_pb.Course.deserializeBinary);
        this.methodInfoGetCourses = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetCourses', grpcWeb.MethodType.UNARY, qf_types_requests_pb.Void, qf_types_types_pb.Courses, function (request) {
            return request.serializeBinary();
        }, qf_types_types_pb.Courses.deserializeBinary);
        this.methodInfoGetCoursesByUser = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetCoursesByUser', grpcWeb.MethodType.UNARY, qf_types_requests_pb.EnrollmentStatusRequest, qf_types_types_pb.Courses, function (request) {
            return request.serializeBinary();
        }, qf_types_types_pb.Courses.deserializeBinary);
        this.methodInfoCreateCourse = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/CreateCourse', grpcWeb.MethodType.UNARY, qf_types_types_pb.Course, qf_types_types_pb.Course, function (request) {
            return request.serializeBinary();
        }, qf_types_types_pb.Course.deserializeBinary);
        this.methodInfoUpdateCourse = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/UpdateCourse', grpcWeb.MethodType.UNARY, qf_types_types_pb.Course, qf_types_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_types_requests_pb.Void.deserializeBinary);
        this.methodInfoUpdateCourseVisibility = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/UpdateCourseVisibility', grpcWeb.MethodType.UNARY, qf_types_types_pb.Enrollment, qf_types_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_types_requests_pb.Void.deserializeBinary);
        this.methodInfoGetAssignments = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetAssignments', grpcWeb.MethodType.UNARY, qf_types_requests_pb.CourseRequest, qf_types_types_pb.Assignments, function (request) {
            return request.serializeBinary();
        }, qf_types_types_pb.Assignments.deserializeBinary);
        this.methodInfoUpdateAssignments = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/UpdateAssignments', grpcWeb.MethodType.UNARY, qf_types_requests_pb.CourseRequest, qf_types_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_types_requests_pb.Void.deserializeBinary);
        this.methodInfoGetEnrollmentsByUser = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetEnrollmentsByUser', grpcWeb.MethodType.UNARY, qf_types_requests_pb.EnrollmentStatusRequest, qf_types_types_pb.Enrollments, function (request) {
            return request.serializeBinary();
        }, qf_types_types_pb.Enrollments.deserializeBinary);
        this.methodInfoGetEnrollmentsByCourse = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetEnrollmentsByCourse', grpcWeb.MethodType.UNARY, qf_types_requests_pb.EnrollmentRequest, qf_types_types_pb.Enrollments, function (request) {
            return request.serializeBinary();
        }, qf_types_types_pb.Enrollments.deserializeBinary);
        this.methodInfoCreateEnrollment = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/CreateEnrollment', grpcWeb.MethodType.UNARY, qf_types_types_pb.Enrollment, qf_types_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_types_requests_pb.Void.deserializeBinary);
        this.methodInfoUpdateEnrollments = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/UpdateEnrollments', grpcWeb.MethodType.UNARY, qf_types_types_pb.Enrollments, qf_types_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_types_requests_pb.Void.deserializeBinary);
        this.methodInfoGetSubmissions = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetSubmissions', grpcWeb.MethodType.UNARY, qf_types_requests_pb.SubmissionRequest, qf_types_types_pb.Submissions, function (request) {
            return request.serializeBinary();
        }, qf_types_types_pb.Submissions.deserializeBinary);
        this.methodInfoGetSubmissionsByCourse = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetSubmissionsByCourse', grpcWeb.MethodType.UNARY, qf_types_requests_pb.SubmissionsForCourseRequest, qf_types_requests_pb.CourseSubmissions, function (request) {
            return request.serializeBinary();
        }, qf_types_requests_pb.CourseSubmissions.deserializeBinary);
        this.methodInfoUpdateSubmission = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/UpdateSubmission', grpcWeb.MethodType.UNARY, qf_types_requests_pb.UpdateSubmissionRequest, qf_types_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_types_requests_pb.Void.deserializeBinary);
        this.methodInfoUpdateSubmissions = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/UpdateSubmissions', grpcWeb.MethodType.UNARY, qf_types_requests_pb.UpdateSubmissionsRequest, qf_types_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_types_requests_pb.Void.deserializeBinary);
        this.methodInfoRebuildSubmissions = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/RebuildSubmissions', grpcWeb.MethodType.UNARY, qf_types_requests_pb.RebuildRequest, qf_types_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_types_requests_pb.Void.deserializeBinary);
        this.methodInfoCreateBenchmark = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/CreateBenchmark', grpcWeb.MethodType.UNARY, qf_types_types_pb.GradingBenchmark, qf_types_types_pb.GradingBenchmark, function (request) {
            return request.serializeBinary();
        }, qf_types_types_pb.GradingBenchmark.deserializeBinary);
        this.methodInfoUpdateBenchmark = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/UpdateBenchmark', grpcWeb.MethodType.UNARY, qf_types_types_pb.GradingBenchmark, qf_types_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_types_requests_pb.Void.deserializeBinary);
        this.methodInfoDeleteBenchmark = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/DeleteBenchmark', grpcWeb.MethodType.UNARY, qf_types_types_pb.GradingBenchmark, qf_types_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_types_requests_pb.Void.deserializeBinary);
        this.methodInfoCreateCriterion = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/CreateCriterion', grpcWeb.MethodType.UNARY, qf_types_types_pb.GradingCriterion, qf_types_types_pb.GradingCriterion, function (request) {
            return request.serializeBinary();
        }, qf_types_types_pb.GradingCriterion.deserializeBinary);
        this.methodInfoUpdateCriterion = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/UpdateCriterion', grpcWeb.MethodType.UNARY, qf_types_types_pb.GradingCriterion, qf_types_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_types_requests_pb.Void.deserializeBinary);
        this.methodInfoDeleteCriterion = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/DeleteCriterion', grpcWeb.MethodType.UNARY, qf_types_types_pb.GradingCriterion, qf_types_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_types_requests_pb.Void.deserializeBinary);
        this.methodInfoCreateReview = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/CreateReview', grpcWeb.MethodType.UNARY, qf_types_requests_pb.ReviewRequest, qf_types_types_pb.Review, function (request) {
            return request.serializeBinary();
        }, qf_types_types_pb.Review.deserializeBinary);
        this.methodInfoUpdateReview = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/UpdateReview', grpcWeb.MethodType.UNARY, qf_types_requests_pb.ReviewRequest, qf_types_types_pb.Review, function (request) {
            return request.serializeBinary();
        }, qf_types_types_pb.Review.deserializeBinary);
        this.methodInfoGetReviewers = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetReviewers', grpcWeb.MethodType.UNARY, qf_types_requests_pb.SubmissionReviewersRequest, qf_types_requests_pb.Reviewers, function (request) {
            return request.serializeBinary();
        }, qf_types_requests_pb.Reviewers.deserializeBinary);
        this.methodInfoGetProviders = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetProviders', grpcWeb.MethodType.UNARY, qf_types_requests_pb.Void, qf_types_requests_pb.Providers, function (request) {
            return request.serializeBinary();
        }, qf_types_requests_pb.Providers.deserializeBinary);
        this.methodInfoGetOrganization = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetOrganization', grpcWeb.MethodType.UNARY, qf_types_requests_pb.OrgRequest, qf_types_requests_pb.Organization, function (request) {
            return request.serializeBinary();
        }, qf_types_requests_pb.Organization.deserializeBinary);
        this.methodInfoGetRepositories = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetRepositories', grpcWeb.MethodType.UNARY, qf_types_requests_pb.URLRequest, qf_types_requests_pb.Repositories, function (request) {
            return request.serializeBinary();
        }, qf_types_requests_pb.Repositories.deserializeBinary);
        this.methodInfoIsEmptyRepo = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/IsEmptyRepo', grpcWeb.MethodType.UNARY, qf_types_requests_pb.RepositoryRequest, qf_types_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_types_requests_pb.Void.deserializeBinary);
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
    QuickFeedServiceClient.prototype.getUser = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetUser', request, metadata || {}, this.methodInfoGetUser, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetUser', request, metadata || {}, this.methodInfoGetUser);
    };
    QuickFeedServiceClient.prototype.getUsers = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetUsers', request, metadata || {}, this.methodInfoGetUsers, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetUsers', request, metadata || {}, this.methodInfoGetUsers);
    };
    QuickFeedServiceClient.prototype.getUserByCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetUserByCourse', request, metadata || {}, this.methodInfoGetUserByCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetUserByCourse', request, metadata || {}, this.methodInfoGetUserByCourse);
    };
    QuickFeedServiceClient.prototype.updateUser = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/UpdateUser', request, metadata || {}, this.methodInfoUpdateUser, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/UpdateUser', request, metadata || {}, this.methodInfoUpdateUser);
    };
    QuickFeedServiceClient.prototype.isAuthorizedTeacher = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/IsAuthorizedTeacher', request, metadata || {}, this.methodInfoIsAuthorizedTeacher, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/IsAuthorizedTeacher', request, metadata || {}, this.methodInfoIsAuthorizedTeacher);
    };
    QuickFeedServiceClient.prototype.getGroup = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetGroup', request, metadata || {}, this.methodInfoGetGroup, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetGroup', request, metadata || {}, this.methodInfoGetGroup);
    };
    QuickFeedServiceClient.prototype.getGroupByUserAndCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetGroupByUserAndCourse', request, metadata || {}, this.methodInfoGetGroupByUserAndCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetGroupByUserAndCourse', request, metadata || {}, this.methodInfoGetGroupByUserAndCourse);
    };
    QuickFeedServiceClient.prototype.getGroupsByCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetGroupsByCourse', request, metadata || {}, this.methodInfoGetGroupsByCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetGroupsByCourse', request, metadata || {}, this.methodInfoGetGroupsByCourse);
    };
    QuickFeedServiceClient.prototype.createGroup = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/CreateGroup', request, metadata || {}, this.methodInfoCreateGroup, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/CreateGroup', request, metadata || {}, this.methodInfoCreateGroup);
    };
    QuickFeedServiceClient.prototype.updateGroup = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/UpdateGroup', request, metadata || {}, this.methodInfoUpdateGroup, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/UpdateGroup', request, metadata || {}, this.methodInfoUpdateGroup);
    };
    QuickFeedServiceClient.prototype.deleteGroup = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/DeleteGroup', request, metadata || {}, this.methodInfoDeleteGroup, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/DeleteGroup', request, metadata || {}, this.methodInfoDeleteGroup);
    };
    QuickFeedServiceClient.prototype.getCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetCourse', request, metadata || {}, this.methodInfoGetCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetCourse', request, metadata || {}, this.methodInfoGetCourse);
    };
    QuickFeedServiceClient.prototype.getCourses = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetCourses', request, metadata || {}, this.methodInfoGetCourses, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetCourses', request, metadata || {}, this.methodInfoGetCourses);
    };
    QuickFeedServiceClient.prototype.getCoursesByUser = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetCoursesByUser', request, metadata || {}, this.methodInfoGetCoursesByUser, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetCoursesByUser', request, metadata || {}, this.methodInfoGetCoursesByUser);
    };
    QuickFeedServiceClient.prototype.createCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/CreateCourse', request, metadata || {}, this.methodInfoCreateCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/CreateCourse', request, metadata || {}, this.methodInfoCreateCourse);
    };
    QuickFeedServiceClient.prototype.updateCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/UpdateCourse', request, metadata || {}, this.methodInfoUpdateCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/UpdateCourse', request, metadata || {}, this.methodInfoUpdateCourse);
    };
    QuickFeedServiceClient.prototype.updateCourseVisibility = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/UpdateCourseVisibility', request, metadata || {}, this.methodInfoUpdateCourseVisibility, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/UpdateCourseVisibility', request, metadata || {}, this.methodInfoUpdateCourseVisibility);
    };
    QuickFeedServiceClient.prototype.getAssignments = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetAssignments', request, metadata || {}, this.methodInfoGetAssignments, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetAssignments', request, metadata || {}, this.methodInfoGetAssignments);
    };
    QuickFeedServiceClient.prototype.updateAssignments = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/UpdateAssignments', request, metadata || {}, this.methodInfoUpdateAssignments, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/UpdateAssignments', request, metadata || {}, this.methodInfoUpdateAssignments);
    };
    QuickFeedServiceClient.prototype.getEnrollmentsByUser = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetEnrollmentsByUser', request, metadata || {}, this.methodInfoGetEnrollmentsByUser, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetEnrollmentsByUser', request, metadata || {}, this.methodInfoGetEnrollmentsByUser);
    };
    QuickFeedServiceClient.prototype.getEnrollmentsByCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetEnrollmentsByCourse', request, metadata || {}, this.methodInfoGetEnrollmentsByCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetEnrollmentsByCourse', request, metadata || {}, this.methodInfoGetEnrollmentsByCourse);
    };
    QuickFeedServiceClient.prototype.createEnrollment = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/CreateEnrollment', request, metadata || {}, this.methodInfoCreateEnrollment, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/CreateEnrollment', request, metadata || {}, this.methodInfoCreateEnrollment);
    };
    QuickFeedServiceClient.prototype.updateEnrollments = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/UpdateEnrollments', request, metadata || {}, this.methodInfoUpdateEnrollments, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/UpdateEnrollments', request, metadata || {}, this.methodInfoUpdateEnrollments);
    };
    QuickFeedServiceClient.prototype.getSubmissions = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetSubmissions', request, metadata || {}, this.methodInfoGetSubmissions, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetSubmissions', request, metadata || {}, this.methodInfoGetSubmissions);
    };
    QuickFeedServiceClient.prototype.getSubmissionsByCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetSubmissionsByCourse', request, metadata || {}, this.methodInfoGetSubmissionsByCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetSubmissionsByCourse', request, metadata || {}, this.methodInfoGetSubmissionsByCourse);
    };
    QuickFeedServiceClient.prototype.updateSubmission = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/UpdateSubmission', request, metadata || {}, this.methodInfoUpdateSubmission, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/UpdateSubmission', request, metadata || {}, this.methodInfoUpdateSubmission);
    };
    QuickFeedServiceClient.prototype.updateSubmissions = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/UpdateSubmissions', request, metadata || {}, this.methodInfoUpdateSubmissions, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/UpdateSubmissions', request, metadata || {}, this.methodInfoUpdateSubmissions);
    };
    QuickFeedServiceClient.prototype.rebuildSubmissions = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/RebuildSubmissions', request, metadata || {}, this.methodInfoRebuildSubmissions, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/RebuildSubmissions', request, metadata || {}, this.methodInfoRebuildSubmissions);
    };
    QuickFeedServiceClient.prototype.createBenchmark = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/CreateBenchmark', request, metadata || {}, this.methodInfoCreateBenchmark, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/CreateBenchmark', request, metadata || {}, this.methodInfoCreateBenchmark);
    };
    QuickFeedServiceClient.prototype.updateBenchmark = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/UpdateBenchmark', request, metadata || {}, this.methodInfoUpdateBenchmark, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/UpdateBenchmark', request, metadata || {}, this.methodInfoUpdateBenchmark);
    };
    QuickFeedServiceClient.prototype.deleteBenchmark = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/DeleteBenchmark', request, metadata || {}, this.methodInfoDeleteBenchmark, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/DeleteBenchmark', request, metadata || {}, this.methodInfoDeleteBenchmark);
    };
    QuickFeedServiceClient.prototype.createCriterion = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/CreateCriterion', request, metadata || {}, this.methodInfoCreateCriterion, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/CreateCriterion', request, metadata || {}, this.methodInfoCreateCriterion);
    };
    QuickFeedServiceClient.prototype.updateCriterion = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/UpdateCriterion', request, metadata || {}, this.methodInfoUpdateCriterion, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/UpdateCriterion', request, metadata || {}, this.methodInfoUpdateCriterion);
    };
    QuickFeedServiceClient.prototype.deleteCriterion = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/DeleteCriterion', request, metadata || {}, this.methodInfoDeleteCriterion, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/DeleteCriterion', request, metadata || {}, this.methodInfoDeleteCriterion);
    };
    QuickFeedServiceClient.prototype.createReview = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/CreateReview', request, metadata || {}, this.methodInfoCreateReview, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/CreateReview', request, metadata || {}, this.methodInfoCreateReview);
    };
    QuickFeedServiceClient.prototype.updateReview = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/UpdateReview', request, metadata || {}, this.methodInfoUpdateReview, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/UpdateReview', request, metadata || {}, this.methodInfoUpdateReview);
    };
    QuickFeedServiceClient.prototype.getReviewers = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetReviewers', request, metadata || {}, this.methodInfoGetReviewers, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetReviewers', request, metadata || {}, this.methodInfoGetReviewers);
    };
    QuickFeedServiceClient.prototype.getProviders = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetProviders', request, metadata || {}, this.methodInfoGetProviders, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetProviders', request, metadata || {}, this.methodInfoGetProviders);
    };
    QuickFeedServiceClient.prototype.getOrganization = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetOrganization', request, metadata || {}, this.methodInfoGetOrganization, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetOrganization', request, metadata || {}, this.methodInfoGetOrganization);
    };
    QuickFeedServiceClient.prototype.getRepositories = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetRepositories', request, metadata || {}, this.methodInfoGetRepositories, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetRepositories', request, metadata || {}, this.methodInfoGetRepositories);
    };
    QuickFeedServiceClient.prototype.isEmptyRepo = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/IsEmptyRepo', request, metadata || {}, this.methodInfoIsEmptyRepo, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/IsEmptyRepo', request, metadata || {}, this.methodInfoIsEmptyRepo);
    };
    return QuickFeedServiceClient;
}());
exports.QuickFeedServiceClient = QuickFeedServiceClient;
