/* eslint-disable */
import Long from 'long'
import { ConfigSet } from '../configset/proto/configset.pb.js'
import { Info } from '../controller.pb.js'
import * as _m0 from 'protobufjs/minimal'

export const protobufPackage = 'controller.exec'

/** ControllerStatus holds basic status for a controller. */
export enum ControllerStatus {
  /** ControllerStatus_UNKNOWN - ControllerStatus_UNKNOWN is unrecognized. */
  ControllerStatus_UNKNOWN = 0,
  /** ControllerStatus_CONFIGURING - ControllerStatus_CONFIGURING indicates the controller is configuring. */
  ControllerStatus_CONFIGURING = 1,
  /** ControllerStatus_RUNNING - ControllerStatus_RUNNING indicates the controller is running. */
  ControllerStatus_RUNNING = 2,
  /** ControllerStatus_ERROR - ControllerStatus_ERROR indicates the controller is terminated with an error. */
  ControllerStatus_ERROR = 3,
  UNRECOGNIZED = -1,
}

export function controllerStatusFromJSON(object: any): ControllerStatus {
  switch (object) {
    case 0:
    case 'ControllerStatus_UNKNOWN':
      return ControllerStatus.ControllerStatus_UNKNOWN
    case 1:
    case 'ControllerStatus_CONFIGURING':
      return ControllerStatus.ControllerStatus_CONFIGURING
    case 2:
    case 'ControllerStatus_RUNNING':
      return ControllerStatus.ControllerStatus_RUNNING
    case 3:
    case 'ControllerStatus_ERROR':
      return ControllerStatus.ControllerStatus_ERROR
    case -1:
    case 'UNRECOGNIZED':
    default:
      return ControllerStatus.UNRECOGNIZED
  }
}

export function controllerStatusToJSON(object: ControllerStatus): string {
  switch (object) {
    case ControllerStatus.ControllerStatus_UNKNOWN:
      return 'ControllerStatus_UNKNOWN'
    case ControllerStatus.ControllerStatus_CONFIGURING:
      return 'ControllerStatus_CONFIGURING'
    case ControllerStatus.ControllerStatus_RUNNING:
      return 'ControllerStatus_RUNNING'
    case ControllerStatus.ControllerStatus_ERROR:
      return 'ControllerStatus_ERROR'
    case ControllerStatus.UNRECOGNIZED:
    default:
      return 'UNRECOGNIZED'
  }
}

/** ExecControllerRequest is a protobuf request to execute a controller. */
export interface ExecControllerRequest {
  /** ConfigSet is the controller config set to execute. */
  configSet: ConfigSet | undefined
  /**
   * ConfigSetYaml is optionally the YAML form of config_set to parse.
   * Merged with config_set.
   */
  configSetYaml: string
  /** ConfigSetYamlOverwrite sets if the yaml portion overwrites the proto portion. */
  configSetYamlOverwrite: boolean
}

/** ExecControllerResponse is a protobuf response stream. */
export interface ExecControllerResponse {
  /** Id is the configset identifier for this status report. */
  id: string
  /** Status is the controller execution status. */
  status: ControllerStatus
  /** ControllerInfo may contain the running controller info. */
  controllerInfo: Info | undefined
  /** ErrorInfo may contain the error information. */
  errorInfo: string
}

function createBaseExecControllerRequest(): ExecControllerRequest {
  return {
    configSet: undefined,
    configSetYaml: '',
    configSetYamlOverwrite: false,
  }
}

export const ExecControllerRequest = {
  encode(
    message: ExecControllerRequest,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.configSet !== undefined) {
      ConfigSet.encode(message.configSet, writer.uint32(10).fork()).ldelim()
    }
    if (message.configSetYaml !== '') {
      writer.uint32(18).string(message.configSetYaml)
    }
    if (message.configSetYamlOverwrite === true) {
      writer.uint32(24).bool(message.configSetYamlOverwrite)
    }
    return writer
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): ExecControllerRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseExecControllerRequest()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          message.configSet = ConfigSet.decode(reader, reader.uint32())
          break
        case 2:
          message.configSetYaml = reader.string()
          break
        case 3:
          message.configSetYamlOverwrite = reader.bool()
          break
        default:
          reader.skipType(tag & 7)
          break
      }
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<ExecControllerRequest, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<ExecControllerRequest | ExecControllerRequest[]>
      | Iterable<ExecControllerRequest | ExecControllerRequest[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [ExecControllerRequest.encode(p).finish()]
        }
      } else {
        yield* [ExecControllerRequest.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, ExecControllerRequest>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<ExecControllerRequest> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [ExecControllerRequest.decode(p)]
        }
      } else {
        yield* [ExecControllerRequest.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): ExecControllerRequest {
    return {
      configSet: isSet(object.configSet)
        ? ConfigSet.fromJSON(object.configSet)
        : undefined,
      configSetYaml: isSet(object.configSetYaml)
        ? String(object.configSetYaml)
        : '',
      configSetYamlOverwrite: isSet(object.configSetYamlOverwrite)
        ? Boolean(object.configSetYamlOverwrite)
        : false,
    }
  },

  toJSON(message: ExecControllerRequest): unknown {
    const obj: any = {}
    message.configSet !== undefined &&
      (obj.configSet = message.configSet
        ? ConfigSet.toJSON(message.configSet)
        : undefined)
    message.configSetYaml !== undefined &&
      (obj.configSetYaml = message.configSetYaml)
    message.configSetYamlOverwrite !== undefined &&
      (obj.configSetYamlOverwrite = message.configSetYamlOverwrite)
    return obj
  },

  fromPartial<I extends Exact<DeepPartial<ExecControllerRequest>, I>>(
    object: I
  ): ExecControllerRequest {
    const message = createBaseExecControllerRequest()
    message.configSet =
      object.configSet !== undefined && object.configSet !== null
        ? ConfigSet.fromPartial(object.configSet)
        : undefined
    message.configSetYaml = object.configSetYaml ?? ''
    message.configSetYamlOverwrite = object.configSetYamlOverwrite ?? false
    return message
  },
}

function createBaseExecControllerResponse(): ExecControllerResponse {
  return { id: '', status: 0, controllerInfo: undefined, errorInfo: '' }
}

export const ExecControllerResponse = {
  encode(
    message: ExecControllerResponse,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.id !== '') {
      writer.uint32(10).string(message.id)
    }
    if (message.status !== 0) {
      writer.uint32(16).int32(message.status)
    }
    if (message.controllerInfo !== undefined) {
      Info.encode(message.controllerInfo, writer.uint32(26).fork()).ldelim()
    }
    if (message.errorInfo !== '') {
      writer.uint32(34).string(message.errorInfo)
    }
    return writer
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): ExecControllerResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseExecControllerResponse()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          message.id = reader.string()
          break
        case 2:
          message.status = reader.int32() as any
          break
        case 3:
          message.controllerInfo = Info.decode(reader, reader.uint32())
          break
        case 4:
          message.errorInfo = reader.string()
          break
        default:
          reader.skipType(tag & 7)
          break
      }
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<ExecControllerResponse, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<ExecControllerResponse | ExecControllerResponse[]>
      | Iterable<ExecControllerResponse | ExecControllerResponse[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [ExecControllerResponse.encode(p).finish()]
        }
      } else {
        yield* [ExecControllerResponse.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, ExecControllerResponse>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<ExecControllerResponse> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [ExecControllerResponse.decode(p)]
        }
      } else {
        yield* [ExecControllerResponse.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): ExecControllerResponse {
    return {
      id: isSet(object.id) ? String(object.id) : '',
      status: isSet(object.status)
        ? controllerStatusFromJSON(object.status)
        : 0,
      controllerInfo: isSet(object.controllerInfo)
        ? Info.fromJSON(object.controllerInfo)
        : undefined,
      errorInfo: isSet(object.errorInfo) ? String(object.errorInfo) : '',
    }
  },

  toJSON(message: ExecControllerResponse): unknown {
    const obj: any = {}
    message.id !== undefined && (obj.id = message.id)
    message.status !== undefined &&
      (obj.status = controllerStatusToJSON(message.status))
    message.controllerInfo !== undefined &&
      (obj.controllerInfo = message.controllerInfo
        ? Info.toJSON(message.controllerInfo)
        : undefined)
    message.errorInfo !== undefined && (obj.errorInfo = message.errorInfo)
    return obj
  },

  fromPartial<I extends Exact<DeepPartial<ExecControllerResponse>, I>>(
    object: I
  ): ExecControllerResponse {
    const message = createBaseExecControllerResponse()
    message.id = object.id ?? ''
    message.status = object.status ?? 0
    message.controllerInfo =
      object.controllerInfo !== undefined && object.controllerInfo !== null
        ? Info.fromPartial(object.controllerInfo)
        : undefined
    message.errorInfo = object.errorInfo ?? ''
    return message
  },
}

type Builtin =
  | Date
  | Function
  | Uint8Array
  | string
  | number
  | boolean
  | undefined

export type DeepPartial<T> = T extends Builtin
  ? T
  : T extends Long
  ? string | number | Long
  : T extends Array<infer U>
  ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U>
  ? ReadonlyArray<DeepPartial<U>>
  : T extends { $case: string }
  ? { [K in keyof Omit<T, '$case'>]?: DeepPartial<T[K]> } & {
      $case: T['$case']
    }
  : T extends {}
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>

type KeysOfUnion<T> = T extends T ? keyof T : never
export type Exact<P, I extends P> = P extends Builtin
  ? P
  : P & { [K in keyof P]: Exact<P[K], I[K]> } & Record<
        Exclude<keyof I, KeysOfUnion<P>>,
        never
      >

if (_m0.util.Long !== Long) {
  _m0.util.Long = Long as any
  _m0.configure()
}

function isSet(value: any): boolean {
  return value !== null && value !== undefined
}
