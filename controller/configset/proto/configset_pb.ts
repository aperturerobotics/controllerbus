// @generated by protoc-gen-es v1.9.0 with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/controllerbus/controller/configset/proto/configset.proto (package configset.proto, syntax proto3)
/* eslint-disable */

import type {
  BinaryReadOptions,
  FieldList,
  JsonReadOptions,
  JsonValue,
  PartialMessage,
  PlainMessage,
} from '@bufbuild/protobuf'
import { Message, proto3, protoInt64 } from '@bufbuild/protobuf'

/**
 * ConfigSet contains a configuration set.
 *
 * @generated from message configset.proto.ConfigSet
 */
export class ConfigSet extends Message<ConfigSet> {
  /**
   * Configs contains the controller configurations.
   *
   * @generated from field: map<string, configset.proto.ControllerConfig> configs = 1;
   */
  configs: { [key: string]: ControllerConfig } = {}

  constructor(data?: PartialMessage<ConfigSet>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'configset.proto.ConfigSet'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    {
      no: 1,
      name: 'configs',
      kind: 'map',
      K: 9 /* ScalarType.STRING */,
      V: { kind: 'message', T: ControllerConfig },
    },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): ConfigSet {
    return new ConfigSet().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): ConfigSet {
    return new ConfigSet().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): ConfigSet {
    return new ConfigSet().fromJsonString(jsonString, options)
  }

  static equals(
    a: ConfigSet | PlainMessage<ConfigSet> | undefined,
    b: ConfigSet | PlainMessage<ConfigSet> | undefined,
  ): boolean {
    return proto3.util.equals(ConfigSet, a, b)
  }
}

/**
 * ControllerConfig contains a controller configuration.
 *
 * protobuf-go-lite:disable-json
 *
 * @generated from message configset.proto.ControllerConfig
 */
export class ControllerConfig extends Message<ControllerConfig> {
  /**
   * Id is the config ID.
   *
   * @generated from field: string id = 1;
   */
  id = ''

  /**
   * Rev is the revision number of the configuration.
   *
   * @generated from field: uint64 rev = 2;
   */
  rev = protoInt64.zero

  /**
   * Config is the configuration object.
   * Proto supports: protobuf (binary) and json (starting with {).
   * Json supports: protobuf (base64) and json (inline object).
   *
   * @generated from field: bytes config = 3;
   */
  config = new Uint8Array(0)

  constructor(data?: PartialMessage<ControllerConfig>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'configset.proto.ControllerConfig'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: 'id', kind: 'scalar', T: 9 /* ScalarType.STRING */ },
    { no: 2, name: 'rev', kind: 'scalar', T: 4 /* ScalarType.UINT64 */ },
    { no: 3, name: 'config', kind: 'scalar', T: 12 /* ScalarType.BYTES */ },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): ControllerConfig {
    return new ControllerConfig().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): ControllerConfig {
    return new ControllerConfig().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): ControllerConfig {
    return new ControllerConfig().fromJsonString(jsonString, options)
  }

  static equals(
    a: ControllerConfig | PlainMessage<ControllerConfig> | undefined,
    b: ControllerConfig | PlainMessage<ControllerConfig> | undefined,
  ): boolean {
    return proto3.util.equals(ControllerConfig, a, b)
  }
}