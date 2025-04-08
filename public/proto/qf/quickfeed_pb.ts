// @generated by protoc-gen-es v2.2.5 with parameter "target=ts"
// @generated from file qf/quickfeed.proto (package qf, syntax proto3)
/* eslint-disable */

import type { GenFile, GenService } from "@bufbuild/protobuf/codegenv1"
import { fileDesc, serviceDesc } from "@bufbuild/protobuf/codegenv1"
import type { AssignmentsSchema, CourseSchema, CoursesSchema, EnrollmentSchema, EnrollmentsSchema, GradingBenchmarkSchema, GradingCriterionSchema, GroupSchema, GroupsSchema, ReviewSchema, SubmissionSchema, SubmissionsSchema, UserSchema, UsersSchema } from "./types_pb"
import { file_qf_types } from "./types_pb"
import type { CourseRequestSchema, CourseSubmissionsSchema, EnrollmentRequestSchema, GroupRequestSchema, RebuildRequestSchema, RepositoriesSchema, RepositoryRequestSchema, ReviewRequestSchema, SubmissionRequestSchema, UpdateSubmissionRequestSchema, UpdateSubmissionsRequestSchema, VoidSchema } from "./requests_pb"
import { file_qf_requests } from "./requests_pb"

/**
 * Describes the file qf/quickfeed.proto.
 */
export const file_qf_quickfeed: GenFile = /*@__PURE__*/
  fileDesc("ChJxZi9xdWlja2ZlZWQucHJvdG8SAnFmMtsNChBRdWlja0ZlZWRTZXJ2aWNlEh8KB0dldFVzZXISCC5xZi5Wb2lkGggucWYuVXNlciIAEiEKCEdldFVzZXJzEggucWYuVm9pZBoJLnFmLlVzZXJzIgASIgoKVXBkYXRlVXNlchIILnFmLlVzZXIaCC5xZi5Wb2lkIgASKQoIR2V0R3JvdXASEC5xZi5Hcm91cFJlcXVlc3QaCS5xZi5Hcm91cCIAEjQKEUdldEdyb3Vwc0J5Q291cnNlEhEucWYuQ291cnNlUmVxdWVzdBoKLnFmLkdyb3VwcyIAEiUKC0NyZWF0ZUdyb3VwEgkucWYuR3JvdXAaCS5xZi5Hcm91cCIAEiUKC1VwZGF0ZUdyb3VwEgkucWYuR3JvdXAaCS5xZi5Hcm91cCIAEisKC0RlbGV0ZUdyb3VwEhAucWYuR3JvdXBSZXF1ZXN0GggucWYuVm9pZCIAEiwKCUdldENvdXJzZRIRLnFmLkNvdXJzZVJlcXVlc3QaCi5xZi5Db3Vyc2UiABIlCgpHZXRDb3Vyc2VzEggucWYuVm9pZBoLLnFmLkNvdXJzZXMiABImCgxVcGRhdGVDb3Vyc2USCi5xZi5Db3Vyc2UaCC5xZi5Wb2lkIgASNAoWVXBkYXRlQ291cnNlVmlzaWJpbGl0eRIOLnFmLkVucm9sbG1lbnQaCC5xZi5Wb2lkIgASNgoOR2V0QXNzaWdubWVudHMSES5xZi5Db3Vyc2VSZXF1ZXN0Gg8ucWYuQXNzaWdubWVudHMiABIyChFVcGRhdGVBc3NpZ25tZW50cxIRLnFmLkNvdXJzZVJlcXVlc3QaCC5xZi5Wb2lkIgASOgoOR2V0RW5yb2xsbWVudHMSFS5xZi5FbnJvbGxtZW50UmVxdWVzdBoPLnFmLkVucm9sbG1lbnRzIgASLgoQQ3JlYXRlRW5yb2xsbWVudBIOLnFmLkVucm9sbG1lbnQaCC5xZi5Wb2lkIgASMAoRVXBkYXRlRW5yb2xsbWVudHMSDy5xZi5FbnJvbGxtZW50cxoILnFmLlZvaWQiABI4Cg1HZXRTdWJtaXNzaW9uEhUucWYuU3VibWlzc2lvblJlcXVlc3QaDi5xZi5TdWJtaXNzaW9uIgASOgoOR2V0U3VibWlzc2lvbnMSFS5xZi5TdWJtaXNzaW9uUmVxdWVzdBoPLnFmLlN1Ym1pc3Npb25zIgASSAoWR2V0U3VibWlzc2lvbnNCeUNvdXJzZRIVLnFmLlN1Ym1pc3Npb25SZXF1ZXN0GhUucWYuQ291cnNlU3VibWlzc2lvbnMiABI7ChBVcGRhdGVTdWJtaXNzaW9uEhsucWYuVXBkYXRlU3VibWlzc2lvblJlcXVlc3QaCC5xZi5Wb2lkIgASPQoRVXBkYXRlU3VibWlzc2lvbnMSHC5xZi5VcGRhdGVTdWJtaXNzaW9uc1JlcXVlc3QaCC5xZi5Wb2lkIgASNAoSUmVidWlsZFN1Ym1pc3Npb25zEhIucWYuUmVidWlsZFJlcXVlc3QaCC5xZi5Wb2lkIgASPwoPQ3JlYXRlQmVuY2htYXJrEhQucWYuR3JhZGluZ0JlbmNobWFyaxoULnFmLkdyYWRpbmdCZW5jaG1hcmsiABIzCg9VcGRhdGVCZW5jaG1hcmsSFC5xZi5HcmFkaW5nQmVuY2htYXJrGggucWYuVm9pZCIAEjMKD0RlbGV0ZUJlbmNobWFyaxIULnFmLkdyYWRpbmdCZW5jaG1hcmsaCC5xZi5Wb2lkIgASPwoPQ3JlYXRlQ3JpdGVyaW9uEhQucWYuR3JhZGluZ0NyaXRlcmlvbhoULnFmLkdyYWRpbmdDcml0ZXJpb24iABIzCg9VcGRhdGVDcml0ZXJpb24SFC5xZi5HcmFkaW5nQ3JpdGVyaW9uGggucWYuVm9pZCIAEjMKD0RlbGV0ZUNyaXRlcmlvbhIULnFmLkdyYWRpbmdDcml0ZXJpb24aCC5xZi5Wb2lkIgASLwoMQ3JlYXRlUmV2aWV3EhEucWYuUmV2aWV3UmVxdWVzdBoKLnFmLlJldmlldyIAEi8KDFVwZGF0ZVJldmlldxIRLnFmLlJldmlld1JlcXVlc3QaCi5xZi5SZXZpZXciABI4Cg9HZXRSZXBvc2l0b3JpZXMSES5xZi5Db3Vyc2VSZXF1ZXN0GhAucWYuUmVwb3NpdG9yaWVzIgASMAoLSXNFbXB0eVJlcG8SFS5xZi5SZXBvc2l0b3J5UmVxdWVzdBoILnFmLlZvaWQiABIwChBTdWJtaXNzaW9uU3RyZWFtEggucWYuVm9pZBoOLnFmLlN1Ym1pc3Npb24iADABQiZaIWdpdGh1Yi5jb20vcXVpY2tmZWVkL3F1aWNrZmVlZC9xZroCAGIGcHJvdG8z", [file_qf_types, file_qf_requests])

/**
 * users //
 *
 * @generated from service qf.QuickFeedService
 */
export const QuickFeedService: GenService<{
  /**
   * @generated from rpc qf.QuickFeedService.GetUser
   */
  getUser: {
    methodKind: "unary"
    input: typeof VoidSchema
    output: typeof UserSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.GetUsers
   */
  getUsers: {
    methodKind: "unary"
    input: typeof VoidSchema
    output: typeof UsersSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.UpdateUser
   */
  updateUser: {
    methodKind: "unary"
    input: typeof UserSchema
    output: typeof VoidSchema
  },
  /**
   * GetGroup returns a group with the given group ID or user ID. Course ID is required.
   *
   * @generated from rpc qf.QuickFeedService.GetGroup
   */
  getGroup: {
    methodKind: "unary"
    input: typeof GroupRequestSchema
    output: typeof GroupSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.GetGroupsByCourse
   */
  getGroupsByCourse: {
    methodKind: "unary"
    input: typeof CourseRequestSchema
    output: typeof GroupsSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.CreateGroup
   */
  createGroup: {
    methodKind: "unary"
    input: typeof GroupSchema
    output: typeof GroupSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.UpdateGroup
   */
  updateGroup: {
    methodKind: "unary"
    input: typeof GroupSchema
    output: typeof GroupSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.DeleteGroup
   */
  deleteGroup: {
    methodKind: "unary"
    input: typeof GroupRequestSchema
    output: typeof VoidSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.GetCourse
   */
  getCourse: {
    methodKind: "unary"
    input: typeof CourseRequestSchema
    output: typeof CourseSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.GetCourses
   */
  getCourses: {
    methodKind: "unary"
    input: typeof VoidSchema
    output: typeof CoursesSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.UpdateCourse
   */
  updateCourse: {
    methodKind: "unary"
    input: typeof CourseSchema
    output: typeof VoidSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.UpdateCourseVisibility
   */
  updateCourseVisibility: {
    methodKind: "unary"
    input: typeof EnrollmentSchema
    output: typeof VoidSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.GetAssignments
   */
  getAssignments: {
    methodKind: "unary"
    input: typeof CourseRequestSchema
    output: typeof AssignmentsSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.UpdateAssignments
   */
  updateAssignments: {
    methodKind: "unary"
    input: typeof CourseRequestSchema
    output: typeof VoidSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.GetEnrollments
   */
  getEnrollments: {
    methodKind: "unary"
    input: typeof EnrollmentRequestSchema
    output: typeof EnrollmentsSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.CreateEnrollment
   */
  createEnrollment: {
    methodKind: "unary"
    input: typeof EnrollmentSchema
    output: typeof VoidSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.UpdateEnrollments
   */
  updateEnrollments: {
    methodKind: "unary"
    input: typeof EnrollmentsSchema
    output: typeof VoidSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.GetSubmission
   */
  getSubmission: {
    methodKind: "unary"
    input: typeof SubmissionRequestSchema
    output: typeof SubmissionSchema
  },
  /**
   * Get latest submissions for all course assignments for a user or a group.
   *
   * @generated from rpc qf.QuickFeedService.GetSubmissions
   */
  getSubmissions: {
    methodKind: "unary"
    input: typeof SubmissionRequestSchema
    output: typeof SubmissionsSchema
  },
  /**
   * Get lab submissions for every course user or every course group
   *
   * @generated from rpc qf.QuickFeedService.GetSubmissionsByCourse
   */
  getSubmissionsByCourse: {
    methodKind: "unary"
    input: typeof SubmissionRequestSchema
    output: typeof CourseSubmissionsSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.UpdateSubmission
   */
  updateSubmission: {
    methodKind: "unary"
    input: typeof UpdateSubmissionRequestSchema
    output: typeof VoidSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.UpdateSubmissions
   */
  updateSubmissions: {
    methodKind: "unary"
    input: typeof UpdateSubmissionRequestSchema
    output: typeof VoidSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.RebuildSubmissions
   */
  rebuildSubmissions: {
    methodKind: "unary"
    input: typeof RebuildRequestSchema
    output: typeof VoidSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.CreateBenchmark
   */
  createBenchmark: {
    methodKind: "unary"
    input: typeof GradingBenchmarkSchema
    output: typeof GradingBenchmarkSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.UpdateBenchmark
   */
  updateBenchmark: {
    methodKind: "unary"
    input: typeof GradingBenchmarkSchema
    output: typeof VoidSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.DeleteBenchmark
   */
  deleteBenchmark: {
    methodKind: "unary"
    input: typeof GradingBenchmarkSchema
    output: typeof VoidSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.CreateCriterion
   */
  createCriterion: {
    methodKind: "unary"
    input: typeof GradingCriterionSchema
    output: typeof GradingCriterionSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.UpdateCriterion
   */
  updateCriterion: {
    methodKind: "unary"
    input: typeof GradingCriterionSchema
    output: typeof VoidSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.DeleteCriterion
   */
  deleteCriterion: {
    methodKind: "unary"
    input: typeof GradingCriterionSchema
    output: typeof VoidSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.CreateReview
   */
  createReview: {
    methodKind: "unary"
    input: typeof ReviewRequestSchema
    output: typeof ReviewSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.UpdateReview
   */
  updateReview: {
    methodKind: "unary"
    input: typeof ReviewRequestSchema
    output: typeof ReviewSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.GetRepositories
   */
  getRepositories: {
    methodKind: "unary"
    input: typeof CourseRequestSchema
    output: typeof RepositoriesSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.IsEmptyRepo
   */
  isEmptyRepo: {
    methodKind: "unary"
    input: typeof RepositoryRequestSchema
    output: typeof VoidSchema
  },
  /**
   * @generated from rpc qf.QuickFeedService.SubmissionStream
   */
  submissionStream: {
    methodKind: "server_streaming"
    input: typeof VoidSchema
    output: typeof SubmissionSchema
  },
}> = /*@__PURE__*/
  serviceDesc(file_qf_quickfeed, 0);
