import * as jspb from 'google-protobuf'

import * as google_protobuf_descriptor_pb from 'google-protobuf/google/protobuf/descriptor_pb';


export class Options extends jspb.Message {
  getName(): string;
  setName(value: string): Options;
  hasName(): boolean;
  clearName(): Options;

  getEmbed(): boolean;
  setEmbed(value: boolean): Options;
  hasEmbed(): boolean;
  clearEmbed(): Options;

  getType(): string;
  setType(value: string): Options;
  hasType(): boolean;
  clearType(): Options;

  getGetter(): string;
  setGetter(value: string): Options;
  hasGetter(): boolean;
  clearGetter(): Options;

  getTags(): string;
  setTags(value: string): Options;
  hasTags(): boolean;
  clearTags(): Options;

  getStringer(): string;
  setStringer(value: string): Options;
  hasStringer(): boolean;
  clearStringer(): Options;

  getStringerName(): string;
  setStringerName(value: string): Options;
  hasStringerName(): boolean;
  clearStringerName(): Options;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Options.AsObject;
  static toObject(includeInstance: boolean, msg: Options): Options.AsObject;
  static serializeBinaryToWriter(message: Options, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Options;
  static deserializeBinaryFromReader(message: Options, reader: jspb.BinaryReader): Options;
}

export namespace Options {
  export type AsObject = {
    name?: string,
    embed?: boolean,
    type?: string,
    getter?: string,
    tags?: string,
    stringer?: string,
    stringerName?: string,
  }
}

export class LintOptions extends jspb.Message {
  getAll(): boolean;
  setAll(value: boolean): LintOptions;
  hasAll(): boolean;
  clearAll(): LintOptions;

  getMessages(): boolean;
  setMessages(value: boolean): LintOptions;
  hasMessages(): boolean;
  clearMessages(): LintOptions;

  getFields(): boolean;
  setFields(value: boolean): LintOptions;
  hasFields(): boolean;
  clearFields(): LintOptions;

  getEnums(): boolean;
  setEnums(value: boolean): LintOptions;
  hasEnums(): boolean;
  clearEnums(): LintOptions;

  getValues(): boolean;
  setValues(value: boolean): LintOptions;
  hasValues(): boolean;
  clearValues(): LintOptions;

  getExtensions(): boolean;
  setExtensions(value: boolean): LintOptions;
  hasExtensions(): boolean;
  clearExtensions(): LintOptions;

  getInitialismsList(): Array<string>;
  setInitialismsList(value: Array<string>): LintOptions;
  clearInitialismsList(): LintOptions;
  addInitialisms(value: string, index?: number): LintOptions;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): LintOptions.AsObject;
  static toObject(includeInstance: boolean, msg: LintOptions): LintOptions.AsObject;
  static serializeBinaryToWriter(message: LintOptions, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): LintOptions;
  static deserializeBinaryFromReader(message: LintOptions, reader: jspb.BinaryReader): LintOptions;
}

export namespace LintOptions {
  export type AsObject = {
    all?: boolean,
    messages?: boolean,
    fields?: boolean,
    enums?: boolean,
    values?: boolean,
    extensions?: boolean,
    initialismsList: Array<string>,
  }
}

