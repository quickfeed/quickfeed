import * as jspb from 'google-protobuf'

import * as qf_types_types_pb from '../../qf/types/types_pb';


export class SubmissionLink extends jspb.Message {
  getAssignment(): qf_types_types_pb.Assignment | undefined;
  setAssignment(value?: qf_types_types_pb.Assignment): SubmissionLink;
  hasAssignment(): boolean;
  clearAssignment(): SubmissionLink;

  getSubmission(): qf_types_types_pb.Submission | undefined;
  setSubmission(value?: qf_types_types_pb.Submission): SubmissionLink;
  hasSubmission(): boolean;
  clearSubmission(): SubmissionLink;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SubmissionLink.AsObject;
  static toObject(includeInstance: boolean, msg: SubmissionLink): SubmissionLink.AsObject;
  static serializeBinaryToWriter(message: SubmissionLink, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SubmissionLink;
  static deserializeBinaryFromReader(message: SubmissionLink, reader: jspb.BinaryReader): SubmissionLink;
}

export namespace SubmissionLink {
  export type AsObject = {
    assignment?: qf_types_types_pb.Assignment.AsObject,
    submission?: qf_types_types_pb.Submission.AsObject,
  }
}

export class EnrollmentLink extends jspb.Message {
  getEnrollment(): qf_types_types_pb.Enrollment | undefined;
  setEnrollment(value?: qf_types_types_pb.Enrollment): EnrollmentLink;
  hasEnrollment(): boolean;
  clearEnrollment(): EnrollmentLink;

  getSubmissionsList(): Array<SubmissionLink>;
  setSubmissionsList(value: Array<SubmissionLink>): EnrollmentLink;
  clearSubmissionsList(): EnrollmentLink;
  addSubmissions(value?: SubmissionLink, index?: number): SubmissionLink;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): EnrollmentLink.AsObject;
  static toObject(includeInstance: boolean, msg: EnrollmentLink): EnrollmentLink.AsObject;
  static serializeBinaryToWriter(message: EnrollmentLink, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): EnrollmentLink;
  static deserializeBinaryFromReader(message: EnrollmentLink, reader: jspb.BinaryReader): EnrollmentLink;
}

export namespace EnrollmentLink {
  export type AsObject = {
    enrollment?: qf_types_types_pb.Enrollment.AsObject,
    submissionsList: Array<SubmissionLink.AsObject>,
  }
}

export class CourseSubmissions extends jspb.Message {
  getCourse(): qf_types_types_pb.Course | undefined;
  setCourse(value?: qf_types_types_pb.Course): CourseSubmissions;
  hasCourse(): boolean;
  clearCourse(): CourseSubmissions;

  getLinksList(): Array<EnrollmentLink>;
  setLinksList(value: Array<EnrollmentLink>): CourseSubmissions;
  clearLinksList(): CourseSubmissions;
  addLinks(value?: EnrollmentLink, index?: number): EnrollmentLink;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CourseSubmissions.AsObject;
  static toObject(includeInstance: boolean, msg: CourseSubmissions): CourseSubmissions.AsObject;
  static serializeBinaryToWriter(message: CourseSubmissions, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CourseSubmissions;
  static deserializeBinaryFromReader(message: CourseSubmissions, reader: jspb.BinaryReader): CourseSubmissions;
}

export namespace CourseSubmissions {
  export type AsObject = {
    course?: qf_types_types_pb.Course.AsObject,
    linksList: Array<EnrollmentLink.AsObject>,
  }
}

export class ReviewRequest extends jspb.Message {
  getCourseid(): number;
  setCourseid(value: number): ReviewRequest;

  getReview(): qf_types_types_pb.Review | undefined;
  setReview(value?: qf_types_types_pb.Review): ReviewRequest;
  hasReview(): boolean;
  clearReview(): ReviewRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ReviewRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ReviewRequest): ReviewRequest.AsObject;
  static serializeBinaryToWriter(message: ReviewRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ReviewRequest;
  static deserializeBinaryFromReader(message: ReviewRequest, reader: jspb.BinaryReader): ReviewRequest;
}

export namespace ReviewRequest {
  export type AsObject = {
    courseid: number,
    review?: qf_types_types_pb.Review.AsObject,
  }
}

export class CourseRequest extends jspb.Message {
  getCourseid(): number;
  setCourseid(value: number): CourseRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CourseRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CourseRequest): CourseRequest.AsObject;
  static serializeBinaryToWriter(message: CourseRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CourseRequest;
  static deserializeBinaryFromReader(message: CourseRequest, reader: jspb.BinaryReader): CourseRequest;
}

export namespace CourseRequest {
  export type AsObject = {
    courseid: number,
  }
}

export class UserRequest extends jspb.Message {
  getUserid(): number;
  setUserid(value: number): UserRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UserRequest.AsObject;
  static toObject(includeInstance: boolean, msg: UserRequest): UserRequest.AsObject;
  static serializeBinaryToWriter(message: UserRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UserRequest;
  static deserializeBinaryFromReader(message: UserRequest, reader: jspb.BinaryReader): UserRequest;
}

export namespace UserRequest {
  export type AsObject = {
    userid: number,
  }
}

export class GetGroupRequest extends jspb.Message {
  getGroupid(): number;
  setGroupid(value: number): GetGroupRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetGroupRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetGroupRequest): GetGroupRequest.AsObject;
  static serializeBinaryToWriter(message: GetGroupRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetGroupRequest;
  static deserializeBinaryFromReader(message: GetGroupRequest, reader: jspb.BinaryReader): GetGroupRequest;
}

export namespace GetGroupRequest {
  export type AsObject = {
    groupid: number,
  }
}

export class GroupRequest extends jspb.Message {
  getUserid(): number;
  setUserid(value: number): GroupRequest;

  getGroupid(): number;
  setGroupid(value: number): GroupRequest;

  getCourseid(): number;
  setCourseid(value: number): GroupRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GroupRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GroupRequest): GroupRequest.AsObject;
  static serializeBinaryToWriter(message: GroupRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GroupRequest;
  static deserializeBinaryFromReader(message: GroupRequest, reader: jspb.BinaryReader): GroupRequest;
}

export namespace GroupRequest {
  export type AsObject = {
    userid: number,
    groupid: number,
    courseid: number,
  }
}

export class Provider extends jspb.Message {
  getProvider(): string;
  setProvider(value: string): Provider;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Provider.AsObject;
  static toObject(includeInstance: boolean, msg: Provider): Provider.AsObject;
  static serializeBinaryToWriter(message: Provider, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Provider;
  static deserializeBinaryFromReader(message: Provider, reader: jspb.BinaryReader): Provider;
}

export namespace Provider {
  export type AsObject = {
    provider: string,
  }
}

export class OrgRequest extends jspb.Message {
  getOrgname(): string;
  setOrgname(value: string): OrgRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): OrgRequest.AsObject;
  static toObject(includeInstance: boolean, msg: OrgRequest): OrgRequest.AsObject;
  static serializeBinaryToWriter(message: OrgRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): OrgRequest;
  static deserializeBinaryFromReader(message: OrgRequest, reader: jspb.BinaryReader): OrgRequest;
}

export namespace OrgRequest {
  export type AsObject = {
    orgname: string,
  }
}

export class Organization extends jspb.Message {
  getId(): number;
  setId(value: number): Organization;

  getPath(): string;
  setPath(value: string): Organization;

  getAvatar(): string;
  setAvatar(value: string): Organization;

  getPaymentplan(): string;
  setPaymentplan(value: string): Organization;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Organization.AsObject;
  static toObject(includeInstance: boolean, msg: Organization): Organization.AsObject;
  static serializeBinaryToWriter(message: Organization, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Organization;
  static deserializeBinaryFromReader(message: Organization, reader: jspb.BinaryReader): Organization;
}

export namespace Organization {
  export type AsObject = {
    id: number,
    path: string,
    avatar: string,
    paymentplan: string,
  }
}

export class Organizations extends jspb.Message {
  getOrganizationsList(): Array<Organization>;
  setOrganizationsList(value: Array<Organization>): Organizations;
  clearOrganizationsList(): Organizations;
  addOrganizations(value?: Organization, index?: number): Organization;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Organizations.AsObject;
  static toObject(includeInstance: boolean, msg: Organizations): Organizations.AsObject;
  static serializeBinaryToWriter(message: Organizations, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Organizations;
  static deserializeBinaryFromReader(message: Organizations, reader: jspb.BinaryReader): Organizations;
}

export namespace Organizations {
  export type AsObject = {
    organizationsList: Array<Organization.AsObject>,
  }
}

export class Reviewers extends jspb.Message {
  getReviewersList(): Array<qf_types_types_pb.User>;
  setReviewersList(value: Array<qf_types_types_pb.User>): Reviewers;
  clearReviewersList(): Reviewers;
  addReviewers(value?: qf_types_types_pb.User, index?: number): qf_types_types_pb.User;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Reviewers.AsObject;
  static toObject(includeInstance: boolean, msg: Reviewers): Reviewers.AsObject;
  static serializeBinaryToWriter(message: Reviewers, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Reviewers;
  static deserializeBinaryFromReader(message: Reviewers, reader: jspb.BinaryReader): Reviewers;
}

export namespace Reviewers {
  export type AsObject = {
    reviewersList: Array<qf_types_types_pb.User.AsObject>,
  }
}

export class EnrollmentRequest extends jspb.Message {
  getCourseid(): number;
  setCourseid(value: number): EnrollmentRequest;

  getIgnoregroupmembers(): boolean;
  setIgnoregroupmembers(value: boolean): EnrollmentRequest;

  getWithactivity(): boolean;
  setWithactivity(value: boolean): EnrollmentRequest;

  getStatusesList(): Array<qf_types_types_pb.Enrollment.UserStatus>;
  setStatusesList(value: Array<qf_types_types_pb.Enrollment.UserStatus>): EnrollmentRequest;
  clearStatusesList(): EnrollmentRequest;
  addStatuses(value: qf_types_types_pb.Enrollment.UserStatus, index?: number): EnrollmentRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): EnrollmentRequest.AsObject;
  static toObject(includeInstance: boolean, msg: EnrollmentRequest): EnrollmentRequest.AsObject;
  static serializeBinaryToWriter(message: EnrollmentRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): EnrollmentRequest;
  static deserializeBinaryFromReader(message: EnrollmentRequest, reader: jspb.BinaryReader): EnrollmentRequest;
}

export namespace EnrollmentRequest {
  export type AsObject = {
    courseid: number,
    ignoregroupmembers: boolean,
    withactivity: boolean,
    statusesList: Array<qf_types_types_pb.Enrollment.UserStatus>,
  }
}

export class EnrollmentStatusRequest extends jspb.Message {
  getUserid(): number;
  setUserid(value: number): EnrollmentStatusRequest;

  getStatusesList(): Array<qf_types_types_pb.Enrollment.UserStatus>;
  setStatusesList(value: Array<qf_types_types_pb.Enrollment.UserStatus>): EnrollmentStatusRequest;
  clearStatusesList(): EnrollmentStatusRequest;
  addStatuses(value: qf_types_types_pb.Enrollment.UserStatus, index?: number): EnrollmentStatusRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): EnrollmentStatusRequest.AsObject;
  static toObject(includeInstance: boolean, msg: EnrollmentStatusRequest): EnrollmentStatusRequest.AsObject;
  static serializeBinaryToWriter(message: EnrollmentStatusRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): EnrollmentStatusRequest;
  static deserializeBinaryFromReader(message: EnrollmentStatusRequest, reader: jspb.BinaryReader): EnrollmentStatusRequest;
}

export namespace EnrollmentStatusRequest {
  export type AsObject = {
    userid: number,
    statusesList: Array<qf_types_types_pb.Enrollment.UserStatus>,
  }
}

export class SubmissionRequest extends jspb.Message {
  getUserid(): number;
  setUserid(value: number): SubmissionRequest;

  getGroupid(): number;
  setGroupid(value: number): SubmissionRequest;

  getCourseid(): number;
  setCourseid(value: number): SubmissionRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SubmissionRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SubmissionRequest): SubmissionRequest.AsObject;
  static serializeBinaryToWriter(message: SubmissionRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SubmissionRequest;
  static deserializeBinaryFromReader(message: SubmissionRequest, reader: jspb.BinaryReader): SubmissionRequest;
}

export namespace SubmissionRequest {
  export type AsObject = {
    userid: number,
    groupid: number,
    courseid: number,
  }
}

export class UpdateSubmissionRequest extends jspb.Message {
  getSubmissionid(): number;
  setSubmissionid(value: number): UpdateSubmissionRequest;

  getCourseid(): number;
  setCourseid(value: number): UpdateSubmissionRequest;

  getScore(): number;
  setScore(value: number): UpdateSubmissionRequest;

  getReleased(): boolean;
  setReleased(value: boolean): UpdateSubmissionRequest;

  getStatus(): qf_types_types_pb.Submission.Status;
  setStatus(value: qf_types_types_pb.Submission.Status): UpdateSubmissionRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UpdateSubmissionRequest.AsObject;
  static toObject(includeInstance: boolean, msg: UpdateSubmissionRequest): UpdateSubmissionRequest.AsObject;
  static serializeBinaryToWriter(message: UpdateSubmissionRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UpdateSubmissionRequest;
  static deserializeBinaryFromReader(message: UpdateSubmissionRequest, reader: jspb.BinaryReader): UpdateSubmissionRequest;
}

export namespace UpdateSubmissionRequest {
  export type AsObject = {
    submissionid: number,
    courseid: number,
    score: number,
    released: boolean,
    status: qf_types_types_pb.Submission.Status,
  }
}

export class UpdateSubmissionsRequest extends jspb.Message {
  getCourseid(): number;
  setCourseid(value: number): UpdateSubmissionsRequest;

  getAssignmentid(): number;
  setAssignmentid(value: number): UpdateSubmissionsRequest;

  getScorelimit(): number;
  setScorelimit(value: number): UpdateSubmissionsRequest;

  getRelease(): boolean;
  setRelease(value: boolean): UpdateSubmissionsRequest;

  getApprove(): boolean;
  setApprove(value: boolean): UpdateSubmissionsRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UpdateSubmissionsRequest.AsObject;
  static toObject(includeInstance: boolean, msg: UpdateSubmissionsRequest): UpdateSubmissionsRequest.AsObject;
  static serializeBinaryToWriter(message: UpdateSubmissionsRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UpdateSubmissionsRequest;
  static deserializeBinaryFromReader(message: UpdateSubmissionsRequest, reader: jspb.BinaryReader): UpdateSubmissionsRequest;
}

export namespace UpdateSubmissionsRequest {
  export type AsObject = {
    courseid: number,
    assignmentid: number,
    scorelimit: number,
    release: boolean,
    approve: boolean,
  }
}

export class SubmissionReviewersRequest extends jspb.Message {
  getSubmissionid(): number;
  setSubmissionid(value: number): SubmissionReviewersRequest;

  getCourseid(): number;
  setCourseid(value: number): SubmissionReviewersRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SubmissionReviewersRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SubmissionReviewersRequest): SubmissionReviewersRequest.AsObject;
  static serializeBinaryToWriter(message: SubmissionReviewersRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SubmissionReviewersRequest;
  static deserializeBinaryFromReader(message: SubmissionReviewersRequest, reader: jspb.BinaryReader): SubmissionReviewersRequest;
}

export namespace SubmissionReviewersRequest {
  export type AsObject = {
    submissionid: number,
    courseid: number,
  }
}

export class Providers extends jspb.Message {
  getProvidersList(): Array<string>;
  setProvidersList(value: Array<string>): Providers;
  clearProvidersList(): Providers;
  addProviders(value: string, index?: number): Providers;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Providers.AsObject;
  static toObject(includeInstance: boolean, msg: Providers): Providers.AsObject;
  static serializeBinaryToWriter(message: Providers, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Providers;
  static deserializeBinaryFromReader(message: Providers, reader: jspb.BinaryReader): Providers;
}

export namespace Providers {
  export type AsObject = {
    providersList: Array<string>,
  }
}

export class URLRequest extends jspb.Message {
  getCourseid(): number;
  setCourseid(value: number): URLRequest;

  getRepotypesList(): Array<qf_types_types_pb.Repository.Type>;
  setRepotypesList(value: Array<qf_types_types_pb.Repository.Type>): URLRequest;
  clearRepotypesList(): URLRequest;
  addRepotypes(value: qf_types_types_pb.Repository.Type, index?: number): URLRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): URLRequest.AsObject;
  static toObject(includeInstance: boolean, msg: URLRequest): URLRequest.AsObject;
  static serializeBinaryToWriter(message: URLRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): URLRequest;
  static deserializeBinaryFromReader(message: URLRequest, reader: jspb.BinaryReader): URLRequest;
}

export namespace URLRequest {
  export type AsObject = {
    courseid: number,
    repotypesList: Array<qf_types_types_pb.Repository.Type>,
  }
}

export class RepositoryRequest extends jspb.Message {
  getUserid(): number;
  setUserid(value: number): RepositoryRequest;

  getGroupid(): number;
  setGroupid(value: number): RepositoryRequest;

  getCourseid(): number;
  setCourseid(value: number): RepositoryRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RepositoryRequest.AsObject;
  static toObject(includeInstance: boolean, msg: RepositoryRequest): RepositoryRequest.AsObject;
  static serializeBinaryToWriter(message: RepositoryRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RepositoryRequest;
  static deserializeBinaryFromReader(message: RepositoryRequest, reader: jspb.BinaryReader): RepositoryRequest;
}

export namespace RepositoryRequest {
  export type AsObject = {
    userid: number,
    groupid: number,
    courseid: number,
  }
}

export class Repositories extends jspb.Message {
  getUrlsMap(): jspb.Map<string, string>;
  clearUrlsMap(): Repositories;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Repositories.AsObject;
  static toObject(includeInstance: boolean, msg: Repositories): Repositories.AsObject;
  static serializeBinaryToWriter(message: Repositories, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Repositories;
  static deserializeBinaryFromReader(message: Repositories, reader: jspb.BinaryReader): Repositories;
}

export namespace Repositories {
  export type AsObject = {
    urlsMap: Array<[string, string]>,
  }
}

export class AuthorizationResponse extends jspb.Message {
  getIsauthorized(): boolean;
  setIsauthorized(value: boolean): AuthorizationResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AuthorizationResponse.AsObject;
  static toObject(includeInstance: boolean, msg: AuthorizationResponse): AuthorizationResponse.AsObject;
  static serializeBinaryToWriter(message: AuthorizationResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AuthorizationResponse;
  static deserializeBinaryFromReader(message: AuthorizationResponse, reader: jspb.BinaryReader): AuthorizationResponse;
}

export namespace AuthorizationResponse {
  export type AsObject = {
    isauthorized: boolean,
  }
}

export class Status extends jspb.Message {
  getCode(): number;
  setCode(value: number): Status;

  getError(): string;
  setError(value: string): Status;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Status.AsObject;
  static toObject(includeInstance: boolean, msg: Status): Status.AsObject;
  static serializeBinaryToWriter(message: Status, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Status;
  static deserializeBinaryFromReader(message: Status, reader: jspb.BinaryReader): Status;
}

export namespace Status {
  export type AsObject = {
    code: number,
    error: string,
  }
}

export class SubmissionsForCourseRequest extends jspb.Message {
  getCourseid(): number;
  setCourseid(value: number): SubmissionsForCourseRequest;

  getType(): SubmissionsForCourseRequest.Type;
  setType(value: SubmissionsForCourseRequest.Type): SubmissionsForCourseRequest;

  getWithbuildinfo(): boolean;
  setWithbuildinfo(value: boolean): SubmissionsForCourseRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SubmissionsForCourseRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SubmissionsForCourseRequest): SubmissionsForCourseRequest.AsObject;
  static serializeBinaryToWriter(message: SubmissionsForCourseRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SubmissionsForCourseRequest;
  static deserializeBinaryFromReader(message: SubmissionsForCourseRequest, reader: jspb.BinaryReader): SubmissionsForCourseRequest;
}

export namespace SubmissionsForCourseRequest {
  export type AsObject = {
    courseid: number,
    type: SubmissionsForCourseRequest.Type,
    withbuildinfo: boolean,
  }

  export enum Type { 
    ALL = 0,
    INDIVIDUAL = 1,
    GROUP = 2,
  }
}

export class RebuildRequest extends jspb.Message {
  getSubmissionid(): number;
  setSubmissionid(value: number): RebuildRequest;

  getCourseid(): number;
  setCourseid(value: number): RebuildRequest;

  getAssignmentid(): number;
  setAssignmentid(value: number): RebuildRequest;

  getRebuildtypeCase(): RebuildRequest.RebuildtypeCase;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RebuildRequest.AsObject;
  static toObject(includeInstance: boolean, msg: RebuildRequest): RebuildRequest.AsObject;
  static serializeBinaryToWriter(message: RebuildRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RebuildRequest;
  static deserializeBinaryFromReader(message: RebuildRequest, reader: jspb.BinaryReader): RebuildRequest;
}

export namespace RebuildRequest {
  export type AsObject = {
    submissionid: number,
    courseid: number,
    assignmentid: number,
  }

  export enum RebuildtypeCase { 
    REBUILDTYPE_NOT_SET = 0,
    SUBMISSIONID = 1,
    COURSEID = 2,
  }
}

export class CourseUserRequest extends jspb.Message {
  getCoursecode(): string;
  setCoursecode(value: string): CourseUserRequest;

  getCourseyear(): number;
  setCourseyear(value: number): CourseUserRequest;

  getUserlogin(): string;
  setUserlogin(value: string): CourseUserRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CourseUserRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CourseUserRequest): CourseUserRequest.AsObject;
  static serializeBinaryToWriter(message: CourseUserRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CourseUserRequest;
  static deserializeBinaryFromReader(message: CourseUserRequest, reader: jspb.BinaryReader): CourseUserRequest;
}

export namespace CourseUserRequest {
  export type AsObject = {
    coursecode: string,
    courseyear: number,
    userlogin: string,
  }
}

export class Void extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Void.AsObject;
  static toObject(includeInstance: boolean, msg: Void): Void.AsObject;
  static serializeBinaryToWriter(message: Void, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Void;
  static deserializeBinaryFromReader(message: Void, reader: jspb.BinaryReader): Void;
}

export namespace Void {
  export type AsObject = {
  }
}

