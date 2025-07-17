import * as $ from "@goscript/builtin/index.js";

export type BinaryAppender = null | {
	// AppendBinary appends the binary representation of itself to the end of b
	// (allocating a larger slice if necessary) and returns the updated slice.
	//
	// Implementations must not retain b, nor mutate any bytes within b[:len(b)].
	AppendBinary(b: $.Bytes): [$.Bytes, $.GoError]
}

$.registerInterfaceType(
  'BinaryAppender',
  null, // Zero value for interface is null
  [{ name: "AppendBinary", args: [{ name: "b", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }], returns: [{ type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }]
);

export type BinaryMarshaler = null | {
	MarshalBinary(): [$.Bytes, $.GoError]
}

$.registerInterfaceType(
  'BinaryMarshaler',
  null, // Zero value for interface is null
  [{ name: "MarshalBinary", args: [], returns: [{ type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }]
);

export type BinaryUnmarshaler = null | {
	UnmarshalBinary(data: $.Bytes): Promise<$.GoError>
}

$.registerInterfaceType(
  'BinaryUnmarshaler',
  null, // Zero value for interface is null
  [{ name: "UnmarshalBinary", args: [{ name: "data", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }], returns: [{ type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }]
);

export type TextAppender = null | {
	// AppendText appends the textual representation of itself to the end of b
	// (allocating a larger slice if necessary) and returns the updated slice.
	//
	// Implementations must not retain b, nor mutate any bytes within b[:len(b)].
	AppendText(b: $.Bytes): [$.Bytes, $.GoError]
}

$.registerInterfaceType(
  'TextAppender',
  null, // Zero value for interface is null
  [{ name: "AppendText", args: [{ name: "b", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }], returns: [{ type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }]
);

export type TextMarshaler = null | {
	MarshalText(): [$.Bytes, $.GoError]
}

$.registerInterfaceType(
  'TextMarshaler',
  null, // Zero value for interface is null
  [{ name: "MarshalText", args: [], returns: [{ type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }]
);

export type TextUnmarshaler = null | {
	UnmarshalText(text: $.Bytes): $.GoError
}

$.registerInterfaceType(
  'TextUnmarshaler',
  null, // Zero value for interface is null
  [{ name: "UnmarshalText", args: [{ name: "text", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }], returns: [{ type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }]
);

