import * as $ from "@goscript/builtin/index.js";
import { nativeEndian } from "./native_endian_little.gs.js";

import * as errors from "@goscript/errors/index.js"

import * as io from "@goscript/io/index.js"

import * as math from "@goscript/math/index.js"

import * as reflect from "@goscript/reflect/index.js"

import * as slices from "@goscript/slices/index.js"

import * as sync from "@goscript/sync/index.js"

export type AppendByteOrder = null | {
	AppendUint16(_p0: $.Bytes, _p1: number): $.Bytes
	AppendUint32(_p0: $.Bytes, _p1: number): $.Bytes
	AppendUint64(_p0: $.Bytes, _p1: number): $.Bytes
	String(): string
}

$.registerInterfaceType(
  'AppendByteOrder',
  null, // Zero value for interface is null
  [{ name: "AppendUint16", args: [{ name: "", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { name: "", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [{ type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }] }, { name: "AppendUint32", args: [{ name: "", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { name: "", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [{ type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }] }, { name: "AppendUint64", args: [{ name: "", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { name: "", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [{ type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }] }, { name: "String", args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: "string" } }] }]
);

export type ByteOrder = null | {
	PutUint16(_p0: $.Bytes, _p1: number): void
	PutUint32(_p0: $.Bytes, _p1: number): void
	PutUint64(_p0: $.Bytes, _p1: number): void
	String(): string
	Uint16(_p0: $.Bytes): number
	Uint32(_p0: $.Bytes): number
	Uint64(_p0: $.Bytes): number
}

$.registerInterfaceType(
  'ByteOrder',
  null, // Zero value for interface is null
  [{ name: "PutUint16", args: [{ name: "", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { name: "", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [] }, { name: "PutUint32", args: [{ name: "", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { name: "", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [] }, { name: "PutUint64", args: [{ name: "", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { name: "", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [] }, { name: "String", args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: "string" } }] }, { name: "Uint16", args: [{ name: "", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }] }, { name: "Uint32", args: [{ name: "", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }] }, { name: "Uint64", args: [{ name: "", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }] }]
);

export class bigEndian {
	public _fields: {
	}

	constructor(init?: Partial<{}>) {
		this._fields = {}
	}

	public clone(): bigEndian {
		const cloned = new bigEndian()
		cloned._fields = {
		}
		return cloned
	}

	// Uint16 returns the uint16 representation of b[0:2].
	public Uint16(b: $.Bytes): number {
		/* _ = */ b![1] // bounds check hint to compiler; see golang.org/issue/14808
		return ((b![1] as number) | ((b![0] as number) << 8))
	}

	// PutUint16 stores v into b[0:2].
	public PutUint16(b: $.Bytes, v: number): void {
		/* _ = */ b![1] // early bounds check to guarantee safety of writes below
		b![0] = $.byte((v >> 8))
		b![1] = $.byte(v)
	}

	// AppendUint16 appends the bytes of v to b and returns the appended slice.
	public AppendUint16(b: $.Bytes, v: number): $.Bytes {
		return $.append(b, $.byte((v >> 8)), $.byte(v))
	}

	// Uint32 returns the uint32 representation of b[0:4].
	public Uint32(b: $.Bytes): number {
		/* _ = */ b![3] // bounds check hint to compiler; see golang.org/issue/14808
		return ((((b![3] as number) | ((b![2] as number) << 8)) | ((b![1] as number) << 16)) | ((b![0] as number) << 24))
	}

	// PutUint32 stores v into b[0:4].
	public PutUint32(b: $.Bytes, v: number): void {
		/* _ = */ b![3] // early bounds check to guarantee safety of writes below
		b![0] = $.byte((v >> 24))
		b![1] = $.byte((v >> 16))
		b![2] = $.byte((v >> 8))
		b![3] = $.byte(v)
	}

	// AppendUint32 appends the bytes of v to b and returns the appended slice.
	public AppendUint32(b: $.Bytes, v: number): $.Bytes {
		return $.append(b, $.byte((v >> 24)), $.byte((v >> 16)), $.byte((v >> 8)), $.byte(v))
	}

	// Uint64 returns the uint64 representation of b[0:8].
	public Uint64(b: $.Bytes): number {
		/* _ = */ b![7] // bounds check hint to compiler; see golang.org/issue/14808
		return ((((((((b![7] as number) | ((b![6] as number) << 8)) | ((b![5] as number) << 16)) | ((b![4] as number) << 24)) | ((b![3] as number) << 32)) | ((b![2] as number) << 40)) | ((b![1] as number) << 48)) | ((b![0] as number) << 56))
	}

	// PutUint64 stores v into b[0:8].
	public PutUint64(b: $.Bytes, v: number): void {
		/* _ = */ b![7] // early bounds check to guarantee safety of writes below
		b![0] = $.byte((v >> 56))
		b![1] = $.byte((v >> 48))
		b![2] = $.byte((v >> 40))
		b![3] = $.byte((v >> 32))
		b![4] = $.byte((v >> 24))
		b![5] = $.byte((v >> 16))
		b![6] = $.byte((v >> 8))
		b![7] = $.byte(v)
	}

	// AppendUint64 appends the bytes of v to b and returns the appended slice.
	public AppendUint64(b: $.Bytes, v: number): $.Bytes {
		return $.append(b, $.byte((v >> 56)), $.byte((v >> 48)), $.byte((v >> 40)), $.byte((v >> 32)), $.byte((v >> 24)), $.byte((v >> 16)), $.byte((v >> 8)), $.byte(v))
	}

	public String(): string {
		return "BigEndian"
	}

	public GoString(): string {
		return "binary.BigEndian"
	}

	// Register this type with the runtime type system
	static __typeInfo = $.registerStructType(
	  'bigEndian',
	  new bigEndian(),
	  [{ name: "Uint16", args: [{ name: "b", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }] }, { name: "PutUint16", args: [{ name: "b", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { name: "v", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [] }, { name: "AppendUint16", args: [{ name: "b", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { name: "v", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [{ type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }] }, { name: "Uint32", args: [{ name: "b", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }] }, { name: "PutUint32", args: [{ name: "b", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { name: "v", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [] }, { name: "AppendUint32", args: [{ name: "b", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { name: "v", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [{ type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }] }, { name: "Uint64", args: [{ name: "b", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }] }, { name: "PutUint64", args: [{ name: "b", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { name: "v", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [] }, { name: "AppendUint64", args: [{ name: "b", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { name: "v", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [{ type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }] }, { name: "String", args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: "string" } }] }, { name: "GoString", args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: "string" } }] }],
	  bigEndian,
	  {}
	);
}

export class littleEndian {
	public _fields: {
	}

	constructor(init?: Partial<{}>) {
		this._fields = {}
	}

	public clone(): littleEndian {
		const cloned = new littleEndian()
		cloned._fields = {
		}
		return cloned
	}

	// Uint16 returns the uint16 representation of b[0:2].
	public Uint16(b: $.Bytes): number {
		/* _ = */ b![1] // bounds check hint to compiler; see golang.org/issue/14808
		return ((b![0] as number) | ((b![1] as number) << 8))
	}

	// PutUint16 stores v into b[0:2].
	public PutUint16(b: $.Bytes, v: number): void {
		/* _ = */ b![1] // early bounds check to guarantee safety of writes below
		b![0] = $.byte(v)
		b![1] = $.byte((v >> 8))
	}

	// AppendUint16 appends the bytes of v to b and returns the appended slice.
	public AppendUint16(b: $.Bytes, v: number): $.Bytes {
		return $.append(b, $.byte(v), $.byte((v >> 8)))
	}

	// Uint32 returns the uint32 representation of b[0:4].
	public Uint32(b: $.Bytes): number {
		/* _ = */ b![3] // bounds check hint to compiler; see golang.org/issue/14808
		return ((((b![0] as number) | ((b![1] as number) << 8)) | ((b![2] as number) << 16)) | ((b![3] as number) << 24))
	}

	// PutUint32 stores v into b[0:4].
	public PutUint32(b: $.Bytes, v: number): void {
		/* _ = */ b![3] // early bounds check to guarantee safety of writes below
		b![0] = $.byte(v)
		b![1] = $.byte((v >> 8))
		b![2] = $.byte((v >> 16))
		b![3] = $.byte((v >> 24))
	}

	// AppendUint32 appends the bytes of v to b and returns the appended slice.
	public AppendUint32(b: $.Bytes, v: number): $.Bytes {
		return $.append(b, $.byte(v), $.byte((v >> 8)), $.byte((v >> 16)), $.byte((v >> 24)))
	}

	// Uint64 returns the uint64 representation of b[0:8].
	public Uint64(b: $.Bytes): number {
		/* _ = */ b![7] // bounds check hint to compiler; see golang.org/issue/14808
		return ((((((((b![0] as number) | ((b![1] as number) << 8)) | ((b![2] as number) << 16)) | ((b![3] as number) << 24)) | ((b![4] as number) << 32)) | ((b![5] as number) << 40)) | ((b![6] as number) << 48)) | ((b![7] as number) << 56))
	}

	// PutUint64 stores v into b[0:8].
	public PutUint64(b: $.Bytes, v: number): void {
		/* _ = */ b![7] // early bounds check to guarantee safety of writes below
		b![0] = $.byte(v)
		b![1] = $.byte((v >> 8))
		b![2] = $.byte((v >> 16))
		b![3] = $.byte((v >> 24))
		b![4] = $.byte((v >> 32))
		b![5] = $.byte((v >> 40))
		b![6] = $.byte((v >> 48))
		b![7] = $.byte((v >> 56))
	}

	// AppendUint64 appends the bytes of v to b and returns the appended slice.
	public AppendUint64(b: $.Bytes, v: number): $.Bytes {
		return $.append(b, $.byte(v), $.byte((v >> 8)), $.byte((v >> 16)), $.byte((v >> 24)), $.byte((v >> 32)), $.byte((v >> 40)), $.byte((v >> 48)), $.byte((v >> 56)))
	}

	public String(): string {
		return "LittleEndian"
	}

	public GoString(): string {
		return "binary.LittleEndian"
	}

	// Register this type with the runtime type system
	static __typeInfo = $.registerStructType(
	  'littleEndian',
	  new littleEndian(),
	  [{ name: "Uint16", args: [{ name: "b", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }] }, { name: "PutUint16", args: [{ name: "b", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { name: "v", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [] }, { name: "AppendUint16", args: [{ name: "b", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { name: "v", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [{ type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }] }, { name: "Uint32", args: [{ name: "b", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }] }, { name: "PutUint32", args: [{ name: "b", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { name: "v", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [] }, { name: "AppendUint32", args: [{ name: "b", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { name: "v", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [{ type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }] }, { name: "Uint64", args: [{ name: "b", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }] }, { name: "PutUint64", args: [{ name: "b", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { name: "v", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [] }, { name: "AppendUint64", args: [{ name: "b", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { name: "v", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [{ type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }] }, { name: "String", args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: "string" } }] }, { name: "GoString", args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: "string" } }] }],
	  littleEndian,
	  {}
	);
}

export class coder {
	public get order(): ByteOrder {
		return this._fields.order.value
	}
	public set order(value: ByteOrder) {
		this._fields.order.value = value
	}

	public get buf(): $.Bytes {
		return this._fields.buf.value
	}
	public set buf(value: $.Bytes) {
		this._fields.buf.value = value
	}

	public get offset(): number {
		return this._fields.offset.value
	}
	public set offset(value: number) {
		this._fields.offset.value = value
	}

	public _fields: {
		order: $.VarRef<ByteOrder>;
		buf: $.VarRef<$.Bytes>;
		offset: $.VarRef<number>;
	}

	constructor(init?: Partial<{buf?: $.Bytes, offset?: number, order?: ByteOrder}>) {
		this._fields = {
			order: $.varRef(init?.order ?? null),
			buf: $.varRef(init?.buf ?? new Uint8Array(0)),
			offset: $.varRef(init?.offset ?? 0)
		}
	}

	public clone(): coder {
		const cloned = new coder()
		cloned._fields = {
			order: $.varRef(this._fields.order.value),
			buf: $.varRef(this._fields.buf.value),
			offset: $.varRef(this._fields.offset.value)
		}
		return cloned
	}

	// Register this type with the runtime type system
	static __typeInfo = $.registerStructType(
	  'coder',
	  new coder(),
	  [],
	  coder,
	  {"order": "ByteOrder", "buf": { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } }, "offset": { kind: $.TypeKind.Basic, name: "number" }}
	);
}

export type decoder = coder;

export function decoder_bool(d: $.VarRef<decoder>): boolean {
	let x = d.buf![d.offset]
	d.offset++
	return x != 0
}

export function decoder_uint8(d: $.VarRef<decoder>): number {
	let x = d.buf![d.offset]
	d.offset++
	return x
}

export function decoder_uint16(d: $.VarRef<decoder>): number {
	let x = d.order!.Uint16($.goSlice(d.buf, d.offset, d.offset + 2))
	d.offset += 2
	return x
}

export function decoder_uint32(d: $.VarRef<decoder>): number {
	let x = d.order!.Uint32($.goSlice(d.buf, d.offset, d.offset + 4))
	d.offset += 4
	return x
}

export function decoder_uint64(d: $.VarRef<decoder>): number {
	let x = d.order!.Uint64($.goSlice(d.buf, d.offset, d.offset + 8))
	d.offset += 8
	return x
}

export function decoder_int8(d: $.VarRef<decoder>): number {
	return (d.uint8() as number)
}

export function decoder_int16(d: $.VarRef<decoder>): number {
	return (d.uint16() as number)
}

export function decoder_int32(d: $.VarRef<decoder>): number {
	return (d.uint32() as number)
}

export function decoder_int64(d: $.VarRef<decoder>): number {
	return (d.uint64() as number)
}

export function decoder_value(d: $.VarRef<decoder>, v: reflect.Value): void {
	switch (v.Kind()) {
		case reflect.Array:
			let l = v.Len()
			for (let i = 0; i < l; i++) {
				d.value(v.Index(i))
			}
			break
		case reflect.Struct:
			let t = v.Type()
			let l = v.NumField()
			for (let i = 0; i < l; i++) {
				// Note: Calling v.CanSet() below is an optimization.
				// It would be sufficient to check the field name,
				// but creating the StructField info for each field is
				// costly (run "go test -bench=ReadStruct" and compare
				// results when making changes to this code).
				const _temp_v = v
				{
					let v = _temp_v.Field(i)
					if (v.CanSet() || t!.Field(i)!.Name != "_") {
						d.value(v)
					}
					 else {
						await d.skip(v)
					}
				}
			}
			break
		case reflect.Slice:
			let l = v.Len()
			for (let i = 0; i < l; i++) {
				d.value(v.Index(i))
			}
			break
		case reflect.Bool:
			v.SetBool(d.bool())
			break
		case reflect.Int8:
			v.SetInt((d.int8() as number))
			break
		case reflect.Int16:
			v.SetInt((d.int16() as number))
			break
		case reflect.Int32:
			v.SetInt((d.int32() as number))
			break
		case reflect.Int64:
			v.SetInt(d.int64())
			break
		case reflect.Uint8:
			v.SetUint((d.uint8() as number))
			break
		case reflect.Uint16:
			v.SetUint((d.uint16() as number))
			break
		case reflect.Uint32:
			v.SetUint((d.uint32() as number))
			break
		case reflect.Uint64:
			v.SetUint(d.uint64())
			break
		case reflect.Float32:
			v.SetFloat((math.Float32frombits(d.uint32()) as number))
			break
		case reflect.Float64:
			v.SetFloat(math.Float64frombits(d.uint64()))
			break
		case reflect.Complex64:
			v.SetComplex(complex((math.Float32frombits(d.uint32()) as number), (math.Float32frombits(d.uint32()) as number)))
			break
		case reflect.Complex128:
			v.SetComplex(complex(math.Float64frombits(d.uint64()), math.Float64frombits(d.uint64())))
			break
	}
}

export function decoder_skip(d: $.VarRef<decoder>, v: reflect.Value): void {
	d.offset += await dataSize(v)
}


export type encoder = coder;

export function encoder_bool(e: $.VarRef<encoder>, x: boolean): void {
	if (x) {
		e.buf![e.offset] = 1
	}
	 else {
		e.buf![e.offset] = 0
	}
	e.offset++
}

export function encoder_uint8(e: $.VarRef<encoder>, x: number): void {
	e.buf![e.offset] = x
	e.offset++
}

export function encoder_uint16(e: $.VarRef<encoder>, x: number): void {
	e.order!.PutUint16($.goSlice(e.buf, e.offset, e.offset + 2), x)
	e.offset += 2
}

export function encoder_uint32(e: $.VarRef<encoder>, x: number): void {
	e.order!.PutUint32($.goSlice(e.buf, e.offset, e.offset + 4), x)
	e.offset += 4
}

export function encoder_uint64(e: $.VarRef<encoder>, x: number): void {
	e.order!.PutUint64($.goSlice(e.buf, e.offset, e.offset + 8), x)
	e.offset += 8
}

export function encoder_int8(e: $.VarRef<encoder>, x: number): void {
	e.uint8((x as number))
}

export function encoder_int16(e: $.VarRef<encoder>, x: number): void {
	e.uint16((x as number))
}

export function encoder_int32(e: $.VarRef<encoder>, x: number): void {
	e.uint32((x as number))
}

export function encoder_int64(e: $.VarRef<encoder>, x: number): void {
	e.uint64((x as number))
}

export function encoder_value(e: $.VarRef<encoder>, v: reflect.Value): void {
	switch (v.Kind()) {
		case reflect.Array:
			let l = v.Len()
			for (let i = 0; i < l; i++) {
				e.value(v.Index(i))
			}
			break
		case reflect.Struct:
			let t = v.Type()
			let l = v.NumField()
			for (let i = 0; i < l; i++) {
				// see comment for corresponding code in decoder.value()
				const _temp_v = v
				{
					let v = _temp_v.Field(i)
					if (v.CanSet() || t!.Field(i)!.Name != "_") {
						e.value(v)
					}
					 else {
						await e.skip(v)
					}
				}
			}
			break
		case reflect.Slice:
			let l = v.Len()
			for (let i = 0; i < l; i++) {
				e.value(v.Index(i))
			}
			break
		case reflect.Bool:
			e.bool(v.Bool())
			break
		case reflect.Int8:
			e.int8((v.Int() as number))
			break
		case reflect.Int16:
			e.int16((v.Int() as number))
			break
		case reflect.Int32:
			e.int32((v.Int() as number))
			break
		case reflect.Int64:
			e.int64(v.Int())
			break
		case reflect.Uint8:
			e.uint8((v.Uint() as number))
			break
		case reflect.Uint16:
			e.uint16((v.Uint() as number))
			break
		case reflect.Uint32:
			e.uint32((v.Uint() as number))
			break
		case reflect.Uint64:
			e.uint64(v.Uint())
			break
		case reflect.Float32:
			e.uint32(math.Float32bits((v.Float() as number)))
			break
		case reflect.Float64:
			e.uint64(math.Float64bits(v.Float()))
			break
		case reflect.Complex64:
			let x = v.Complex()
			e.uint32(math.Float32bits((real(x) as number)))
			e.uint32(math.Float32bits((imag(x) as number)))
			break
		case reflect.Complex128:
			let x = v.Complex()
			e.uint64(math.Float64bits(real(x)))
			e.uint64(math.Float64bits(imag(x)))
			break
	}
}

export function encoder_skip(e: $.VarRef<encoder>, v: reflect.Value): void {
	let n = await dataSize(v)
	clear($.goSlice(e.buf, e.offset, e.offset + n))
	e.offset += n
}


let errBufferTooSmall: $.GoError = errors.New("buffer too small")

// map[reflect.Type]int
let structSize: sync.Map = new sync.Map()

export let BigEndian: bigEndian = new bigEndian({})

export let LittleEndian: littleEndian = new littleEndian({})

// Read reads structured binary data from r into data.
// Data must be a pointer to a fixed-size value or a slice
// of fixed-size values.
// Bytes read from r are decoded using the specified byte order
// and written to successive fields of the data.
// When decoding boolean values, a zero byte is decoded as false, and
// any other non-zero byte is decoded as true.
// When reading into structs, the field data for fields with
// blank (_) field names is skipped; i.e., blank field names
// may be used for padding.
// When reading into a struct, all non-blank fields must be exported
// or Read may panic.
//
// The error is [io.EOF] only if no bytes were read.
// If an [io.EOF] happens after reading some but not all the bytes,
// Read returns [io.ErrUnexpectedEOF].
export async function Read(r: io.Reader, order: ByteOrder, data: null | any): Promise<$.GoError> {
	// Fast path for basic types and slices.
	{
		let [n, ] = intDataSize(data)
		if (n != 0) {
			let bs = new Uint8Array(n)
			{
				let [, err] = io.ReadFull(r, bs)
				if (err != null) {
					return err
				}
			}

			if (decodeFast(bs, order, data)) {
				return null
			}
		}
	}

	// Fallback to reflect-based decoding.
	let v = $.markAsStructValue(reflect.ValueOf(data).clone())
	let size = -1
	switch (v.Kind()) {
		case reflect.Pointer:
			v = $.markAsStructValue(v.Elem().clone())
			size = await dataSize(v)
			break
		case reflect.Slice:
			size = await dataSize(v)
			break
	}
	if (size < 0) {
		return errors.New("binary.Read: invalid type " + reflect.TypeOf(data)!.String())
	}

	let d = new decoder({buf: new Uint8Array(size), order: order})
	{
		let [, err] = io.ReadFull(r, d.buf)
		if (err != null) {
			return err
		}
	}
	d.value(v)
	return null
}

// Decode decodes binary data from buf into data according to
// the given byte order.
// It returns an error if buf is too small, otherwise the number of
// bytes consumed from buf.
export async function Decode(buf: $.Bytes, order: ByteOrder, data: null | any): Promise<[number, $.GoError]> {
	{
		let [n, ] = intDataSize(data)
		if (n != 0) {
			if ($.len(buf) < n) {
				return [0, errBufferTooSmall]
			}

			if (decodeFast(buf, order, data)) {
				return [n, null]
			}
		}
	}

	// Fallback to reflect-based decoding.
	let v = $.markAsStructValue(reflect.ValueOf(data).clone())
	let size = -1
	switch (v.Kind()) {
		case reflect.Pointer:
			v = $.markAsStructValue(v.Elem().clone())
			size = await dataSize(v)
			break
		case reflect.Slice:
			size = await dataSize(v)
			break
	}
	if (size < 0) {
		return [0, errors.New("binary.Decode: invalid type " + reflect.TypeOf(data)!.String())]
	}

	if ($.len(buf) < size) {
		return [0, errBufferTooSmall]
	}
	let d = new decoder({buf: $.goSlice(buf, undefined, size), order: order})
	d.value(v)
	return [size, null]
}

export function decodeFast(bs: $.Bytes, order: ByteOrder, data: null | any): boolean {

	// Easier to loop over the input for 8-bit values.
	$.typeSwitch(data, [{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'boolean'}}], body: (data) => {
		data!.value = bs![0] != 0
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		data!.value = (bs![0] as number)
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		data!.value = bs![0]
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		data!.value = (order!.Uint16(bs) as number)
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		data!.value = order!.Uint16(bs)
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		data!.value = (order!.Uint32(bs) as number)
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		data!.value = order!.Uint32(bs)
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		data!.value = (order!.Uint64(bs) as number)
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		data!.value = order!.Uint64(bs)
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		data!.value = math.Float32frombits(order!.Uint32(bs))
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		data!.value = math.Float64frombits(order!.Uint64(bs))
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'boolean'}}], body: (data) => {
		for (let i = 0; i < $.len(bs); i++) {
			const x = bs![i]
			{
				// Easier to loop over the input for 8-bit values.
				data![i] = x != 0
			}
		}
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		for (let i = 0; i < $.len(bs); i++) {
			const x = bs![i]
			{
				data![i] = (x as number)
			}
		}
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		$.copy(data, bs)
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		for (let i = 0; i < $.len(data); i++) {
			{
				data![i] = (order!.Uint16($.goSlice(bs, 2 * i, undefined)) as number)
			}
		}
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		for (let i = 0; i < $.len(data); i++) {
			{
				data![i] = order!.Uint16($.goSlice(bs, 2 * i, undefined))
			}
		}
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		for (let i = 0; i < $.len(data); i++) {
			{
				data![i] = (order!.Uint32($.goSlice(bs, 4 * i, undefined)) as number)
			}
		}
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		for (let i = 0; i < $.len(data); i++) {
			{
				data![i] = order!.Uint32($.goSlice(bs, 4 * i, undefined))
			}
		}
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		for (let i = 0; i < $.len(data); i++) {
			{
				data![i] = (order!.Uint64($.goSlice(bs, 8 * i, undefined)) as number)
			}
		}
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		for (let i = 0; i < $.len(data); i++) {
			{
				data![i] = order!.Uint64($.goSlice(bs, 8 * i, undefined))
			}
		}
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		for (let i = 0; i < $.len(data); i++) {
			{
				data![i] = math.Float32frombits(order!.Uint32($.goSlice(bs, 4 * i, undefined)))
			}
		}
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		for (let i = 0; i < $.len(data); i++) {
			{
				data![i] = math.Float64frombits(order!.Uint64($.goSlice(bs, 8 * i, undefined)))
			}
		}
	}}], () => {
		return false
	})
	return true
}

// Write writes the binary representation of data into w.
// Data must be a fixed-size value or a slice of fixed-size
// values, or a pointer to such data.
// Boolean values encode as one byte: 1 for true, and 0 for false.
// Bytes written to w are encoded using the specified byte order
// and read from successive fields of the data.
// When writing structs, zero values are written for fields
// with blank (_) field names.
export async function Write(w: io.Writer, order: ByteOrder, data: null | any): Promise<$.GoError> {
	// Fast path for basic types and slices.
	{
		let [n, bs] = intDataSize(data)
		if (n != 0) {
			if (bs == null) {
				bs = new Uint8Array(n)
				encodeFast(bs, order, data)
			}

			let [, err] = w!.Write(bs)
			return err
		}
	}

	// Fallback to reflect-based encoding.
	let v = $.markAsStructValue(reflect.Indirect(reflect.ValueOf(data)).clone())
	let size = await dataSize(v)
	if (size < 0) {
		return errors.New("binary.Write: some values are not fixed-sized in type " + reflect.TypeOf(data)!.String())
	}

	let buf = new Uint8Array(size)
	let e = new encoder({buf: buf, order: order})
	e.value(v)
	let [, err] = w!.Write(buf)
	return err
}

// Encode encodes the binary representation of data into buf according to
// the given byte order.
// It returns an error if buf is too small, otherwise the number of
// bytes written into buf.
export async function Encode(buf: $.Bytes, order: ByteOrder, data: null | any): Promise<[number, $.GoError]> {
	// Fast path for basic types and slices.
	{
		let [n, ] = intDataSize(data)
		if (n != 0) {
			if ($.len(buf) < n) {
				return [0, errBufferTooSmall]
			}

			encodeFast(buf, order, data)
			return [n, null]
		}
	}

	// Fallback to reflect-based encoding.
	let v = $.markAsStructValue(reflect.Indirect(reflect.ValueOf(data)).clone())
	let size = await dataSize(v)
	if (size < 0) {
		return [0, errors.New("binary.Encode: some values are not fixed-sized in type " + reflect.TypeOf(data)!.String())]
	}

	if ($.len(buf) < size) {
		return [0, errBufferTooSmall]
	}
	let e = new encoder({buf: buf, order: order})
	e.value(v)
	return [size, null]
}

// Append appends the binary representation of data to buf.
// buf may be nil, in which case a new buffer will be allocated.
// See [Write] on which data are acceptable.
// It returns the (possibly extended) buffer containing data or an error.
export async function Append(buf: $.Bytes, order: ByteOrder, data: null | any): Promise<[$.Bytes, $.GoError]> {
	// Fast path for basic types and slices.
	{
		let [n, ] = intDataSize(data)
		if (n != 0) {
			let [buf, pos] = ensure(buf, n)
			encodeFast(pos, order, data)
			return [buf, null]
		}
	}

	// Fallback to reflect-based encoding.
	let v = $.markAsStructValue(reflect.Indirect(reflect.ValueOf(data)).clone())
	let size = await dataSize(v)
	if (size < 0) {
		return [null, errors.New("binary.Append: some values are not fixed-sized in type " + reflect.TypeOf(data)!.String())]
	}

	let pos: $.Bytes
	[buf, pos] = ensure(buf, size)
	let e = new encoder({buf: pos, order: order})
	e.value(v)
	return [buf, null]
}

export function encodeFast(bs: $.Bytes, order: ByteOrder, data: null | any): void {
	$.typeSwitch(data, [{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'boolean'}}], body: (v) => {
		if (v!.value) {
			bs![0] = 1
		}
		 else {
			bs![0] = 0
		}
	}},
	{ types: [{kind: $.TypeKind.Basic, name: 'boolean'}], body: (v) => {
		if (v) {
			bs![0] = 1
		}
		 else {
			bs![0] = 0
		}
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'boolean'}}], body: (v) => {
		for (let i = 0; i < $.len(v); i++) {
			const x = v![i]
			{
				if (x) {
					bs![i] = 1
				}
				 else {
					bs![i] = 0
				}
			}
		}
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (v) => {
		bs![0] = $.byte(v!.value)
	}},
	{ types: [{kind: $.TypeKind.Basic, name: 'number'}], body: (v) => {
		bs![0] = $.byte(v)
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (v) => {
		for (let i = 0; i < $.len(v); i++) {
			const x = v![i]
			{
				bs![i] = $.byte(x)
			}
		}
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (v) => {
		bs![0] = v!.value
	}},
	{ types: [{kind: $.TypeKind.Basic, name: 'number'}], body: (v) => {
		bs![0] = v
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (v) => {
		$.copy(bs, v)
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (v) => {
		order!.PutUint16(bs, (v!.value as number))
	}},
	{ types: [{kind: $.TypeKind.Basic, name: 'number'}], body: (v) => {
		order!.PutUint16(bs, (v as number))
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (v) => {
		for (let i = 0; i < $.len(v); i++) {
			const x = v![i]
			{
				order!.PutUint16($.goSlice(bs, 2 * i, undefined), (x as number))
			}
		}
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (v) => {
		order!.PutUint16(bs, v!.value)
	}},
	{ types: [{kind: $.TypeKind.Basic, name: 'number'}], body: (v) => {
		order!.PutUint16(bs, v)
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (v) => {
		for (let i = 0; i < $.len(v); i++) {
			const x = v![i]
			{
				order!.PutUint16($.goSlice(bs, 2 * i, undefined), x)
			}
		}
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (v) => {
		order!.PutUint32(bs, (v!.value as number))
	}},
	{ types: [{kind: $.TypeKind.Basic, name: 'number'}], body: (v) => {
		order!.PutUint32(bs, (v as number))
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (v) => {
		for (let i = 0; i < $.len(v); i++) {
			const x = v![i]
			{
				order!.PutUint32($.goSlice(bs, 4 * i, undefined), (x as number))
			}
		}
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (v) => {
		order!.PutUint32(bs, v!.value)
	}},
	{ types: [{kind: $.TypeKind.Basic, name: 'number'}], body: (v) => {
		order!.PutUint32(bs, v)
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (v) => {
		for (let i = 0; i < $.len(v); i++) {
			const x = v![i]
			{
				order!.PutUint32($.goSlice(bs, 4 * i, undefined), x)
			}
		}
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (v) => {
		order!.PutUint64(bs, (v!.value as number))
	}},
	{ types: [{kind: $.TypeKind.Basic, name: 'number'}], body: (v) => {
		order!.PutUint64(bs, (v as number))
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (v) => {
		for (let i = 0; i < $.len(v); i++) {
			const x = v![i]
			{
				order!.PutUint64($.goSlice(bs, 8 * i, undefined), (x as number))
			}
		}
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (v) => {
		order!.PutUint64(bs, v!.value)
	}},
	{ types: [{kind: $.TypeKind.Basic, name: 'number'}], body: (v) => {
		order!.PutUint64(bs, v)
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (v) => {
		for (let i = 0; i < $.len(v); i++) {
			const x = v![i]
			{
				order!.PutUint64($.goSlice(bs, 8 * i, undefined), x)
			}
		}
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (v) => {
		order!.PutUint32(bs, math.Float32bits(v!.value))
	}},
	{ types: [{kind: $.TypeKind.Basic, name: 'number'}], body: (v) => {
		order!.PutUint32(bs, math.Float32bits(v))
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (v) => {
		for (let i = 0; i < $.len(v); i++) {
			const x = v![i]
			{
				order!.PutUint32($.goSlice(bs, 4 * i, undefined), math.Float32bits(x))
			}
		}
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (v) => {
		order!.PutUint64(bs, math.Float64bits(v!.value))
	}},
	{ types: [{kind: $.TypeKind.Basic, name: 'number'}], body: (v) => {
		order!.PutUint64(bs, math.Float64bits(v))
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (v) => {
		for (let i = 0; i < $.len(v); i++) {
			const x = v![i]
			{
				order!.PutUint64($.goSlice(bs, 8 * i, undefined), math.Float64bits(x))
			}
		}
	}}])
}

// Size returns how many bytes [Write] would generate to encode the value v, which
// must be a fixed-size value or a slice of fixed-size values, or a pointer to such data.
// If v is neither of these, Size returns -1.
export async function Size(v: null | any): Promise<number> {
	$.typeSwitch(v, [{ types: [{kind: $.TypeKind.Basic, name: 'boolean'}, {kind: $.TypeKind.Basic, name: 'number'}, {kind: $.TypeKind.Basic, name: 'number'}], body: (data) => {
		return 1
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'boolean'}}], body: (data) => {
		if (data == null) {
			return -1
		}
		return 1
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		if (data == null) {
			return -1
		}
		return 1
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		if (data == null) {
			return -1
		}
		return 1
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'boolean'}}], body: (data) => {
		return $.len(data)
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return $.len(data)
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return $.len(data)
	}},
	{ types: [{kind: $.TypeKind.Basic, name: 'number'}, {kind: $.TypeKind.Basic, name: 'number'}], body: (data) => {
		return 2
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		if (data == null) {
			return -1
		}
		return 2
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		if (data == null) {
			return -1
		}
		return 2
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return 2 * $.len(data)
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return 2 * $.len(data)
	}},
	{ types: [{kind: $.TypeKind.Basic, name: 'number'}, {kind: $.TypeKind.Basic, name: 'number'}], body: (data) => {
		return 4
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		if (data == null) {
			return -1
		}
		return 4
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		if (data == null) {
			return -1
		}
		return 4
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return 4 * $.len(data)
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return 4 * $.len(data)
	}},
	{ types: [{kind: $.TypeKind.Basic, name: 'number'}, {kind: $.TypeKind.Basic, name: 'number'}], body: (data) => {
		return 8
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		if (data == null) {
			return -1
		}
		return 8
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		if (data == null) {
			return -1
		}
		return 8
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return 8 * $.len(data)
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return 8 * $.len(data)
	}},
	{ types: [{kind: $.TypeKind.Basic, name: 'number'}], body: (data) => {
		return 4
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		if (data == null) {
			return -1
		}
		return 4
	}},
	{ types: [{kind: $.TypeKind.Basic, name: 'number'}], body: (data) => {
		return 8
	}},
	{ types: [{kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		if (data == null) {
			return -1
		}
		return 8
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return 4 * $.len(data)
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return 8 * $.len(data)
	}}])
	return await dataSize(reflect.Indirect(reflect.ValueOf(v)))
}

// dataSize returns the number of bytes the actual data represented by v occupies in memory.
// For compound structures, it sums the sizes of the elements. Thus, for instance, for a slice
// it returns the length of the slice times the element size and does not count the memory
// occupied by the header. If the type of v is not acceptable, dataSize returns -1.
export async function dataSize(v: reflect.Value): Promise<number> {
	switch (v.Kind()) {
		case reflect.Slice:
		case reflect.Array:
			let t = v.Type()!.Elem()
			{
				let [size, ok] = await structSize.Load(t)
				if (ok) {
					return $.mustTypeAssert<number>(size, {kind: $.TypeKind.Basic, name: 'number'}) * v.Len()
				}
			}
			let size = sizeof(t)
			if (size >= 0) {
				if (t!.Kind() == reflect.Struct) {
					await structSize.Store(t, size)
				}
				return size * v.Len()
			}
			break
		case reflect.Struct:
			let t = v.Type()
			{
				let [size, ok] = await structSize.Load(t)
				if (ok) {
					return $.mustTypeAssert<number>(size, {kind: $.TypeKind.Basic, name: 'number'})
				}
			}
			let size = sizeof(t)
			await structSize.Store(t, size)
			return size
			break
		default:
			if (v.IsValid()) {
				return sizeof(v.Type())
			}
			break
	}

	return -1
}

// sizeof returns the size >= 0 of variables for the given type or -1 if the type is not acceptable.
export function sizeof(t: reflect.Type): number {
	switch (t!.Kind()) {
		case reflect.Array:
			{
				let s = sizeof(t!.Elem())
				if (s >= 0) {
					return s * t!.Len()
				}
			}
			break
		case reflect.Struct:
			let sum = 0
			for (let i = 0, n = t!.NumField(); i < n; i++) {
				let s = sizeof(t!.Field(i)!.Type)
				if (s < 0) {
					return -1
				}
				sum += s
			}
			return sum
			break
		case reflect.Bool:
		case reflect.Uint8:
		case reflect.Uint16:
		case reflect.Uint32:
		case reflect.Uint64:
		case reflect.Int8:
		case reflect.Int16:
		case reflect.Int32:
		case reflect.Int64:
		case reflect.Float32:
		case reflect.Float64:
		case reflect.Complex64:
		case reflect.Complex128:
			return $.int(t!.Size())
			break
	}

	return -1
}

// intDataSize returns the size of the data required to represent the data when encoded,
// and optionally a byte slice containing the encoded data if no conversion is necessary.
// It returns zero, nil if the type cannot be implemented by the fast path in Read or Write.
export function intDataSize(data: null | any): [number, $.Bytes] {
	$.typeSwitch(data, [{ types: [{kind: $.TypeKind.Basic, name: 'boolean'}, {kind: $.TypeKind.Basic, name: 'number'}, {kind: $.TypeKind.Basic, name: 'number'}, {kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'boolean'}}, {kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}, {kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return [1, null]
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'boolean'}}], body: (data) => {
		return [$.len(data), null]
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return [$.len(data), null]
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return [$.len(data), data]
	}},
	{ types: [{kind: $.TypeKind.Basic, name: 'number'}, {kind: $.TypeKind.Basic, name: 'number'}, {kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}, {kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return [2, null]
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return [2 * $.len(data), null]
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return [2 * $.len(data), null]
	}},
	{ types: [{kind: $.TypeKind.Basic, name: 'number'}, {kind: $.TypeKind.Basic, name: 'number'}, {kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}, {kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return [4, null]
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return [4 * $.len(data), null]
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return [4 * $.len(data), null]
	}},
	{ types: [{kind: $.TypeKind.Basic, name: 'number'}, {kind: $.TypeKind.Basic, name: 'number'}, {kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}, {kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return [8, null]
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return [8 * $.len(data), null]
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return [8 * $.len(data), null]
	}},
	{ types: [{kind: $.TypeKind.Basic, name: 'number'}, {kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return [4, null]
	}},
	{ types: [{kind: $.TypeKind.Basic, name: 'number'}, {kind: $.TypeKind.Pointer, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return [8, null]
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return [4 * $.len(data), null]
	}},
	{ types: [{kind: $.TypeKind.Slice, elemType: {kind: $.TypeKind.Basic, name: 'number'}}], body: (data) => {
		return [8 * $.len(data), null]
	}}])
	return [0, null]
}

// ensure grows buf to length len(buf) + n and returns the grown buffer
// and a slice starting at the original length of buf (that is, buf2[len(buf):]).
export function ensure(buf: $.Bytes, n: number): $.Bytes {
	let buf2: $.Bytes = new Uint8Array(0)
	let pos: $.Bytes = new Uint8Array(0)
	{
		let l = $.len(buf)
		buf = $.goSlice(slices.Grow(buf, n), undefined, l + n)
		return [buf, $.goSlice(buf, l, undefined)]
	}
}

