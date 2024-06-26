// @generated by protoc-gen-es v1.10.0 with parameter "target=ts"
// @generated from file kit/score/score.proto (package score, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3, protoInt64, Timestamp } from "@bufbuild/protobuf";

/**
 * Score give the score for a single test named TestName.
 *
 * @generated from message score.Score
 */
export class Score extends Message<Score> {
  /**
   * @generated from field: uint64 ID = 1;
   */
  ID = protoInt64.zero;

  /**
   * @generated from field: uint64 SubmissionID = 2;
   */
  SubmissionID = protoInt64.zero;

  /**
   * the unique identifier for a scoring session
   *
   * @generated from field: string Secret = 3;
   */
  Secret = "";

  /**
   * name of the test
   *
   * @generated from field: string TestName = 4;
   */
  TestName = "";

  /**
   * name of task this score belongs to
   *
   * @generated from field: string TaskName = 5;
   */
  TaskName = "";

  /**
   * the score obtained
   *
   * @generated from field: int32 Score = 6;
   */
  Score = 0;

  /**
   * max score possible to get on this specific test
   *
   * @generated from field: int32 MaxScore = 7;
   */
  MaxScore = 0;

  /**
   * the weight of this test; used to compute final grade
   *
   * @generated from field: int32 Weight = 8;
   */
  Weight = 0;

  /**
   * if populated, the frontend may display these details
   *
   * @generated from field: string TestDetails = 9;
   */
  TestDetails = "";

  constructor(data?: PartialMessage<Score>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "score.Score";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "ID", kind: "scalar", T: 4 /* ScalarType.UINT64 */ },
    { no: 2, name: "SubmissionID", kind: "scalar", T: 4 /* ScalarType.UINT64 */ },
    { no: 3, name: "Secret", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "TestName", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 5, name: "TaskName", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 6, name: "Score", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
    { no: 7, name: "MaxScore", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
    { no: 8, name: "Weight", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
    { no: 9, name: "TestDetails", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Score {
    return new Score().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Score {
    return new Score().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Score {
    return new Score().fromJsonString(jsonString, options);
  }

  static equals(a: Score | PlainMessage<Score> | undefined, b: Score | PlainMessage<Score> | undefined): boolean {
    return proto3.util.equals(Score, a, b);
  }
}

/**
 * BuildInfo holds build data for an assignment's test execution.
 *
 * @generated from message score.BuildInfo
 */
export class BuildInfo extends Message<BuildInfo> {
  /**
   * @generated from field: uint64 ID = 1;
   */
  ID = protoInt64.zero;

  /**
   * @generated from field: uint64 SubmissionID = 2;
   */
  SubmissionID = protoInt64.zero;

  /**
   * @generated from field: string BuildLog = 3;
   */
  BuildLog = "";

  /**
   * @generated from field: int64 ExecTime = 4;
   */
  ExecTime = protoInt64.zero;

  /**
   * @generated from field: google.protobuf.Timestamp BuildDate = 5;
   */
  BuildDate?: Timestamp;

  /**
   * @generated from field: google.protobuf.Timestamp SubmissionDate = 6;
   */
  SubmissionDate?: Timestamp;

  constructor(data?: PartialMessage<BuildInfo>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "score.BuildInfo";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "ID", kind: "scalar", T: 4 /* ScalarType.UINT64 */ },
    { no: 2, name: "SubmissionID", kind: "scalar", T: 4 /* ScalarType.UINT64 */ },
    { no: 3, name: "BuildLog", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "ExecTime", kind: "scalar", T: 3 /* ScalarType.INT64 */ },
    { no: 5, name: "BuildDate", kind: "message", T: Timestamp },
    { no: 6, name: "SubmissionDate", kind: "message", T: Timestamp },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): BuildInfo {
    return new BuildInfo().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): BuildInfo {
    return new BuildInfo().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): BuildInfo {
    return new BuildInfo().fromJsonString(jsonString, options);
  }

  static equals(a: BuildInfo | PlainMessage<BuildInfo> | undefined, b: BuildInfo | PlainMessage<BuildInfo> | undefined): boolean {
    return proto3.util.equals(BuildInfo, a, b);
  }
}

