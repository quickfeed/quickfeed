import * as jspb from 'google-protobuf'

import * as google_protobuf_timestamp_pb from 'google-protobuf/google/protobuf/timestamp_pb';


export class Score extends jspb.Message {
  getId(): number;
  setId(value: number): Score;

  getSubmissionid(): number;
  setSubmissionid(value: number): Score;

  getSecret(): string;
  setSecret(value: string): Score;

  getTestname(): string;
  setTestname(value: string): Score;

  getTaskname(): string;
  setTaskname(value: string): Score;

  getScore(): number;
  setScore(value: number): Score;

  getMaxscore(): number;
  setMaxscore(value: number): Score;

  getWeight(): number;
  setWeight(value: number): Score;

  getTestdetails(): string;
  setTestdetails(value: string): Score;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Score.AsObject;
  static toObject(includeInstance: boolean, msg: Score): Score.AsObject;
  static serializeBinaryToWriter(message: Score, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Score;
  static deserializeBinaryFromReader(message: Score, reader: jspb.BinaryReader): Score;
}

export namespace Score {
  export type AsObject = {
    id: number,
    submissionid: number,
    secret: string,
    testname: string,
    taskname: string,
    score: number,
    maxscore: number,
    weight: number,
    testdetails: string,
  }
}

export class BuildInfo extends jspb.Message {
  getId(): number;
  setId(value: number): BuildInfo;

  getSubmissionid(): number;
  setSubmissionid(value: number): BuildInfo;

  getBuilddate(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setBuilddate(value?: google_protobuf_timestamp_pb.Timestamp): BuildInfo;
  hasBuilddate(): boolean;
  clearBuilddate(): BuildInfo;

  getBuildlog(): string;
  setBuildlog(value: string): BuildInfo;

  getExectime(): number;
  setExectime(value: number): BuildInfo;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): BuildInfo.AsObject;
  static toObject(includeInstance: boolean, msg: BuildInfo): BuildInfo.AsObject;
  static serializeBinaryToWriter(message: BuildInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): BuildInfo;
  static deserializeBinaryFromReader(message: BuildInfo, reader: jspb.BinaryReader): BuildInfo;
}

export namespace BuildInfo {
  export type AsObject = {
    id: number,
    submissionid: number,
    builddate?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    buildlog: string,
    exectime: number,
  }
}

