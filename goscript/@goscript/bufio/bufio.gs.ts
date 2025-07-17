import * as $ from "@goscript/builtin/index.js";

import * as bytes from "@goscript/bytes/index.js"

import * as errors from "@goscript/errors/index.js"

import * as io from "@goscript/io/index.js"

import * as strings from "@goscript/strings/index.js"

import * as utf8 from "@goscript/unicode/utf8/index.js"

let defaultBufSize: number = 4096

let minReadBufferSize: number = 16

let maxConsecutiveEmptyReads: number = 100

export class Reader {
	public get buf(): $.Bytes {
		return this._fields.buf.value
	}
	public set buf(value: $.Bytes) {
		this._fields.buf.value = value
	}

	// reader provided by the client
	public get rd(): io.Reader {
		return this._fields.rd.value
	}
	public set rd(value: io.Reader) {
		this._fields.rd.value = value
	}

	// buf read and write positions
	public get r(): number {
		return this._fields.r.value
	}
	public set r(value: number) {
		this._fields.r.value = value
	}

	// buf read and write positions
	public get w(): number {
		return this._fields.w.value
	}
	public set w(value: number) {
		this._fields.w.value = value
	}

	public get err(): $.GoError {
		return this._fields.err.value
	}
	public set err(value: $.GoError) {
		this._fields.err.value = value
	}

	// last byte read for UnreadByte; -1 means invalid
	public get lastByte(): number {
		return this._fields.lastByte.value
	}
	public set lastByte(value: number) {
		this._fields.lastByte.value = value
	}

	// size of last rune read for UnreadRune; -1 means invalid
	public get lastRuneSize(): number {
		return this._fields.lastRuneSize.value
	}
	public set lastRuneSize(value: number) {
		this._fields.lastRuneSize.value = value
	}

	public _fields: {
		buf: $.VarRef<$.Bytes>;
		rd: $.VarRef<io.Reader>;
		r: $.VarRef<number>;
		w: $.VarRef<number>;
		err: $.VarRef<$.GoError>;
		lastByte: $.VarRef<number>;
		lastRuneSize: $.VarRef<number>;
	}

	constructor(init?: Partial<{buf?: $.Bytes, err?: $.GoError, lastByte?: number, lastRuneSize?: number, r?: number, rd?: io.Reader, w?: number}>) {
		this._fields = {
			buf: $.varRef(init?.buf ?? new Uint8Array(0)),
			rd: $.varRef(init?.rd ?? null),
			r: $.varRef(init?.r ?? 0),
			w: $.varRef(init?.w ?? 0),
			err: $.varRef(init?.err ?? null),
			lastByte: $.varRef(init?.lastByte ?? 0),
			lastRuneSize: $.varRef(init?.lastRuneSize ?? 0)
		}
	}

	public clone(): Reader {
		const cloned = new Reader()
		cloned._fields = {
			buf: $.varRef(this._fields.buf.value),
			rd: $.varRef(this._fields.rd.value),
			r: $.varRef(this._fields.r.value),
			w: $.varRef(this._fields.w.value),
			err: $.varRef(this._fields.err.value),
			lastByte: $.varRef(this._fields.lastByte.value),
			lastRuneSize: $.varRef(this._fields.lastRuneSize.value)
		}
		return cloned
	}

	// Size returns the size of the underlying buffer in bytes.
	public Size(): number {
		const b = this
		return $.len(b.buf)
	}

	// Reset discards any buffered data, resets all state, and switches
	// the buffered reader to read from r.
	// Calling Reset on the zero value of [Reader] initializes the internal buffer
	// to the default size.
	// Calling b.Reset(b) (that is, resetting a [Reader] to itself) does nothing.
	public Reset(r: io.Reader): void {
		const b = this
		if (b == r) {
			return 
		}
		if (b.buf == null) {
			b.buf = new Uint8Array(4096)
		}
		b.reset(b.buf, r)
	}

	public reset(buf: $.Bytes, r: io.Reader): void {
		const b = this
		b!.value = $.markAsStructValue(new Reader({buf: buf, lastByte: -1, lastRuneSize: -1, rd: r}))
	}

	// fill reads a new chunk into the buffer.
	public fill(): void {
		const b = this
		if (b.r > 0) {
			$.copy(b.buf, $.goSlice(b.buf, b.r, b.w))
			b.w -= b.r
			b.r = 0
		}
		if (b.w >= $.len(b.buf)) {
			$.panic("bufio: tried to fill full buffer")
		}
		for (let i = 100; i > 0; i--) {
			let [n, err] = b.rd!.Read($.goSlice(b.buf, b.w, undefined))
			if (n < 0) {
				$.panic(errNegativeRead)
			}
			b.w += n
			if (err != null) {
				b.err = err
				return 
			}
			if (n > 0) {
				return 
			}
		}
		b.err = io.ErrNoProgress
	}

	public readErr(): $.GoError {
		const b = this
		let err = b.err
		b.err = null
		return err
	}

	// Peek returns the next n bytes without advancing the reader. The bytes stop
	// being valid at the next read call. If necessary, Peek will read more bytes
	// into the buffer in order to make n bytes available. If Peek returns fewer
	// than n bytes, it also returns an error explaining why the read is short.
	// The error is [ErrBufferFull] if n is larger than b's buffer size.
	//
	// Calling Peek prevents a [Reader.UnreadByte] or [Reader.UnreadRune] call from succeeding
	// until the next read operation.
	public Peek(n: number): [$.Bytes, $.GoError] {
		const b = this
		if (n < 0) {
			return [null, ErrNegativeCount]
		}
		b.lastByte = -1
		b.lastRuneSize = -1
		for (; b.w - b.r < n && b.w - b.r < $.len(b.buf) && b.err == null; ) {
			b.fill() // b.w-b.r < len(b.buf) => buffer is not full
		}
		if (n > $.len(b.buf)) {
			return [$.goSlice(b.buf, b.r, b.w), ErrBufferFull]
		}
		let err: $.GoError = null
		{
			let avail = b.w - b.r
			if (avail < n) {
				// not enough data in buffer
				n = avail
				err = b.readErr()
				if (err == null) {
					err = ErrBufferFull
				}
			}
		}
		return [$.goSlice(b.buf, b.r, b.r + n), err]
	}

	// Discard skips the next n bytes, returning the number of bytes discarded.
	//
	// If Discard skips fewer than n bytes, it also returns an error.
	// If 0 <= n <= b.Buffered(), Discard is guaranteed to succeed without
	// reading from the underlying io.Reader.
	public Discard(n: number): [number, $.GoError] {
		const b = this
		let discarded: number = 0
		let err: $.GoError = null
		if (n < 0) {
			return [0, ErrNegativeCount]
		}
		if (n == 0) {
			return [discarded, err]
		}
		b.lastByte = -1
		b.lastRuneSize = -1
		let remain = n
		for (; ; ) {
			let skip = b.Buffered()
			if (skip == 0) {
				b.fill()
				skip = b.Buffered()
			}
			if (skip > remain) {
				skip = remain
			}
			b.r += skip
			remain -= skip
			if (remain == 0) {
				return [n, null]
			}
			if (b.err != null) {
				return [n - remain, b.readErr()]
			}
		}
	}

	// Read reads data into p.
	// It returns the number of bytes read into p.
	// The bytes are taken from at most one Read on the underlying [Reader],
	// hence n may be less than len(p).
	// To read exactly len(p) bytes, use io.ReadFull(b, p).
	// If the underlying [Reader] can return a non-zero count with io.EOF,
	// then this Read method can do so as well; see the [io.Reader] docs.
	public Read(p: $.Bytes): [number, $.GoError] {
		const b = this
		let n: number = 0
		let err: $.GoError = null
		n = $.len(p)
		if (n == 0) {
			if (b.Buffered() > 0) {
				return [0, null]
			}
			return [0, b.readErr()]
		}
		if (b.r == b.w) {
			if (b.err != null) {
				return [0, b.readErr()]
			}

			// Large read, empty buffer.
			// Read directly into p to avoid copy.
			if ($.len(p) >= $.len(b.buf)) {
				// Large read, empty buffer.
				// Read directly into p to avoid copy.
				{
				  const _tmp = b.rd!.Read(p)
				  n = _tmp[0]
				  b.err = _tmp[1]
				}
				if (n < 0) {
					$.panic(errNegativeRead)
				}
				if (n > 0) {
					b.lastByte = $.int(p![n - 1])
					b.lastRuneSize = -1
				}
				return [n, b.readErr()]
			}
			// One read.
			// Do not use b.fill, which will loop.
			b.r = 0
			b.w = 0
			{
			  const _tmp = b.rd!.Read(b.buf)
			  n = _tmp[0]
			  b.err = _tmp[1]
			}
			if (n < 0) {
				$.panic(errNegativeRead)
			}
			if (n == 0) {
				return [0, b.readErr()]
			}
			b.w += n
		}
		n = $.copy(p, $.goSlice(b.buf, b.r, b.w))
		b.r += n
		b.lastByte = $.int(b.buf![b.r - 1])
		b.lastRuneSize = -1
		return [n, null]
	}

	// ReadByte reads and returns a single byte.
	// If no byte is available, returns an error.
	public ReadByte(): [number, $.GoError] {
		const b = this
		b.lastRuneSize = -1
		for (; b.r == b.w; ) {
			if (b.err != null) {
				return [0, b.readErr()]
			}
			b.fill() // buffer is empty
		}
		let c = b.buf![b.r]
		b.r++
		b.lastByte = $.int(c)
		return [c, null]
	}

	// UnreadByte unreads the last byte. Only the most recently read byte can be unread.
	//
	// UnreadByte returns an error if the most recent method called on the
	// [Reader] was not a read operation. Notably, [Reader.Peek], [Reader.Discard], and [Reader.WriteTo] are not
	// considered read operations.
	public UnreadByte(): $.GoError {
		const b = this
		if (b.lastByte < 0 || b.r == 0 && b.w > 0) {
			return ErrInvalidUnreadByte
		}
		if (b.r > 0) {
			b.r--
		}
		 else {
			// b.r == 0 && b.w == 0
			b.w = 1
		}
		b.buf![b.r] = $.byte(b.lastByte)
		b.lastByte = -1
		b.lastRuneSize = -1
		return null
	}

	// ReadRune reads a single UTF-8 encoded Unicode character and returns the
	// rune and its size in bytes. If the encoded rune is invalid, it consumes one byte
	// and returns unicode.ReplacementChar (U+FFFD) with a size of 1.
	public ReadRune(): [number, number, $.GoError] {
		const b = this
		let r: number = 0
		let size: number = 0
		let err: $.GoError = null
		for (; b.r + utf8.UTFMax > b.w && !utf8.FullRune($.goSlice(b.buf, b.r, b.w)) && b.err == null && b.w - b.r < $.len(b.buf); ) {
			b.fill() // b.w-b.r < len(buf) => buffer is not full
		}
		b.lastRuneSize = -1
		if (b.r == b.w) {
			return [0, 0, b.readErr()]
		}
		;[r, size] = [(b.buf![b.r] as number), 1]
		if (r >= utf8.RuneSelf) {
			;[r, size] = utf8.DecodeRune($.goSlice(b.buf, b.r, b.w))
		}
		b.r += size
		b.lastByte = $.int(b.buf![b.r - 1])
		b.lastRuneSize = size
		return [r, size, null]
	}

	// UnreadRune unreads the last rune. If the most recent method called on
	// the [Reader] was not a [Reader.ReadRune], [Reader.UnreadRune] returns an error. (In this
	// regard it is stricter than [Reader.UnreadByte], which will unread the last byte
	// from any read operation.)
	public UnreadRune(): $.GoError {
		const b = this
		if (b.lastRuneSize < 0 || b.r < b.lastRuneSize) {
			return ErrInvalidUnreadRune
		}
		b.r -= b.lastRuneSize
		b.lastByte = -1
		b.lastRuneSize = -1
		return null
	}

	// Buffered returns the number of bytes that can be read from the current buffer.
	public Buffered(): number {
		const b = this
		return b.w - b.r
	}

	// ReadSlice reads until the first occurrence of delim in the input,
	// returning a slice pointing at the bytes in the buffer.
	// The bytes stop being valid at the next read.
	// If ReadSlice encounters an error before finding a delimiter,
	// it returns all the data in the buffer and the error itself (often io.EOF).
	// ReadSlice fails with error [ErrBufferFull] if the buffer fills without a delim.
	// Because the data returned from ReadSlice will be overwritten
	// by the next I/O operation, most clients should use
	// [Reader.ReadBytes] or ReadString instead.
	// ReadSlice returns err != nil if and only if line does not end in delim.
	public ReadSlice(delim: number): [$.Bytes, $.GoError] {
		const b = this
		let line: $.Bytes = new Uint8Array(0)
		let err: $.GoError = null
		let s = 0 // search start index
		for (; ; ) {
			// Search buffer.
			{
				let i = bytes.IndexByte($.goSlice(b.buf, b.r + s, b.w), delim)
				if (i >= 0) {
					i += s
					line = $.goSlice(b.buf, b.r, b.r + i + 1)
					b.r += i + 1
					break
				}
			}

			// Pending error?
			if (b.err != null) {
				line = $.goSlice(b.buf, b.r, b.w)
				b.r = b.w
				err = b.readErr()
				break
			}

			// Buffer full?
			if (b.Buffered() >= $.len(b.buf)) {
				b.r = b.w
				line = b.buf
				err = ErrBufferFull
				break
			}

			s = b.w - b.r // do not rescan area we scanned before

			b.fill() // buffer is not full
		}
		{
			let i = $.len(line) - 1
			if (i >= 0) {
				b.lastByte = $.int(line![i])
				b.lastRuneSize = -1
			}
		}
		return [line, err]
	}

	// ReadLine is a low-level line-reading primitive. Most callers should use
	// [Reader.ReadBytes]('\n') or [Reader.ReadString]('\n') instead or use a [Scanner].
	//
	// ReadLine tries to return a single line, not including the end-of-line bytes.
	// If the line was too long for the buffer then isPrefix is set and the
	// beginning of the line is returned. The rest of the line will be returned
	// from future calls. isPrefix will be false when returning the last fragment
	// of the line. The returned buffer is only valid until the next call to
	// ReadLine. ReadLine either returns a non-nil line or it returns an error,
	// never both.
	//
	// The text returned from ReadLine does not include the line end ("\r\n" or "\n").
	// No indication or error is given if the input ends without a final line end.
	// Calling [Reader.UnreadByte] after ReadLine will always unread the last byte read
	// (possibly a character belonging to the line end) even if that byte is not
	// part of the line returned by ReadLine.
	public ReadLine(): [$.Bytes, boolean, $.GoError] {
		const b = this
		let line: $.Bytes = new Uint8Array(0)
		let isPrefix: boolean = false
		let err: $.GoError = null
		;[line, err] = b.ReadSlice(10)
		if (err == ErrBufferFull) {
			// Handle the case where "\r\n" straddles the buffer.

			// Put the '\r' back on buf and drop it from line.
			// Let the next call to ReadLine check for "\r\n".

			// should be unreachable
			if ($.len(line) > 0 && line![$.len(line) - 1] == 13) {
				// Put the '\r' back on buf and drop it from line.
				// Let the next call to ReadLine check for "\r\n".

				// should be unreachable
				if (b.r == 0) {
					// should be unreachable
					$.panic("bufio: tried to rewind past start of buffer")
				}
				b.r--
				line = $.goSlice(line, undefined, $.len(line) - 1)
			}
			return [line, true, null]
		}
		if ($.len(line) == 0) {
			if (err != null) {
				line = null
			}
			return [line, isPrefix, err]
		}
		err = null
		if (line![$.len(line) - 1] == 10) {
			let drop = 1
			if ($.len(line) > 1 && line![$.len(line) - 2] == 13) {
				drop = 2
			}
			line = $.goSlice(line, undefined, $.len(line) - drop)
		}
		return [line, isPrefix, err]
	}

	// collectFragments reads until the first occurrence of delim in the input. It
	// returns (slice of full buffers, remaining bytes before delim, total number
	// of bytes in the combined first two elements, error).
	// The complete result is equal to
	// `bytes.Join(append(fullBuffers, finalFragment), nil)`, which has a
	// length of `totalLen`. The result is structured in this way to allow callers
	// to minimize allocations and copies.
	public collectFragments(delim: number): [$.Slice<$.Bytes>, $.Bytes, number, $.GoError] {
		const b = this
		let fullBuffers: $.Slice<$.Bytes> = null
		let finalFragment: $.Bytes = new Uint8Array(0)
		let totalLen: number = 0
		let err: $.GoError = null
		let frag: $.Bytes = new Uint8Array(0)
		for (; ; ) {
			let e: $.GoError = null
			;[frag, e] = b.ReadSlice(delim)
			// got final fragment
			if (e == null) {
				// got final fragment
				break
			}
			// unexpected error
			if (e != ErrBufferFull) {
				// unexpected error
				err = e
				break
			}

			// Make a copy of the buffer.
			let buf = bytes.Clone(frag)
			fullBuffers = $.append(fullBuffers, buf)
			totalLen += $.len(buf)
		}
		totalLen += $.len(frag)
		return [fullBuffers, frag, totalLen, err]
	}

	// ReadBytes reads until the first occurrence of delim in the input,
	// returning a slice containing the data up to and including the delimiter.
	// If ReadBytes encounters an error before finding a delimiter,
	// it returns the data read before the error and the error itself (often io.EOF).
	// ReadBytes returns err != nil if and only if the returned data does not end in
	// delim.
	// For simple uses, a Scanner may be more convenient.
	public ReadBytes(delim: number): [$.Bytes, $.GoError] {
		const b = this
		let [full, frag, n, err] = b.collectFragments(delim)
		let buf = new Uint8Array(n)
		n = 0
		for (let i = 0; i < $.len(full); i++) {
			{
				n += $.copy($.goSlice(buf, n, undefined), full![i])
			}
		}
		$.copy($.goSlice(buf, n, undefined), frag)
		return [buf, err]
	}

	// ReadString reads until the first occurrence of delim in the input,
	// returning a string containing the data up to and including the delimiter.
	// If ReadString encounters an error before finding a delimiter,
	// it returns the data read before the error and the error itself (often io.EOF).
	// ReadString returns err != nil if and only if the returned data does not end in
	// delim.
	// For simple uses, a Scanner may be more convenient.
	public ReadString(delim: number): [string, $.GoError] {
		const b = this
		let [full, frag, n, err] = b.collectFragments(delim)
		let buf: strings.Builder = new strings.Builder()
		buf.Grow(n)
		for (let _i = 0; _i < $.len(full); _i++) {
			const fb = full![_i]
			{
				buf.Write(fb)
			}
		}
		buf.Write(frag)
		return [buf.String(), err]
	}

	// WriteTo implements io.WriterTo.
	// This may make multiple calls to the [Reader.Read] method of the underlying [Reader].
	// If the underlying reader supports the [Reader.WriteTo] method,
	// this calls the underlying [Reader.WriteTo] without buffering.
	public WriteTo(w: io.Writer): [number, $.GoError] {
		const b = this
		let n: number = 0
		let err: $.GoError = null
		b.lastByte = -1
		b.lastRuneSize = -1
		;[n, err] = b.writeBuf(w)
		if (err != null) {
			return [n, err]
		}
		{
			let { value: r, ok: ok } = $.typeAssert<io.WriterTo>(b.rd, 'io.WriterTo')
			if (ok) {
				let [m, err] = r!.WriteTo(w)
				n += m
				return [n, err]
			}
		}
		const _temp_w = w
		{
			let { value: w, ok: ok } = $.typeAssert<io.ReaderFrom>(_temp_w, 'io.ReaderFrom')
			if (ok) {
				let [m, err] = w!.ReadFrom(b.rd)
				n += m
				return [n, err]
			}
		}
		if (b.w - b.r < $.len(b.buf)) {
			b.fill() // buffer not full
		}
		for (; b.r < b.w; ) {
			// b.r < b.w => buffer is not empty
			let [m, err] = b.writeBuf(w)
			n += m
			if (err != null) {
				return [n, err]
			}
			b.fill() // buffer is empty
		}
		if (b.err == io.EOF) {
			b.err = null
		}
		return [n, b.readErr()]
	}

	// writeBuf writes the [Reader]'s buffer to the writer.
	public writeBuf(w: io.Writer): [number, $.GoError] {
		const b = this
		let [n, err] = w!.Write($.goSlice(b.buf, b.r, b.w))
		if (n < 0) {
			$.panic(errNegativeWrite)
		}
		b.r += n
		return [(n as number), err]
	}

	// Register this type with the runtime type system
	static __typeInfo = $.registerStructType(
	  'Reader',
	  new Reader(),
	  [{ name: "Size", args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }] }, { name: "Reset", args: [{ name: "r", type: "Reader" }], returns: [] }, { name: "reset", args: [{ name: "buf", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { name: "r", type: "Reader" }], returns: [] }, { name: "fill", args: [], returns: [] }, { name: "readErr", args: [], returns: [{ type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }, { name: "Peek", args: [{ name: "n", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [{ type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }, { name: "Discard", args: [{ name: "n", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }, { type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }, { name: "Read", args: [{ name: "p", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }, { type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }, { name: "ReadByte", args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }, { type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }, { name: "UnreadByte", args: [], returns: [{ type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }, { name: "ReadRune", args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }, { type: { kind: $.TypeKind.Basic, name: "number" } }, { type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }, { name: "UnreadRune", args: [], returns: [{ type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }, { name: "Buffered", args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }] }, { name: "ReadSlice", args: [{ name: "delim", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [{ type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }, { name: "ReadLine", args: [], returns: [{ type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { type: { kind: $.TypeKind.Basic, name: "boolean" } }, { type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }, { name: "collectFragments", args: [{ name: "delim", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [{ type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } } }, { type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { type: { kind: $.TypeKind.Basic, name: "number" } }, { type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }, { name: "ReadBytes", args: [{ name: "delim", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [{ type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }, { type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }, { name: "ReadString", args: [{ name: "delim", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [{ type: { kind: $.TypeKind.Basic, name: "string" } }, { type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }, { name: "WriteTo", args: [{ name: "w", type: "Writer" }], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }, { type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }, { name: "writeBuf", args: [{ name: "w", type: "Writer" }], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }, { type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }],
	  Reader,
	  {"buf": { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } }, "rd": "Reader", "r": { kind: $.TypeKind.Basic, name: "number" }, "w": { kind: $.TypeKind.Basic, name: "number" }, "err": { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] }, "lastByte": { kind: $.TypeKind.Basic, name: "number" }, "lastRuneSize": { kind: $.TypeKind.Basic, name: "number" }}
	);
}

export class Writer {
	public get err(): $.GoError {
		return this._fields.err.value
	}
	public set err(value: $.GoError) {
		this._fields.err.value = value
	}

	public get buf(): $.Bytes {
		return this._fields.buf.value
	}
	public set buf(value: $.Bytes) {
		this._fields.buf.value = value
	}

	public get n(): number {
		return this._fields.n.value
	}
	public set n(value: number) {
		this._fields.n.value = value
	}

	public get wr(): io.Writer {
		return this._fields.wr.value
	}
	public set wr(value: io.Writer) {
		this._fields.wr.value = value
	}

	public _fields: {
		err: $.VarRef<$.GoError>;
		buf: $.VarRef<$.Bytes>;
		n: $.VarRef<number>;
		wr: $.VarRef<io.Writer>;
	}

	constructor(init?: Partial<{buf?: $.Bytes, err?: $.GoError, n?: number, wr?: io.Writer}>) {
		this._fields = {
			err: $.varRef(init?.err ?? null),
			buf: $.varRef(init?.buf ?? new Uint8Array(0)),
			n: $.varRef(init?.n ?? 0),
			wr: $.varRef(init?.wr ?? null)
		}
	}

	public clone(): Writer {
		const cloned = new Writer()
		cloned._fields = {
			err: $.varRef(this._fields.err.value),
			buf: $.varRef(this._fields.buf.value),
			n: $.varRef(this._fields.n.value),
			wr: $.varRef(this._fields.wr.value)
		}
		return cloned
	}

	// Size returns the size of the underlying buffer in bytes.
	public Size(): number {
		const b = this
		return $.len(b.buf)
	}

	// Reset discards any unflushed buffered data, clears any error, and
	// resets b to write its output to w.
	// Calling Reset on the zero value of [Writer] initializes the internal buffer
	// to the default size.
	// Calling w.Reset(w) (that is, resetting a [Writer] to itself) does nothing.
	public Reset(w: io.Writer): void {
		const b = this
		if (b == w) {
			return 
		}
		if (b.buf == null) {
			b.buf = new Uint8Array(4096)
		}
		b.err = null
		b.n = 0
		b.wr = w
	}

	// Flush writes any buffered data to the underlying [io.Writer].
	public Flush(): $.GoError {
		const b = this
		if (b.err != null) {
			return b.err
		}
		if (b.n == 0) {
			return null
		}
		let [n, err] = b.wr!.Write($.goSlice(b.buf, 0, b.n))
		if (n < b.n && err == null) {
			err = io.ErrShortWrite
		}
		if (err != null) {
			if (n > 0 && n < b.n) {
				$.copy($.goSlice(b.buf, 0, b.n - n), $.goSlice(b.buf, n, b.n))
			}
			b.n -= n
			b.err = err
			return err
		}
		b.n = 0
		return null
	}

	// Available returns how many bytes are unused in the buffer.
	public Available(): number {
		const b = this
		return $.len(b.buf) - b.n
	}

	// AvailableBuffer returns an empty buffer with b.Available() capacity.
	// This buffer is intended to be appended to and
	// passed to an immediately succeeding [Writer.Write] call.
	// The buffer is only valid until the next write operation on b.
	public AvailableBuffer(): $.Bytes {
		const b = this
		return $.goSlice($.goSlice(b.buf, b.n, undefined), undefined, 0)
	}

	// Buffered returns the number of bytes that have been written into the current buffer.
	public Buffered(): number {
		const b = this
		return b.n
	}

	// Write writes the contents of p into the buffer.
	// It returns the number of bytes written.
	// If nn < len(p), it also returns an error explaining
	// why the write is short.
	public Write(p: $.Bytes): [number, $.GoError] {
		const b = this
		let nn: number = 0
		let err: $.GoError = null
		for (; $.len(p) > b.Available() && b.err == null; ) {
			let n: number = 0

			// Large write, empty buffer.
			// Write directly from p to avoid copy.
			if (b.Buffered() == 0) {
				// Large write, empty buffer.
				// Write directly from p to avoid copy.
				{
				  const _tmp = b.wr!.Write(p)
				  n = _tmp[0]
				  b.err = _tmp[1]
				}
			}
			 else {
				n = $.copy($.goSlice(b.buf, b.n, undefined), p)
				b.n += n
				b.Flush()
			}
			nn += n
			p = $.goSlice(p, n, undefined)
		}
		if (b.err != null) {
			return [nn, b.err]
		}
		let n = $.copy($.goSlice(b.buf, b.n, undefined), p)
		b.n += n
		nn += n
		return [nn, null]
	}

	// WriteByte writes a single byte.
	public WriteByte(c: number): $.GoError {
		const b = this
		if (b.err != null) {
			return b.err
		}
		if (b.Available() <= 0 && b.Flush() != null) {
			return b.err
		}
		b.buf![b.n] = c
		b.n++
		return null
	}

	// WriteRune writes a single Unicode code point, returning
	// the number of bytes written and any error.
	public WriteRune(r: number): [number, $.GoError] {
		const b = this
		let size: number = 0
		let err: $.GoError = null
		if ((r as number) < utf8.RuneSelf) {
			err = b.WriteByte($.byte(r))
			if (err != null) {
				return [0, err]
			}
			return [1, null]
		}
		if (b.err != null) {
			return [0, b.err]
		}
		let n = b.Available()
		if (n < utf8.UTFMax) {
			{
				b.Flush()
				if (b.err != null) {
					return [0, b.err]
				}
			}
			n = b.Available()

			// Can only happen if buffer is silly small.
			if (n < utf8.UTFMax) {
				// Can only happen if buffer is silly small.
				return b.WriteString($.runeOrStringToString(r))
			}
		}
		size = utf8.EncodeRune($.goSlice(b.buf, b.n, undefined), r)
		b.n += size
		return [size, null]
	}

	// WriteString writes a string.
	// It returns the number of bytes written.
	// If the count is less than len(s), it also returns an error explaining
	// why the write is short.
	public WriteString(s: string): [number, $.GoError] {
		const b = this
		let sw: io.StringWriter = null
		let tryStringWriter = true
		let nn = 0
		for (; $.len(s) > b.Available() && b.err == null; ) {
			let n: number = 0

			// Check at most once whether b.wr is a StringWriter.
			if (b.Buffered() == 0 && sw == null && tryStringWriter) {
				// Check at most once whether b.wr is a StringWriter.
				({ value: sw, ok: tryStringWriter } = $.typeAssert<io.StringWriter>(b.wr, 'io.StringWriter'))
			}

			// Large write, empty buffer, and the underlying writer supports
			// WriteString: forward the write to the underlying StringWriter.
			// This avoids an extra copy.
			if (b.Buffered() == 0 && tryStringWriter) {
				// Large write, empty buffer, and the underlying writer supports
				// WriteString: forward the write to the underlying StringWriter.
				// This avoids an extra copy.
				{
				  const _tmp = sw!.WriteString(s)
				  n = _tmp[0]
				  b.err = _tmp[1]
				}
			}
			 else {
				n = $.copy($.goSlice(b.buf, b.n, undefined), s)
				b.n += n
				b.Flush()
			}
			nn += n
			s = $.sliceString(s, n, undefined)
		}
		if (b.err != null) {
			return [nn, b.err]
		}
		let n = $.copy($.goSlice(b.buf, b.n, undefined), s)
		b.n += n
		nn += n
		return [nn, null]
	}

	// ReadFrom implements [io.ReaderFrom]. If the underlying writer
	// supports the ReadFrom method, this calls the underlying ReadFrom.
	// If there is buffered data and an underlying ReadFrom, this fills
	// the buffer and writes it before calling ReadFrom.
	public ReadFrom(r: io.Reader): [number, $.GoError] {
		const b = this
		let n: number = 0
		let err: $.GoError = null
		if (b.err != null) {
			return [0, b.err]
		}
		let { value: readerFrom, ok: readerFromOK } = $.typeAssert<io.ReaderFrom>(b.wr, 'io.ReaderFrom')
		let m: number = 0
		for (; ; ) {
			if (b.Available() == 0) {
				{
					let err1 = b.Flush()
					if (err1 != null) {
						return [n, err1]
					}
				}
			}
			if (readerFromOK && b.Buffered() == 0) {
				let [nn, err] = readerFrom!.ReadFrom(r)
				b.err = err
				n += nn
				return [n, err]
			}
			let nr = 0
			for (; nr < 100; ) {
				;[m, err] = r!.Read($.goSlice(b.buf, b.n, undefined))
				if (m != 0 || err != null) {
					break
				}
				nr++
			}
			if (nr == 100) {
				return [n, io.ErrNoProgress]
			}
			b.n += m
			n += (m as number)
			if (err != null) {
				break
			}
		}
		if (err == io.EOF) {
			// If we filled the buffer exactly, flush preemptively.
			if (b.Available() == 0) {
				err = b.Flush()
			}
			 else {
				err = null
			}
		}
		return [n, err]
	}

	// Register this type with the runtime type system
	static __typeInfo = $.registerStructType(
	  'Writer',
	  new Writer(),
	  [{ name: "Size", args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }] }, { name: "Reset", args: [{ name: "w", type: "Writer" }], returns: [] }, { name: "Flush", args: [], returns: [{ type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }, { name: "Available", args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }] }, { name: "AvailableBuffer", args: [], returns: [{ type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }] }, { name: "Buffered", args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }] }, { name: "Write", args: [{ name: "p", type: { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } } }], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }, { type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }, { name: "WriteByte", args: [{ name: "c", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [{ type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }, { name: "WriteRune", args: [{ name: "r", type: { kind: $.TypeKind.Basic, name: "number" } }], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }, { type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }, { name: "WriteString", args: [{ name: "s", type: { kind: $.TypeKind.Basic, name: "string" } }], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }, { type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }, { name: "ReadFrom", args: [{ name: "r", type: "Reader" }], returns: [{ type: { kind: $.TypeKind.Basic, name: "number" } }, { type: { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] } }] }],
	  Writer,
	  {"err": { kind: $.TypeKind.Interface, name: 'GoError', methods: [{ name: 'Error', args: [], returns: [{ type: { kind: $.TypeKind.Basic, name: 'string' } }] }] }, "buf": { kind: $.TypeKind.Slice, elemType: { kind: $.TypeKind.Basic, name: "number" } }, "n": { kind: $.TypeKind.Basic, name: "number" }, "wr": "Writer"}
	);
}

export class ReadWriter {
	public get Reader(): Reader | null {
		return this._fields.Reader.value
	}
	public set Reader(value: Reader | null) {
		this._fields.Reader.value = value
	}

	public get Writer(): Writer | null {
		return this._fields.Writer.value
	}
	public set Writer(value: Writer | null) {
		this._fields.Writer.value = value
	}

	public _fields: {
		Reader: $.VarRef<Reader | null>;
		Writer: $.VarRef<Writer | null>;
	}

	constructor(init?: Partial<{Reader?: Partial<ConstructorParameters<typeof Reader>[0]>, Writer?: Partial<ConstructorParameters<typeof Writer>[0]>}>) {
		this._fields = {
			Reader: $.varRef(init?.Reader ?? null),
			Writer: $.varRef(init?.Writer ?? null)
		}
	}

	public clone(): ReadWriter {
		const cloned = new ReadWriter()
		cloned._fields = {
			Reader: $.varRef(this._fields.Reader.value),
			Writer: $.varRef(this._fields.Writer.value)
		}
		return cloned
	}

	public get buf(): $.Bytes {
		return this.Reader.buf
	}
	public set buf(value: $.Bytes) {
		this.Reader.buf = value
	}

	public get rd(): io.Reader {
		return this.Reader.rd
	}
	public set rd(value: io.Reader) {
		this.Reader.rd = value
	}

	public get r(): number {
		return this.Reader.r
	}
	public set r(value: number) {
		this.Reader.r = value
	}

	public get w(): number {
		return this.Reader.w
	}
	public set w(value: number) {
		this.Reader.w = value
	}

	public get err(): $.GoError {
		return this.Reader.err
	}
	public set err(value: $.GoError) {
		this.Reader.err = value
	}

	public get lastByte(): number {
		return this.Reader.lastByte
	}
	public set lastByte(value: number) {
		this.Reader.lastByte = value
	}

	public get lastRuneSize(): number {
		return this.Reader.lastRuneSize
	}
	public set lastRuneSize(value: number) {
		this.Reader.lastRuneSize = value
	}

	public Buffered(): number {
		return this.Reader.Buffered()
	}

	public Discard(n: number): [number, $.GoError] {
		return this.Reader.Discard(n)
	}

	public Peek(n: number): [$.Bytes, $.GoError] {
		return this.Reader.Peek(n)
	}

	public Read(p: $.Bytes): [number, $.GoError] {
		return this.Reader.Read(p)
	}

	public ReadByte(): [number, $.GoError] {
		return this.Reader.ReadByte()
	}

	public ReadBytes(delim: number): [$.Bytes, $.GoError] {
		return this.Reader.ReadBytes(delim)
	}

	public ReadLine(): [$.Bytes, boolean, $.GoError] {
		return this.Reader.ReadLine()
	}

	public ReadRune(): [number, number, $.GoError] {
		return this.Reader.ReadRune()
	}

	public ReadSlice(delim: number): [$.Bytes, $.GoError] {
		return this.Reader.ReadSlice(delim)
	}

	public ReadString(delim: number): [string, $.GoError] {
		return this.Reader.ReadString(delim)
	}

	public Reset(r: io.Reader): void {
		this.Reader.Reset(r)
	}

	public Size(): number {
		return this.Reader.Size()
	}

	public UnreadByte(): $.GoError {
		return this.Reader.UnreadByte()
	}

	public UnreadRune(): $.GoError {
		return this.Reader.UnreadRune()
	}

	public WriteTo(w: io.Writer): [number, $.GoError] {
		return this.Reader.WriteTo(w)
	}

	public collectFragments(delim: number): [$.Slice<$.Bytes>, $.Bytes, number, $.GoError] {
		return this.Reader.collectFragments(delim)
	}

	public fill(): void {
		this.Reader.fill()
	}

	public readErr(): $.GoError {
		return this.Reader.readErr()
	}

	public reset(buf: $.Bytes, r: io.Reader): void {
		this.Reader.reset(buf, r)
	}

	public writeBuf(w: io.Writer): [number, $.GoError] {
		return this.Reader.writeBuf(w)
	}

	public get n(): number {
		return this.Writer.n
	}
	public set n(value: number) {
		this.Writer.n = value
	}

	public get wr(): io.Writer {
		return this.Writer.wr
	}
	public set wr(value: io.Writer) {
		this.Writer.wr = value
	}

	public Available(): number {
		return this.Writer.Available()
	}

	public AvailableBuffer(): $.Bytes {
		return this.Writer.AvailableBuffer()
	}

	public Flush(): $.GoError {
		return this.Writer.Flush()
	}

	public ReadFrom(r: io.Reader): [number, $.GoError] {
		return this.Writer.ReadFrom(r)
	}

	public Write(p: $.Bytes): [number, $.GoError] {
		return this.Writer.Write(p)
	}

	public WriteByte(c: number): $.GoError {
		return this.Writer.WriteByte(c)
	}

	public WriteRune(r: number): [number, $.GoError] {
		return this.Writer.WriteRune(r)
	}

	public WriteString(s: string): [number, $.GoError] {
		return this.Writer.WriteString(s)
	}

	// Register this type with the runtime type system
	static __typeInfo = $.registerStructType(
	  'ReadWriter',
	  new ReadWriter(),
	  [],
	  ReadWriter,
	  {"Reader": { kind: $.TypeKind.Pointer, elemType: "Reader" }, "Writer": { kind: $.TypeKind.Pointer, elemType: "Writer" }}
	);
}

export let ErrBufferFull: $.GoError = errors.New("bufio: buffer full")

export let ErrInvalidUnreadByte: $.GoError = errors.New("bufio: invalid use of UnreadByte")

export let ErrInvalidUnreadRune: $.GoError = errors.New("bufio: invalid use of UnreadRune")

export let ErrNegativeCount: $.GoError = errors.New("bufio: negative count")

let errNegativeRead: $.GoError = errors.New("bufio: reader returned negative count from Read")

let errNegativeWrite: $.GoError = errors.New("bufio: writer returned negative count from Write")

// NewReaderSize returns a new [Reader] whose buffer has at least the specified
// size. If the argument io.Reader is already a [Reader] with large enough
// size, it returns the underlying [Reader].
export function NewReaderSize(rd: io.Reader, size: number): Reader | null {
	// Is it already a Reader?
	let { value: b, ok: ok } = $.typeAssert<Reader | null>(rd, {kind: $.TypeKind.Pointer, elemType: 'Reader'})
	if (ok && $.len(b.buf) >= size) {
		return b
	}
	let r = new Reader()
	r.reset(new Uint8Array(max(size, 16)), rd)
	return r
}

// NewReader returns a new [Reader] whose buffer has the default size.
export function NewReader(rd: io.Reader): Reader | null {
	return NewReaderSize(rd, 4096)
}

// NewWriterSize returns a new [Writer] whose buffer has at least the specified
// size. If the argument io.Writer is already a [Writer] with large enough
// size, it returns the underlying [Writer].
export function NewWriterSize(w: io.Writer, size: number): Writer | null {
	// Is it already a Writer?
	let { value: b, ok: ok } = $.typeAssert<Writer | null>(w, {kind: $.TypeKind.Pointer, elemType: 'Writer'})
	if (ok && $.len(b.buf) >= size) {
		return b
	}
	if (size <= 0) {
		size = 4096
	}
	return new Writer({buf: new Uint8Array(size), wr: w})
}

// NewWriter returns a new [Writer] whose buffer has the default size.
// If the argument io.Writer is already a [Writer] with large enough buffer size,
// it returns the underlying [Writer].
export function NewWriter(w: io.Writer): Writer | null {
	return NewWriterSize(w, 4096)
}

// NewReadWriter allocates a new [ReadWriter] that dispatches to r and w.
export function NewReadWriter(r: Reader | null, w: Writer | null): ReadWriter | null {
	return new ReadWriter({})
}

