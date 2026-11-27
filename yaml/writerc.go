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

// The writer appends the emitter's output to memory, so a write cannot fail.
// It tracks the line and column of the next character and whether the last
// one written was whitespace or indentation, which the emitter reads to lay
// out what follows.

// Put a character to the output.
func put(writer *yaml_writer_t, value byte) {
	writer.out = append(writer.out, value)
	writer.column++
}

// Write the UTF-8 byte order mark.
func write_bom(writer *yaml_writer_t) {
	writer.out = append(writer.out, '\xEF', '\xBB', '\xBF')
}

// Put a line break to the output.
func put_break(writer *yaml_writer_t) {
	switch writer.line_break {
	case yaml_CR_BREAK:
		writer.out = append(writer.out, '\r')
	case yaml_LN_BREAK:
		writer.out = append(writer.out, '\n')
	case yaml_CRLN_BREAK:
		writer.out = append(writer.out, '\r', '\n')
	default:
		panic("unknown line break setting")
	}
	end_line(writer)
}

// end_line moves the writer to the start of the next line.
func end_line(writer *yaml_writer_t) {
	if writer.column == 0 {
		writer.space_above = true
	}
	writer.column = 0
	writer.line++
	writer.indention = true
}

// Copy the character at s[*i] to the output and advance *i past it.
func write(writer *yaml_writer_t, s []byte, i *int) {
	w := width(s[*i])
	if w < 1 || w > 4 {
		panic("unknown character width")
	}
	writer.out = append(writer.out, s[*i:*i+w]...)
	writer.column++
	*i += w
}

// Write a whole string to the output.
func write_all(writer *yaml_writer_t, s []byte) {
	for i := 0; i < len(s); {
		write(writer, s, &i)
	}
}

// Copy the line break at s[*i] to the output and advance *i past it.
func write_break(writer *yaml_writer_t, s []byte, i *int) {
	if s[*i] == '\n' {
		put_break(writer)
		*i++
		return
	}
	write(writer, s, i)
	end_line(writer)
}
