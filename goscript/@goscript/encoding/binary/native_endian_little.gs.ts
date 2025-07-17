import * as $ from "@goscript/builtin/index.js";
import { littleEndian } from "./binary.gs.js";

export class nativeEndian {
	public get littleEndian(): littleEndian {
		return this._fields.littleEndian.value
	}
	public set littleEndian(value: littleEndian) {
		this._fields.littleEndian.value = value
	}

	public _fields: {
		littleEndian: $.VarRef<littleEndian>;
	}

	constructor(init?: Partial<{littleEndian?: Partial<ConstructorParameters<typeof littleEndian>[0]>}>) {
		this._fields = {
			littleEndian: $.varRef(new littleEndian(init?.littleEndian))
		}
	}

	public clone(): nativeEndian {
		const cloned = new nativeEndian()
		cloned._fields = {
			littleEndian: $.varRef($.markAsStructValue(this._fields.littleEndian.value.clone()))
		}
		return cloned
	}

	public String(): string {
		return "NativeEndian"
	}

	public GoString(): string {
		return "binary.NativeEndian"
	}

	public AppendUint16(b: $.Bytes, v: number): $.Bytes {
		return this.littleEndian.AppendUint16(b, v)
	}

	public AppendUint32(b: $.Bytes, v: number): $.Bytes {
		return this.littleEndian.AppendUint32(b, v)
	}

	public AppendUint64(b: $.Bytes, v: number): $.Bytes {
		return this.littleEndian.AppendUint64(b, v)
	}

	public PutUint16(b: $.Bytes, v: number): void {
		this.littleEndian.PutUint16(b, v)
	}

	public PutUint32(b: $.Bytes, v: number): void {
		this.littleEndian.PutUint32(b, v)
	}

	public PutUint64(b: $.Bytes, v: number): void {
		this.littleEndian.PutUint64(b, v)
	}

	public Uint16(b: $.Bytes): number {
		return this.littleEndian.Uint16(b)
	}

	public Uint32(b: $.Bytes): number {
		return this.littleEndian.Uint32(b)
	}

	public Uint64(b: $.Bytes): number {
		return this.littleEndian.Uint64(b)
	}

	// Register this type with the runtime type system
	static __typeInfo = $.registerStructType(
	  'nativeEndian',
	  new nativeEndian(),
	  [{ name: "String", args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: "string" } }] }, { name: "GoString", args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: "string" } }] }],
	  nativeEndian,
	  {"littleEndian": "littleEndian"}
	);
}

export let NativeEndian: nativeEndian = new nativeEndian({})

