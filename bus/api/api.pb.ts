/* eslint-disable */
import Long from "long";
import _m0 from "protobufjs/minimal.js";
import { Info } from "../../controller/controller.pb.js";
import { ExecControllerRequest, ExecControllerResponse } from "../../controller/exec/exec.pb.js";
import { DirectiveState } from "../../directive/directive.pb.js";

export const protobufPackage = "bus.api";

/** Config are configuration arguments. */
export interface Config {
  /** EnableExecController enables the exec controller API. */
  enableExecController: boolean;
}

/** GetBusInfoRequest is the request type for GetBusInfo. */
export interface GetBusInfoRequest {
}

/** GetBusInfoResponse is the response type for GetBusInfo. */
export interface GetBusInfoResponse {
  /** RunningControllers is the list of running controllers. */
  runningControllers: Info[];
  /** RunningDirectives is the list of running directives. */
  runningDirectives: DirectiveState[];
}

function createBaseConfig(): Config {
  return { enableExecController: false };
}

export const Config = {
  encode(message: Config, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.enableExecController === true) {
      writer.uint32(8).bool(message.enableExecController);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Config {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseConfig();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.enableExecController = reader.bool();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  // encodeTransform encodes a source of message objects.
  // Transform<Config, Uint8Array>
  async *encodeTransform(
    source: AsyncIterable<Config | Config[]> | Iterable<Config | Config[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Config.encode(p).finish()];
        }
      } else {
        yield* [Config.encode(pkt).finish()];
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Config>
  async *decodeTransform(
    source: AsyncIterable<Uint8Array | Uint8Array[]> | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<Config> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Config.decode(p)];
        }
      } else {
        yield* [Config.decode(pkt)];
      }
    }
  },

  fromJSON(object: any): Config {
    return { enableExecController: isSet(object.enableExecController) ? Boolean(object.enableExecController) : false };
  },

  toJSON(message: Config): unknown {
    const obj: any = {};
    message.enableExecController !== undefined && (obj.enableExecController = message.enableExecController);
    return obj;
  },

  create<I extends Exact<DeepPartial<Config>, I>>(base?: I): Config {
    return Config.fromPartial(base ?? {});
  },

  fromPartial<I extends Exact<DeepPartial<Config>, I>>(object: I): Config {
    const message = createBaseConfig();
    message.enableExecController = object.enableExecController ?? false;
    return message;
  },
};

function createBaseGetBusInfoRequest(): GetBusInfoRequest {
  return {};
}

export const GetBusInfoRequest = {
  encode(_: GetBusInfoRequest, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetBusInfoRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseGetBusInfoRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  // encodeTransform encodes a source of message objects.
  // Transform<GetBusInfoRequest, Uint8Array>
  async *encodeTransform(
    source: AsyncIterable<GetBusInfoRequest | GetBusInfoRequest[]> | Iterable<GetBusInfoRequest | GetBusInfoRequest[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [GetBusInfoRequest.encode(p).finish()];
        }
      } else {
        yield* [GetBusInfoRequest.encode(pkt).finish()];
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, GetBusInfoRequest>
  async *decodeTransform(
    source: AsyncIterable<Uint8Array | Uint8Array[]> | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<GetBusInfoRequest> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [GetBusInfoRequest.decode(p)];
        }
      } else {
        yield* [GetBusInfoRequest.decode(pkt)];
      }
    }
  },

  fromJSON(_: any): GetBusInfoRequest {
    return {};
  },

  toJSON(_: GetBusInfoRequest): unknown {
    const obj: any = {};
    return obj;
  },

  create<I extends Exact<DeepPartial<GetBusInfoRequest>, I>>(base?: I): GetBusInfoRequest {
    return GetBusInfoRequest.fromPartial(base ?? {});
  },

  fromPartial<I extends Exact<DeepPartial<GetBusInfoRequest>, I>>(_: I): GetBusInfoRequest {
    const message = createBaseGetBusInfoRequest();
    return message;
  },
};

function createBaseGetBusInfoResponse(): GetBusInfoResponse {
  return { runningControllers: [], runningDirectives: [] };
}

export const GetBusInfoResponse = {
  encode(message: GetBusInfoResponse, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    for (const v of message.runningControllers) {
      Info.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    for (const v of message.runningDirectives) {
      DirectiveState.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetBusInfoResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseGetBusInfoResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.runningControllers.push(Info.decode(reader, reader.uint32()));
          break;
        case 2:
          message.runningDirectives.push(DirectiveState.decode(reader, reader.uint32()));
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  // encodeTransform encodes a source of message objects.
  // Transform<GetBusInfoResponse, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<GetBusInfoResponse | GetBusInfoResponse[]>
      | Iterable<GetBusInfoResponse | GetBusInfoResponse[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [GetBusInfoResponse.encode(p).finish()];
        }
      } else {
        yield* [GetBusInfoResponse.encode(pkt).finish()];
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, GetBusInfoResponse>
  async *decodeTransform(
    source: AsyncIterable<Uint8Array | Uint8Array[]> | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<GetBusInfoResponse> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [GetBusInfoResponse.decode(p)];
        }
      } else {
        yield* [GetBusInfoResponse.decode(pkt)];
      }
    }
  },

  fromJSON(object: any): GetBusInfoResponse {
    return {
      runningControllers: Array.isArray(object?.runningControllers)
        ? object.runningControllers.map((e: any) => Info.fromJSON(e))
        : [],
      runningDirectives: Array.isArray(object?.runningDirectives)
        ? object.runningDirectives.map((e: any) => DirectiveState.fromJSON(e))
        : [],
    };
  },

  toJSON(message: GetBusInfoResponse): unknown {
    const obj: any = {};
    if (message.runningControllers) {
      obj.runningControllers = message.runningControllers.map((e) => e ? Info.toJSON(e) : undefined);
    } else {
      obj.runningControllers = [];
    }
    if (message.runningDirectives) {
      obj.runningDirectives = message.runningDirectives.map((e) => e ? DirectiveState.toJSON(e) : undefined);
    } else {
      obj.runningDirectives = [];
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<GetBusInfoResponse>, I>>(base?: I): GetBusInfoResponse {
    return GetBusInfoResponse.fromPartial(base ?? {});
  },

  fromPartial<I extends Exact<DeepPartial<GetBusInfoResponse>, I>>(object: I): GetBusInfoResponse {
    const message = createBaseGetBusInfoResponse();
    message.runningControllers = object.runningControllers?.map((e) => Info.fromPartial(e)) || [];
    message.runningDirectives = object.runningDirectives?.map((e) => DirectiveState.fromPartial(e)) || [];
    return message;
  },
};

/** ControllerBusService is a generic controller bus lookup api. */
export interface ControllerBusService {
  /** GetBusInfo requests information about the controller bus. */
  GetBusInfo(request: GetBusInfoRequest): Promise<GetBusInfoResponse>;
  /** ExecController executes a controller configuration on the bus. */
  ExecController(request: ExecControllerRequest): AsyncIterable<ExecControllerResponse>;
}

export class ControllerBusServiceClientImpl implements ControllerBusService {
  private readonly rpc: Rpc;
  private readonly service: string;
  constructor(rpc: Rpc, opts?: { service?: string }) {
    this.service = opts?.service || "bus.api.ControllerBusService";
    this.rpc = rpc;
    this.GetBusInfo = this.GetBusInfo.bind(this);
    this.ExecController = this.ExecController.bind(this);
  }
  GetBusInfo(request: GetBusInfoRequest): Promise<GetBusInfoResponse> {
    const data = GetBusInfoRequest.encode(request).finish();
    const promise = this.rpc.request(this.service, "GetBusInfo", data);
    return promise.then((data) => GetBusInfoResponse.decode(new _m0.Reader(data)));
  }

  ExecController(request: ExecControllerRequest): AsyncIterable<ExecControllerResponse> {
    const data = ExecControllerRequest.encode(request).finish();
    const result = this.rpc.serverStreamingRequest(this.service, "ExecController", data);
    return ExecControllerResponse.decodeTransform(result);
  }
}

/** ControllerBusService is a generic controller bus lookup api. */
export type ControllerBusServiceDefinition = typeof ControllerBusServiceDefinition;
export const ControllerBusServiceDefinition = {
  name: "ControllerBusService",
  fullName: "bus.api.ControllerBusService",
  methods: {
    /** GetBusInfo requests information about the controller bus. */
    getBusInfo: {
      name: "GetBusInfo",
      requestType: GetBusInfoRequest,
      requestStream: false,
      responseType: GetBusInfoResponse,
      responseStream: false,
      options: {},
    },
    /** ExecController executes a controller configuration on the bus. */
    execController: {
      name: "ExecController",
      requestType: ExecControllerRequest,
      requestStream: false,
      responseType: ExecControllerResponse,
      responseStream: true,
      options: {},
    },
  },
} as const;

interface Rpc {
  request(service: string, method: string, data: Uint8Array): Promise<Uint8Array>;
  clientStreamingRequest(service: string, method: string, data: AsyncIterable<Uint8Array>): Promise<Uint8Array>;
  serverStreamingRequest(service: string, method: string, data: Uint8Array): AsyncIterable<Uint8Array>;
  bidirectionalStreamingRequest(
    service: string,
    method: string,
    data: AsyncIterable<Uint8Array>,
  ): AsyncIterable<Uint8Array>;
}

type Builtin = Date | Function | Uint8Array | string | number | boolean | undefined;

export type DeepPartial<T> = T extends Builtin ? T
  : T extends Long ? string | number | Long : T extends Array<infer U> ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>>
  : T extends { $case: string } ? { [K in keyof Omit<T, "$case">]?: DeepPartial<T[K]> } & { $case: T["$case"] }
  : T extends {} ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;

type KeysOfUnion<T> = T extends T ? keyof T : never;
export type Exact<P, I extends P> = P extends Builtin ? P
  : P & { [K in keyof P]: Exact<P[K], I[K]> } & { [K in Exclude<keyof I, KeysOfUnion<P>>]: never };

if (_m0.util.Long !== Long) {
  _m0.util.Long = Long as any;
  _m0.configure();
}

function isSet(value: any): boolean {
  return value !== null && value !== undefined;
}
