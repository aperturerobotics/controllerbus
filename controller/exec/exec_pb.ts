// @generated by protoc-gen-es v1.8.0 with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/controllerbus/controller/exec/exec.proto (package controller.exec, syntax proto3)
/* eslint-disable */

import type {
  BinaryReadOptions,
  FieldList,
  JsonReadOptions,
  JsonValue,
  PartialMessage,
  PlainMessage,
} from '@bufbuild/protobuf'
import { Message, proto3 } from '@bufbuild/protobuf'
import { ConfigSet } from '../configset/proto/configset_pb.js'
import { Info } from '../controller_pb.js'

/**
 * ControllerStatus holds basic status for a controller.
 *
 * @generated from enum controller.exec.ControllerStatus
 */
export enum ControllerStatus {
  /**
   * ControllerStatus_UNKNOWN is unrecognized.
   *
   * @generated from enum value: ControllerStatus_UNKNOWN = 0;
   */
  ControllerStatus_UNKNOWN = 0,

  /**
   * ControllerStatus_CONFIGURING indicates the controller is configuring.
   *
   * @generated from enum value: ControllerStatus_CONFIGURING = 1;
   */
  ControllerStatus_CONFIGURING = 1,

  /**
   * ControllerStatus_RUNNING indicates the controller is running.
   *
   * @generated from enum value: ControllerStatus_RUNNING = 2;
   */
  ControllerStatus_RUNNING = 2,

  /**
   * ControllerStatus_ERROR indicates the controller is terminated with an error.
   *
   * @generated from enum value: ControllerStatus_ERROR = 3;
   */
  ControllerStatus_ERROR = 3,
}
// Retrieve enum metadata with: proto3.getEnumType(ControllerStatus)
proto3.util.setEnumType(ControllerStatus, 'controller.exec.ControllerStatus', [
  { no: 0, name: 'ControllerStatus_UNKNOWN' },
  { no: 1, name: 'ControllerStatus_CONFIGURING' },
  { no: 2, name: 'ControllerStatus_RUNNING' },
  { no: 3, name: 'ControllerStatus_ERROR' },
])

/**
 * ExecControllerRequest is a protobuf request to execute a controller.
 *
 * @generated from message controller.exec.ExecControllerRequest
 */
export class ExecControllerRequest extends Message<ExecControllerRequest> {
  /**
   * ConfigSet is the controller config set to execute.
   *
   * @generated from field: configset.proto.ConfigSet config_set = 1;
   */
  configSet?: ConfigSet

  /**
   * ConfigSetYaml is optionally the YAML form of config_set to parse.
   * Merged with config_set.
   *
   * @generated from field: string config_set_yaml = 2;
   */
  configSetYaml = ''

  /**
   * ConfigSetYamlOverwrite sets if the yaml portion overwrites the proto portion.
   *
   * @generated from field: bool config_set_yaml_overwrite = 3;
   */
  configSetYamlOverwrite = false

  constructor(data?: PartialMessage<ExecControllerRequest>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'controller.exec.ExecControllerRequest'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: 'config_set', kind: 'message', T: ConfigSet },
    {
      no: 2,
      name: 'config_set_yaml',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
    },
    {
      no: 3,
      name: 'config_set_yaml_overwrite',
      kind: 'scalar',
      T: 8 /* ScalarType.BOOL */,
    },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): ExecControllerRequest {
    return new ExecControllerRequest().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): ExecControllerRequest {
    return new ExecControllerRequest().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): ExecControllerRequest {
    return new ExecControllerRequest().fromJsonString(jsonString, options)
  }

  static equals(
    a: ExecControllerRequest | PlainMessage<ExecControllerRequest> | undefined,
    b: ExecControllerRequest | PlainMessage<ExecControllerRequest> | undefined,
  ): boolean {
    return proto3.util.equals(ExecControllerRequest, a, b)
  }
}

/**
 * ExecControllerResponse is a protobuf response stream.
 *
 * @generated from message controller.exec.ExecControllerResponse
 */
export class ExecControllerResponse extends Message<ExecControllerResponse> {
  /**
   * Id is the configset identifier for this status report.
   *
   * @generated from field: string id = 1;
   */
  id = ''

  /**
   * Status is the controller execution status.
   *
   * @generated from field: controller.exec.ControllerStatus status = 2;
   */
  status = ControllerStatus.ControllerStatus_UNKNOWN

  /**
   * ControllerInfo may contain the running controller info.
   *
   * @generated from field: controller.Info controller_info = 3;
   */
  controllerInfo?: Info

  /**
   * ErrorInfo may contain the error information.
   *
   * @generated from field: string error_info = 4;
   */
  errorInfo = ''

  constructor(data?: PartialMessage<ExecControllerResponse>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'controller.exec.ExecControllerResponse'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: 'id', kind: 'scalar', T: 9 /* ScalarType.STRING */ },
    {
      no: 2,
      name: 'status',
      kind: 'enum',
      T: proto3.getEnumType(ControllerStatus),
    },
    { no: 3, name: 'controller_info', kind: 'message', T: Info },
    { no: 4, name: 'error_info', kind: 'scalar', T: 9 /* ScalarType.STRING */ },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): ExecControllerResponse {
    return new ExecControllerResponse().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): ExecControllerResponse {
    return new ExecControllerResponse().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): ExecControllerResponse {
    return new ExecControllerResponse().fromJsonString(jsonString, options)
  }

  static equals(
    a:
      | ExecControllerResponse
      | PlainMessage<ExecControllerResponse>
      | undefined,
    b:
      | ExecControllerResponse
      | PlainMessage<ExecControllerResponse>
      | undefined,
  ): boolean {
    return proto3.util.equals(ExecControllerResponse, a, b)
  }
}
