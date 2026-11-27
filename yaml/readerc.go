//
// Copyright (c) 2011-2019 Canonical Ltd
// Copyright (c) 2006-2010 Kirill Simonov
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
// of the Software, and to permit persons to whom the Software is furnished to do
// so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package yaml

// Set the reader error and return 0.
func yaml_scanner_set_reader_error(scanner *yaml_scanner_t, problem string) bool {
	scanner.error = yaml_READER_ERROR
	scanner.problem = problem
	return false
}

// Byte order marks.
const (
	bom_UTF8    = "\xef\xbb\xbf"
	bom_UTF16LE = "\xff\xfe"
	bom_UTF16BE = "\xfe\xff"
)

// Determine the input stream encoding by checking the BOM symbol. If no BOM is
// found, the UTF-8 encoding is assumed. Return 1 on success, 0 on failure.
func yaml_scanner_determine_encoding(scanner *yaml_scanner_t) bool {
	// Ensure that we had enough bytes in the raw buffer.
	for !scanner.eof && len(scanner.raw_buffer)-scanner.raw_buffer_pos < 3 {
		if !yaml_scanner_update_raw_buffer(scanner) {
			return false
		}
	}

	// Determine the encoding.
	buf := scanner.raw_buffer
	pos := scanner.raw_buffer_pos
	avail := len(buf) - pos
	if avail >= 2 && buf[pos] == bom_UTF16LE[0] && buf[pos+1] == bom_UTF16LE[1] {
		scanner.encoding = yaml_UTF16LE_ENCODING
		scanner.raw_buffer_pos += 2
		scanner.offset += 2
	} else if avail >= 2 && buf[pos] == bom_UTF16BE[0] && buf[pos+1] == bom_UTF16BE[1] {
		scanner.encoding = yaml_UTF16BE_ENCODING
		scanner.raw_buffer_pos += 2
		scanner.offset += 2
	} else if avail >= 3 && buf[pos] == bom_UTF8[0] && buf[pos+1] == bom_UTF8[1] && buf[pos+2] == bom_UTF8[2] {
		scanner.encoding = yaml_UTF8_ENCODING
		scanner.raw_buffer_pos += 3
		scanner.offset += 3
	} else {
		scanner.encoding = yaml_UTF8_ENCODING
	}
	return true
}

// Update the raw buffer.
func yaml_scanner_update_raw_buffer(scanner *yaml_scanner_t) bool {
	// Return if the raw buffer is full.
	if scanner.raw_buffer_pos == 0 && len(scanner.raw_buffer) == cap(scanner.raw_buffer) {
		return true
	}

	// Return on EOF.
	if scanner.eof {
		return true
	}

	// Move the remaining bytes in the raw buffer to the beginning.
	if scanner.raw_buffer_pos > 0 && scanner.raw_buffer_pos < len(scanner.raw_buffer) {
		copy(scanner.raw_buffer, scanner.raw_buffer[scanner.raw_buffer_pos:])
	}
	scanner.raw_buffer = scanner.raw_buffer[:len(scanner.raw_buffer)-scanner.raw_buffer_pos]
	scanner.raw_buffer_pos = 0

	// Copy the next input bytes into the free end of the raw buffer. The
	// input is in memory, so EOF comes on the first call after all of it is
	// copied.
	if scanner.input_pos == len(scanner.input) {
		scanner.eof = true
		return true
	}
	free := scanner.raw_buffer[len(scanner.raw_buffer):cap(scanner.raw_buffer)]
	n := copy(free, scanner.input[scanner.input_pos:])
	scanner.input_pos += n
	scanner.raw_buffer = scanner.raw_buffer[:len(scanner.raw_buffer)+n]
	return true
}

// Ensure that the buffer contains at least `length` characters.
// Return true on success, false on failure.
//
// The length is supposed to be significantly less that the buffer size.
func yaml_scanner_update_buffer(scanner *yaml_scanner_t, length int) bool {
	// [Go] This function was changed to guarantee the requested length size at EOF.
	// The fact we need to do this is pretty awful, but the description above implies
	// for that to be the case, and there are tests

	// Return if the buffer contains enough characters.
	if scanner.unread >= length {
		return true
	}

	// Determine the input encoding if it is not known yet.
	if scanner.encoding == yaml_ANY_ENCODING {
		if !yaml_scanner_determine_encoding(scanner) {
			return false
		}
	}

	// Move the unread characters to the beginning of the buffer.
	buffer_len := len(scanner.buffer)
	if scanner.buffer_pos > 0 && scanner.buffer_pos < buffer_len {
		copy(scanner.buffer, scanner.buffer[scanner.buffer_pos:])
		buffer_len -= scanner.buffer_pos
		scanner.buffer_pos = 0
	} else if scanner.buffer_pos == buffer_len {
		buffer_len = 0
		scanner.buffer_pos = 0
	}

	// Open the whole buffer for writing, and cut it before returning.
	scanner.buffer = scanner.buffer[:cap(scanner.buffer)]

	// Fill the buffer until it has enough characters.
	first := true
	for scanner.unread < length {

		// Fill the raw buffer if necessary.
		if !first || scanner.raw_buffer_pos == len(scanner.raw_buffer) {
			if !yaml_scanner_update_raw_buffer(scanner) {
				scanner.buffer = scanner.buffer[:buffer_len]
				return false
			}
		}
		first = false

		// Decode the raw buffer.
	inner:
		for scanner.raw_buffer_pos != len(scanner.raw_buffer) {
			var value rune
			var width int

			raw_unread := len(scanner.raw_buffer) - scanner.raw_buffer_pos

			// Decode the next character.
			switch scanner.encoding {
			case yaml_UTF8_ENCODING:
				// Decode a UTF-8 character.  Check RFC 3629
				// (http://www.ietf.org/rfc/rfc3629.txt) for more details.
				//
				// The following table (taken from the RFC) is used for
				// decoding.
				//
				//    Char. number range |        UTF-8 octet sequence
				//      (hexadecimal)    |              (binary)
				//   --------------------+------------------------------------
				//   0000 0000-0000 007F | 0xxxxxxx
				//   0000 0080-0000 07FF | 110xxxxx 10xxxxxx
				//   0000 0800-0000 FFFF | 1110xxxx 10xxxxxx 10xxxxxx
				//   0001 0000-0010 FFFF | 11110xxx 10xxxxxx 10xxxxxx 10xxxxxx
				//
				// Additionally, the characters in the range 0xD800-0xDFFF
				// are prohibited as they are reserved for use with UTF-16
				// surrogate pairs.

				// Determine the length of the UTF-8 sequence.
				octet := scanner.raw_buffer[scanner.raw_buffer_pos]
				switch {
				case octet&0x80 == 0x00:
					width = 1
				case octet&0xE0 == 0xC0:
					width = 2
				case octet&0xF0 == 0xE0:
					width = 3
				case octet&0xF8 == 0xF0:
					width = 4
				default:
					// The leading octet is invalid.
					return yaml_scanner_set_reader_error(scanner,
						"invalid leading UTF-8 octet")
				}

				// Check if the raw buffer contains an incomplete character.
				if width > raw_unread {
					if scanner.eof {
						return yaml_scanner_set_reader_error(scanner,
							"incomplete UTF-8 octet sequence")
					}
					break inner
				}

				// Decode the leading octet.
				switch {
				case octet&0x80 == 0x00:
					value = rune(octet & 0x7F)
				case octet&0xE0 == 0xC0:
					value = rune(octet & 0x1F)
				case octet&0xF0 == 0xE0:
					value = rune(octet & 0x0F)
				case octet&0xF8 == 0xF0:
					value = rune(octet & 0x07)
				default:
					value = 0
				}

				// Check and decode the trailing octets.
				for k := 1; k < width; k++ {
					octet = scanner.raw_buffer[scanner.raw_buffer_pos+k]

					// Check if the octet is valid.
					if (octet & 0xC0) != 0x80 {
						return yaml_scanner_set_reader_error(scanner,
							"invalid trailing UTF-8 octet")
					}

					// Decode the octet.
					value = (value << 6) + rune(octet&0x3F)
				}

				// Check the length of the sequence against the value.
				switch {
				case width == 1:
				case width == 2 && value >= 0x80:
				case width == 3 && value >= 0x800:
				case width == 4 && value >= 0x10000:
				default:
					return yaml_scanner_set_reader_error(scanner,
						"invalid length of a UTF-8 sequence")
				}

				// Check the range of the value.
				if value >= 0xD800 && value <= 0xDFFF || value > 0x10FFFF {
					return yaml_scanner_set_reader_error(scanner,
						"invalid Unicode character")
				}

			case yaml_UTF16LE_ENCODING, yaml_UTF16BE_ENCODING:
				var low, high int
				if scanner.encoding == yaml_UTF16LE_ENCODING {
					low, high = 0, 1
				} else {
					low, high = 1, 0
				}

				// The UTF-16 encoding is not as simple as one might
				// naively think.  Check RFC 2781
				// (http://www.ietf.org/rfc/rfc2781.txt).
				//
				// Normally, two subsequent bytes describe a Unicode
				// character.  However a special technique (called a
				// surrogate pair) is used for specifying character
				// values larger than 0xFFFF.
				//
				// A surrogate pair consists of two pseudo-characters:
				//      high surrogate area (0xD800-0xDBFF)
				//      low surrogate area (0xDC00-0xDFFF)
				//
				// The following formulas are used for decoding
				// and encoding characters using surrogate pairs:
				//
				//  U  = U' + 0x10000   (0x01 00 00 <= U <= 0x10 FF FF)
				//  U' = yyyyyyyyyyxxxxxxxxxx   (0 <= U' <= 0x0F FF FF)
				//  W1 = 110110yyyyyyyyyy
				//  W2 = 110111xxxxxxxxxx
				//
				// where U is the character value, W1 is the high surrogate
				// area, W2 is the low surrogate area.

				// Check for incomplete UTF-16 character.
				if raw_unread < 2 {
					if scanner.eof {
						return yaml_scanner_set_reader_error(scanner,
							"incomplete UTF-16 character")
					}
					break inner
				}

				// Get the character.
				value = rune(scanner.raw_buffer[scanner.raw_buffer_pos+low]) +
					(rune(scanner.raw_buffer[scanner.raw_buffer_pos+high]) << 8)

				// Check for unexpected low surrogate area.
				if value&0xFC00 == 0xDC00 {
					return yaml_scanner_set_reader_error(scanner,
						"unexpected low surrogate area")
				}

				// Check for a high surrogate area.
				if value&0xFC00 == 0xD800 {
					width = 4

					// Check for incomplete surrogate pair.
					if raw_unread < 4 {
						if scanner.eof {
							return yaml_scanner_set_reader_error(scanner,
								"incomplete UTF-16 surrogate pair")
						}
						break inner
					}

					// Get the next character.
					value2 := rune(scanner.raw_buffer[scanner.raw_buffer_pos+low+2]) +
						(rune(scanner.raw_buffer[scanner.raw_buffer_pos+high+2]) << 8)

					// Check for a low surrogate area.
					if value2&0xFC00 != 0xDC00 {
						return yaml_scanner_set_reader_error(scanner,
							"expected low surrogate area")
					}

					// Generate the value of the surrogate pair.
					value = 0x10000 + ((value & 0x3FF) << 10) + (value2 & 0x3FF)
				} else {
					width = 2
				}

			default:
				panic("impossible")
			}

			// Check if the character is in the allowed range:
			//      #x9 | #xA | #xD | [#x20-#x7E]               (8 bit)
			//      | #x85 | [#xA0-#xD7FF] | [#xE000-#xFFFD]    (16 bit)
			//      | [#x10000-#x10FFFF]                        (32 bit)
			switch {
			case value == 0x09:
			case value == 0x0A:
			case value == 0x0D:
			case value >= 0x20 && value <= 0x7E:
			case value == 0x85:
			case value >= 0xA0 && value <= 0xD7FF:
			case value >= 0xE000 && value <= 0xFFFD:
			case value >= 0x10000 && value <= 0x10FFFF:
			default:
				return yaml_scanner_set_reader_error(scanner,
					"control characters are not allowed")
			}

			// Move the raw pointers.
			scanner.raw_buffer_pos += width
			scanner.offset += width

			// Finally put the character into the buffer.
			if value <= 0x7F {
				// 0000 0000-0000 007F . 0xxxxxxx
				scanner.buffer[buffer_len+0] = byte(value)
				buffer_len += 1
			} else if value <= 0x7FF {
				// 0000 0080-0000 07FF . 110xxxxx 10xxxxxx
				scanner.buffer[buffer_len+0] = byte(0xC0 + (value >> 6))
				scanner.buffer[buffer_len+1] = byte(0x80 + (value & 0x3F))
				buffer_len += 2
			} else if value <= 0xFFFF {
				// 0000 0800-0000 FFFF . 1110xxxx 10xxxxxx 10xxxxxx
				scanner.buffer[buffer_len+0] = byte(0xE0 + (value >> 12))
				scanner.buffer[buffer_len+1] = byte(0x80 + ((value >> 6) & 0x3F))
				scanner.buffer[buffer_len+2] = byte(0x80 + (value & 0x3F))
				buffer_len += 3
			} else {
				// 0001 0000-0010 FFFF . 11110xxx 10xxxxxx 10xxxxxx 10xxxxxx
				scanner.buffer[buffer_len+0] = byte(0xF0 + (value >> 18))
				scanner.buffer[buffer_len+1] = byte(0x80 + ((value >> 12) & 0x3F))
				scanner.buffer[buffer_len+2] = byte(0x80 + ((value >> 6) & 0x3F))
				scanner.buffer[buffer_len+3] = byte(0x80 + (value & 0x3F))
				buffer_len += 4
			}

			scanner.unread++
		}

		// On EOF, put NUL into the buffer and return.
		if scanner.eof {
			scanner.buffer[buffer_len] = 0
			buffer_len++
			scanner.unread++
			break
		}
	}
	// [Go] Read the documentation of this function above. To return true,
	// we need to have the given length in the buffer. Not doing that means
	// every single check that calls this function to make sure the buffer
	// has a given length is Go) panicking; or C) accessing invalid memory.
	// This happens here due to the EOF above breaking early.
	for buffer_len < length {
		scanner.buffer[buffer_len] = 0
		buffer_len++
	}
	scanner.buffer = scanner.buffer[:buffer_len]
	return true
}
