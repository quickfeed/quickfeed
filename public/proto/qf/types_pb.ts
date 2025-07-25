// @generated by protoc-gen-es v2.6.0 with parameter "target=ts"
// @generated from file qf/types.proto (package qf, syntax proto3)
/* eslint-disable */

import type { GenEnum, GenFile, GenMessage } from "@bufbuild/protobuf/codegenv2";
import { enumDesc, fileDesc, messageDesc } from "@bufbuild/protobuf/codegenv2";
import type { Timestamp } from "@bufbuild/protobuf/wkt";
import { file_google_protobuf_timestamp } from "@bufbuild/protobuf/wkt";
import type { BuildInfo, Score } from "../kit/score/score_pb";
import { file_kit_score_score } from "../kit/score/score_pb";
import type { Message } from "@bufbuild/protobuf";

/**
 * Describes the file qf/types.proto.
 */
export const file_qf_types: GenFile = /*@__PURE__*/
  fileDesc("Cg5xZi90eXBlcy5wcm90bxICcWYi2gEKBFVzZXISCgoCSUQYASABKAQSDwoHSXNBZG1pbhgCIAEoCBIMCgROYW1lGAMgASgJEhEKCVN0dWRlbnRJRBgEIAEoCRINCgVFbWFpbBgFIAEoCRIRCglBdmF0YXJVUkwYBiABKAkSDQoFTG9naW4YByABKAkSEwoLVXBkYXRlVG9rZW4YCCABKAgSEwoLU2NtUmVtb3RlSUQYCSABKAQSFAoMUmVmcmVzaFRva2VuGAogASgJEiMKC0Vucm9sbG1lbnRzGAsgAygLMg4ucWYuRW5yb2xsbWVudCIgCgVVc2VycxIXCgV1c2VycxgBIAMoCzIILnFmLlVzZXIiqgIKBUdyb3VwEgoKAklEGAEgASgEEi0KBG5hbWUYAiABKAlCH8q1AxuiARhnb3JtOiJ1bmlxdWVJbmRleDpncm91cCISMQoIY291cnNlSUQYAyABKARCH8q1AxuiARhnb3JtOiJ1bmlxdWVJbmRleDpncm91cCISJQoGc3RhdHVzGAUgASgOMhUucWYuR3JvdXAuR3JvdXBTdGF0dXMSPQoFdXNlcnMYBiADKAsyCC5xZi5Vc2VyQiTKtQMgogEdZ29ybToibWFueTJtYW55Omdyb3VwX3VzZXJzOyISIwoLZW5yb2xsbWVudHMYByADKAsyDi5xZi5FbnJvbGxtZW50IigKC0dyb3VwU3RhdHVzEgsKB1BFTkRJTkcQABIMCghBUFBST1ZFRBABIiMKBkdyb3VwcxIZCgZncm91cHMYASADKAsyCS5xZi5Hcm91cCKvAwoGQ291cnNlEgoKAklEGAEgASgEEhcKD2NvdXJzZUNyZWF0b3JJRBgCIAEoBBIMCgRuYW1lGAMgASgJEi4KBGNvZGUYBCABKAlCIMq1AxyiARlnb3JtOiJ1bmlxdWVJbmRleDpjb3Vyc2UiEi4KBHllYXIYBSABKA1CIMq1AxyiARlnb3JtOiJ1bmlxdWVJbmRleDpjb3Vyc2UiEgsKA3RhZxgGIAEoCRIZChFTY21Pcmdhbml6YXRpb25JRBgIIAEoBBIbChNTY21Pcmdhbml6YXRpb25OYW1lGAkgASgJEhAKCHNsaXBEYXlzGAogASgNEhgKEERvY2tlcmZpbGVEaWdlc3QYCyABKAkSPAoIZW5yb2xsZWQYDCABKA4yGS5xZi5FbnJvbGxtZW50LlVzZXJTdGF0dXNCD8q1AwuiAQhnb3JtOiItIhIjCgtlbnJvbGxtZW50cxgNIAMoCzIOLnFmLkVucm9sbG1lbnQSIwoLYXNzaWdubWVudHMYDiADKAsyDi5xZi5Bc3NpZ25tZW50EhkKBmdyb3VwcxgPIAMoCzIJLnFmLkdyb3VwIiYKB0NvdXJzZXMSGwoHY291cnNlcxgBIAMoCzIKLnFmLkNvdXJzZSKlAwoKUmVwb3NpdG9yeRIKCgJJRBgBIAEoBBI/ChFTY21Pcmdhbml6YXRpb25JRBgCIAEoBEIkyrUDIKIBHWdvcm06InVuaXF1ZUluZGV4OnJlcG9zaXRvcnkiEhcKD1NjbVJlcG9zaXRvcnlJRBgDIAEoBBI0CgZ1c2VySUQYBCABKARCJMq1AyCiAR1nb3JtOiJ1bmlxdWVJbmRleDpyZXBvc2l0b3J5IhI1Cgdncm91cElEGAUgASgEQiTKtQMgogEdZ29ybToidW5pcXVlSW5kZXg6cmVwb3NpdG9yeSISDwoHSFRNTFVSTBgGIAEoCRJLCghyZXBvVHlwZRgHIAEoDjITLnFmLlJlcG9zaXRvcnkuVHlwZUIkyrUDIKIBHWdvcm06InVuaXF1ZUluZGV4OnJlcG9zaXRvcnkiEhkKBmlzc3VlcxgIIAMoCzIJLnFmLklzc3VlIksKBFR5cGUSCAoETk9ORRAAEggKBElORk8QARIPCgtBU1NJR05NRU5UUxACEgkKBVRFU1RTEAMSCAoEVVNFUhAEEgkKBUdST1VQEAUikAUKCkVucm9sbG1lbnQSCgoCSUQYASABKAQSNgoIY291cnNlSUQYAiABKARCJMq1AyCiAR1nb3JtOiJ1bmlxdWVJbmRleDplbnJvbGxtZW50IhI0CgZ1c2VySUQYAyABKARCJMq1AyCiAR1nb3JtOiJ1bmlxdWVJbmRleDplbnJvbGxtZW50IhIPCgdncm91cElEGAQgASgEEhYKBHVzZXIYBSABKAsyCC5xZi5Vc2VyEhoKBmNvdXJzZRgGIAEoCzIKLnFmLkNvdXJzZRIYCgVncm91cBgHIAEoCzIJLnFmLkdyb3VwEikKBnN0YXR1cxgIIAEoDjIZLnFmLkVucm9sbG1lbnQuVXNlclN0YXR1cxIqCgVzdGF0ZRgJIAEoDjIbLnFmLkVucm9sbG1lbnQuRGlzcGxheVN0YXRlEioKEXNsaXBEYXlzUmVtYWluaW5nGAogASgNQg/KtQMLogEIZ29ybToiLSISZgoQbGFzdEFjdGl2aXR5RGF0ZRgLIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBCMMq1AyyiASlnb3JtOiJzZXJpYWxpemVyOnRpbWVzdGFtcDt0eXBlOmRhdGV0aW1lIhIVCg10b3RhbEFwcHJvdmVkGAwgASgEEiYKDHVzZWRTbGlwRGF5cxgNIAMoCzIQLnFmLlVzZWRTbGlwRGF5cyI9CgpVc2VyU3RhdHVzEggKBE5PTkUQABILCgdQRU5ESU5HEAESCwoHU1RVREVOVBACEgsKB1RFQUNIRVIQAyJACgxEaXNwbGF5U3RhdGUSCQoFVU5TRVQQABIKCgZISURERU4QARILCgdWSVNJQkxFEAISDAoIRkFWT1JJVEUQAyJYCgxVc2VkU2xpcERheXMSCgoCSUQYASABKAQSFAoMZW5yb2xsbWVudElEGAIgASgEEhQKDGFzc2lnbm1lbnRJRBgDIAEoBBIQCgh1c2VkRGF5cxgEIAEoDSIyCgtFbnJvbGxtZW50cxIjCgtlbnJvbGxtZW50cxgBIAMoCzIOLnFmLkVucm9sbG1lbnQipQMKCkFzc2lnbm1lbnQSCgoCSUQYASABKAQSEAoIQ291cnNlSUQYAiABKAQSDAoEbmFtZRgDIAEoCRJeCghkZWFkbGluZRgEIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBCMMq1AyyiASlnb3JtOiJzZXJpYWxpemVyOnRpbWVzdGFtcDt0eXBlOmRhdGV0aW1lIhITCgthdXRvQXBwcm92ZRgFIAEoCBINCgVvcmRlchgGIAEoDRISCgppc0dyb3VwTGFiGAcgASgIEhIKCnNjb3JlTGltaXQYCCABKA0SEQoJcmV2aWV3ZXJzGAkgASgNEhgKEGNvbnRhaW5lclRpbWVvdXQYCiABKA0SIwoLc3VibWlzc2lvbnMYCyADKAsyDi5xZi5TdWJtaXNzaW9uEhcKBXRhc2tzGAwgAygLMggucWYuVGFzaxIvChFncmFkaW5nQmVuY2htYXJrcxgNIAMoCzIULnFmLkdyYWRpbmdCZW5jaG1hcmsSIwoNRXhwZWN0ZWRUZXN0cxgOIAMoCzIMLnFmLlRlc3RJbmZvIrkBCghUZXN0SW5mbxIKCgJJRBgBIAEoBBI4CgxBc3NpZ25tZW50SUQYAiABKARCIsq1Ax6iARtnb3JtOiJ1bmlxdWVJbmRleDp0ZXN0aW5mbyISNAoIVGVzdE5hbWUYAyABKAlCIsq1Ax6iARtnb3JtOiJ1bmlxdWVJbmRleDp0ZXN0aW5mbyISEAoITWF4U2NvcmUYBCABKAUSDgoGV2VpZ2h0GAUgASgFEg8KB0RldGFpbHMYBiABKAkihwEKBFRhc2sSCgoCSUQYASABKAQSFAoMYXNzaWdubWVudElEGAIgASgEEhcKD2Fzc2lnbm1lbnRPcmRlchgDIAEoDRINCgV0aXRsZRgEIAEoCRIMCgRib2R5GAUgASgJEgwKBG5hbWUYBiABKAkSGQoGaXNzdWVzGAcgAygLMgkucWYuSXNzdWUiUQoFSXNzdWUSCgoCSUQYASABKAQSFAoMcmVwb3NpdG9yeUlEGAIgASgEEg4KBnRhc2tJRBgDIAEoBBIWCg5TY21Jc3N1ZU51bWJlchgEIAEoBCL9AQoLUHVsbFJlcXVlc3QSCgoCSUQYASABKAQSFwoPU2NtUmVwb3NpdG9yeUlEGAIgASgEEg4KBnRhc2tJRBgDIAEoBBIPCgdpc3N1ZUlEGAQgASgEEg4KBnVzZXJJRBgFIAEoBBIUCgxTY21Db21tZW50SUQYBiABKAQSFAoMc291cmNlQnJhbmNoGAcgASgJEg4KBm51bWJlchgIIAEoBBIkCgVzdGFnZRgJIAEoDjIVLnFmLlB1bGxSZXF1ZXN0LlN0YWdlIjYKBVN0YWdlEggKBE5PTkUQABIJCgVEUkFGVBABEgoKBlJFVklFVxACEgwKCEFQUFJPVkVEEAMiMgoLQXNzaWdubWVudHMSIwoLYXNzaWdubWVudHMYASADKAsyDi5xZi5Bc3NpZ25tZW50IqEDCgpTdWJtaXNzaW9uEgoKAklEGAEgASgEEhQKDEFzc2lnbm1lbnRJRBgCIAEoBBIOCgZ1c2VySUQYAyABKAQSDwoHZ3JvdXBJRBgEIAEoBBINCgVzY29yZRgFIAEoDRISCgpjb21taXRIYXNoGAYgASgJEhAKCHJlbGVhc2VkGAcgASgIEhkKBkdyYWRlcxgIIAMoCzIJLnFmLkdyYWRlEmIKDGFwcHJvdmVkRGF0ZRgJIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBCMMq1AyyiASlnb3JtOiJzZXJpYWxpemVyOnRpbWVzdGFtcDt0eXBlOmRhdGV0aW1lIhIbCgdyZXZpZXdzGAogAygLMgoucWYuUmV2aWV3EiMKCUJ1aWxkSW5mbxgLIAEoCzIQLnNjb3JlLkJ1aWxkSW5mbxIcCgZTY29yZXMYDCADKAsyDC5zY29yZS5TY29yZSI8CgZTdGF0dXMSCAoETk9ORRAAEgwKCEFQUFJPVkVEEAESDAoIUkVKRUNURUQQAhIMCghSRVZJU0lPThADIjIKC1N1Ym1pc3Npb25zEiMKC3N1Ym1pc3Npb25zGAEgAygLMg4ucWYuU3VibWlzc2lvbiKWAQoFR3JhZGUSNQoMU3VibWlzc2lvbklEGAEgASgEQh/KtQMbogEYZ29ybToidW5pcXVlSW5kZXg6Z3JhZGUiEi8KBlVzZXJJRBgCIAEoBEIfyrUDG6IBGGdvcm06InVuaXF1ZUluZGV4OmdyYWRlIhIlCgZTdGF0dXMYAyABKA4yFS5xZi5TdWJtaXNzaW9uLlN0YXR1cyLIAQoQR3JhZGluZ0JlbmNobWFyaxIKCgJJRBgBIAEoBBIQCghDb3Vyc2VJRBgCIAEoBBIUCgxBc3NpZ25tZW50SUQYAyABKAQSEAoIUmV2aWV3SUQYBCABKAQSDwoHaGVhZGluZxgFIAEoCRIPCgdjb21tZW50GAYgASgJEkwKCGNyaXRlcmlhGAcgAygLMhQucWYuR3JhZGluZ0NyaXRlcmlvbkIkyrUDIKIBHWdvcm06ImZvcmVpZ25LZXk6QmVuY2htYXJrSUQiIjYKCkJlbmNobWFya3MSKAoKYmVuY2htYXJrcxgBIAMoCzIULnFmLkdyYWRpbmdCZW5jaG1hcmsi0QEKEEdyYWRpbmdDcml0ZXJpb24SCgoCSUQYASABKAQSEwoLQmVuY2htYXJrSUQYAiABKAQSEAoIQ291cnNlSUQYAyABKAQSDgoGcG9pbnRzGAQgASgEEhMKC2Rlc2NyaXB0aW9uGAUgASgJEikKBWdyYWRlGAYgASgOMhoucWYuR3JhZGluZ0NyaXRlcmlvbi5HcmFkZRIPCgdjb21tZW50GAcgASgJIikKBUdyYWRlEggKBE5PTkUQABIKCgZGQUlMRUQQARIKCgZQQVNTRUQQAiKgAgoGUmV2aWV3EgoKAklEGAEgASgEEhQKDFN1Ym1pc3Npb25JRBgCIAEoBBISCgpSZXZpZXdlcklEGAMgASgEEhAKCGZlZWRiYWNrGAQgASgJEg0KBXJlYWR5GAUgASgIEg0KBXNjb3JlGAYgASgNElIKEWdyYWRpbmdCZW5jaG1hcmtzGAcgAygLMhQucWYuR3JhZGluZ0JlbmNobWFya0IhyrUDHaIBGmdvcm06ImZvcmVpZ25LZXk6UmV2aWV3SUQiElwKBmVkaXRlZBgIIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBCMMq1AyyiASlnb3JtOiJzZXJpYWxpemVyOnRpbWVzdGFtcDt0eXBlOmRhdGV0aW1lIkImWiFnaXRodWIuY29tL3F1aWNrZmVlZC9xdWlja2ZlZWQvcWa6AgBiBnByb3RvMw", [file_google_protobuf_timestamp, file_kit_score_score]);

/**
 * @generated from message qf.User
 */
export type User = Message<"qf.User"> & {
  /**
   * @generated from field: uint64 ID = 1;
   */
  ID: bigint;

  /**
   * @generated from field: bool IsAdmin = 2;
   */
  IsAdmin: boolean;

  /**
   * @generated from field: string Name = 3;
   */
  Name: string;

  /**
   * @generated from field: string StudentID = 4;
   */
  StudentID: string;

  /**
   * @generated from field: string Email = 5;
   */
  Email: string;

  /**
   * @generated from field: string AvatarURL = 6;
   */
  AvatarURL: string;

  /**
   * @generated from field: string Login = 7;
   */
  Login: string;

  /**
   * Filter; True if user's JWT token needs to be updated.
   *
   * @generated from field: bool UpdateToken = 8;
   */
  UpdateToken: boolean;

  /**
   * Filter; The user's ID on the remote provider.
   *
   * @generated from field: uint64 ScmRemoteID = 9;
   */
  ScmRemoteID: bigint;

  /**
   * Filter; The user's refresh token that may be exchanged for an access token.
   *
   * @generated from field: string RefreshToken = 10;
   */
  RefreshToken: string;

  /**
   * @generated from field: repeated qf.Enrollment Enrollments = 11;
   */
  Enrollments: Enrollment[];
};

/**
 * Describes the message qf.User.
 * Use `create(UserSchema)` to create a new message.
 */
export const UserSchema: GenMessage<User> = /*@__PURE__*/
  messageDesc(file_qf_types, 0);

/**
 * @generated from message qf.Users
 */
export type Users = Message<"qf.Users"> & {
  /**
   * @generated from field: repeated qf.User users = 1;
   */
  users: User[];
};

/**
 * Describes the message qf.Users.
 * Use `create(UsersSchema)` to create a new message.
 */
export const UsersSchema: GenMessage<Users> = /*@__PURE__*/
  messageDesc(file_qf_types, 1);

/**
 * @generated from message qf.Group
 */
export type Group = Message<"qf.Group"> & {
  /**
   * @generated from field: uint64 ID = 1;
   */
  ID: bigint;

  /**
   * @generated from field: string name = 2;
   */
  name: string;

  /**
   * @generated from field: uint64 courseID = 3;
   */
  courseID: bigint;

  /**
   * @generated from field: qf.Group.GroupStatus status = 5;
   */
  status: Group_GroupStatus;

  /**
   * @generated from field: repeated qf.User users = 6;
   */
  users: User[];

  /**
   * @generated from field: repeated qf.Enrollment enrollments = 7;
   */
  enrollments: Enrollment[];
};

/**
 * Describes the message qf.Group.
 * Use `create(GroupSchema)` to create a new message.
 */
export const GroupSchema: GenMessage<Group> = /*@__PURE__*/
  messageDesc(file_qf_types, 2);

/**
 * @generated from enum qf.Group.GroupStatus
 */
export enum Group_GroupStatus {
  /**
   * @generated from enum value: PENDING = 0;
   */
  PENDING = 0,

  /**
   * @generated from enum value: APPROVED = 1;
   */
  APPROVED = 1,
}

/**
 * Describes the enum qf.Group.GroupStatus.
 */
export const Group_GroupStatusSchema: GenEnum<Group_GroupStatus> = /*@__PURE__*/
  enumDesc(file_qf_types, 2, 0);

/**
 * @generated from message qf.Groups
 */
export type Groups = Message<"qf.Groups"> & {
  /**
   * @generated from field: repeated qf.Group groups = 1;
   */
  groups: Group[];
};

/**
 * Describes the message qf.Groups.
 * Use `create(GroupsSchema)` to create a new message.
 */
export const GroupsSchema: GenMessage<Groups> = /*@__PURE__*/
  messageDesc(file_qf_types, 3);

/**
 * @generated from message qf.Course
 */
export type Course = Message<"qf.Course"> & {
  /**
   * @generated from field: uint64 ID = 1;
   */
  ID: bigint;

  /**
   * @generated from field: uint64 courseCreatorID = 2;
   */
  courseCreatorID: bigint;

  /**
   * @generated from field: string name = 3;
   */
  name: string;

  /**
   * @generated from field: string code = 4;
   */
  code: string;

  /**
   * @generated from field: uint32 year = 5;
   */
  year: number;

  /**
   * @generated from field: string tag = 6;
   */
  tag: string;

  /**
   * @generated from field: uint64 ScmOrganizationID = 8;
   */
  ScmOrganizationID: bigint;

  /**
   * The organization's SCM name, e.g., dat520-2020.
   *
   * @generated from field: string ScmOrganizationName = 9;
   */
  ScmOrganizationName: string;

  /**
   * @generated from field: uint32 slipDays = 10;
   */
  slipDays: number;

  /**
   * Digest of the dockerfile used to build the course's docker image.
   *
   * @generated from field: string DockerfileDigest = 11;
   */
  DockerfileDigest: string;

  /**
   * @generated from field: qf.Enrollment.UserStatus enrolled = 12;
   */
  enrolled: Enrollment_UserStatus;

  /**
   * @generated from field: repeated qf.Enrollment enrollments = 13;
   */
  enrollments: Enrollment[];

  /**
   * @generated from field: repeated qf.Assignment assignments = 14;
   */
  assignments: Assignment[];

  /**
   * @generated from field: repeated qf.Group groups = 15;
   */
  groups: Group[];
};

/**
 * Describes the message qf.Course.
 * Use `create(CourseSchema)` to create a new message.
 */
export const CourseSchema: GenMessage<Course> = /*@__PURE__*/
  messageDesc(file_qf_types, 4);

/**
 * @generated from message qf.Courses
 */
export type Courses = Message<"qf.Courses"> & {
  /**
   * @generated from field: repeated qf.Course courses = 1;
   */
  courses: Course[];
};

/**
 * Describes the message qf.Courses.
 * Use `create(CoursesSchema)` to create a new message.
 */
export const CoursesSchema: GenMessage<Courses> = /*@__PURE__*/
  messageDesc(file_qf_types, 5);

/**
 * @generated from message qf.Repository
 */
export type Repository = Message<"qf.Repository"> & {
  /**
   * @generated from field: uint64 ID = 1;
   */
  ID: bigint;

  /**
   * @generated from field: uint64 ScmOrganizationID = 2;
   */
  ScmOrganizationID: bigint;

  /**
   * ID of a github repository
   *
   * @generated from field: uint64 ScmRepositoryID = 3;
   */
  ScmRepositoryID: bigint;

  /**
   * @generated from field: uint64 userID = 4;
   */
  userID: bigint;

  /**
   * @generated from field: uint64 groupID = 5;
   */
  groupID: bigint;

  /**
   * @generated from field: string HTMLURL = 6;
   */
  HTMLURL: string;

  /**
   * @generated from field: qf.Repository.Type repoType = 7;
   */
  repoType: Repository_Type;

  /**
   * Issues associated with this repository
   *
   * @generated from field: repeated qf.Issue issues = 8;
   */
  issues: Issue[];
};

/**
 * Describes the message qf.Repository.
 * Use `create(RepositorySchema)` to create a new message.
 */
export const RepositorySchema: GenMessage<Repository> = /*@__PURE__*/
  messageDesc(file_qf_types, 6);

/**
 * @generated from enum qf.Repository.Type
 */
export enum Repository_Type {
  /**
   * @generated from enum value: NONE = 0;
   */
  NONE = 0,

  /**
   * @generated from enum value: INFO = 1;
   */
  INFO = 1,

  /**
   * @generated from enum value: ASSIGNMENTS = 2;
   */
  ASSIGNMENTS = 2,

  /**
   * @generated from enum value: TESTS = 3;
   */
  TESTS = 3,

  /**
   * @generated from enum value: USER = 4;
   */
  USER = 4,

  /**
   * @generated from enum value: GROUP = 5;
   */
  GROUP = 5,
}

/**
 * Describes the enum qf.Repository.Type.
 */
export const Repository_TypeSchema: GenEnum<Repository_Type> = /*@__PURE__*/
  enumDesc(file_qf_types, 6, 0);

/**
 * @generated from message qf.Enrollment
 */
export type Enrollment = Message<"qf.Enrollment"> & {
  /**
   * @generated from field: uint64 ID = 1;
   */
  ID: bigint;

  /**
   * @generated from field: uint64 courseID = 2;
   */
  courseID: bigint;

  /**
   * @generated from field: uint64 userID = 3;
   */
  userID: bigint;

  /**
   * @generated from field: uint64 groupID = 4;
   */
  groupID: bigint;

  /**
   * @generated from field: qf.User user = 5;
   */
  user?: User;

  /**
   * @generated from field: qf.Course course = 6;
   */
  course?: Course;

  /**
   * @generated from field: qf.Group group = 7;
   */
  group?: Group;

  /**
   * @generated from field: qf.Enrollment.UserStatus status = 8;
   */
  status: Enrollment_UserStatus;

  /**
   * @generated from field: qf.Enrollment.DisplayState state = 9;
   */
  state: Enrollment_DisplayState;

  /**
   * @generated from field: uint32 slipDaysRemaining = 10;
   */
  slipDaysRemaining: number;

  /**
   * @generated from field: google.protobuf.Timestamp lastActivityDate = 11;
   */
  lastActivityDate?: Timestamp;

  /**
   * @generated from field: uint64 totalApproved = 12;
   */
  totalApproved: bigint;

  /**
   * @generated from field: repeated qf.UsedSlipDays usedSlipDays = 13;
   */
  usedSlipDays: UsedSlipDays[];
};

/**
 * Describes the message qf.Enrollment.
 * Use `create(EnrollmentSchema)` to create a new message.
 */
export const EnrollmentSchema: GenMessage<Enrollment> = /*@__PURE__*/
  messageDesc(file_qf_types, 7);

/**
 * @generated from enum qf.Enrollment.UserStatus
 */
export enum Enrollment_UserStatus {
  /**
   * @generated from enum value: NONE = 0;
   */
  NONE = 0,

  /**
   * @generated from enum value: PENDING = 1;
   */
  PENDING = 1,

  /**
   * @generated from enum value: STUDENT = 2;
   */
  STUDENT = 2,

  /**
   * @generated from enum value: TEACHER = 3;
   */
  TEACHER = 3,
}

/**
 * Describes the enum qf.Enrollment.UserStatus.
 */
export const Enrollment_UserStatusSchema: GenEnum<Enrollment_UserStatus> = /*@__PURE__*/
  enumDesc(file_qf_types, 7, 0);

/**
 * @generated from enum qf.Enrollment.DisplayState
 */
export enum Enrollment_DisplayState {
  /**
   * @generated from enum value: UNSET = 0;
   */
  UNSET = 0,

  /**
   * @generated from enum value: HIDDEN = 1;
   */
  HIDDEN = 1,

  /**
   * @generated from enum value: VISIBLE = 2;
   */
  VISIBLE = 2,

  /**
   * @generated from enum value: FAVORITE = 3;
   */
  FAVORITE = 3,
}

/**
 * Describes the enum qf.Enrollment.DisplayState.
 */
export const Enrollment_DisplayStateSchema: GenEnum<Enrollment_DisplayState> = /*@__PURE__*/
  enumDesc(file_qf_types, 7, 1);

/**
 * @generated from message qf.UsedSlipDays
 */
export type UsedSlipDays = Message<"qf.UsedSlipDays"> & {
  /**
   * @generated from field: uint64 ID = 1;
   */
  ID: bigint;

  /**
   * @generated from field: uint64 enrollmentID = 2;
   */
  enrollmentID: bigint;

  /**
   * @generated from field: uint64 assignmentID = 3;
   */
  assignmentID: bigint;

  /**
   * @generated from field: uint32 usedDays = 4;
   */
  usedDays: number;
};

/**
 * Describes the message qf.UsedSlipDays.
 * Use `create(UsedSlipDaysSchema)` to create a new message.
 */
export const UsedSlipDaysSchema: GenMessage<UsedSlipDays> = /*@__PURE__*/
  messageDesc(file_qf_types, 8);

/**
 * @generated from message qf.Enrollments
 */
export type Enrollments = Message<"qf.Enrollments"> & {
  /**
   * @generated from field: repeated qf.Enrollment enrollments = 1;
   */
  enrollments: Enrollment[];
};

/**
 * Describes the message qf.Enrollments.
 * Use `create(EnrollmentsSchema)` to create a new message.
 */
export const EnrollmentsSchema: GenMessage<Enrollments> = /*@__PURE__*/
  messageDesc(file_qf_types, 9);

/**
 * @generated from message qf.Assignment
 */
export type Assignment = Message<"qf.Assignment"> & {
  /**
   * @generated from field: uint64 ID = 1;
   */
  ID: bigint;

  /**
   * foreign key
   *
   * @generated from field: uint64 CourseID = 2;
   */
  CourseID: bigint;

  /**
   * @generated from field: string name = 3;
   */
  name: string;

  /**
   * @generated from field: google.protobuf.Timestamp deadline = 4;
   */
  deadline?: Timestamp;

  /**
   * @generated from field: bool autoApprove = 5;
   */
  autoApprove: boolean;

  /**
   * @generated from field: uint32 order = 6;
   */
  order: number;

  /**
   * @generated from field: bool isGroupLab = 7;
   */
  isGroupLab: boolean;

  /**
   * minimal score limit for auto approval
   *
   * @generated from field: uint32 scoreLimit = 8;
   */
  scoreLimit: number;

  /**
   * number of reviewers that will review submissions for this assignment
   *
   * @generated from field: uint32 reviewers = 9;
   */
  reviewers: number;

  /**
   * container timeout for this assignment
   *
   * @generated from field: uint32 containerTimeout = 10;
   */
  containerTimeout: number;

  /**
   * submissions produced for this assignment
   *
   * @generated from field: repeated qf.Submission submissions = 11;
   */
  submissions: Submission[];

  /**
   * tasks associated with this assignment
   *
   * @generated from field: repeated qf.Task tasks = 12;
   */
  tasks: Task[];

  /**
   * grading benchmarks for this assignment
   *
   * @generated from field: repeated qf.GradingBenchmark gradingBenchmarks = 13;
   */
  gradingBenchmarks: GradingBenchmark[];

  /**
   * list of expected tests for this assignment
   *
   * @generated from field: repeated qf.TestInfo ExpectedTests = 14;
   */
  ExpectedTests: TestInfo[];
};

/**
 * Describes the message qf.Assignment.
 * Use `create(AssignmentSchema)` to create a new message.
 */
export const AssignmentSchema: GenMessage<Assignment> = /*@__PURE__*/
  messageDesc(file_qf_types, 10);

/**
 * @generated from message qf.TestInfo
 */
export type TestInfo = Message<"qf.TestInfo"> & {
  /**
   * @generated from field: uint64 ID = 1;
   */
  ID: bigint;

  /**
   * foreign key
   *
   * @generated from field: uint64 AssignmentID = 2;
   */
  AssignmentID: bigint;

  /**
   * name of the test
   *
   * @generated from field: string TestName = 3;
   */
  TestName: string;

  /**
   * max score possible to get on this test
   *
   * @generated from field: int32 MaxScore = 4;
   */
  MaxScore: number;

  /**
   * the weight of this test; used to compute final grade
   *
   * @generated from field: int32 Weight = 5;
   */
  Weight: number;

  /**
   * if populated, the frontend may display these details
   *
   * @generated from field: string Details = 6;
   */
  Details: string;
};

/**
 * Describes the message qf.TestInfo.
 * Use `create(TestInfoSchema)` to create a new message.
 */
export const TestInfoSchema: GenMessage<TestInfo> = /*@__PURE__*/
  messageDesc(file_qf_types, 11);

/**
 * @generated from message qf.Task
 */
export type Task = Message<"qf.Task"> & {
  /**
   * @generated from field: uint64 ID = 1;
   */
  ID: bigint;

  /**
   * foreign key
   *
   * @generated from field: uint64 assignmentID = 2;
   */
  assignmentID: bigint;

  /**
   * @generated from field: uint32 assignmentOrder = 3;
   */
  assignmentOrder: number;

  /**
   * @generated from field: string title = 4;
   */
  title: string;

  /**
   * @generated from field: string body = 5;
   */
  body: string;

  /**
   * @generated from field: string name = 6;
   */
  name: string;

  /**
   * Issues that use this task as a benchmark
   *
   * @generated from field: repeated qf.Issue issues = 7;
   */
  issues: Issue[];
};

/**
 * Describes the message qf.Task.
 * Use `create(TaskSchema)` to create a new message.
 */
export const TaskSchema: GenMessage<Task> = /*@__PURE__*/
  messageDesc(file_qf_types, 12);

/**
 * @generated from message qf.Issue
 */
export type Issue = Message<"qf.Issue"> & {
  /**
   * @generated from field: uint64 ID = 1;
   */
  ID: bigint;

  /**
   * Represents the internal ID of a repository
   *
   * @generated from field: uint64 repositoryID = 2;
   */
  repositoryID: bigint;

  /**
   * Task that this issue draws its content from
   *
   * @generated from field: uint64 taskID = 3;
   */
  taskID: bigint;

  /**
   * Issue number on scm. Needed for associating db issue with scm issue
   *
   * @generated from field: uint64 ScmIssueNumber = 4;
   */
  ScmIssueNumber: bigint;
};

/**
 * Describes the message qf.Issue.
 * Use `create(IssueSchema)` to create a new message.
 */
export const IssueSchema: GenMessage<Issue> = /*@__PURE__*/
  messageDesc(file_qf_types, 13);

/**
 * @generated from message qf.PullRequest
 */
export type PullRequest = Message<"qf.PullRequest"> & {
  /**
   * @generated from field: uint64 ID = 1;
   */
  ID: bigint;

  /**
   * Represents the scm repository ID
   *
   * @generated from field: uint64 ScmRepositoryID = 2;
   */
  ScmRepositoryID: bigint;

  /**
   * Foreign key
   *
   * @generated from field: uint64 taskID = 3;
   */
  taskID: bigint;

  /**
   * Foreign key
   *
   * @generated from field: uint64 issueID = 4;
   */
  issueID: bigint;

  /**
   * The user who owns this PR
   *
   * @generated from field: uint64 userID = 5;
   */
  userID: bigint;

  /**
   * Scm ID of the comment used for automatic feedback
   *
   * @generated from field: uint64 ScmCommentID = 6;
   */
  ScmCommentID: bigint;

  /**
   * The source branch for this pull request
   *
   * @generated from field: string sourceBranch = 7;
   */
  sourceBranch: string;

  /**
   * Pull request number
   *
   * @generated from field: uint64 number = 8;
   */
  number: bigint;

  /**
   * @generated from field: qf.PullRequest.Stage stage = 9;
   */
  stage: PullRequest_Stage;
};

/**
 * Describes the message qf.PullRequest.
 * Use `create(PullRequestSchema)` to create a new message.
 */
export const PullRequestSchema: GenMessage<PullRequest> = /*@__PURE__*/
  messageDesc(file_qf_types, 14);

/**
 * @generated from enum qf.PullRequest.Stage
 */
export enum PullRequest_Stage {
  /**
   * @generated from enum value: NONE = 0;
   */
  NONE = 0,

  /**
   * @generated from enum value: DRAFT = 1;
   */
  DRAFT = 1,

  /**
   * @generated from enum value: REVIEW = 2;
   */
  REVIEW = 2,

  /**
   * @generated from enum value: APPROVED = 3;
   */
  APPROVED = 3,
}

/**
 * Describes the enum qf.PullRequest.Stage.
 */
export const PullRequest_StageSchema: GenEnum<PullRequest_Stage> = /*@__PURE__*/
  enumDesc(file_qf_types, 14, 0);

/**
 * @generated from message qf.Assignments
 */
export type Assignments = Message<"qf.Assignments"> & {
  /**
   * @generated from field: repeated qf.Assignment assignments = 1;
   */
  assignments: Assignment[];
};

/**
 * Describes the message qf.Assignments.
 * Use `create(AssignmentsSchema)` to create a new message.
 */
export const AssignmentsSchema: GenMessage<Assignments> = /*@__PURE__*/
  messageDesc(file_qf_types, 15);

/**
 * @generated from message qf.Submission
 */
export type Submission = Message<"qf.Submission"> & {
  /**
   * @generated from field: uint64 ID = 1;
   */
  ID: bigint;

  /**
   * foreign key
   *
   * @generated from field: uint64 AssignmentID = 2;
   */
  AssignmentID: bigint;

  /**
   * @generated from field: uint64 userID = 3;
   */
  userID: bigint;

  /**
   * @generated from field: uint64 groupID = 4;
   */
  groupID: bigint;

  /**
   * @generated from field: uint32 score = 5;
   */
  score: number;

  /**
   * @generated from field: string commitHash = 6;
   */
  commitHash: string;

  /**
   * true => feedback is visible to the student or group members
   *
   * @generated from field: bool released = 7;
   */
  released: boolean;

  /**
   * @generated from field: repeated qf.Grade Grades = 8;
   */
  Grades: Grade[];

  /**
   * @generated from field: google.protobuf.Timestamp approvedDate = 9;
   */
  approvedDate?: Timestamp;

  /**
   * reviews produced for this submission
   *
   * @generated from field: repeated qf.Review reviews = 10;
   */
  reviews: Review[];

  /**
   * build info for tests
   *
   * @generated from field: score.BuildInfo BuildInfo = 11;
   */
  BuildInfo?: BuildInfo;

  /**
   * list of scores for different tests
   *
   * @generated from field: repeated score.Score Scores = 12;
   */
  Scores: Score[];
};

/**
 * Describes the message qf.Submission.
 * Use `create(SubmissionSchema)` to create a new message.
 */
export const SubmissionSchema: GenMessage<Submission> = /*@__PURE__*/
  messageDesc(file_qf_types, 16);

/**
 * @generated from enum qf.Submission.Status
 */
export enum Submission_Status {
  /**
   * @generated from enum value: NONE = 0;
   */
  NONE = 0,

  /**
   * @generated from enum value: APPROVED = 1;
   */
  APPROVED = 1,

  /**
   * @generated from enum value: REJECTED = 2;
   */
  REJECTED = 2,

  /**
   * @generated from enum value: REVISION = 3;
   */
  REVISION = 3,
}

/**
 * Describes the enum qf.Submission.Status.
 */
export const Submission_StatusSchema: GenEnum<Submission_Status> = /*@__PURE__*/
  enumDesc(file_qf_types, 16, 0);

/**
 * @generated from message qf.Submissions
 */
export type Submissions = Message<"qf.Submissions"> & {
  /**
   * @generated from field: repeated qf.Submission submissions = 1;
   */
  submissions: Submission[];
};

/**
 * Describes the message qf.Submissions.
 * Use `create(SubmissionsSchema)` to create a new message.
 */
export const SubmissionsSchema: GenMessage<Submissions> = /*@__PURE__*/
  messageDesc(file_qf_types, 17);

/**
 * @generated from message qf.Grade
 */
export type Grade = Message<"qf.Grade"> & {
  /**
   * @generated from field: uint64 SubmissionID = 1;
   */
  SubmissionID: bigint;

  /**
   * @generated from field: uint64 UserID = 2;
   */
  UserID: bigint;

  /**
   * @generated from field: qf.Submission.Status Status = 3;
   */
  Status: Submission_Status;
};

/**
 * Describes the message qf.Grade.
 * Use `create(GradeSchema)` to create a new message.
 */
export const GradeSchema: GenMessage<Grade> = /*@__PURE__*/
  messageDesc(file_qf_types, 18);

/**
 * @generated from message qf.GradingBenchmark
 */
export type GradingBenchmark = Message<"qf.GradingBenchmark"> & {
  /**
   * @generated from field: uint64 ID = 1;
   */
  ID: bigint;

  /**
   * foreign key
   *
   * @generated from field: uint64 CourseID = 2;
   */
  CourseID: bigint;

  /**
   * foreign key
   *
   * @generated from field: uint64 AssignmentID = 3;
   */
  AssignmentID: bigint;

  /**
   * foreign key
   *
   * @generated from field: uint64 ReviewID = 4;
   */
  ReviewID: bigint;

  /**
   * @generated from field: string heading = 5;
   */
  heading: string;

  /**
   * @generated from field: string comment = 6;
   */
  comment: string;

  /**
   * @generated from field: repeated qf.GradingCriterion criteria = 7;
   */
  criteria: GradingCriterion[];
};

/**
 * Describes the message qf.GradingBenchmark.
 * Use `create(GradingBenchmarkSchema)` to create a new message.
 */
export const GradingBenchmarkSchema: GenMessage<GradingBenchmark> = /*@__PURE__*/
  messageDesc(file_qf_types, 19);

/**
 * @generated from message qf.Benchmarks
 */
export type Benchmarks = Message<"qf.Benchmarks"> & {
  /**
   * @generated from field: repeated qf.GradingBenchmark benchmarks = 1;
   */
  benchmarks: GradingBenchmark[];
};

/**
 * Describes the message qf.Benchmarks.
 * Use `create(BenchmarksSchema)` to create a new message.
 */
export const BenchmarksSchema: GenMessage<Benchmarks> = /*@__PURE__*/
  messageDesc(file_qf_types, 20);

/**
 * @generated from message qf.GradingCriterion
 */
export type GradingCriterion = Message<"qf.GradingCriterion"> & {
  /**
   * @generated from field: uint64 ID = 1;
   */
  ID: bigint;

  /**
   * foreign key
   *
   * @generated from field: uint64 BenchmarkID = 2;
   */
  BenchmarkID: bigint;

  /**
   * foreign key
   *
   * @generated from field: uint64 CourseID = 3;
   */
  CourseID: bigint;

  /**
   * @generated from field: uint64 points = 4;
   */
  points: bigint;

  /**
   * @generated from field: string description = 5;
   */
  description: string;

  /**
   * @generated from field: qf.GradingCriterion.Grade grade = 6;
   */
  grade: GradingCriterion_Grade;

  /**
   * @generated from field: string comment = 7;
   */
  comment: string;
};

/**
 * Describes the message qf.GradingCriterion.
 * Use `create(GradingCriterionSchema)` to create a new message.
 */
export const GradingCriterionSchema: GenMessage<GradingCriterion> = /*@__PURE__*/
  messageDesc(file_qf_types, 21);

/**
 * @generated from enum qf.GradingCriterion.Grade
 */
export enum GradingCriterion_Grade {
  /**
   * @generated from enum value: NONE = 0;
   */
  NONE = 0,

  /**
   * @generated from enum value: FAILED = 1;
   */
  FAILED = 1,

  /**
   * @generated from enum value: PASSED = 2;
   */
  PASSED = 2,
}

/**
 * Describes the enum qf.GradingCriterion.Grade.
 */
export const GradingCriterion_GradeSchema: GenEnum<GradingCriterion_Grade> = /*@__PURE__*/
  enumDesc(file_qf_types, 21, 0);

/**
 * @generated from message qf.Review
 */
export type Review = Message<"qf.Review"> & {
  /**
   * @generated from field: uint64 ID = 1;
   */
  ID: bigint;

  /**
   * foreign key
   *
   * @generated from field: uint64 SubmissionID = 2;
   */
  SubmissionID: bigint;

  /**
   * UserID of the reviewer
   *
   * @generated from field: uint64 ReviewerID = 3;
   */
  ReviewerID: bigint;

  /**
   * @generated from field: string feedback = 4;
   */
  feedback: string;

  /**
   * @generated from field: bool ready = 5;
   */
  ready: boolean;

  /**
   * @generated from field: uint32 score = 6;
   */
  score: number;

  /**
   * @generated from field: repeated qf.GradingBenchmark gradingBenchmarks = 7;
   */
  gradingBenchmarks: GradingBenchmark[];

  /**
   * @generated from field: google.protobuf.Timestamp edited = 8;
   */
  edited?: Timestamp;
};

/**
 * Describes the message qf.Review.
 * Use `create(ReviewSchema)` to create a new message.
 */
export const ReviewSchema: GenMessage<Review> = /*@__PURE__*/
  messageDesc(file_qf_types, 22);

