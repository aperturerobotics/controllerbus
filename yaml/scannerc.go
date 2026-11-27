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
	"fmt"
)

// Introduction
// ************
//
// The following notes assume that you are familiar with the YAML specification
// (http://yaml.org/spec/1.2/spec.html).  We mostly follow it, although in
// some cases we are less restrictive that it requires.
//
// The process of transforming a YAML stream into a sequence of events is
// divided on two steps: Scanning and Parsing.
//
// The Scanner transforms the input stream into a sequence of tokens, while the
// parser transform the sequence of tokens produced by the Scanner into a
// sequence of parsing events.
//
// The Scanner is rather clever and complicated. The Parser, on the contrary,
// is a straightforward implementation of a recursive-descendant parser (or,
// LL(1) parser, as it is usually called).
//
// Actually there are two issues of Scanning that might be called "clever", the
// rest is quite straightforward.  The issues are "block collection start" and
// "simple keys".  Both issues are explained below in details.
//
// Here the Scanning step is explained and implemented.  We start with the list
// of all the tokens produced by the Scanner together with short descriptions.
//
// Now, tokens:
//
//      STREAM-START(encoding)          # The stream start.
//      STREAM-END                      # The stream end.
//      VERSION-DIRECTIVE(major,minor)  # The '%YAML' directive.
//      TAG-DIRECTIVE(handle,prefix)    # The '%TAG' directive.
//      DOCUMENT-START                  # '---'
//      DOCUMENT-END                    # '...'
//      BLOCK-SEQUENCE-START            # Indentation increase denoting a block
//      BLOCK-MAPPING-START             # sequence or a block mapping.
//      BLOCK-END                       # Indentation decrease.
//      FLOW-SEQUENCE-START             # '['
//      FLOW-SEQUENCE-END               # ']'
//      BLOCK-SEQUENCE-START            # '{'
//      BLOCK-SEQUENCE-END              # '}'
//      BLOCK-ENTRY                     # '-'
//      FLOW-ENTRY                      # ','
//      KEY                             # '?' or nothing (simple keys).
//      VALUE                           # ':'
//      ALIAS(anchor)                   # '*anchor'
//      ANCHOR(anchor)                  # '&anchor'
//      TAG(handle,suffix)              # '!handle!suffix'
//      SCALAR(value,style)             # A scalar.
//
// The following two tokens are "virtual" tokens denoting the beginning and the
// end of the stream:
//
//      STREAM-START(encoding)
//      STREAM-END
//
// We pass the information about the input stream encoding with the
// STREAM-START token.
//
// The next two tokens are responsible for tags:
//
//      VERSION-DIRECTIVE(major,minor)
//      TAG-DIRECTIVE(handle,prefix)
//
// Example:
//
//      %YAML   1.1
//      %TAG    !   !foo
//      %TAG    !yaml!  tag:yaml.org,2002:
//      ---
//
// The correspoding sequence of tokens:
//
//      STREAM-START(utf-8)
//      VERSION-DIRECTIVE(1,1)
//      TAG-DIRECTIVE("!","!foo")
//      TAG-DIRECTIVE("!yaml","tag:yaml.org,2002:")
//      DOCUMENT-START
//      STREAM-END
//
// Note that the VERSION-DIRECTIVE and TAG-DIRECTIVE tokens occupy a whole
// line.
//
// The document start and end indicators are represented by:
//
//      DOCUMENT-START
//      DOCUMENT-END
//
// Note that if a YAML stream contains an implicit document (without '---'
// and '...' indicators), no DOCUMENT-START and DOCUMENT-END tokens will be
// produced.
//
// In the following examples, we present whole documents together with the
// produced tokens.
//
//      1. An implicit document:
//
//          'a scalar'
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          SCALAR("a scalar",single-quoted)
//          STREAM-END
//
//      2. An explicit document:
//
//          ---
//          'a scalar'
//          ...
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          DOCUMENT-START
//          SCALAR("a scalar",single-quoted)
//          DOCUMENT-END
//          STREAM-END
//
//      3. Several documents in a stream:
//
//          'a scalar'
//          ---
//          'another scalar'
//          ---
//          'yet another scalar'
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          SCALAR("a scalar",single-quoted)
//          DOCUMENT-START
//          SCALAR("another scalar",single-quoted)
//          DOCUMENT-START
//          SCALAR("yet another scalar",single-quoted)
//          STREAM-END
//
// We have already introduced the SCALAR token above.  The following tokens are
// used to describe aliases, anchors, tag, and scalars:
//
//      ALIAS(anchor)
//      ANCHOR(anchor)
//      TAG(handle,suffix)
//      SCALAR(value,style)
//
// The following series of examples illustrate the usage of these tokens:
//
//      1. A recursive sequence:
//
//          &A [ *A ]
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          ANCHOR("A")
//          FLOW-SEQUENCE-START
//          ALIAS("A")
//          FLOW-SEQUENCE-END
//          STREAM-END
//
//      2. A tagged scalar:
//
//          !!float "3.14"  # A good approximation.
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          TAG("!!","float")
//          SCALAR("3.14",double-quoted)
//          STREAM-END
//
//      3. Various scalar styles:
//
//          --- # Implicit empty plain scalars do not produce tokens.
//          --- a plain scalar
//          --- 'a single-quoted scalar'
//          --- "a double-quoted scalar"
//          --- |-
//            a literal scalar
//          --- >-
//            a folded
//            scalar
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          DOCUMENT-START
//          DOCUMENT-START
//          SCALAR("a plain scalar",plain)
//          DOCUMENT-START
//          SCALAR("a single-quoted scalar",single-quoted)
//          DOCUMENT-START
//          SCALAR("a double-quoted scalar",double-quoted)
//          DOCUMENT-START
//          SCALAR("a literal scalar",literal)
//          DOCUMENT-START
//          SCALAR("a folded scalar",folded)
//          STREAM-END
//
// Now it's time to review collection-related tokens. We will start with
// flow collections:
//
//      FLOW-SEQUENCE-START
//      FLOW-SEQUENCE-END
//      FLOW-MAPPING-START
//      FLOW-MAPPING-END
//      FLOW-ENTRY
//      KEY
//      VALUE
//
// The tokens FLOW-SEQUENCE-START, FLOW-SEQUENCE-END, FLOW-MAPPING-START, and
// FLOW-MAPPING-END represent the indicators '[', ']', '{', and '}'
// correspondingly.  FLOW-ENTRY represent the ',' indicator.  Finally the
// indicators '?' and ':', which are used for denoting mapping keys and values,
// are represented by the KEY and VALUE tokens.
//
// The following examples show flow collections:
//
//      1. A flow sequence:
//
//          [item 1, item 2, item 3]
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          FLOW-SEQUENCE-START
//          SCALAR("item 1",plain)
//          FLOW-ENTRY
//          SCALAR("item 2",plain)
//          FLOW-ENTRY
//          SCALAR("item 3",plain)
//          FLOW-SEQUENCE-END
//          STREAM-END
//
//      2. A flow mapping:
//
//          {
//              a simple key: a value,  # Note that the KEY token is produced.
//              ? a complex key: another value,
//          }
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          FLOW-MAPPING-START
//          KEY
//          SCALAR("a simple key",plain)
//          VALUE
//          SCALAR("a value",plain)
//          FLOW-ENTRY
//          KEY
//          SCALAR("a complex key",plain)
//          VALUE
//          SCALAR("another value",plain)
//          FLOW-ENTRY
//          FLOW-MAPPING-END
//          STREAM-END
//
// A simple key is a key which is not denoted by the '?' indicator.  Note that
// the Scanner still produce the KEY token whenever it encounters a simple key.
//
// For scanning block collections, the following tokens are used (note that we
// repeat KEY and VALUE here):
//
//      BLOCK-SEQUENCE-START
//      BLOCK-MAPPING-START
//      BLOCK-END
//      BLOCK-ENTRY
//      KEY
//      VALUE
//
// The tokens BLOCK-SEQUENCE-START and BLOCK-MAPPING-START denote indentation
// increase that precedes a block collection (cf. the INDENT token in Python).
// The token BLOCK-END denote indentation decrease that ends a block collection
// (cf. the DEDENT token in Python).  However YAML has some syntax pecularities
// that makes detections of these tokens more complex.
//
// The tokens BLOCK-ENTRY, KEY, and VALUE are used to represent the indicators
// '-', '?', and ':' correspondingly.
//
// The following examples show how the tokens BLOCK-SEQUENCE-START,
// BLOCK-MAPPING-START, and BLOCK-END are emitted by the Scanner:
//
//      1. Block sequences:
//
//          - item 1
//          - item 2
//          -
//            - item 3.1
//            - item 3.2
//          -
//            key 1: value 1
//            key 2: value 2
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          BLOCK-SEQUENCE-START
//          BLOCK-ENTRY
//          SCALAR("item 1",plain)
//          BLOCK-ENTRY
//          SCALAR("item 2",plain)
//          BLOCK-ENTRY
//          BLOCK-SEQUENCE-START
//          BLOCK-ENTRY
//          SCALAR("item 3.1",plain)
//          BLOCK-ENTRY
//          SCALAR("item 3.2",plain)
//          BLOCK-END
//          BLOCK-ENTRY
//          BLOCK-MAPPING-START
//          KEY
//          SCALAR("key 1",plain)
//          VALUE
//          SCALAR("value 1",plain)
//          KEY
//          SCALAR("key 2",plain)
//          VALUE
//          SCALAR("value 2",plain)
//          BLOCK-END
//          BLOCK-END
//          STREAM-END
//
//      2. Block mappings:
//
//          a simple key: a value   # The KEY token is produced here.
//          ? a complex key
//          : another value
//          a mapping:
//            key 1: value 1
//            key 2: value 2
//          a sequence:
//            - item 1
//            - item 2
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          BLOCK-MAPPING-START
//          KEY
//          SCALAR("a simple key",plain)
//          VALUE
//          SCALAR("a value",plain)
//          KEY
//          SCALAR("a complex key",plain)
//          VALUE
//          SCALAR("another value",plain)
//          KEY
//          SCALAR("a mapping",plain)
//          BLOCK-MAPPING-START
//          KEY
//          SCALAR("key 1",plain)
//          VALUE
//          SCALAR("value 1",plain)
//          KEY
//          SCALAR("key 2",plain)
//          VALUE
//          SCALAR("value 2",plain)
//          BLOCK-END
//          KEY
//          SCALAR("a sequence",plain)
//          VALUE
//          BLOCK-SEQUENCE-START
//          BLOCK-ENTRY
//          SCALAR("item 1",plain)
//          BLOCK-ENTRY
//          SCALAR("item 2",plain)
//          BLOCK-END
//          BLOCK-END
//          STREAM-END
//
// YAML does not always require to start a new block collection from a new
// line.  If the current line contains only '-', '?', and ':' indicators, a new
// block collection may start at the current line.  The following examples
// illustrate this case:
//
//      1. Collections in a sequence:
//
//          - - item 1
//            - item 2
//          - key 1: value 1
//            key 2: value 2
//          - ? complex key
//            : complex value
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          BLOCK-SEQUENCE-START
//          BLOCK-ENTRY
//          BLOCK-SEQUENCE-START
//          BLOCK-ENTRY
//          SCALAR("item 1",plain)
//          BLOCK-ENTRY
//          SCALAR("item 2",plain)
//          BLOCK-END
//          BLOCK-ENTRY
//          BLOCK-MAPPING-START
//          KEY
//          SCALAR("key 1",plain)
//          VALUE
//          SCALAR("value 1",plain)
//          KEY
//          SCALAR("key 2",plain)
//          VALUE
//          SCALAR("value 2",plain)
//          BLOCK-END
//          BLOCK-ENTRY
//          BLOCK-MAPPING-START
//          KEY
//          SCALAR("complex key")
//          VALUE
//          SCALAR("complex value")
//          BLOCK-END
//          BLOCK-END
//          STREAM-END
//
//      2. Collections in a mapping:
//
//          ? a sequence
//          : - item 1
//            - item 2
//          ? a mapping
//          : key 1: value 1
//            key 2: value 2
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          BLOCK-MAPPING-START
//          KEY
//          SCALAR("a sequence",plain)
//          VALUE
//          BLOCK-SEQUENCE-START
//          BLOCK-ENTRY
//          SCALAR("item 1",plain)
//          BLOCK-ENTRY
//          SCALAR("item 2",plain)
//          BLOCK-END
//          KEY
//          SCALAR("a mapping",plain)
//          VALUE
//          BLOCK-MAPPING-START
//          KEY
//          SCALAR("key 1",plain)
//          VALUE
//          SCALAR("value 1",plain)
//          KEY
//          SCALAR("key 2",plain)
//          VALUE
//          SCALAR("value 2",plain)
//          BLOCK-END
//          BLOCK-END
//          STREAM-END
//
// YAML also permits non-indented sequences if they are included into a block
// mapping.  In this case, the token BLOCK-SEQUENCE-START is not produced:
//
//      key:
//      - item 1    # BLOCK-SEQUENCE-START is NOT produced here.
//      - item 2
//
// Tokens:
//
//      STREAM-START(utf-8)
//      BLOCK-MAPPING-START
//      KEY
//      SCALAR("key",plain)
//      VALUE
//      BLOCK-ENTRY
//      SCALAR("item 1",plain)
//      BLOCK-ENTRY
//      SCALAR("item 2",plain)
//      BLOCK-END
//

// Advance the buffer pointer.
func skip(scanner *yaml_scanner_t) {
	if !is_blank(scanner.buffer, scanner.buffer_pos) {
		scanner.newlines = 0
	}
	scanner.mark.index++
	scanner.mark.column++
	scanner.unread--
	scanner.buffer_pos += width(scanner.buffer[scanner.buffer_pos])
}

func skip_line(scanner *yaml_scanner_t) {
	if is_crlf(scanner.buffer, scanner.buffer_pos) {
		scanner.mark.index += 2
		scanner.mark.column = 0
		scanner.mark.line++
		scanner.unread -= 2
		scanner.buffer_pos += 2
		scanner.newlines++
	} else if is_break(scanner.buffer, scanner.buffer_pos) {
		scanner.mark.index++
		scanner.mark.column = 0
		scanner.mark.line++
		scanner.unread--
		scanner.buffer_pos += width(scanner.buffer[scanner.buffer_pos])
		scanner.newlines++
	}
}

// Copy a character to a string buffer and advance pointers.
func read(scanner *yaml_scanner_t, s []byte) []byte {
	if !is_blank(scanner.buffer, scanner.buffer_pos) {
		scanner.newlines = 0
	}
	w := width(scanner.buffer[scanner.buffer_pos])
	if w == 0 {
		panic("invalid character sequence")
	}
	if len(s) == 0 {
		s = make([]byte, 0, 32)
	}
	if w == 1 && len(s)+w <= cap(s) {
		s = s[:len(s)+1]
		s[len(s)-1] = scanner.buffer[scanner.buffer_pos]
		scanner.buffer_pos++
	} else {
		s = append(s, scanner.buffer[scanner.buffer_pos:scanner.buffer_pos+w]...)
		scanner.buffer_pos += w
	}
	scanner.mark.index++
	scanner.mark.column++
	scanner.unread--
	return s
}

// Copy a line break character to a string buffer and advance pointers.
func read_line(scanner *yaml_scanner_t, s []byte) []byte {
	buf := scanner.buffer
	pos := scanner.buffer_pos
	switch {
	case buf[pos] == '\r' && buf[pos+1] == '\n':
		// CR LF . LF
		s = append(s, '\n')
		scanner.buffer_pos += 2
		scanner.mark.index++
		scanner.unread--
	case buf[pos] == '\r' || buf[pos] == '\n':
		// CR|LF . LF
		s = append(s, '\n')
		scanner.buffer_pos += 1
	case buf[pos] == '\xC2' && buf[pos+1] == '\x85':
		// NEL . LF
		s = append(s, '\n')
		scanner.buffer_pos += 2
	case buf[pos] == '\xE2' && buf[pos+1] == '\x80' && (buf[pos+2] == '\xA8' || buf[pos+2] == '\xA9'):
		// LS|PS . LS|PS
		s = append(s, buf[scanner.buffer_pos:pos+3]...)
		scanner.buffer_pos += 3
	default:
		return s
	}
	scanner.mark.index++
	scanner.mark.column = 0
	scanner.mark.line++
	scanner.unread--
	scanner.newlines++
	return s
}

// Set the scanner error and return false.
func yaml_scanner_set_error(scanner *yaml_scanner_t, context string, context_mark yaml_mark_t, problem string) bool {
	scanner.error = yaml_SCANNER_ERROR
	scanner.context = context
	scanner.context_mark = context_mark
	scanner.problem = problem
	scanner.problem_mark = scanner.mark
	return false
}

func yaml_scanner_set_tag_error(scanner *yaml_scanner_t, directive bool, context_mark yaml_mark_t, problem string) bool {
	context := "while parsing a tag"
	if directive {
		context = "while parsing a %TAG directive"
	}
	return yaml_scanner_set_error(scanner, context, context_mark, problem)
}

// yaml_scanner_take_comments removes the queued comments that precede token
// and returns their head, line and foot texts, each joined by line breaks. A
// head comment before a BLOCK-END stays queued for the token after it.
func yaml_scanner_take_comments(scanner *yaml_scanner_t, token *yaml_token_t) (head, line, foot []byte) {
	for scanner.comments_head < len(scanner.comments) && token.start_mark.index >= scanner.comments[scanner.comments_head].token_mark.index {
		comment := &scanner.comments[scanner.comments_head]
		if len(comment.head) > 0 {
			if token.typ == yaml_BLOCK_END_TOKEN {
				// No heads on ends, so keep comment.head for a follow up token.
				break
			}
			head = join_comment(head, comment.head)
		}
		foot = join_comment(foot, comment.foot)
		line = join_comment(line, comment.line)
		*comment = yaml_comment_t{}
		scanner.comments_head++
	}
	return head, line, foot
}

// yaml_scanner_skip_token removes the head token of the queue, which
// yaml_scanner_fetch_more_tokens made available.
func yaml_scanner_skip_token(scanner *yaml_scanner_t) {
	scanner.token_available = false
	scanner.tokens_parsed++
	scanner.tokens_head++
}

// yaml_insert_token inserts token into the queue pos places after its head,
// or appends it when pos is negative. It first moves the queue to the front
// of its array when the array is full and the head has moved.
func yaml_insert_token(scanner *yaml_scanner_t, pos int, token *yaml_token_t) {
	if scanner.tokens_head > 0 && len(scanner.tokens) == cap(scanner.tokens) {
		if scanner.tokens_head != len(scanner.tokens) {
			copy(scanner.tokens, scanner.tokens[scanner.tokens_head:])
		}
		scanner.tokens = scanner.tokens[:len(scanner.tokens)-scanner.tokens_head]
		scanner.tokens_head = 0
	}
	scanner.tokens = append(scanner.tokens, *token)
	if pos < 0 {
		return
	}
	copy(scanner.tokens[scanner.tokens_head+pos+1:], scanner.tokens[scanner.tokens_head+pos:])
	scanner.tokens[scanner.tokens_head+pos] = *token
}

// Ensure that the tokens queue contains at least one token which can be
// returned to the Parser.
func yaml_scanner_fetch_more_tokens(scanner *yaml_scanner_t) bool {
	// While we need more tokens to fetch, do it.
	for {
		// [Go] The comment parsing logic requires a lookahead of two tokens
		// so that foot comments may be parsed in time of associating them
		// with the tokens that are parsed before them, and also for line
		// comments to be transformed into head comments in some edge cases.
		if scanner.tokens_head < len(scanner.tokens)-2 {
			// If a potential simple key is at the head position, we need to fetch
			// the next token to disambiguate it.
			head_tok_idx, ok := scanner.simple_keys_by_tok[scanner.tokens_parsed]
			if !ok {
				break
			} else if valid, ok := yaml_simple_key_is_valid(scanner, &scanner.simple_keys[head_tok_idx]); !ok {
				return false
			} else if !valid {
				break
			}
		}
		// Fetch the next token.
		if !yaml_scanner_fetch_next_token(scanner) {
			return false
		}
	}

	scanner.token_available = true
	return true
}

// The dispatcher for token fetchers.
func yaml_scanner_fetch_next_token(scanner *yaml_scanner_t) (ok bool) {
	// Ensure that the buffer is initialized.
	if scanner.unread < 1 && !yaml_scanner_update_buffer(scanner, 1) {
		return false
	}

	// Check if we just started scanning.  Fetch STREAM-START then.
	if !scanner.stream_start_produced {
		return yaml_scanner_fetch_stream_start(scanner)
	}

	scan_mark := scanner.mark

	// Eat whitespaces and comments until we reach the next token.
	if !yaml_scanner_scan_to_next_token(scanner) {
		return false
	}

	// [Go] While unrolling indents, transform the head comments of prior
	// indentation levels observed after scan_start into foot comments at
	// the respective indexes.

	// Check the indentation level against the current column.
	if !yaml_scanner_unroll_indent(scanner, scanner.mark.column, scan_mark) {
		return false
	}

	// Ensure that the buffer contains at least 4 characters.  4 is the length
	// of the longest indicators ('--- ' and '... ').
	if scanner.unread < 4 && !yaml_scanner_update_buffer(scanner, 4) {
		return false
	}

	// Is it the end of the stream?
	if is_z(scanner.buffer, scanner.buffer_pos) {
		return yaml_scanner_fetch_stream_end(scanner)
	}

	// Is it a directive?
	if scanner.mark.column == 0 && scanner.buffer[scanner.buffer_pos] == '%' {
		return yaml_scanner_fetch_directive(scanner)
	}

	buf := scanner.buffer
	pos := scanner.buffer_pos

	// Is it the document start indicator?
	if scanner.mark.column == 0 && buf[pos] == '-' && buf[pos+1] == '-' && buf[pos+2] == '-' && is_blankz(buf, pos+3) {
		return yaml_scanner_fetch_document_indicator(scanner, yaml_DOCUMENT_START_TOKEN)
	}

	// Is it the document end indicator?
	if scanner.mark.column == 0 && buf[pos] == '.' && buf[pos+1] == '.' && buf[pos+2] == '.' && is_blankz(buf, pos+3) {
		return yaml_scanner_fetch_document_indicator(scanner, yaml_DOCUMENT_END_TOKEN)
	}

	comment_mark := scanner.mark
	if len(scanner.tokens) > 0 && (scanner.flow_level == 0 && buf[pos] == ':' || scanner.flow_level > 0 && buf[pos] == ',') {
		// Associate any following comments with the prior token.
		comment_mark = scanner.tokens[len(scanner.tokens)-1].start_mark
	}
	defer func() {
		if !ok {
			return
		}
		if len(scanner.tokens) > 0 && scanner.tokens[len(scanner.tokens)-1].typ == yaml_BLOCK_ENTRY_TOKEN {
			// Sequence indicators alone have no line comments. It becomes
			// a head comment for whatever follows.
			return
		}
		if !yaml_scanner_scan_line_comment(scanner, comment_mark) {
			ok = false
			return
		}
	}()

	// Is it the flow sequence start indicator?
	if buf[pos] == '[' {
		return yaml_scanner_fetch_flow_collection_start(scanner, yaml_FLOW_SEQUENCE_START_TOKEN)
	}

	// Is it the flow mapping start indicator?
	if scanner.buffer[scanner.buffer_pos] == '{' {
		return yaml_scanner_fetch_flow_collection_start(scanner, yaml_FLOW_MAPPING_START_TOKEN)
	}

	// Is it the flow sequence end indicator?
	if scanner.buffer[scanner.buffer_pos] == ']' {
		return yaml_scanner_fetch_flow_collection_end(scanner,
			yaml_FLOW_SEQUENCE_END_TOKEN)
	}

	// Is it the flow mapping end indicator?
	if scanner.buffer[scanner.buffer_pos] == '}' {
		return yaml_scanner_fetch_flow_collection_end(scanner,
			yaml_FLOW_MAPPING_END_TOKEN)
	}

	// Is it the flow entry indicator?
	if scanner.buffer[scanner.buffer_pos] == ',' {
		return yaml_scanner_fetch_flow_entry(scanner)
	}

	// Is it the block entry indicator?
	if scanner.buffer[scanner.buffer_pos] == '-' && is_blankz(scanner.buffer, scanner.buffer_pos+1) {
		return yaml_scanner_fetch_block_entry(scanner)
	}

	// Is it the key indicator?
	if scanner.buffer[scanner.buffer_pos] == '?' && (scanner.flow_level > 0 || is_blankz(scanner.buffer, scanner.buffer_pos+1)) {
		return yaml_scanner_fetch_key(scanner)
	}

	// Is it the value indicator?
	if scanner.buffer[scanner.buffer_pos] == ':' && (scanner.flow_level > 0 || is_blankz(scanner.buffer, scanner.buffer_pos+1)) {
		return yaml_scanner_fetch_value(scanner)
	}

	// Is it an alias?
	if scanner.buffer[scanner.buffer_pos] == '*' {
		return yaml_scanner_fetch_anchor(scanner, yaml_ALIAS_TOKEN)
	}

	// Is it an anchor?
	if scanner.buffer[scanner.buffer_pos] == '&' {
		return yaml_scanner_fetch_anchor(scanner, yaml_ANCHOR_TOKEN)
	}

	// Is it a tag?
	if scanner.buffer[scanner.buffer_pos] == '!' {
		return yaml_scanner_fetch_tag(scanner)
	}

	// Is it a literal scalar?
	if scanner.buffer[scanner.buffer_pos] == '|' && scanner.flow_level == 0 {
		return yaml_scanner_fetch_block_scalar(scanner, true)
	}

	// Is it a folded scalar?
	if scanner.buffer[scanner.buffer_pos] == '>' && scanner.flow_level == 0 {
		return yaml_scanner_fetch_block_scalar(scanner, false)
	}

	// Is it a single-quoted scalar?
	if scanner.buffer[scanner.buffer_pos] == '\'' {
		return yaml_scanner_fetch_flow_scalar(scanner, true)
	}

	// Is it a double-quoted scalar?
	if scanner.buffer[scanner.buffer_pos] == '"' {
		return yaml_scanner_fetch_flow_scalar(scanner, false)
	}

	// Is it a plain scalar?
	//
	// A plain scalar may start with any non-blank characters except
	//
	//      '-', '?', ':', ',', '[', ']', '{', '}',
	//      '#', '&', '*', '!', '|', '>', '\'', '\"',
	//      '%', '@', '`'.
	//
	// In the block context (and, for the '-' indicator, in the flow context
	// too), it may also start with the characters
	//
	//      '-', '?', ':'
	//
	// if it is followed by a non-space character.
	//
	// The last rule is more restrictive than the specification requires.
	// [Go] TODO Make this logic more reasonable.
	//switch parser.buffer[parser.buffer_pos] {
	//case '-', '?', ':', ',', '?', '-', ',', ':', ']', '[', '}', '{', '&', '#', '!', '*', '>', '|', '"', '\'', '@', '%', '-', '`':
	//}
	if (!is_blankz(scanner.buffer, scanner.buffer_pos) && scanner.buffer[scanner.buffer_pos] != '-' &&
		scanner.buffer[scanner.buffer_pos] != '?' && scanner.buffer[scanner.buffer_pos] != ':' &&
		scanner.buffer[scanner.buffer_pos] != ',' && scanner.buffer[scanner.buffer_pos] != '[' &&
		scanner.buffer[scanner.buffer_pos] != ']' && scanner.buffer[scanner.buffer_pos] != '{' &&
		scanner.buffer[scanner.buffer_pos] != '}' && scanner.buffer[scanner.buffer_pos] != '#' &&
		scanner.buffer[scanner.buffer_pos] != '&' && scanner.buffer[scanner.buffer_pos] != '*' &&
		scanner.buffer[scanner.buffer_pos] != '!' && scanner.buffer[scanner.buffer_pos] != '|' &&
		scanner.buffer[scanner.buffer_pos] != '>' && scanner.buffer[scanner.buffer_pos] != '\'' &&
		scanner.buffer[scanner.buffer_pos] != '"' && scanner.buffer[scanner.buffer_pos] != '%' &&
		scanner.buffer[scanner.buffer_pos] != '@' && scanner.buffer[scanner.buffer_pos] != '`') ||
		(scanner.buffer[scanner.buffer_pos] == '-' && !is_blank(scanner.buffer, scanner.buffer_pos+1)) ||
		(scanner.flow_level == 0 &&
			(scanner.buffer[scanner.buffer_pos] == '?' || scanner.buffer[scanner.buffer_pos] == ':') &&
			!is_blankz(scanner.buffer, scanner.buffer_pos+1)) {
		return yaml_scanner_fetch_plain_scalar(scanner)
	}

	// If we don't determine the token type so far, it is an error.
	return yaml_scanner_set_error(scanner,
		"while scanning for the next token", scanner.mark,
		"found character that cannot start any token")
}

func yaml_simple_key_is_valid(scanner *yaml_scanner_t, simple_key *yaml_simple_key_t) (valid, ok bool) {
	if !simple_key.possible {
		return false, true
	}

	// The 1.2 specification says:
	//
	//     "If the ? indicator is omitted, parsing needs to see past the
	//     implicit key to recognize it as such. To limit the amount of
	//     lookahead required, the “:” indicator must appear at most 1024
	//     Unicode characters beyond the start of the key. In addition, the key
	//     is restricted to a single line."
	//
	if simple_key.mark.line < scanner.mark.line || simple_key.mark.index+1024 < scanner.mark.index {
		// Check if the potential simple key to be removed is required.
		if simple_key.required {
			return false, yaml_scanner_set_error(scanner,
				"while scanning a simple key", simple_key.mark,
				"could not find expected ':'")
		}
		simple_key.possible = false
		return false, true
	}
	return true, true
}

// Check if a simple key may start at the current position and add it if
// needed.
func yaml_scanner_save_simple_key(scanner *yaml_scanner_t) bool {
	// A simple key is required at the current position if the scanner is in
	// the block context and the current column coincides with the indentation
	// level.

	required := scanner.flow_level == 0 && scanner.indent == scanner.mark.column

	//
	// If the current position may start a simple key, save it.
	//
	if scanner.simple_key_allowed {
		simple_key := yaml_simple_key_t{
			possible:     true,
			required:     required,
			token_number: scanner.tokens_parsed + (len(scanner.tokens) - scanner.tokens_head),
			mark:         scanner.mark,
		}

		if !yaml_scanner_remove_simple_key(scanner) {
			return false
		}
		scanner.simple_keys[len(scanner.simple_keys)-1] = simple_key
		scanner.simple_keys_by_tok[simple_key.token_number] = len(scanner.simple_keys) - 1
	}
	return true
}

// Remove a potential simple key at the current flow level.
func yaml_scanner_remove_simple_key(scanner *yaml_scanner_t) bool {
	i := len(scanner.simple_keys) - 1
	if scanner.simple_keys[i].possible {
		// If the key is required, it is an error.
		if scanner.simple_keys[i].required {
			return yaml_scanner_set_error(scanner,
				"while scanning a simple key", scanner.simple_keys[i].mark,
				"could not find expected ':'")
		}
		// Remove the key from the stack.
		scanner.simple_keys[i].possible = false
		delete(scanner.simple_keys_by_tok, scanner.simple_keys[i].token_number)
	}
	return true
}

// max_flow_level limits the flow_level
const max_flow_level = 10000

// Increase the flow level and resize the simple key list if needed.
func yaml_scanner_increase_flow_level(scanner *yaml_scanner_t) bool {
	// Reset the simple key on the next level.
	scanner.simple_keys = append(scanner.simple_keys, yaml_simple_key_t{
		possible:     false,
		required:     false,
		token_number: scanner.tokens_parsed + (len(scanner.tokens) - scanner.tokens_head),
		mark:         scanner.mark,
	})

	// Increase the flow level.
	scanner.flow_level++
	if scanner.flow_level > max_flow_level {
		return yaml_scanner_set_error(scanner,
			"while increasing flow level", scanner.simple_keys[len(scanner.simple_keys)-1].mark,
			fmt.Sprintf("exceeded max depth of %d", max_flow_level))
	}
	return true
}

// Decrease the flow level.
func yaml_scanner_decrease_flow_level(scanner *yaml_scanner_t) bool {
	if scanner.flow_level > 0 {
		scanner.flow_level--
		last := len(scanner.simple_keys) - 1
		delete(scanner.simple_keys_by_tok, scanner.simple_keys[last].token_number)
		scanner.simple_keys = scanner.simple_keys[:last]
	}
	return true
}

// max_indents limits the indents stack size
const max_indents = 10000

// Push the current indentation level to the stack and set the new level
// the current column is greater than the indentation level.  In this case,
// append or insert the specified token into the token queue.
func yaml_scanner_roll_indent(scanner *yaml_scanner_t, column, number int, typ yaml_token_type_t, mark yaml_mark_t) bool {
	// In the flow context, do nothing.
	if scanner.flow_level > 0 {
		return true
	}

	if scanner.indent < column {
		// Push the current indentation level to the stack and set the new
		// indentation level.
		scanner.indents = append(scanner.indents, scanner.indent)
		scanner.indent = column
		if len(scanner.indents) > max_indents {
			return yaml_scanner_set_error(scanner,
				"while increasing indent level", scanner.simple_keys[len(scanner.simple_keys)-1].mark,
				fmt.Sprintf("exceeded max depth of %d", max_indents))
		}

		// Create a token and insert it into the queue.
		token := yaml_token_t{
			typ:        typ,
			start_mark: mark,
			end_mark:   mark,
		}
		if number > -1 {
			number -= scanner.tokens_parsed
		}
		yaml_insert_token(scanner, number, &token)
	}
	return true
}

// Pop indentation levels from the indents stack until the current level
// becomes less or equal to the column.  For each indentation level, append
// the BLOCK-END token.
func yaml_scanner_unroll_indent(scanner *yaml_scanner_t, column int, scan_mark yaml_mark_t) bool {
	// In the flow context, do nothing.
	if scanner.flow_level > 0 {
		return true
	}

	block_mark := scan_mark
	block_mark.index--

	// Loop through the indentation levels in the stack.
	for scanner.indent > column {

		// [Go] Reposition the end token before potential following
		//      foot comments of parent blocks. For that, search
		//      backwards for recent comments that were at the same
		//      indent as the block that is ending now.
		stop_index := block_mark.index
		for i := len(scanner.comments) - 1; i >= 0; i-- {
			comment := &scanner.comments[i]

			if comment.end_mark.index < stop_index {
				// Don't go back beyond the start of the comment/whitespace scan, unless column < 0.
				// If requested indent column is < 0, then the document is over and everything else
				// is a foot anyway.
				break
			}
			if comment.start_mark.column == scanner.indent+1 {
				// This is a good match. But maybe there's a former comment
				// at that same indent level, so keep searching.
				block_mark = comment.start_mark
			}

			// While the end of the former comment matches with
			// the start of the following one, we know there's
			// nothing in between and scanning is still safe.
			stop_index = comment.scan_mark.index
		}

		// Create a token and append it to the queue.
		token := yaml_token_t{
			typ:        yaml_BLOCK_END_TOKEN,
			start_mark: block_mark,
			end_mark:   block_mark,
		}
		yaml_insert_token(scanner, -1, &token)

		// Pop the indentation level.
		scanner.indent = scanner.indents[len(scanner.indents)-1]
		scanner.indents = scanner.indents[:len(scanner.indents)-1]
	}
	return true
}

// Initialize the scanner and produce the STREAM-START token.
func yaml_scanner_fetch_stream_start(scanner *yaml_scanner_t) bool {

	// Set the initial indentation.
	scanner.indent = -1

	// Initialize the simple key stack.
	scanner.simple_keys = append(scanner.simple_keys, yaml_simple_key_t{})

	scanner.simple_keys_by_tok = make(map[int]int)

	// A simple key is allowed at the beginning of the stream.
	scanner.simple_key_allowed = true

	// We have started.
	scanner.stream_start_produced = true

	// Create the STREAM-START token and append it to the queue.
	token := yaml_token_t{
		typ:        yaml_STREAM_START_TOKEN,
		start_mark: scanner.mark,
		end_mark:   scanner.mark,
		encoding:   scanner.encoding,
	}
	yaml_insert_token(scanner, -1, &token)
	return true
}

// Produce the STREAM-END token and shut down the scanner.
func yaml_scanner_fetch_stream_end(scanner *yaml_scanner_t) bool {

	// Force new line.
	if scanner.mark.column != 0 {
		scanner.mark.column = 0
		scanner.mark.line++
	}

	// Reset the indentation level.
	if !yaml_scanner_unroll_indent(scanner, -1, scanner.mark) {
		return false
	}

	// Reset simple keys.
	if !yaml_scanner_remove_simple_key(scanner) {
		return false
	}

	scanner.simple_key_allowed = false

	// Create the STREAM-END token and append it to the queue.
	token := yaml_token_t{
		typ:        yaml_STREAM_END_TOKEN,
		start_mark: scanner.mark,
		end_mark:   scanner.mark,
	}
	yaml_insert_token(scanner, -1, &token)
	return true
}

// Produce a VERSION-DIRECTIVE or TAG-DIRECTIVE token.
func yaml_scanner_fetch_directive(scanner *yaml_scanner_t) bool {
	// Reset the indentation level.
	if !yaml_scanner_unroll_indent(scanner, -1, scanner.mark) {
		return false
	}

	// Reset simple keys.
	if !yaml_scanner_remove_simple_key(scanner) {
		return false
	}

	scanner.simple_key_allowed = false

	// Create the YAML-DIRECTIVE or TAG-DIRECTIVE token.
	token := yaml_token_t{}
	if !yaml_scanner_scan_directive(scanner, &token) {
		return false
	}
	// Append the token to the queue.
	yaml_insert_token(scanner, -1, &token)
	return true
}

// Produce the DOCUMENT-START or DOCUMENT-END token.
func yaml_scanner_fetch_document_indicator(scanner *yaml_scanner_t, typ yaml_token_type_t) bool {
	// Reset the indentation level.
	if !yaml_scanner_unroll_indent(scanner, -1, scanner.mark) {
		return false
	}

	// Reset simple keys.
	if !yaml_scanner_remove_simple_key(scanner) {
		return false
	}

	scanner.simple_key_allowed = false

	// Consume the token.
	start_mark := scanner.mark

	skip(scanner)
	skip(scanner)
	skip(scanner)

	end_mark := scanner.mark

	// Create the DOCUMENT-START or DOCUMENT-END token.
	token := yaml_token_t{
		typ:        typ,
		start_mark: start_mark,
		end_mark:   end_mark,
	}
	// Append the token to the queue.
	yaml_insert_token(scanner, -1, &token)
	return true
}

// Produce the FLOW-SEQUENCE-START or FLOW-MAPPING-START token.
func yaml_scanner_fetch_flow_collection_start(scanner *yaml_scanner_t, typ yaml_token_type_t) bool {

	// The indicators '[' and '{' may start a simple key.
	if !yaml_scanner_save_simple_key(scanner) {
		return false
	}

	// Increase the flow level.
	if !yaml_scanner_increase_flow_level(scanner) {
		return false
	}

	// A simple key may follow the indicators '[' and '{'.
	scanner.simple_key_allowed = true

	// Consume the token.
	start_mark := scanner.mark
	skip(scanner)
	end_mark := scanner.mark

	// Create the FLOW-SEQUENCE-START of FLOW-MAPPING-START token.
	token := yaml_token_t{
		typ:        typ,
		start_mark: start_mark,
		end_mark:   end_mark,
	}
	// Append the token to the queue.
	yaml_insert_token(scanner, -1, &token)
	return true
}

// Produce the FLOW-SEQUENCE-END or FLOW-MAPPING-END token.
func yaml_scanner_fetch_flow_collection_end(scanner *yaml_scanner_t, typ yaml_token_type_t) bool {
	// Reset any potential simple key on the current flow level.
	if !yaml_scanner_remove_simple_key(scanner) {
		return false
	}

	// Decrease the flow level.
	if !yaml_scanner_decrease_flow_level(scanner) {
		return false
	}

	// No simple keys after the indicators ']' and '}'.
	scanner.simple_key_allowed = false

	// Consume the token.

	start_mark := scanner.mark
	skip(scanner)
	end_mark := scanner.mark

	// Create the FLOW-SEQUENCE-END of FLOW-MAPPING-END token.
	token := yaml_token_t{
		typ:        typ,
		start_mark: start_mark,
		end_mark:   end_mark,
	}
	// Append the token to the queue.
	yaml_insert_token(scanner, -1, &token)
	return true
}

// Produce the FLOW-ENTRY token.
func yaml_scanner_fetch_flow_entry(scanner *yaml_scanner_t) bool {
	// Reset any potential simple keys on the current flow level.
	if !yaml_scanner_remove_simple_key(scanner) {
		return false
	}

	// Simple keys are allowed after ','.
	scanner.simple_key_allowed = true

	// Consume the token.
	start_mark := scanner.mark
	skip(scanner)
	end_mark := scanner.mark

	// Create the FLOW-ENTRY token and append it to the queue.
	token := yaml_token_t{
		typ:        yaml_FLOW_ENTRY_TOKEN,
		start_mark: start_mark,
		end_mark:   end_mark,
	}
	yaml_insert_token(scanner, -1, &token)
	return true
}

// Produce the BLOCK-ENTRY token.
func yaml_scanner_fetch_block_entry(scanner *yaml_scanner_t) bool {
	// Check if the scanner is in the block context.
	if scanner.flow_level == 0 {
		// Check if we are allowed to start a new entry.
		if !scanner.simple_key_allowed {
			return yaml_scanner_set_error(scanner, "", scanner.mark,
				"block sequence entries are not allowed in this context")
		}
		// Add the BLOCK-SEQUENCE-START token if needed.
		if !yaml_scanner_roll_indent(scanner, scanner.mark.column, -1, yaml_BLOCK_SEQUENCE_START_TOKEN, scanner.mark) {
			return false
		}
	}

	// Reset any potential simple keys on the current flow level.
	if !yaml_scanner_remove_simple_key(scanner) {
		return false
	}

	// Simple keys are allowed after '-'.
	scanner.simple_key_allowed = true

	// Consume the token.
	start_mark := scanner.mark
	skip(scanner)
	end_mark := scanner.mark

	// Create the BLOCK-ENTRY token and append it to the queue.
	token := yaml_token_t{
		typ:        yaml_BLOCK_ENTRY_TOKEN,
		start_mark: start_mark,
		end_mark:   end_mark,
	}
	yaml_insert_token(scanner, -1, &token)
	return true
}

// Produce the KEY token.
func yaml_scanner_fetch_key(scanner *yaml_scanner_t) bool {

	// In the block context, additional checks are required.
	if scanner.flow_level == 0 {
		// Check if we are allowed to start a new key (not nessesary simple).
		if !scanner.simple_key_allowed {
			return yaml_scanner_set_error(scanner, "", scanner.mark,
				"mapping keys are not allowed in this context")
		}
		// Add the BLOCK-MAPPING-START token if needed.
		if !yaml_scanner_roll_indent(scanner, scanner.mark.column, -1, yaml_BLOCK_MAPPING_START_TOKEN, scanner.mark) {
			return false
		}
	}

	// Reset any potential simple keys on the current flow level.
	if !yaml_scanner_remove_simple_key(scanner) {
		return false
	}

	// Simple keys are allowed after '?' in the block context.
	scanner.simple_key_allowed = scanner.flow_level == 0

	// Consume the token.
	start_mark := scanner.mark
	skip(scanner)
	end_mark := scanner.mark

	// Create the KEY token and append it to the queue.
	token := yaml_token_t{
		typ:        yaml_KEY_TOKEN,
		start_mark: start_mark,
		end_mark:   end_mark,
	}
	yaml_insert_token(scanner, -1, &token)
	return true
}

// Produce the VALUE token.
func yaml_scanner_fetch_value(scanner *yaml_scanner_t) bool {

	simple_key := &scanner.simple_keys[len(scanner.simple_keys)-1]

	// Have we found a simple key?
	if valid, ok := yaml_simple_key_is_valid(scanner, simple_key); !ok {
		return false

	} else if valid {

		// Create the KEY token and insert it into the queue.
		token := yaml_token_t{
			typ:        yaml_KEY_TOKEN,
			start_mark: simple_key.mark,
			end_mark:   simple_key.mark,
		}
		yaml_insert_token(scanner, simple_key.token_number-scanner.tokens_parsed, &token)

		// In the block context, we may need to add the BLOCK-MAPPING-START token.
		if !yaml_scanner_roll_indent(scanner, simple_key.mark.column,
			simple_key.token_number,
			yaml_BLOCK_MAPPING_START_TOKEN, simple_key.mark) {
			return false
		}

		// Remove the simple key.
		simple_key.possible = false
		delete(scanner.simple_keys_by_tok, simple_key.token_number)

		// A simple key cannot follow another simple key.
		scanner.simple_key_allowed = false

	} else {
		// The ':' indicator follows a complex key.

		// In the block context, extra checks are required.
		if scanner.flow_level == 0 {

			// Check if we are allowed to start a complex value.
			if !scanner.simple_key_allowed {
				return yaml_scanner_set_error(scanner, "", scanner.mark,
					"mapping values are not allowed in this context")
			}

			// Add the BLOCK-MAPPING-START token if needed.
			if !yaml_scanner_roll_indent(scanner, scanner.mark.column, -1, yaml_BLOCK_MAPPING_START_TOKEN, scanner.mark) {
				return false
			}
		}

		// Simple keys after ':' are allowed in the block context.
		scanner.simple_key_allowed = scanner.flow_level == 0
	}

	// Consume the token.
	start_mark := scanner.mark
	skip(scanner)
	end_mark := scanner.mark

	// Create the VALUE token and append it to the queue.
	token := yaml_token_t{
		typ:        yaml_VALUE_TOKEN,
		start_mark: start_mark,
		end_mark:   end_mark,
	}
	yaml_insert_token(scanner, -1, &token)
	return true
}

// Produce the ALIAS or ANCHOR token.
func yaml_scanner_fetch_anchor(scanner *yaml_scanner_t, typ yaml_token_type_t) bool {
	// An anchor or an alias could be a simple key.
	if !yaml_scanner_save_simple_key(scanner) {
		return false
	}

	// A simple key cannot follow an anchor or an alias.
	scanner.simple_key_allowed = false

	// Create the ALIAS or ANCHOR token and append it to the queue.
	var token yaml_token_t
	if !yaml_scanner_scan_anchor(scanner, &token, typ) {
		return false
	}
	yaml_insert_token(scanner, -1, &token)
	return true
}

// Produce the TAG token.
func yaml_scanner_fetch_tag(scanner *yaml_scanner_t) bool {
	// A tag could be a simple key.
	if !yaml_scanner_save_simple_key(scanner) {
		return false
	}

	// A simple key cannot follow a tag.
	scanner.simple_key_allowed = false

	// Create the TAG token and append it to the queue.
	var token yaml_token_t
	if !yaml_scanner_scan_tag(scanner, &token) {
		return false
	}
	yaml_insert_token(scanner, -1, &token)
	return true
}

// Produce the SCALAR(...,literal) or SCALAR(...,folded) tokens.
func yaml_scanner_fetch_block_scalar(scanner *yaml_scanner_t, literal bool) bool {
	// Remove any potential simple keys.
	if !yaml_scanner_remove_simple_key(scanner) {
		return false
	}

	// A simple key may follow a block scalar.
	scanner.simple_key_allowed = true

	// Create the SCALAR token and append it to the queue.
	var token yaml_token_t
	if !yaml_scanner_scan_block_scalar(scanner, &token, literal) {
		return false
	}
	yaml_insert_token(scanner, -1, &token)
	return true
}

// Produce the SCALAR(...,single-quoted) or SCALAR(...,double-quoted) tokens.
func yaml_scanner_fetch_flow_scalar(scanner *yaml_scanner_t, single bool) bool {
	// A plain scalar could be a simple key.
	if !yaml_scanner_save_simple_key(scanner) {
		return false
	}

	// A simple key cannot follow a flow scalar.
	scanner.simple_key_allowed = false

	// Create the SCALAR token and append it to the queue.
	var token yaml_token_t
	if !yaml_scanner_scan_flow_scalar(scanner, &token, single) {
		return false
	}
	yaml_insert_token(scanner, -1, &token)
	return true
}

// Produce the SCALAR(...,plain) token.
func yaml_scanner_fetch_plain_scalar(scanner *yaml_scanner_t) bool {
	// A plain scalar could be a simple key.
	if !yaml_scanner_save_simple_key(scanner) {
		return false
	}

	// A simple key cannot follow a flow scalar.
	scanner.simple_key_allowed = false

	// Create the SCALAR token and append it to the queue.
	var token yaml_token_t
	if !yaml_scanner_scan_plain_scalar(scanner, &token) {
		return false
	}
	yaml_insert_token(scanner, -1, &token)
	return true
}
