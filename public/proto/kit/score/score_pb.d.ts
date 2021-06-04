import * as jspb from "google-protobuf"

export class Score extends jspb.Message {
  getSecret(): string;
  setSecret(value: string): Score;

  getTestname(): string;
  setTestname(value: string): Score;

  getScore(): number;
  setScore(value: number): Score;

  getMaxscore(): number;
  setMaxscore(value: number): Score;

  getWeight(): number;
  setWeight(value: number): Score;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Score.AsObject;
  static toObject(includeInstance: boolean, msg: Score): Score.AsObject;
  static serializeBinaryToWriter(message: Score, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Score;
  static deserializeBinaryFromReader(message: Score, reader: jspb.BinaryReader): Score;
}

export namespace Score {
  export type AsObject = {
    secret: string,
    testname: string,
    score: number,
    maxscore: number,
    weight: number,
  }
}

export class Results extends jspb.Message {
  getBuildinfo(): BuildInfo | undefined;
  setBuildinfo(value?: BuildInfo): Results;
  hasBuildinfo(): boolean;
  clearBuildinfo(): Results;

  getTestnamesList(): Array<string>;
  setTestnamesList(value: Array<string>): Results;
  clearTestnamesList(): Results;
  addTestnames(value: string, index?: number): Results;

  getScoremapMap(): jspb.Map<string, Score>;
  clearScoremapMap(): Results;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Results.AsObject;
  static toObject(includeInstance: boolean, msg: Results): Results.AsObject;
  static serializeBinaryToWriter(message: Results, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Results;
  static deserializeBinaryFromReader(message: Results, reader: jspb.BinaryReader): Results;
}

export namespace Results {
  export type AsObject = {
    buildinfo?: BuildInfo.AsObject,
    testnamesList: Array<string>,
    scoremapMap: Array<[string, Score.AsObject]>,
  }
}

export class BuildInfo extends jspb.Message {
  getBuildid(): number;
  setBuildid(value: number): BuildInfo;

  getBuilddate(): string;
  setBuilddate(value: string): BuildInfo;

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
    buildid: number,
    builddate: string,
    buildlog: string,
    exectime: number,
  }
}

