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
var qf_types_pb = require("../qf/types_pb");
var qf_requests_pb = require("../qf/requests_pb");
var QuickFeedServiceClient = /** @class */ (function () {
    function QuickFeedServiceClient(hostname, credentials, options) {
        this.methodDescriptorGetUser = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetUser', grpcWeb.MethodType.UNARY, qf_requests_pb.Void, qf_types_pb.User, function (request) {
            return request.serializeBinary();
        }, qf_types_pb.User.deserializeBinary);
        this.methodDescriptorGetUsers = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetUsers', grpcWeb.MethodType.UNARY, qf_requests_pb.Void, qf_types_pb.Users, function (request) {
            return request.serializeBinary();
        }, qf_types_pb.Users.deserializeBinary);
        this.methodDescriptorGetUserByCourse = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetUserByCourse', grpcWeb.MethodType.UNARY, qf_requests_pb.CourseUserRequest, qf_types_pb.User, function (request) {
            return request.serializeBinary();
        }, qf_types_pb.User.deserializeBinary);
        this.methodDescriptorUpdateUser = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/UpdateUser', grpcWeb.MethodType.UNARY, qf_types_pb.User, qf_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_requests_pb.Void.deserializeBinary);
        this.methodDescriptorGetGroup = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetGroup', grpcWeb.MethodType.UNARY, qf_requests_pb.GetGroupRequest, qf_types_pb.Group, function (request) {
            return request.serializeBinary();
        }, qf_types_pb.Group.deserializeBinary);
        this.methodDescriptorGetGroupByUserAndCourse = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetGroupByUserAndCourse', grpcWeb.MethodType.UNARY, qf_requests_pb.GroupRequest, qf_types_pb.Group, function (request) {
            return request.serializeBinary();
        }, qf_types_pb.Group.deserializeBinary);
        this.methodDescriptorGetGroupsByCourse = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetGroupsByCourse', grpcWeb.MethodType.UNARY, qf_requests_pb.CourseRequest, qf_types_pb.Groups, function (request) {
            return request.serializeBinary();
        }, qf_types_pb.Groups.deserializeBinary);
        this.methodDescriptorCreateGroup = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/CreateGroup', grpcWeb.MethodType.UNARY, qf_types_pb.Group, qf_types_pb.Group, function (request) {
            return request.serializeBinary();
        }, qf_types_pb.Group.deserializeBinary);
        this.methodDescriptorUpdateGroup = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/UpdateGroup', grpcWeb.MethodType.UNARY, qf_types_pb.Group, qf_types_pb.Group, function (request) {
            return request.serializeBinary();
        }, qf_types_pb.Group.deserializeBinary);
        this.methodDescriptorDeleteGroup = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/DeleteGroup', grpcWeb.MethodType.UNARY, qf_requests_pb.GroupRequest, qf_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_requests_pb.Void.deserializeBinary);
        this.methodDescriptorGetCourse = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetCourse', grpcWeb.MethodType.UNARY, qf_requests_pb.CourseRequest, qf_types_pb.Course, function (request) {
            return request.serializeBinary();
        }, qf_types_pb.Course.deserializeBinary);
        this.methodDescriptorGetCourses = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetCourses', grpcWeb.MethodType.UNARY, qf_requests_pb.Void, qf_types_pb.Courses, function (request) {
            return request.serializeBinary();
        }, qf_types_pb.Courses.deserializeBinary);
        this.methodDescriptorGetCoursesByUser = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetCoursesByUser', grpcWeb.MethodType.UNARY, qf_requests_pb.EnrollmentStatusRequest, qf_types_pb.Courses, function (request) {
            return request.serializeBinary();
        }, qf_types_pb.Courses.deserializeBinary);
        this.methodDescriptorCreateCourse = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/CreateCourse', grpcWeb.MethodType.UNARY, qf_types_pb.Course, qf_types_pb.Course, function (request) {
            return request.serializeBinary();
        }, qf_types_pb.Course.deserializeBinary);
        this.methodDescriptorUpdateCourse = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/UpdateCourse', grpcWeb.MethodType.UNARY, qf_types_pb.Course, qf_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_requests_pb.Void.deserializeBinary);
        this.methodDescriptorUpdateCourseVisibility = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/UpdateCourseVisibility', grpcWeb.MethodType.UNARY, qf_types_pb.Enrollment, qf_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_requests_pb.Void.deserializeBinary);
        this.methodDescriptorGetAssignments = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetAssignments', grpcWeb.MethodType.UNARY, qf_requests_pb.CourseRequest, qf_types_pb.Assignments, function (request) {
            return request.serializeBinary();
        }, qf_types_pb.Assignments.deserializeBinary);
        this.methodDescriptorUpdateAssignments = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/UpdateAssignments', grpcWeb.MethodType.UNARY, qf_requests_pb.CourseRequest, qf_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_requests_pb.Void.deserializeBinary);
        this.methodDescriptorGetEnrollmentsByUser = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetEnrollmentsByUser', grpcWeb.MethodType.UNARY, qf_requests_pb.EnrollmentStatusRequest, qf_types_pb.Enrollments, function (request) {
            return request.serializeBinary();
        }, qf_types_pb.Enrollments.deserializeBinary);
        this.methodDescriptorGetEnrollmentsByCourse = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetEnrollmentsByCourse', grpcWeb.MethodType.UNARY, qf_requests_pb.EnrollmentRequest, qf_types_pb.Enrollments, function (request) {
            return request.serializeBinary();
        }, qf_types_pb.Enrollments.deserializeBinary);
        this.methodDescriptorCreateEnrollment = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/CreateEnrollment', grpcWeb.MethodType.UNARY, qf_types_pb.Enrollment, qf_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_requests_pb.Void.deserializeBinary);
        this.methodDescriptorUpdateEnrollments = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/UpdateEnrollments', grpcWeb.MethodType.UNARY, qf_types_pb.Enrollments, qf_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_requests_pb.Void.deserializeBinary);
        this.methodDescriptorGetSubmissions = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetSubmissions', grpcWeb.MethodType.UNARY, qf_requests_pb.SubmissionRequest, qf_types_pb.Submissions, function (request) {
            return request.serializeBinary();
        }, qf_types_pb.Submissions.deserializeBinary);
        this.methodDescriptorGetSubmissionsByCourse = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetSubmissionsByCourse', grpcWeb.MethodType.UNARY, qf_requests_pb.SubmissionsForCourseRequest, qf_requests_pb.CourseSubmissions, function (request) {
            return request.serializeBinary();
        }, qf_requests_pb.CourseSubmissions.deserializeBinary);
        this.methodDescriptorUpdateSubmission = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/UpdateSubmission', grpcWeb.MethodType.UNARY, qf_requests_pb.UpdateSubmissionRequest, qf_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_requests_pb.Void.deserializeBinary);
        this.methodDescriptorUpdateSubmissions = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/UpdateSubmissions', grpcWeb.MethodType.UNARY, qf_requests_pb.UpdateSubmissionsRequest, qf_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_requests_pb.Void.deserializeBinary);
        this.methodDescriptorRebuildSubmissions = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/RebuildSubmissions', grpcWeb.MethodType.UNARY, qf_requests_pb.RebuildRequest, qf_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_requests_pb.Void.deserializeBinary);
        this.methodDescriptorCreateBenchmark = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/CreateBenchmark', grpcWeb.MethodType.UNARY, qf_types_pb.GradingBenchmark, qf_types_pb.GradingBenchmark, function (request) {
            return request.serializeBinary();
        }, qf_types_pb.GradingBenchmark.deserializeBinary);
        this.methodDescriptorUpdateBenchmark = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/UpdateBenchmark', grpcWeb.MethodType.UNARY, qf_types_pb.GradingBenchmark, qf_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_requests_pb.Void.deserializeBinary);
        this.methodDescriptorDeleteBenchmark = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/DeleteBenchmark', grpcWeb.MethodType.UNARY, qf_types_pb.GradingBenchmark, qf_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_requests_pb.Void.deserializeBinary);
        this.methodDescriptorCreateCriterion = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/CreateCriterion', grpcWeb.MethodType.UNARY, qf_types_pb.GradingCriterion, qf_types_pb.GradingCriterion, function (request) {
            return request.serializeBinary();
        }, qf_types_pb.GradingCriterion.deserializeBinary);
        this.methodDescriptorUpdateCriterion = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/UpdateCriterion', grpcWeb.MethodType.UNARY, qf_types_pb.GradingCriterion, qf_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_requests_pb.Void.deserializeBinary);
        this.methodDescriptorDeleteCriterion = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/DeleteCriterion', grpcWeb.MethodType.UNARY, qf_types_pb.GradingCriterion, qf_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_requests_pb.Void.deserializeBinary);
        this.methodDescriptorCreateReview = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/CreateReview', grpcWeb.MethodType.UNARY, qf_requests_pb.ReviewRequest, qf_types_pb.Review, function (request) {
            return request.serializeBinary();
        }, qf_types_pb.Review.deserializeBinary);
        this.methodDescriptorUpdateReview = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/UpdateReview', grpcWeb.MethodType.UNARY, qf_requests_pb.ReviewRequest, qf_types_pb.Review, function (request) {
            return request.serializeBinary();
        }, qf_types_pb.Review.deserializeBinary);
        this.methodDescriptorGetReviewers = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetReviewers', grpcWeb.MethodType.UNARY, qf_requests_pb.SubmissionReviewersRequest, qf_requests_pb.Reviewers, function (request) {
            return request.serializeBinary();
        }, qf_requests_pb.Reviewers.deserializeBinary);
        this.methodDescriptorGetOrganization = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetOrganization', grpcWeb.MethodType.UNARY, qf_requests_pb.OrgRequest, qf_requests_pb.Organization, function (request) {
            return request.serializeBinary();
        }, qf_requests_pb.Organization.deserializeBinary);
        this.methodDescriptorGetRepositories = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/GetRepositories', grpcWeb.MethodType.UNARY, qf_requests_pb.URLRequest, qf_requests_pb.Repositories, function (request) {
            return request.serializeBinary();
        }, qf_requests_pb.Repositories.deserializeBinary);
        this.methodDescriptorIsEmptyRepo = new grpcWeb.MethodDescriptor('/qf.QuickFeedService/IsEmptyRepo', grpcWeb.MethodType.UNARY, qf_requests_pb.RepositoryRequest, qf_requests_pb.Void, function (request) {
            return request.serializeBinary();
        }, qf_requests_pb.Void.deserializeBinary);
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
                '/qf.QuickFeedService/GetUser', request, metadata || {}, this.methodDescriptorGetUser, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetUser', request, metadata || {}, this.methodDescriptorGetUser);
    };
    QuickFeedServiceClient.prototype.getUsers = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetUsers', request, metadata || {}, this.methodDescriptorGetUsers, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetUsers', request, metadata || {}, this.methodDescriptorGetUsers);
    };
    QuickFeedServiceClient.prototype.getUserByCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetUserByCourse', request, metadata || {}, this.methodDescriptorGetUserByCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetUserByCourse', request, metadata || {}, this.methodDescriptorGetUserByCourse);
    };
    QuickFeedServiceClient.prototype.updateUser = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/UpdateUser', request, metadata || {}, this.methodDescriptorUpdateUser, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/UpdateUser', request, metadata || {}, this.methodDescriptorUpdateUser);
    };
    QuickFeedServiceClient.prototype.getGroup = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetGroup', request, metadata || {}, this.methodDescriptorGetGroup, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetGroup', request, metadata || {}, this.methodDescriptorGetGroup);
    };
    QuickFeedServiceClient.prototype.getGroupByUserAndCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetGroupByUserAndCourse', request, metadata || {}, this.methodDescriptorGetGroupByUserAndCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetGroupByUserAndCourse', request, metadata || {}, this.methodDescriptorGetGroupByUserAndCourse);
    };
    QuickFeedServiceClient.prototype.getGroupsByCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetGroupsByCourse', request, metadata || {}, this.methodDescriptorGetGroupsByCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetGroupsByCourse', request, metadata || {}, this.methodDescriptorGetGroupsByCourse);
    };
    QuickFeedServiceClient.prototype.createGroup = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/CreateGroup', request, metadata || {}, this.methodDescriptorCreateGroup, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/CreateGroup', request, metadata || {}, this.methodDescriptorCreateGroup);
    };
    QuickFeedServiceClient.prototype.updateGroup = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/UpdateGroup', request, metadata || {}, this.methodDescriptorUpdateGroup, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/UpdateGroup', request, metadata || {}, this.methodDescriptorUpdateGroup);
    };
    QuickFeedServiceClient.prototype.deleteGroup = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/DeleteGroup', request, metadata || {}, this.methodDescriptorDeleteGroup, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/DeleteGroup', request, metadata || {}, this.methodDescriptorDeleteGroup);
    };
    QuickFeedServiceClient.prototype.getCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetCourse', request, metadata || {}, this.methodDescriptorGetCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetCourse', request, metadata || {}, this.methodDescriptorGetCourse);
    };
    QuickFeedServiceClient.prototype.getCourses = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetCourses', request, metadata || {}, this.methodDescriptorGetCourses, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetCourses', request, metadata || {}, this.methodDescriptorGetCourses);
    };
    QuickFeedServiceClient.prototype.getCoursesByUser = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetCoursesByUser', request, metadata || {}, this.methodDescriptorGetCoursesByUser, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetCoursesByUser', request, metadata || {}, this.methodDescriptorGetCoursesByUser);
    };
    QuickFeedServiceClient.prototype.createCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/CreateCourse', request, metadata || {}, this.methodDescriptorCreateCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/CreateCourse', request, metadata || {}, this.methodDescriptorCreateCourse);
    };
    QuickFeedServiceClient.prototype.updateCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/UpdateCourse', request, metadata || {}, this.methodDescriptorUpdateCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/UpdateCourse', request, metadata || {}, this.methodDescriptorUpdateCourse);
    };
    QuickFeedServiceClient.prototype.updateCourseVisibility = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/UpdateCourseVisibility', request, metadata || {}, this.methodDescriptorUpdateCourseVisibility, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/UpdateCourseVisibility', request, metadata || {}, this.methodDescriptorUpdateCourseVisibility);
    };
    QuickFeedServiceClient.prototype.getAssignments = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetAssignments', request, metadata || {}, this.methodDescriptorGetAssignments, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetAssignments', request, metadata || {}, this.methodDescriptorGetAssignments);
    };
    QuickFeedServiceClient.prototype.updateAssignments = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/UpdateAssignments', request, metadata || {}, this.methodDescriptorUpdateAssignments, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/UpdateAssignments', request, metadata || {}, this.methodDescriptorUpdateAssignments);
    };
    QuickFeedServiceClient.prototype.getEnrollmentsByUser = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetEnrollmentsByUser', request, metadata || {}, this.methodDescriptorGetEnrollmentsByUser, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetEnrollmentsByUser', request, metadata || {}, this.methodDescriptorGetEnrollmentsByUser);
    };
    QuickFeedServiceClient.prototype.getEnrollmentsByCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetEnrollmentsByCourse', request, metadata || {}, this.methodDescriptorGetEnrollmentsByCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetEnrollmentsByCourse', request, metadata || {}, this.methodDescriptorGetEnrollmentsByCourse);
    };
    QuickFeedServiceClient.prototype.createEnrollment = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/CreateEnrollment', request, metadata || {}, this.methodDescriptorCreateEnrollment, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/CreateEnrollment', request, metadata || {}, this.methodDescriptorCreateEnrollment);
    };
    QuickFeedServiceClient.prototype.updateEnrollments = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/UpdateEnrollments', request, metadata || {}, this.methodDescriptorUpdateEnrollments, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/UpdateEnrollments', request, metadata || {}, this.methodDescriptorUpdateEnrollments);
    };
    QuickFeedServiceClient.prototype.getSubmissions = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetSubmissions', request, metadata || {}, this.methodDescriptorGetSubmissions, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetSubmissions', request, metadata || {}, this.methodDescriptorGetSubmissions);
    };
    QuickFeedServiceClient.prototype.getSubmissionsByCourse = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetSubmissionsByCourse', request, metadata || {}, this.methodDescriptorGetSubmissionsByCourse, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetSubmissionsByCourse', request, metadata || {}, this.methodDescriptorGetSubmissionsByCourse);
    };
    QuickFeedServiceClient.prototype.updateSubmission = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/UpdateSubmission', request, metadata || {}, this.methodDescriptorUpdateSubmission, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/UpdateSubmission', request, metadata || {}, this.methodDescriptorUpdateSubmission);
    };
    QuickFeedServiceClient.prototype.updateSubmissions = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/UpdateSubmissions', request, metadata || {}, this.methodDescriptorUpdateSubmissions, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/UpdateSubmissions', request, metadata || {}, this.methodDescriptorUpdateSubmissions);
    };
    QuickFeedServiceClient.prototype.rebuildSubmissions = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/RebuildSubmissions', request, metadata || {}, this.methodDescriptorRebuildSubmissions, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/RebuildSubmissions', request, metadata || {}, this.methodDescriptorRebuildSubmissions);
    };
    QuickFeedServiceClient.prototype.createBenchmark = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/CreateBenchmark', request, metadata || {}, this.methodDescriptorCreateBenchmark, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/CreateBenchmark', request, metadata || {}, this.methodDescriptorCreateBenchmark);
    };
    QuickFeedServiceClient.prototype.updateBenchmark = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/UpdateBenchmark', request, metadata || {}, this.methodDescriptorUpdateBenchmark, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/UpdateBenchmark', request, metadata || {}, this.methodDescriptorUpdateBenchmark);
    };
    QuickFeedServiceClient.prototype.deleteBenchmark = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/DeleteBenchmark', request, metadata || {}, this.methodDescriptorDeleteBenchmark, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/DeleteBenchmark', request, metadata || {}, this.methodDescriptorDeleteBenchmark);
    };
    QuickFeedServiceClient.prototype.createCriterion = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/CreateCriterion', request, metadata || {}, this.methodDescriptorCreateCriterion, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/CreateCriterion', request, metadata || {}, this.methodDescriptorCreateCriterion);
    };
    QuickFeedServiceClient.prototype.updateCriterion = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/UpdateCriterion', request, metadata || {}, this.methodDescriptorUpdateCriterion, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/UpdateCriterion', request, metadata || {}, this.methodDescriptorUpdateCriterion);
    };
    QuickFeedServiceClient.prototype.deleteCriterion = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/DeleteCriterion', request, metadata || {}, this.methodDescriptorDeleteCriterion, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/DeleteCriterion', request, metadata || {}, this.methodDescriptorDeleteCriterion);
    };
    QuickFeedServiceClient.prototype.createReview = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/CreateReview', request, metadata || {}, this.methodDescriptorCreateReview, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/CreateReview', request, metadata || {}, this.methodDescriptorCreateReview);
    };
    QuickFeedServiceClient.prototype.updateReview = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/UpdateReview', request, metadata || {}, this.methodDescriptorUpdateReview, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/UpdateReview', request, metadata || {}, this.methodDescriptorUpdateReview);
    };
    QuickFeedServiceClient.prototype.getReviewers = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetReviewers', request, metadata || {}, this.methodDescriptorGetReviewers, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetReviewers', request, metadata || {}, this.methodDescriptorGetReviewers);
    };
    QuickFeedServiceClient.prototype.getOrganization = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetOrganization', request, metadata || {}, this.methodDescriptorGetOrganization, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetOrganization', request, metadata || {}, this.methodDescriptorGetOrganization);
    };
    QuickFeedServiceClient.prototype.getRepositories = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/GetRepositories', request, metadata || {}, this.methodDescriptorGetRepositories, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/GetRepositories', request, metadata || {}, this.methodDescriptorGetRepositories);
    };
    QuickFeedServiceClient.prototype.isEmptyRepo = function (request, metadata, callback) {
        if (callback !== undefined) {
            return this.client_.rpcCall(this.hostname_ +
                '/qf.QuickFeedService/IsEmptyRepo', request, metadata || {}, this.methodDescriptorIsEmptyRepo, callback);
        }
        return this.client_.unaryCall(this.hostname_ +
            '/qf.QuickFeedService/IsEmptyRepo', request, metadata || {}, this.methodDescriptorIsEmptyRepo);
    };
    return QuickFeedServiceClient;
}());
exports.QuickFeedServiceClient = QuickFeedServiceClient;
