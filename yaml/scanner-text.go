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

import (
	"bytes"
	"encoding/hex"
	"math"
)

// The functions here read the text of one token from the reader's buffer:
// the whitespace and comments between tokens, directives, anchors, tags and
// the three kinds of scalar. The fetchers in scannerc.go decide which token
// comes next and queue what these return.

// Eat whitespaces and comments until the next token is found.
func yaml_scanner_scan_to_next_token(scanner *yaml_scanner_t) bool {

	scan_mark := scanner.mark

	// Until the next token is not found.
	for {
		// Allow the BOM mark to start a line.
		if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
			return false
		}
		if scanner.mark.column == 0 && is_bom(scanner.buffer, scanner.buffer_pos) {
			skip(scanner)
		}

		// Eat whitespaces.
		// Tabs are allowed:
		//  - in the flow context
		//  - in the block context, but not at the beginning of the line or
		//  after '-', '?', or ':' (complex value).
		if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
			return false
		}

		for scanner.buffer[scanner.buffer_pos] == ' ' || ((scanner.flow_level > 0 || !scanner.simple_key_allowed) && scanner.buffer[scanner.buffer_pos] == '\t') {
			skip(scanner)
			if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
				return false
			}
		}

		// Check if we just had a line comment under a sequence entry that
		// looks more like a header to the following content. Similar to this:
		//
		// - # The comment
		//   - Some data
		//
		// If so, transform the line comment to a head comment and reposition.
		if len(scanner.comments) > 0 && len(scanner.tokens) > 1 {
			tokenA := scanner.tokens[len(scanner.tokens)-2]
			tokenB := scanner.tokens[len(scanner.tokens)-1]
			comment := &scanner.comments[len(scanner.comments)-1]
			if tokenA.typ == yaml_BLOCK_SEQUENCE_START_TOKEN && tokenB.typ == yaml_BLOCK_ENTRY_TOKEN && len(comment.line) > 0 && !is_break(scanner.buffer, scanner.buffer_pos) {
				// If it was in the prior line, reposition so it becomes a
				// header of the follow up token. Otherwise, keep it in place
				// so it becomes a header of the former.
				comment.head = comment.line
				comment.line = nil
				if comment.start_mark.line == scanner.mark.line-1 {
					comment.token_mark = scanner.mark
				}
			}
		}

		// Eat a comment until a line break.
		if scanner.buffer[scanner.buffer_pos] == '#' {
			if !yaml_scanner_scan_comments(scanner, scan_mark) {
				return false
			}
		}

		// If it is a line break, eat it.
		if is_break(scanner.buffer, scanner.buffer_pos) {
			if scanner.unread < 2 && !yaml_scanner_update_buffer(scanner, 2) {
				return false
			}
			skip_line(scanner)

			// In the block context, a new line may start a simple key.
			if scanner.flow_level == 0 {
				scanner.simple_key_allowed = true
			}
		} else {
			break // We have found a token.
		}
	}

	return true
}

// Scan a YAML-DIRECTIVE or TAG-DIRECTIVE token.
//
// Scope:
//
//	%YAML    1.1    # a comment \n
//	^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
//	%TAG    !yaml!  tag:yaml.org,2002:  \n
//	^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
func yaml_scanner_scan_directive(scanner *yaml_scanner_t, token *yaml_token_t) bool {
	// Eat '%'.
	start_mark := scanner.mark
	skip(scanner)

	// Scan the directive name.
	var name []byte
	if !yaml_scanner_scan_directive_name(scanner, start_mark, &name) {
		return false
	}

	// Is it a YAML directive?
	if bytes.Equal(name, []byte("YAML")) {
		// Scan the VERSION directive value.
		var major, minor int8
		if !yaml_scanner_scan_version_directive_value(scanner, start_mark, &major, &minor) {
			return false
		}
		end_mark := scanner.mark

		// Create a VERSION-DIRECTIVE token.
		*token = yaml_token_t{
			typ:        yaml_VERSION_DIRECTIVE_TOKEN,
			start_mark: start_mark,
			end_mark:   end_mark,
			major:      major,
			minor:      minor,
		}

		// Is it a TAG directive?
	} else if bytes.Equal(name, []byte("TAG")) {
		// Scan the TAG directive value.
		var handle, prefix []byte
		if !yaml_scanner_scan_tag_directive_value(scanner, start_mark, &handle, &prefix) {
			return false
		}
		end_mark := scanner.mark

		// Create a TAG-DIRECTIVE token.
		*token = yaml_token_t{
			typ:        yaml_TAG_DIRECTIVE_TOKEN,
			start_mark: start_mark,
			end_mark:   end_mark,
			value:      handle,
			prefix:     prefix,
		}

		// Unknown directive.
	} else {
		yaml_scanner_set_error(scanner, "while scanning a directive",
			start_mark, "found unknown directive name")
		return false
	}

	// Eat the rest of the line including any comments.
	if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
		return false
	}

	for is_blank(scanner.buffer, scanner.buffer_pos) {
		skip(scanner)
		if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
			return false
		}
	}

	if scanner.buffer[scanner.buffer_pos] == '#' {
		// A directive keeps no comment, so skip it to the line break.
		for !is_breakz(scanner.buffer, scanner.buffer_pos) {
			skip(scanner)
			if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
				return false
			}
		}
	}

	// Check if we are at the end of the line.
	if !is_breakz(scanner.buffer, scanner.buffer_pos) {
		yaml_scanner_set_error(scanner, "while scanning a directive",
			start_mark, "did not find expected comment or line break")
		return false
	}

	// Eat a line break.
	if is_break(scanner.buffer, scanner.buffer_pos) {
		if scanner.unread < 2 && !yaml_scanner_update_buffer(scanner, 2) {
			return false
		}
		skip_line(scanner)
	}

	return true
}

// Scan the directive name.
//
// Scope:
//
//	%YAML   1.1     # a comment \n
//	 ^^^^
//	%TAG    !yaml!  tag:yaml.org,2002:  \n
//	 ^^^
func yaml_scanner_scan_directive_name(scanner *yaml_scanner_t, start_mark yaml_mark_t, name *[]byte) bool {
	// Consume the directive name.
	if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
		return false
	}

	var s []byte
	for is_alpha(scanner.buffer, scanner.buffer_pos) {
		s = read(scanner, s)
		if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
			return false
		}
	}

	// Check if the name is empty.
	if len(s) == 0 {
		yaml_scanner_set_error(scanner, "while scanning a directive",
			start_mark, "could not find expected directive name")
		return false
	}

	// Check for an blank character after the name.
	if !is_blankz(scanner.buffer, scanner.buffer_pos) {
		yaml_scanner_set_error(scanner, "while scanning a directive",
			start_mark, "found unexpected non-alphabetical character")
		return false
	}
	*name = s
	return true
}

// Scan the value of VERSION-DIRECTIVE.
//
// Scope:
//
//	%YAML   1.1     # a comment \n
//	     ^^^^^^
func yaml_scanner_scan_version_directive_value(scanner *yaml_scanner_t, start_mark yaml_mark_t, major, minor *int8) bool {
	// Eat whitespaces.
	if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
		return false
	}
	for is_blank(scanner.buffer, scanner.buffer_pos) {
		skip(scanner)
		if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
			return false
		}
	}

	// Consume the major version number.
	if !yaml_scanner_scan_version_directive_number(scanner, start_mark, major) {
		return false
	}

	// Eat '.'.
	if scanner.buffer[scanner.buffer_pos] != '.' {
		return yaml_scanner_set_error(scanner, "while scanning a %YAML directive",
			start_mark, "did not find expected digit or '.' character")
	}

	skip(scanner)

	// Consume the minor version number.
	if !yaml_scanner_scan_version_directive_number(scanner, start_mark, minor) {
		return false
	}
	return true
}

const max_number_length = 2

// Scan the version number of VERSION-DIRECTIVE.
//
// Scope:
//
//	%YAML   1.1     # a comment \n
//	        ^
//	%YAML   1.1     # a comment \n
//	          ^
func yaml_scanner_scan_version_directive_number(scanner *yaml_scanner_t, start_mark yaml_mark_t, number *int8) bool {

	// Repeat while the next character is digit.
	if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
		return false
	}
	var value, length int
	for is_digit(scanner.buffer, scanner.buffer_pos) {
		// Check if the number is too long.
		length++
		if length > max_number_length {
			return yaml_scanner_set_error(scanner, "while scanning a %YAML directive",
				start_mark, "found extremely long version number")
		}
		value = value*10 + as_digit(scanner.buffer, scanner.buffer_pos)
		skip(scanner)
		if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
			return false
		}
	}

	// Check if the number was present.
	if length == 0 {
		return yaml_scanner_set_error(scanner, "while scanning a %YAML directive",
			start_mark, "did not find expected version number")
	}
	if value > math.MaxInt8 {
		return yaml_scanner_set_error(scanner, "while scanning a %YAML directive",
			start_mark, "version number exceeds supported range")
	}
	*number = int8(value)
	return true
}

// Scan the value of a TAG-DIRECTIVE token.
//
// Scope:
//
//	%TAG    !yaml!  tag:yaml.org,2002:  \n
//	    ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
func yaml_scanner_scan_tag_directive_value(scanner *yaml_scanner_t, start_mark yaml_mark_t, handle, prefix *[]byte) bool {
	var handle_value, prefix_value []byte

	// Eat whitespaces.
	if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
		return false
	}

	for is_blank(scanner.buffer, scanner.buffer_pos) {
		skip(scanner)
		if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
			return false
		}
	}

	// Scan a handle.
	if !yaml_scanner_scan_tag_handle(scanner, true, start_mark, &handle_value) {
		return false
	}

	// Expect a whitespace.
	if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
		return false
	}
	if !is_blank(scanner.buffer, scanner.buffer_pos) {
		yaml_scanner_set_error(scanner, "while scanning a %TAG directive",
			start_mark, "did not find expected whitespace")
		return false
	}

	// Eat whitespaces.
	for is_blank(scanner.buffer, scanner.buffer_pos) {
		skip(scanner)
		if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
			return false
		}
	}

	// Scan a prefix.
	if !yaml_scanner_scan_tag_uri(scanner, true, nil, start_mark, &prefix_value) {
		return false
	}

	// Expect a whitespace or line break.
	if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
		return false
	}
	if !is_blankz(scanner.buffer, scanner.buffer_pos) {
		yaml_scanner_set_error(scanner, "while scanning a %TAG directive",
			start_mark, "did not find expected whitespace or line break")
		return false
	}

	*handle = handle_value
	*prefix = prefix_value
	return true
}

func yaml_scanner_scan_anchor(scanner *yaml_scanner_t, token *yaml_token_t, typ yaml_token_type_t) bool {
	var s []byte

	// Eat the indicator character.
	start_mark := scanner.mark
	skip(scanner)

	// Consume the value.
	if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
		return false
	}

	for is_alpha(scanner.buffer, scanner.buffer_pos) {
		s = read(scanner, s)
		if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
			return false
		}
	}

	end_mark := scanner.mark

	/*
	 * Check if length of the anchor is greater than 0 and it is followed by
	 * a whitespace character or one of the indicators:
	 *
	 *      '?', ':', ',', ']', '}', '%', '@', '`'.
	 */

	if len(s) == 0 ||
		(!is_blankz(scanner.buffer, scanner.buffer_pos) && scanner.buffer[scanner.buffer_pos] != '?' &&
			scanner.buffer[scanner.buffer_pos] != ':' && scanner.buffer[scanner.buffer_pos] != ',' &&
			scanner.buffer[scanner.buffer_pos] != ']' && scanner.buffer[scanner.buffer_pos] != '}' &&
			scanner.buffer[scanner.buffer_pos] != '%' && scanner.buffer[scanner.buffer_pos] != '@' &&
			scanner.buffer[scanner.buffer_pos] != '`') {
		context := "while scanning an alias"
		if typ == yaml_ANCHOR_TOKEN {
			context = "while scanning an anchor"
		}
		yaml_scanner_set_error(scanner, context, start_mark,
			"did not find expected alphabetic or numeric character")
		return false
	}

	// Create a token.
	*token = yaml_token_t{
		typ:        typ,
		start_mark: start_mark,
		end_mark:   end_mark,
		value:      s,
	}

	return true
}

/*
 * Scan a TAG token.
 */

func yaml_scanner_scan_tag(scanner *yaml_scanner_t, token *yaml_token_t) bool {
	var handle, suffix []byte

	start_mark := scanner.mark

	// Check if the tag is in the canonical form.
	if scanner.unread < 2 && !yaml_scanner_update_buffer(scanner, 2) {
		return false
	}

	if scanner.buffer[scanner.buffer_pos+1] == '<' {
		// Keep the handle as ''

		// Eat '!<'
		skip(scanner)
		skip(scanner)

		// Consume the tag value.
		if !yaml_scanner_scan_tag_uri(scanner, false, nil, start_mark, &suffix) {
			return false
		}

		// Check for '>' and eat it.
		if scanner.buffer[scanner.buffer_pos] != '>' {
			yaml_scanner_set_error(scanner, "while scanning a tag",
				start_mark, "did not find the expected '>'")
			return false
		}

		skip(scanner)
	} else {
		// The tag has either the '!suffix' or the '!handle!suffix' form.

		// First, try to scan a handle.
		if !yaml_scanner_scan_tag_handle(scanner, false, start_mark, &handle) {
			return false
		}

		// Check if it is, indeed, handle.
		if handle[0] == '!' && len(handle) > 1 && handle[len(handle)-1] == '!' {
			// Scan the suffix now.
			if !yaml_scanner_scan_tag_uri(scanner, false, nil, start_mark, &suffix) {
				return false
			}
		} else {
			// It wasn't a handle after all.  Scan the rest of the tag.
			if !yaml_scanner_scan_tag_uri(scanner, false, handle, start_mark, &suffix) {
				return false
			}

			// Set the handle to '!'.
			handle = []byte{'!'}

			// A special case: the '!' tag.  Set the handle to '' and the
			// suffix to '!'.
			if len(suffix) == 0 {
				handle, suffix = suffix, handle
			}
		}
	}

	// Check the character which ends the tag.
	if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
		return false
	}
	if !is_blankz(scanner.buffer, scanner.buffer_pos) {
		yaml_scanner_set_error(scanner, "while scanning a tag",
			start_mark, "did not find expected whitespace or line break")
		return false
	}

	end_mark := scanner.mark

	// Create a token.
	*token = yaml_token_t{
		typ:        yaml_TAG_TOKEN,
		start_mark: start_mark,
		end_mark:   end_mark,
		value:      handle,
		suffix:     suffix,
	}
	return true
}

// Scan a tag handle.
func yaml_scanner_scan_tag_handle(scanner *yaml_scanner_t, directive bool, start_mark yaml_mark_t, handle *[]byte) bool {
	// Check the initial '!' character.
	if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
		return false
	}
	if scanner.buffer[scanner.buffer_pos] != '!' {
		yaml_scanner_set_tag_error(scanner, directive,
			start_mark, "did not find expected '!'")
		return false
	}

	var s []byte

	// Copy the '!' character.
	s = read(scanner, s)

	// Copy all subsequent alphabetical and numerical characters.
	if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
		return false
	}
	for is_alpha(scanner.buffer, scanner.buffer_pos) {
		s = read(scanner, s)
		if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
			return false
		}
	}

	// Check if the trailing character is '!' and copy it.
	if scanner.buffer[scanner.buffer_pos] == '!' {
		s = read(scanner, s)
	} else {
		// It's either the '!' tag or not really a tag handle.  If it's a %TAG
		// directive, it's an error.  If it's a tag token, it must be a part of URI.
		if directive && string(s) != "!" {
			yaml_scanner_set_tag_error(scanner, directive,
				start_mark, "did not find expected '!'")
			return false
		}
	}

	*handle = s
	return true
}

// Scan a tag.
func yaml_scanner_scan_tag_uri(scanner *yaml_scanner_t, directive bool, head []byte, start_mark yaml_mark_t, uri *[]byte) bool {
	//size_t length = head ? strlen((char *)head) : 0
	var s []byte
	hasTag := len(head) > 0

	// Copy the head if needed.
	//
	// Note that we don't copy the leading '!' character.
	if len(head) > 1 {
		s = append(s, head[1:]...)
	}

	// Scan the tag.
	if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
		return false
	}

	// The set of characters that may appear in URI is as follows:
	//
	//      '0'-'9', 'A'-'Z', 'a'-'z', '_', '-', ';', '/', '?', ':', '@', '&',
	//      '=', '+', '$', ',', '.', '!', '~', '*', '\'', '(', ')', '[', ']',
	//      '%'.
	// [Go] TODO Convert this into more reasonable logic.
	for is_alpha(scanner.buffer, scanner.buffer_pos) || scanner.buffer[scanner.buffer_pos] == ';' ||
		scanner.buffer[scanner.buffer_pos] == '/' || scanner.buffer[scanner.buffer_pos] == '?' ||
		scanner.buffer[scanner.buffer_pos] == ':' || scanner.buffer[scanner.buffer_pos] == '@' ||
		scanner.buffer[scanner.buffer_pos] == '&' || scanner.buffer[scanner.buffer_pos] == '=' ||
		scanner.buffer[scanner.buffer_pos] == '+' || scanner.buffer[scanner.buffer_pos] == '$' ||
		scanner.buffer[scanner.buffer_pos] == ',' || scanner.buffer[scanner.buffer_pos] == '.' ||
		scanner.buffer[scanner.buffer_pos] == '!' || scanner.buffer[scanner.buffer_pos] == '~' ||
		scanner.buffer[scanner.buffer_pos] == '*' || scanner.buffer[scanner.buffer_pos] == '\'' ||
		scanner.buffer[scanner.buffer_pos] == '(' || scanner.buffer[scanner.buffer_pos] == ')' ||
		scanner.buffer[scanner.buffer_pos] == '[' || scanner.buffer[scanner.buffer_pos] == ']' ||
		scanner.buffer[scanner.buffer_pos] == '%' {
		// Check if it is a URI-escape sequence.
		if scanner.buffer[scanner.buffer_pos] == '%' {
			if !yaml_scanner_scan_uri_escapes(scanner, directive, start_mark, &s) {
				return false
			}
		} else {
			s = read(scanner, s)
		}
		if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
			return false
		}
		hasTag = true
	}

	if !hasTag {
		yaml_scanner_set_tag_error(scanner, directive,
			start_mark, "did not find expected tag URI")
		return false
	}
	*uri = s
	return true
}

// Decode an URI-escape sequence corresponding to a single UTF-8 character.
func yaml_scanner_scan_uri_escapes(scanner *yaml_scanner_t, directive bool, start_mark yaml_mark_t, s *[]byte) bool {

	// Decode the required number of characters.
	w := 1024
	for w > 0 {
		// Check for a URI-escaped octet.
		if scanner.unread < 3 && !yaml_scanner_update_buffer(scanner, 3) {
			return false
		}

		if scanner.buffer[scanner.buffer_pos] != '%' || !is_hex(scanner.buffer, scanner.buffer_pos+1) || !is_hex(scanner.buffer, scanner.buffer_pos+2) {
			return yaml_scanner_set_tag_error(scanner, directive,
				start_mark, "did not find URI escaped octet")
		}

		// Decode the checked hexadecimal pair into one octet.
		var decoded [1]byte
		if _, err := hex.Decode(decoded[:], scanner.buffer[scanner.buffer_pos+1:scanner.buffer_pos+3]); err != nil {
			return yaml_scanner_set_tag_error(scanner, directive,
				start_mark, "did not find URI escaped octet")
		}
		octet := decoded[0]

		// If it is the leading octet, determine the length of the UTF-8 sequence.
		if w == 1024 {
			w = width(octet)
			if w == 0 {
				return yaml_scanner_set_tag_error(scanner, directive,
					start_mark, "found an incorrect leading UTF-8 octet")
			}
		} else {
			// Check if the trailing octet is correct.
			if octet&0xC0 != 0x80 {
				return yaml_scanner_set_tag_error(scanner, directive,
					start_mark, "found an incorrect trailing UTF-8 octet")
			}
		}

		// Copy the octet and move the pointers.
		*s = append(*s, octet)
		skip(scanner)
		skip(scanner)
		skip(scanner)
		w--
	}
	return true
}

// Scan a block scalar.
func yaml_scanner_scan_block_scalar(scanner *yaml_scanner_t, token *yaml_token_t, literal bool) bool {
	// Eat the indicator '|' or '>'.
	start_mark := scanner.mark
	skip(scanner)

	// Scan the additional block scalar indicators.
	if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
		return false
	}

	// Check for a chomping indicator.
	var chomping, increment int
	if scanner.buffer[scanner.buffer_pos] == '+' || scanner.buffer[scanner.buffer_pos] == '-' {
		// Set the chomping method and eat the indicator.
		if scanner.buffer[scanner.buffer_pos] == '+' {
			chomping = +1
		} else {
			chomping = -1
		}
		skip(scanner)

		// Check for an indentation indicator.
		if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
			return false
		}
		if is_digit(scanner.buffer, scanner.buffer_pos) {
			// Check that the indentation is greater than 0.
			if scanner.buffer[scanner.buffer_pos] == '0' {
				yaml_scanner_set_error(scanner, "while scanning a block scalar",
					start_mark, "found an indentation indicator equal to 0")
				return false
			}

			// Get the indentation level and eat the indicator.
			increment = as_digit(scanner.buffer, scanner.buffer_pos)
			skip(scanner)
		}

	} else if is_digit(scanner.buffer, scanner.buffer_pos) {
		// Do the same as above, but in the opposite order.

		if scanner.buffer[scanner.buffer_pos] == '0' {
			yaml_scanner_set_error(scanner, "while scanning a block scalar",
				start_mark, "found an indentation indicator equal to 0")
			return false
		}
		increment = as_digit(scanner.buffer, scanner.buffer_pos)
		skip(scanner)

		if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
			return false
		}
		if scanner.buffer[scanner.buffer_pos] == '+' || scanner.buffer[scanner.buffer_pos] == '-' {
			if scanner.buffer[scanner.buffer_pos] == '+' {
				chomping = +1
			} else {
				chomping = -1
			}
			skip(scanner)
		}
	}

	// Eat whitespaces and comments to the end of the line.
	if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
		return false
	}
	for is_blank(scanner.buffer, scanner.buffer_pos) {
		skip(scanner)
		if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
			return false
		}
	}
	if scanner.buffer[scanner.buffer_pos] == '#' {
		if !yaml_scanner_scan_line_comment(scanner, start_mark) {
			return false
		}
		for !is_breakz(scanner.buffer, scanner.buffer_pos) {
			skip(scanner)
			if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
				return false
			}
		}
	}

	// Check if we are at the end of the line.
	if !is_breakz(scanner.buffer, scanner.buffer_pos) {
		yaml_scanner_set_error(scanner, "while scanning a block scalar",
			start_mark, "did not find expected comment or line break")
		return false
	}

	// Eat a line break.
	if is_break(scanner.buffer, scanner.buffer_pos) {
		if scanner.unread < 2 && !yaml_scanner_update_buffer(scanner, 2) {
			return false
		}
		skip_line(scanner)
	}

	end_mark := scanner.mark

	// Set the indentation level if it was specified.
	var indent int
	if increment > 0 {
		if scanner.indent >= 0 {
			indent = scanner.indent + increment
		} else {
			indent = increment
		}
	}

	// Scan the leading line breaks and determine the indentation level if needed.
	var s, leading_break, trailing_breaks []byte
	if !yaml_scanner_scan_block_scalar_breaks(scanner, &indent, &trailing_breaks, start_mark, &end_mark) {
		return false
	}

	// Scan the block scalar content.
	if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
		return false
	}
	var leading_blank, trailing_blank bool
	for scanner.mark.column == indent && !is_z(scanner.buffer, scanner.buffer_pos) {
		// We are at the beginning of a non-empty line.

		// Is it a trailing whitespace?
		trailing_blank = is_blank(scanner.buffer, scanner.buffer_pos)

		// Check if we need to fold the leading line break.
		if !literal && !leading_blank && !trailing_blank && len(leading_break) > 0 && leading_break[0] == '\n' {
			// Do we need to join the lines by space?
			if len(trailing_breaks) == 0 {
				s = append(s, ' ')
			}
		} else {
			s = append(s, leading_break...)
		}
		leading_break = leading_break[:0]

		// Append the remaining line breaks.
		s = append(s, trailing_breaks...)
		trailing_breaks = trailing_breaks[:0]

		// Is it a leading whitespace?
		leading_blank = is_blank(scanner.buffer, scanner.buffer_pos)

		// Consume the current line.
		for !is_breakz(scanner.buffer, scanner.buffer_pos) {
			s = read(scanner, s)
			if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
				return false
			}
		}

		// Consume the line break.
		if scanner.unread < 2 && !yaml_scanner_update_buffer(scanner, 2) {
			return false
		}

		leading_break = read_line(scanner, leading_break)

		// Eat the following indentation spaces and line breaks.
		if !yaml_scanner_scan_block_scalar_breaks(scanner, &indent, &trailing_breaks, start_mark, &end_mark) {
			return false
		}
	}

	// Chomp the tail.
	if chomping != -1 {
		s = append(s, leading_break...)
	}
	if chomping == 1 {
		s = append(s, trailing_breaks...)
	}

	// Create a token.
	*token = yaml_token_t{
		typ:        yaml_SCALAR_TOKEN,
		start_mark: start_mark,
		end_mark:   end_mark,
		value:      s,
		style:      yaml_LITERAL_SCALAR_STYLE,
	}
	if !literal {
		token.style = yaml_FOLDED_SCALAR_STYLE
	}
	return true
}

// Scan indentation spaces and line breaks for a block scalar.  Determine the
// indentation level if needed.
func yaml_scanner_scan_block_scalar_breaks(scanner *yaml_scanner_t, indent *int, breaks *[]byte, start_mark yaml_mark_t, end_mark *yaml_mark_t) bool {
	*end_mark = scanner.mark

	// Eat the indentation spaces and line breaks.
	max_indent := 0
	for {
		// Eat the indentation spaces.
		if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
			return false
		}
		for (*indent == 0 || scanner.mark.column < *indent) && is_space(scanner.buffer, scanner.buffer_pos) {
			skip(scanner)
			if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
				return false
			}
		}
		if scanner.mark.column > max_indent {
			max_indent = scanner.mark.column
		}

		// Check for a tab character messing the indentation.
		if (*indent == 0 || scanner.mark.column < *indent) && is_tab(scanner.buffer, scanner.buffer_pos) {
			return yaml_scanner_set_error(scanner, "while scanning a block scalar",
				start_mark, "found a tab character where an indentation space is expected")
		}

		// Have we found a non-empty line?
		if !is_break(scanner.buffer, scanner.buffer_pos) {
			break
		}

		// Consume the line break.
		if scanner.unread < 2 && !yaml_scanner_update_buffer(scanner, 2) {
			return false
		}
		// [Go] Should really be returning breaks instead.
		*breaks = read_line(scanner, *breaks)
		*end_mark = scanner.mark
	}

	// Determine the indentation level if needed.
	if *indent == 0 {
		*indent = max_indent
		if *indent < scanner.indent+1 {
			*indent = scanner.indent + 1
		}
		if *indent < 1 {
			*indent = 1
		}
	}
	return true
}

// Scan a quoted scalar.
func yaml_scanner_scan_flow_scalar(scanner *yaml_scanner_t, token *yaml_token_t, single bool) bool {
	// Eat the left quote.
	start_mark := scanner.mark
	skip(scanner)

	// Consume the content of the quoted scalar.
	var s, leading_break, trailing_breaks, whitespaces []byte
	for {
		// Check that there are no document indicators at the beginning of the line.
		if scanner.unread < 4 && !yaml_scanner_update_buffer(scanner, 4) {
			return false
		}

		if scanner.mark.column == 0 &&
			((scanner.buffer[scanner.buffer_pos+0] == '-' &&
				scanner.buffer[scanner.buffer_pos+1] == '-' &&
				scanner.buffer[scanner.buffer_pos+2] == '-') ||
				(scanner.buffer[scanner.buffer_pos+0] == '.' &&
					scanner.buffer[scanner.buffer_pos+1] == '.' &&
					scanner.buffer[scanner.buffer_pos+2] == '.')) &&
			is_blankz(scanner.buffer, scanner.buffer_pos+3) {
			yaml_scanner_set_error(scanner, "while scanning a quoted scalar",
				start_mark, "found unexpected document indicator")
			return false
		}

		// Check for EOF.
		if is_z(scanner.buffer, scanner.buffer_pos) {
			yaml_scanner_set_error(scanner, "while scanning a quoted scalar",
				start_mark, "found unexpected end of stream")
			return false
		}

		// Consume non-blank characters.
		leading_blanks := false
		for !is_blankz(scanner.buffer, scanner.buffer_pos) {
			if single && scanner.buffer[scanner.buffer_pos] == '\'' && scanner.buffer[scanner.buffer_pos+1] == '\'' {
				// Is is an escaped single quote.
				s = append(s, '\'')
				skip(scanner)
				skip(scanner)

			} else if single && scanner.buffer[scanner.buffer_pos] == '\'' {
				// It is a right single quote.
				break
			} else if !single && scanner.buffer[scanner.buffer_pos] == '"' {
				// It is a right double quote.
				break

			} else if !single && scanner.buffer[scanner.buffer_pos] == '\\' && is_break(scanner.buffer, scanner.buffer_pos+1) {
				// It is an escaped line break.
				if scanner.unread < 3 && !yaml_scanner_update_buffer(scanner, 3) {
					return false
				}
				skip(scanner)
				skip_line(scanner)
				leading_blanks = true
				break

			} else if !single && scanner.buffer[scanner.buffer_pos] == '\\' {
				// It is an escape sequence.
				code_length := 0

				// Check the escape character.
				switch scanner.buffer[scanner.buffer_pos+1] {
				case '0':
					s = append(s, 0)
				case 'a':
					s = append(s, '\x07')
				case 'b':
					s = append(s, '\x08')
				case 't', '\t':
					s = append(s, '\x09')
				case 'n':
					s = append(s, '\x0A')
				case 'v':
					s = append(s, '\x0B')
				case 'f':
					s = append(s, '\x0C')
				case 'r':
					s = append(s, '\x0D')
				case 'e':
					s = append(s, '\x1B')
				case ' ':
					s = append(s, '\x20')
				case '"':
					s = append(s, '"')
				case '\'':
					s = append(s, '\'')
				case '\\':
					s = append(s, '\\')
				case 'N': // NEL (#x85)
					s = append(s, '\xC2')
					s = append(s, '\x85')
				case '_': // #xA0
					s = append(s, '\xC2')
					s = append(s, '\xA0')
				case 'L': // LS (#x2028)
					s = append(s, '\xE2')
					s = append(s, '\x80')
					s = append(s, '\xA8')
				case 'P': // PS (#x2029)
					s = append(s, '\xE2')
					s = append(s, '\x80')
					s = append(s, '\xA9')
				case 'x':
					code_length = 2
				case 'u':
					code_length = 4
				case 'U':
					code_length = 8
				default:
					yaml_scanner_set_error(scanner, "while parsing a quoted scalar",
						start_mark, "found unknown escape character")
					return false
				}

				skip(scanner)
				skip(scanner)

				// Consume an arbitrary escape code.
				if code_length > 0 {
					var value int

					// Scan the character value.
					if scanner.unread < code_length && !yaml_scanner_update_buffer(scanner, code_length) {
						return false
					}
					for k := 0; k < code_length; k++ {
						if !is_hex(scanner.buffer, scanner.buffer_pos+k) {
							yaml_scanner_set_error(scanner, "while parsing a quoted scalar",
								start_mark, "did not find expected hexdecimal number")
							return false
						}
						value = (value << 4) + as_hex(scanner.buffer, scanner.buffer_pos+k)
					}

					// Check the value and write the character.
					if (value >= 0xD800 && value <= 0xDFFF) || value > 0x10FFFF {
						yaml_scanner_set_error(scanner, "while parsing a quoted scalar",
							start_mark, "found invalid Unicode character escape code")
						return false
					}
					if value <= 0x7F {
						s = append(s, byte(value))
					} else if value <= 0x7FF {
						s = append(s, byte(0xC0+(value>>6)))
						s = append(s, byte(0x80+(value&0x3F)))
					} else if value <= 0xFFFF {
						s = append(s, byte(0xE0+(value>>12)))
						s = append(s, byte(0x80+((value>>6)&0x3F)))
						s = append(s, byte(0x80+(value&0x3F)))
					} else {
						s = append(s, byte(0xF0+(value>>18)))
						s = append(s, byte(0x80+((value>>12)&0x3F)))
						s = append(s, byte(0x80+((value>>6)&0x3F)))
						s = append(s, byte(0x80+(value&0x3F)))
					}

					// Advance the pointer.
					for k := 0; k < code_length; k++ {
						skip(scanner)
					}
				}
			} else {
				// It is a non-escaped non-blank character.
				s = read(scanner, s)
			}
			if scanner.unread < 2 && !yaml_scanner_update_buffer(scanner, 2) {
				return false
			}
		}

		if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
			return false
		}

		// Check if we are at the end of the scalar.
		if single {
			if scanner.buffer[scanner.buffer_pos] == '\'' {
				break
			}
		} else {
			if scanner.buffer[scanner.buffer_pos] == '"' {
				break
			}
		}

		// Consume blank characters.
		for is_blank(scanner.buffer, scanner.buffer_pos) || is_break(scanner.buffer, scanner.buffer_pos) {
			if is_blank(scanner.buffer, scanner.buffer_pos) {
				// Consume a space or a tab character.
				if !leading_blanks {
					whitespaces = read(scanner, whitespaces)
				} else {
					skip(scanner)
				}
			} else {
				if scanner.unread < 2 && !yaml_scanner_update_buffer(scanner, 2) {
					return false
				}

				// Check if it is a first line break.
				if !leading_blanks {
					whitespaces = whitespaces[:0]
					leading_break = read_line(scanner, leading_break)
					leading_blanks = true
				} else {
					trailing_breaks = read_line(scanner, trailing_breaks)
				}
			}
			if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
				return false
			}
		}

		// Join the whitespaces or fold line breaks.
		if leading_blanks {
			// Do we need to fold line breaks?
			if len(leading_break) > 0 && leading_break[0] == '\n' {
				if len(trailing_breaks) == 0 {
					s = append(s, ' ')
				} else {
					s = append(s, trailing_breaks...)
				}
			} else {
				s = append(s, leading_break...)
				s = append(s, trailing_breaks...)
			}
			trailing_breaks = trailing_breaks[:0]
			leading_break = leading_break[:0]
		} else {
			s = append(s, whitespaces...)
			whitespaces = whitespaces[:0]
		}
	}

	// Eat the right quote.
	skip(scanner)
	end_mark := scanner.mark

	// Create a token.
	*token = yaml_token_t{
		typ:        yaml_SCALAR_TOKEN,
		start_mark: start_mark,
		end_mark:   end_mark,
		value:      s,
		style:      yaml_SINGLE_QUOTED_SCALAR_STYLE,
	}
	if !single {
		token.style = yaml_DOUBLE_QUOTED_SCALAR_STYLE
	}
	return true
}

// Scan a plain scalar.
func yaml_scanner_scan_plain_scalar(scanner *yaml_scanner_t, token *yaml_token_t) bool {

	var s, leading_break, trailing_breaks, whitespaces []byte
	var leading_blanks bool
	var indent = scanner.indent + 1

	start_mark := scanner.mark
	end_mark := scanner.mark

	// Consume the content of the plain scalar.
	for {
		// Check for a document indicator.
		if scanner.unread < 4 && !yaml_scanner_update_buffer(scanner, 4) {
			return false
		}
		if scanner.mark.column == 0 &&
			((scanner.buffer[scanner.buffer_pos+0] == '-' &&
				scanner.buffer[scanner.buffer_pos+1] == '-' &&
				scanner.buffer[scanner.buffer_pos+2] == '-') ||
				(scanner.buffer[scanner.buffer_pos+0] == '.' &&
					scanner.buffer[scanner.buffer_pos+1] == '.' &&
					scanner.buffer[scanner.buffer_pos+2] == '.')) &&
			is_blankz(scanner.buffer, scanner.buffer_pos+3) {
			break
		}

		// Check for a comment.
		if scanner.buffer[scanner.buffer_pos] == '#' {
			break
		}

		// Consume non-blank characters.
		for !is_blankz(scanner.buffer, scanner.buffer_pos) {

			// Check for indicators that may end a plain scalar.
			if (scanner.buffer[scanner.buffer_pos] == ':' && is_blankz(scanner.buffer, scanner.buffer_pos+1)) ||
				(scanner.flow_level > 0 &&
					(scanner.buffer[scanner.buffer_pos] == ',' ||
						scanner.buffer[scanner.buffer_pos] == '?' || scanner.buffer[scanner.buffer_pos] == '[' ||
						scanner.buffer[scanner.buffer_pos] == ']' || scanner.buffer[scanner.buffer_pos] == '{' ||
						scanner.buffer[scanner.buffer_pos] == '}')) {
				break
			}

			// Check if we need to join whitespaces and breaks.
			if leading_blanks || len(whitespaces) > 0 {
				if leading_blanks {
					// Do we need to fold line breaks?
					if leading_break[0] == '\n' {
						if len(trailing_breaks) == 0 {
							s = append(s, ' ')
						} else {
							s = append(s, trailing_breaks...)
						}
					} else {
						s = append(s, leading_break...)
						s = append(s, trailing_breaks...)
					}
					trailing_breaks = trailing_breaks[:0]
					leading_break = leading_break[:0]
					leading_blanks = false
				} else {
					s = append(s, whitespaces...)
					whitespaces = whitespaces[:0]
				}
			}

			// Copy the character.
			s = read(scanner, s)

			end_mark = scanner.mark
			if scanner.unread < 2 && !yaml_scanner_update_buffer(scanner, 2) {
				return false
			}
		}

		// Is it the end?
		if !is_blank(scanner.buffer, scanner.buffer_pos) && !is_break(scanner.buffer, scanner.buffer_pos) {
			break
		}

		// Consume blank characters.
		if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
			return false
		}

		for is_blank(scanner.buffer, scanner.buffer_pos) || is_break(scanner.buffer, scanner.buffer_pos) {
			if is_blank(scanner.buffer, scanner.buffer_pos) {

				// Check for tab characters that abuse indentation.
				if leading_blanks && scanner.mark.column < indent && is_tab(scanner.buffer, scanner.buffer_pos) {
					yaml_scanner_set_error(scanner, "while scanning a plain scalar",
						start_mark, "found a tab character that violates indentation")
					return false
				}

				// Consume a space or a tab character.
				if !leading_blanks {
					whitespaces = read(scanner, whitespaces)
				} else {
					skip(scanner)
				}
			} else {
				if scanner.unread < 2 && !yaml_scanner_update_buffer(scanner, 2) {
					return false
				}

				// Check if it is a first line break.
				if !leading_blanks {
					whitespaces = whitespaces[:0]
					leading_break = read_line(scanner, leading_break)
					leading_blanks = true
				} else {
					trailing_breaks = read_line(scanner, trailing_breaks)
				}
			}
			if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
				return false
			}
		}

		// Check indentation level.
		if scanner.flow_level == 0 && scanner.mark.column < indent {
			break
		}
	}

	// Create a token.
	*token = yaml_token_t{
		typ:        yaml_SCALAR_TOKEN,
		start_mark: start_mark,
		end_mark:   end_mark,
		value:      s,
		style:      yaml_PLAIN_SCALAR_STYLE,
	}

	// Note that we change the 'simple_key_allowed' flag.
	if leading_blanks {
		scanner.simple_key_allowed = true
	}
	return true
}

func yaml_scanner_scan_line_comment(scanner *yaml_scanner_t, token_mark yaml_mark_t) bool {
	if scanner.newlines > 0 {
		return true
	}

	var start_mark yaml_mark_t
	var text []byte

	for peek := 0; peek < 512; peek++ {
		if scanner.unread < peek+1 && !yaml_scanner_update_buffer(scanner, peek+1) {
			break
		}
		if is_blank(scanner.buffer, scanner.buffer_pos+peek) {
			continue
		}
		if scanner.buffer[scanner.buffer_pos+peek] == '#' {
			seen := scanner.mark.index + peek
			for {
				if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
					return false
				}
				if is_breakz(scanner.buffer, scanner.buffer_pos) {
					if scanner.mark.index >= seen {
						break
					}
					if scanner.unread < 2 && !yaml_scanner_update_buffer(scanner, 2) {
						return false
					}
					skip_line(scanner)
				} else if scanner.mark.index >= seen {
					if len(text) == 0 {
						start_mark = scanner.mark
					}
					text = read(scanner, text)
				} else {
					skip(scanner)
				}
			}
		}
		break
	}
	if len(text) > 0 {
		scanner.comments = append(scanner.comments, yaml_comment_t{
			token_mark: token_mark,
			start_mark: start_mark,
			line:       text,
		})
	}
	return true
}

func yaml_scanner_scan_comments(scanner *yaml_scanner_t, scan_mark yaml_mark_t) bool {
	token := scanner.tokens[len(scanner.tokens)-1]

	if token.typ == yaml_FLOW_ENTRY_TOKEN && len(scanner.tokens) > 1 {
		token = scanner.tokens[len(scanner.tokens)-2]
	}

	var token_mark = token.start_mark
	var start_mark yaml_mark_t
	var next_indent = scanner.indent
	if next_indent < 0 {
		next_indent = 0
	}

	var recent_empty = false
	var first_empty = scanner.newlines <= 1

	var line = scanner.mark.line
	var column = scanner.mark.column

	var text []byte

	// The foot line is the place where a comment must start to
	// still be considered as a foot of the prior content.
	// If there's some content in the currently parsed line, then
	// the foot is the line below it.
	var foot_line = -1
	if scan_mark.line > 0 {
		foot_line = scanner.mark.line - scanner.newlines + 1
		if scanner.newlines == 0 && scanner.mark.column > 1 {
			foot_line++
		}
	}

	var peek = 0
	for ; peek < 512; peek++ {
		if scanner.unread < peek+1 && !yaml_scanner_update_buffer(scanner, peek+1) {
			break
		}
		column++
		if is_blank(scanner.buffer, scanner.buffer_pos+peek) {
			continue
		}
		c := scanner.buffer[scanner.buffer_pos+peek]
		var close_flow = scanner.flow_level > 0 && (c == ']' || c == '}')
		if close_flow || is_breakz(scanner.buffer, scanner.buffer_pos+peek) {
			// Got line break or terminator.
			if close_flow || !recent_empty {
				if close_flow || first_empty && (start_mark.line == foot_line && token.typ != yaml_VALUE_TOKEN || start_mark.column-1 < next_indent) {
					// This is the first empty line and there were no empty lines before,
					// so this initial part of the comment is a foot of the prior token
					// instead of being a head for the following one. Split it up.
					// Alternatively, this might also be the last comment inside a flow
					// scope, so it must be a footer.
					if len(text) > 0 {
						if start_mark.column-1 < next_indent {
							// If dedented it's unrelated to the prior token.
							token_mark = start_mark
						}
						scanner.comments = append(scanner.comments, yaml_comment_t{
							scan_mark:  scan_mark,
							token_mark: token_mark,
							start_mark: start_mark,
							end_mark:   yaml_mark_t{scanner.mark.index + peek, line, column},
							foot:       text,
						})
						scan_mark = yaml_mark_t{scanner.mark.index + peek, line, column}
						token_mark = scan_mark
						text = nil
					}
				} else {
					if len(text) > 0 && scanner.buffer[scanner.buffer_pos+peek] != 0 {
						text = append(text, '\n')
					}
				}
			}
			if !is_break(scanner.buffer, scanner.buffer_pos+peek) {
				break
			}
			first_empty = false
			recent_empty = true
			column = 0
			line++
			continue
		}

		if len(text) > 0 && (close_flow || column-1 < next_indent && column != start_mark.column) {
			// The comment at the different indentation is a foot of the
			// preceding data rather than a head of the upcoming one.
			scanner.comments = append(scanner.comments, yaml_comment_t{
				scan_mark:  scan_mark,
				token_mark: token_mark,
				start_mark: start_mark,
				end_mark:   yaml_mark_t{scanner.mark.index + peek, line, column},
				foot:       text,
			})
			scan_mark = yaml_mark_t{scanner.mark.index + peek, line, column}
			token_mark = scan_mark
			text = nil
		}

		if scanner.buffer[scanner.buffer_pos+peek] != '#' {
			break
		}

		if len(text) == 0 {
			start_mark = yaml_mark_t{scanner.mark.index + peek, line, column}
		} else {
			text = append(text, '\n')
		}

		recent_empty = false

		// Consume until after the consumed comment line.
		seen := scanner.mark.index + peek
		for {
			if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
				return false
			}
			if is_breakz(scanner.buffer, scanner.buffer_pos) {
				if scanner.mark.index >= seen {
					break
				}
				if scanner.unread < 2 && !yaml_scanner_update_buffer(scanner, 2) {
					return false
				}
				skip_line(scanner)
			} else if scanner.mark.index >= seen {
				text = read(scanner, text)
			} else {
				skip(scanner)
			}
		}

		peek = 0
		column = 0
		line = scanner.mark.line
		next_indent = scanner.indent
		if next_indent < 0 {
			next_indent = 0
		}
	}

	if len(text) > 0 {
		scanner.comments = append(scanner.comments, yaml_comment_t{
			scan_mark:  scan_mark,
			token_mark: start_mark,
			start_mark: start_mark,
			end_mark:   yaml_mark_t{scanner.mark.index + peek - 1, line, column},
			head:       text,
		})
	}
	return true
}
