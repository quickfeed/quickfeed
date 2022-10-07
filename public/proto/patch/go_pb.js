// source: patch/go.proto
/**
 * @fileoverview
 * @enhanceable
 * @suppress {missingRequire} reports error on implicit type usages.
 * @suppress {messageConventions} JS Compiler reports an error if a variable or
 *     field starts with 'MSG_' and isn't a translatable message.
 * @public
 */
// GENERATED CODE -- DO NOT EDIT!
/* eslint-disable */
// @ts-nocheck

var jspb = require('google-protobuf');
var goog = jspb;
var global = (function() { return this || window || global || self || Function('return this')(); }).call(null);

var google_protobuf_descriptor_pb = require('google-protobuf/google/protobuf/descriptor_pb.js');
goog.object.extend(proto, google_protobuf_descriptor_pb);
goog.exportSymbol('proto.go.LintOptions', null, global);
goog.exportSymbol('proto.go.Options', null, global);
goog.exportSymbol('proto.go.field', null, global);
goog.exportSymbol('proto.go.lint', null, global);
goog.exportSymbol('proto.go.message', null, global);
goog.exportSymbol('proto.go.oneof', null, global);
goog.exportSymbol('proto.go.pb_enum', null, global);
goog.exportSymbol('proto.go.value', null, global);
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.go.Options = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.go.Options, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.go.Options.displayName = 'proto.go.Options';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.go.LintOptions = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.go.LintOptions.repeatedFields_, null);
};
goog.inherits(proto.go.LintOptions, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.go.LintOptions.displayName = 'proto.go.LintOptions';
}



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.go.Options.prototype.toObject = function(opt_includeInstance) {
  return proto.go.Options.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.go.Options} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.go.Options.toObject = function(includeInstance, msg) {
  var f, obj = {
    name: (f = jspb.Message.getField(msg, 1)) == null ? undefined : f,
    embed: (f = jspb.Message.getBooleanField(msg, 2)) == null ? undefined : f,
    type: (f = jspb.Message.getField(msg, 3)) == null ? undefined : f,
    getter: (f = jspb.Message.getField(msg, 10)) == null ? undefined : f,
    tags: (f = jspb.Message.getField(msg, 20)) == null ? undefined : f,
    stringer: (f = jspb.Message.getField(msg, 30)) == null ? undefined : f,
    stringerName: (f = jspb.Message.getField(msg, 31)) == null ? undefined : f
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.go.Options}
 */
proto.go.Options.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.go.Options;
  return proto.go.Options.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.go.Options} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.go.Options}
 */
proto.go.Options.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    case 2:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setEmbed(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setType(value);
      break;
    case 10:
      var value = /** @type {string} */ (reader.readString());
      msg.setGetter(value);
      break;
    case 20:
      var value = /** @type {string} */ (reader.readString());
      msg.setTags(value);
      break;
    case 30:
      var value = /** @type {string} */ (reader.readString());
      msg.setStringer(value);
      break;
    case 31:
      var value = /** @type {string} */ (reader.readString());
      msg.setStringerName(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.go.Options.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.go.Options.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.go.Options} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.go.Options.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = /** @type {string} */ (jspb.Message.getField(message, 1));
  if (f != null) {
    writer.writeString(
      1,
      f
    );
  }
  f = /** @type {boolean} */ (jspb.Message.getField(message, 2));
  if (f != null) {
    writer.writeBool(
      2,
      f
    );
  }
  f = /** @type {string} */ (jspb.Message.getField(message, 3));
  if (f != null) {
    writer.writeString(
      3,
      f
    );
  }
  f = /** @type {string} */ (jspb.Message.getField(message, 10));
  if (f != null) {
    writer.writeString(
      10,
      f
    );
  }
  f = /** @type {string} */ (jspb.Message.getField(message, 20));
  if (f != null) {
    writer.writeString(
      20,
      f
    );
  }
  f = /** @type {string} */ (jspb.Message.getField(message, 30));
  if (f != null) {
    writer.writeString(
      30,
      f
    );
  }
  f = /** @type {string} */ (jspb.Message.getField(message, 31));
  if (f != null) {
    writer.writeString(
      31,
      f
    );
  }
};


/**
 * optional string name = 1;
 * @return {string}
 */
proto.go.Options.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.go.Options} returns this
 */
proto.go.Options.prototype.setName = function(value) {
  return jspb.Message.setField(this, 1, value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.go.Options} returns this
 */
proto.go.Options.prototype.clearName = function() {
  return jspb.Message.setField(this, 1, undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.go.Options.prototype.hasName = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional bool embed = 2;
 * @return {boolean}
 */
proto.go.Options.prototype.getEmbed = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 2, false));
};


/**
 * @param {boolean} value
 * @return {!proto.go.Options} returns this
 */
proto.go.Options.prototype.setEmbed = function(value) {
  return jspb.Message.setField(this, 2, value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.go.Options} returns this
 */
proto.go.Options.prototype.clearEmbed = function() {
  return jspb.Message.setField(this, 2, undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.go.Options.prototype.hasEmbed = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional string type = 3;
 * @return {string}
 */
proto.go.Options.prototype.getType = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.go.Options} returns this
 */
proto.go.Options.prototype.setType = function(value) {
  return jspb.Message.setField(this, 3, value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.go.Options} returns this
 */
proto.go.Options.prototype.clearType = function() {
  return jspb.Message.setField(this, 3, undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.go.Options.prototype.hasType = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional string getter = 10;
 * @return {string}
 */
proto.go.Options.prototype.getGetter = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 10, ""));
};


/**
 * @param {string} value
 * @return {!proto.go.Options} returns this
 */
proto.go.Options.prototype.setGetter = function(value) {
  return jspb.Message.setField(this, 10, value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.go.Options} returns this
 */
proto.go.Options.prototype.clearGetter = function() {
  return jspb.Message.setField(this, 10, undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.go.Options.prototype.hasGetter = function() {
  return jspb.Message.getField(this, 10) != null;
};


/**
 * optional string tags = 20;
 * @return {string}
 */
proto.go.Options.prototype.getTags = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 20, ""));
};


/**
 * @param {string} value
 * @return {!proto.go.Options} returns this
 */
proto.go.Options.prototype.setTags = function(value) {
  return jspb.Message.setField(this, 20, value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.go.Options} returns this
 */
proto.go.Options.prototype.clearTags = function() {
  return jspb.Message.setField(this, 20, undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.go.Options.prototype.hasTags = function() {
  return jspb.Message.getField(this, 20) != null;
};


/**
 * optional string stringer = 30;
 * @return {string}
 */
proto.go.Options.prototype.getStringer = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 30, ""));
};


/**
 * @param {string} value
 * @return {!proto.go.Options} returns this
 */
proto.go.Options.prototype.setStringer = function(value) {
  return jspb.Message.setField(this, 30, value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.go.Options} returns this
 */
proto.go.Options.prototype.clearStringer = function() {
  return jspb.Message.setField(this, 30, undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.go.Options.prototype.hasStringer = function() {
  return jspb.Message.getField(this, 30) != null;
};


/**
 * optional string stringer_name = 31;
 * @return {string}
 */
proto.go.Options.prototype.getStringerName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 31, ""));
};


/**
 * @param {string} value
 * @return {!proto.go.Options} returns this
 */
proto.go.Options.prototype.setStringerName = function(value) {
  return jspb.Message.setField(this, 31, value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.go.Options} returns this
 */
proto.go.Options.prototype.clearStringerName = function() {
  return jspb.Message.setField(this, 31, undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.go.Options.prototype.hasStringerName = function() {
  return jspb.Message.getField(this, 31) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.go.LintOptions.repeatedFields_ = [10];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.go.LintOptions.prototype.toObject = function(opt_includeInstance) {
  return proto.go.LintOptions.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.go.LintOptions} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.go.LintOptions.toObject = function(includeInstance, msg) {
  var f, obj = {
    all: (f = jspb.Message.getBooleanField(msg, 1)) == null ? undefined : f,
    messages: (f = jspb.Message.getBooleanField(msg, 2)) == null ? undefined : f,
    fields: (f = jspb.Message.getBooleanField(msg, 3)) == null ? undefined : f,
    enums: (f = jspb.Message.getBooleanField(msg, 4)) == null ? undefined : f,
    values: (f = jspb.Message.getBooleanField(msg, 5)) == null ? undefined : f,
    extensions: (f = jspb.Message.getBooleanField(msg, 6)) == null ? undefined : f,
    initialismsList: (f = jspb.Message.getRepeatedField(msg, 10)) == null ? undefined : f
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.go.LintOptions}
 */
proto.go.LintOptions.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.go.LintOptions;
  return proto.go.LintOptions.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.go.LintOptions} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.go.LintOptions}
 */
proto.go.LintOptions.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setAll(value);
      break;
    case 2:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setMessages(value);
      break;
    case 3:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setFields(value);
      break;
    case 4:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setEnums(value);
      break;
    case 5:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setValues(value);
      break;
    case 6:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setExtensions(value);
      break;
    case 10:
      var value = /** @type {string} */ (reader.readString());
      msg.addInitialisms(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.go.LintOptions.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.go.LintOptions.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.go.LintOptions} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.go.LintOptions.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = /** @type {boolean} */ (jspb.Message.getField(message, 1));
  if (f != null) {
    writer.writeBool(
      1,
      f
    );
  }
  f = /** @type {boolean} */ (jspb.Message.getField(message, 2));
  if (f != null) {
    writer.writeBool(
      2,
      f
    );
  }
  f = /** @type {boolean} */ (jspb.Message.getField(message, 3));
  if (f != null) {
    writer.writeBool(
      3,
      f
    );
  }
  f = /** @type {boolean} */ (jspb.Message.getField(message, 4));
  if (f != null) {
    writer.writeBool(
      4,
      f
    );
  }
  f = /** @type {boolean} */ (jspb.Message.getField(message, 5));
  if (f != null) {
    writer.writeBool(
      5,
      f
    );
  }
  f = /** @type {boolean} */ (jspb.Message.getField(message, 6));
  if (f != null) {
    writer.writeBool(
      6,
      f
    );
  }
  f = message.getInitialismsList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      10,
      f
    );
  }
};


/**
 * optional bool all = 1;
 * @return {boolean}
 */
proto.go.LintOptions.prototype.getAll = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.go.LintOptions} returns this
 */
proto.go.LintOptions.prototype.setAll = function(value) {
  return jspb.Message.setField(this, 1, value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.go.LintOptions} returns this
 */
proto.go.LintOptions.prototype.clearAll = function() {
  return jspb.Message.setField(this, 1, undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.go.LintOptions.prototype.hasAll = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional bool messages = 2;
 * @return {boolean}
 */
proto.go.LintOptions.prototype.getMessages = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 2, false));
};


/**
 * @param {boolean} value
 * @return {!proto.go.LintOptions} returns this
 */
proto.go.LintOptions.prototype.setMessages = function(value) {
  return jspb.Message.setField(this, 2, value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.go.LintOptions} returns this
 */
proto.go.LintOptions.prototype.clearMessages = function() {
  return jspb.Message.setField(this, 2, undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.go.LintOptions.prototype.hasMessages = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional bool fields = 3;
 * @return {boolean}
 */
proto.go.LintOptions.prototype.getFields = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 3, false));
};


/**
 * @param {boolean} value
 * @return {!proto.go.LintOptions} returns this
 */
proto.go.LintOptions.prototype.setFields = function(value) {
  return jspb.Message.setField(this, 3, value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.go.LintOptions} returns this
 */
proto.go.LintOptions.prototype.clearFields = function() {
  return jspb.Message.setField(this, 3, undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.go.LintOptions.prototype.hasFields = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional bool enums = 4;
 * @return {boolean}
 */
proto.go.LintOptions.prototype.getEnums = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 4, false));
};


/**
 * @param {boolean} value
 * @return {!proto.go.LintOptions} returns this
 */
proto.go.LintOptions.prototype.setEnums = function(value) {
  return jspb.Message.setField(this, 4, value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.go.LintOptions} returns this
 */
proto.go.LintOptions.prototype.clearEnums = function() {
  return jspb.Message.setField(this, 4, undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.go.LintOptions.prototype.hasEnums = function() {
  return jspb.Message.getField(this, 4) != null;
};


/**
 * optional bool values = 5;
 * @return {boolean}
 */
proto.go.LintOptions.prototype.getValues = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 5, false));
};


/**
 * @param {boolean} value
 * @return {!proto.go.LintOptions} returns this
 */
proto.go.LintOptions.prototype.setValues = function(value) {
  return jspb.Message.setField(this, 5, value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.go.LintOptions} returns this
 */
proto.go.LintOptions.prototype.clearValues = function() {
  return jspb.Message.setField(this, 5, undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.go.LintOptions.prototype.hasValues = function() {
  return jspb.Message.getField(this, 5) != null;
};


/**
 * optional bool extensions = 6;
 * @return {boolean}
 */
proto.go.LintOptions.prototype.getExtensions = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 6, false));
};


/**
 * @param {boolean} value
 * @return {!proto.go.LintOptions} returns this
 */
proto.go.LintOptions.prototype.setExtensions = function(value) {
  return jspb.Message.setField(this, 6, value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.go.LintOptions} returns this
 */
proto.go.LintOptions.prototype.clearExtensions = function() {
  return jspb.Message.setField(this, 6, undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.go.LintOptions.prototype.hasExtensions = function() {
  return jspb.Message.getField(this, 6) != null;
};


/**
 * repeated string initialisms = 10;
 * @return {!Array<string>}
 */
proto.go.LintOptions.prototype.getInitialismsList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 10));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.go.LintOptions} returns this
 */
proto.go.LintOptions.prototype.setInitialismsList = function(value) {
  return jspb.Message.setField(this, 10, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.go.LintOptions} returns this
 */
proto.go.LintOptions.prototype.addInitialisms = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 10, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.go.LintOptions} returns this
 */
proto.go.LintOptions.prototype.clearInitialismsList = function() {
  return this.setInitialismsList([]);
};



/**
 * A tuple of {field number, class constructor} for the extension
 * field named `message`.
 * @type {!jspb.ExtensionFieldInfo<!proto.go.Options>}
 */
proto.go.message = new jspb.ExtensionFieldInfo(
    7001,
    {message: 0},
    proto.go.Options,
     /** @type {?function((boolean|undefined),!jspb.Message=): !Object} */ (
         proto.go.Options.toObject),
    0);

google_protobuf_descriptor_pb.MessageOptions.extensionsBinary[7001] = new jspb.ExtensionFieldBinaryInfo(
    proto.go.message,
    jspb.BinaryReader.prototype.readMessage,
    jspb.BinaryWriter.prototype.writeMessage,
    proto.go.Options.serializeBinaryToWriter,
    proto.go.Options.deserializeBinaryFromReader,
    false);
// This registers the extension field with the extended class, so that
// toObject() will function correctly.
google_protobuf_descriptor_pb.MessageOptions.extensions[7001] = proto.go.message;


/**
 * A tuple of {field number, class constructor} for the extension
 * field named `field`.
 * @type {!jspb.ExtensionFieldInfo<!proto.go.Options>}
 */
proto.go.field = new jspb.ExtensionFieldInfo(
    7001,
    {field: 0},
    proto.go.Options,
     /** @type {?function((boolean|undefined),!jspb.Message=): !Object} */ (
         proto.go.Options.toObject),
    0);

google_protobuf_descriptor_pb.FieldOptions.extensionsBinary[7001] = new jspb.ExtensionFieldBinaryInfo(
    proto.go.field,
    jspb.BinaryReader.prototype.readMessage,
    jspb.BinaryWriter.prototype.writeMessage,
    proto.go.Options.serializeBinaryToWriter,
    proto.go.Options.deserializeBinaryFromReader,
    false);
// This registers the extension field with the extended class, so that
// toObject() will function correctly.
google_protobuf_descriptor_pb.FieldOptions.extensions[7001] = proto.go.field;


/**
 * A tuple of {field number, class constructor} for the extension
 * field named `oneof`.
 * @type {!jspb.ExtensionFieldInfo<!proto.go.Options>}
 */
proto.go.oneof = new jspb.ExtensionFieldInfo(
    7001,
    {oneof: 0},
    proto.go.Options,
     /** @type {?function((boolean|undefined),!jspb.Message=): !Object} */ (
         proto.go.Options.toObject),
    0);

google_protobuf_descriptor_pb.OneofOptions.extensionsBinary[7001] = new jspb.ExtensionFieldBinaryInfo(
    proto.go.oneof,
    jspb.BinaryReader.prototype.readMessage,
    jspb.BinaryWriter.prototype.writeMessage,
    proto.go.Options.serializeBinaryToWriter,
    proto.go.Options.deserializeBinaryFromReader,
    false);
// This registers the extension field with the extended class, so that
// toObject() will function correctly.
google_protobuf_descriptor_pb.OneofOptions.extensions[7001] = proto.go.oneof;


/**
 * A tuple of {field number, class constructor} for the extension
 * field named `pb_enum`.
 * @type {!jspb.ExtensionFieldInfo<!proto.go.Options>}
 */
proto.go.pb_enum = new jspb.ExtensionFieldInfo(
    7001,
    {pb_enum: 0},
    proto.go.Options,
     /** @type {?function((boolean|undefined),!jspb.Message=): !Object} */ (
         proto.go.Options.toObject),
    0);

google_protobuf_descriptor_pb.EnumOptions.extensionsBinary[7001] = new jspb.ExtensionFieldBinaryInfo(
    proto.go.pb_enum,
    jspb.BinaryReader.prototype.readMessage,
    jspb.BinaryWriter.prototype.writeMessage,
    proto.go.Options.serializeBinaryToWriter,
    proto.go.Options.deserializeBinaryFromReader,
    false);
// This registers the extension field with the extended class, so that
// toObject() will function correctly.
google_protobuf_descriptor_pb.EnumOptions.extensions[7001] = proto.go.pb_enum;


/**
 * A tuple of {field number, class constructor} for the extension
 * field named `value`.
 * @type {!jspb.ExtensionFieldInfo<!proto.go.Options>}
 */
proto.go.value = new jspb.ExtensionFieldInfo(
    7001,
    {value: 0},
    proto.go.Options,
     /** @type {?function((boolean|undefined),!jspb.Message=): !Object} */ (
         proto.go.Options.toObject),
    0);

google_protobuf_descriptor_pb.EnumValueOptions.extensionsBinary[7001] = new jspb.ExtensionFieldBinaryInfo(
    proto.go.value,
    jspb.BinaryReader.prototype.readMessage,
    jspb.BinaryWriter.prototype.writeMessage,
    proto.go.Options.serializeBinaryToWriter,
    proto.go.Options.deserializeBinaryFromReader,
    false);
// This registers the extension field with the extended class, so that
// toObject() will function correctly.
google_protobuf_descriptor_pb.EnumValueOptions.extensions[7001] = proto.go.value;


/**
 * A tuple of {field number, class constructor} for the extension
 * field named `lint`.
 * @type {!jspb.ExtensionFieldInfo<!proto.go.LintOptions>}
 */
proto.go.lint = new jspb.ExtensionFieldInfo(
    7001,
    {lint: 0},
    proto.go.LintOptions,
     /** @type {?function((boolean|undefined),!jspb.Message=): !Object} */ (
         proto.go.LintOptions.toObject),
    0);

google_protobuf_descriptor_pb.FileOptions.extensionsBinary[7001] = new jspb.ExtensionFieldBinaryInfo(
    proto.go.lint,
    jspb.BinaryReader.prototype.readMessage,
    jspb.BinaryWriter.prototype.writeMessage,
    proto.go.LintOptions.serializeBinaryToWriter,
    proto.go.LintOptions.deserializeBinaryFromReader,
    false);
// This registers the extension field with the extended class, so that
// toObject() will function correctly.
google_protobuf_descriptor_pb.FileOptions.extensions[7001] = proto.go.lint;

goog.object.extend(exports, proto.go);
