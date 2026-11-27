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
)

// The parser implements the following grammar:
//
// stream               ::= STREAM-START implicit_document? explicit_document* STREAM-END
// implicit_document    ::= block_node DOCUMENT-END*
// explicit_document    ::= DIRECTIVE* DOCUMENT-START block_node? DOCUMENT-END*
// block_node_or_indentless_sequence    ::=
//                          ALIAS
//                          | properties (block_content | indentless_block_sequence)?
//                          | block_content
//                          | indentless_block_sequence
// block_node           ::= ALIAS
//                          | properties block_content?
//                          | block_content
// flow_node            ::= ALIAS
//                          | properties flow_content?
//                          | flow_content
// properties           ::= TAG ANCHOR? | ANCHOR TAG?
// block_content        ::= block_collection | flow_collection | SCALAR
// flow_content         ::= flow_collection | SCALAR
// block_collection     ::= block_sequence | block_mapping
// flow_collection      ::= flow_sequence | flow_mapping
// block_sequence       ::= BLOCK-SEQUENCE-START (BLOCK-ENTRY block_node?)* BLOCK-END
// indentless_sequence  ::= (BLOCK-ENTRY block_node?)+
// block_mapping        ::= BLOCK-MAPPING_START
//                          ((KEY block_node_or_indentless_sequence?)?
//                          (VALUE block_node_or_indentless_sequence?)?)*
//                          BLOCK-END
// flow_sequence        ::= FLOW-SEQUENCE-START
//                          (flow_sequence_entry FLOW-ENTRY)*
//                          flow_sequence_entry?
//                          FLOW-SEQUENCE-END
// flow_sequence_entry  ::= flow_node | KEY flow_node? (VALUE flow_node?)?
// flow_mapping         ::= FLOW-MAPPING-START
//                          (flow_mapping_entry FLOW-ENTRY)*
//                          flow_mapping_entry?
//                          FLOW-MAPPING-END
// flow_mapping_entry   ::= flow_node | KEY flow_node? (VALUE flow_node?)?

// yaml_new_parser returns a parser over input, which it reads in place.
func yaml_new_parser(input []byte) yaml_parser_t {
	return yaml_parser_t{scanner: yaml_scanner_t{
		raw_buffer: make([]byte, 0, input_raw_buffer_size),
		buffer:     make([]byte, 0, input_buffer_size),
		input:      input,
	}}
}

// peek_token returns the next token, fetching it when the queue holds none,
// and takes the comments that precede it from the scanner. It returns nil
// when the scanner fails, and the parser then holds the scanner's problem.
func (parser *yaml_parser_t) peek_token() *yaml_token_t {
	scanner := &parser.scanner
	if !scanner.token_available && !yaml_scanner_fetch_more_tokens(scanner) {
		parser.yaml_problem_t = scanner.yaml_problem_t
		return nil
	}
	token := &scanner.tokens[scanner.tokens_head]
	head, line, foot := yaml_scanner_take_comments(scanner, token)
	parser.head_comment = join_comment(parser.head_comment, head)
	parser.line_comment = join_comment(parser.line_comment, line)
	parser.foot_comment = join_comment(parser.foot_comment, foot)
	return token
}

// skip_token removes the token peek_token returned.
func (parser *yaml_parser_t) skip_token() {
	scanner := &parser.scanner
	parser.stream_end_produced = scanner.tokens[scanner.tokens_head].typ == yaml_STREAM_END_TOKEN
	yaml_scanner_skip_token(scanner)
}

// join_comment appends text to comments, after a line break when both hold
// text.
func join_comment(comments, text []byte) []byte {
	if len(text) == 0 {
		return comments
	}
	if len(comments) > 0 {
		comments = append(comments, '\n')
	}
	return append(comments, text...)
}

// Get the next event.
func (parser *yaml_parser_t) parse(event *yaml_event_t) bool {
	// Erase the event object.
	*event = yaml_event_t{}

	// No events after the end of the stream or error.
	if parser.stream_end_produced || parser.error != yaml_NO_ERROR || parser.state == yaml_PARSE_END_STATE {
		return true
	}

	// Generate the next event.
	return parser.state_machine(event)
}

// Set parser error.
func (parser *yaml_parser_t) set_parser_error(problem string, problem_mark yaml_mark_t) bool {
	parser.error = yaml_PARSER_ERROR
	parser.problem = problem
	parser.problem_mark = problem_mark
	return false
}

func (parser *yaml_parser_t) set_parser_error_context(context string, context_mark yaml_mark_t, problem string, problem_mark yaml_mark_t) bool {
	parser.error = yaml_PARSER_ERROR
	parser.context = context
	parser.context_mark = context_mark
	parser.problem = problem
	parser.problem_mark = problem_mark
	return false
}

// State dispatcher.
func (parser *yaml_parser_t) state_machine(event *yaml_event_t) bool {

	switch parser.state {
	case yaml_PARSE_STREAM_START_STATE:
		return parser.parse_stream_start(event)

	case yaml_PARSE_IMPLICIT_DOCUMENT_START_STATE:
		return parser.parse_document_start(event, true)

	case yaml_PARSE_DOCUMENT_START_STATE:
		return parser.parse_document_start(event, false)

	case yaml_PARSE_DOCUMENT_CONTENT_STATE:
		return parser.parse_document_content(event)

	case yaml_PARSE_DOCUMENT_END_STATE:
		return parser.parse_document_end(event)

	case yaml_PARSE_BLOCK_NODE_STATE:
		return parser.parse_node(event, true, false)

	case yaml_PARSE_BLOCK_NODE_OR_INDENTLESS_SEQUENCE_STATE:
		return parser.parse_node(event, true, true)

	case yaml_PARSE_FLOW_NODE_STATE:
		return parser.parse_node(event, false, false)

	case yaml_PARSE_BLOCK_SEQUENCE_FIRST_ENTRY_STATE:
		return parser.parse_block_sequence_entry(event, true)

	case yaml_PARSE_BLOCK_SEQUENCE_ENTRY_STATE:
		return parser.parse_block_sequence_entry(event, false)

	case yaml_PARSE_INDENTLESS_SEQUENCE_ENTRY_STATE:
		return parser.parse_indentless_sequence_entry(event)

	case yaml_PARSE_BLOCK_MAPPING_FIRST_KEY_STATE:
		return parser.parse_block_mapping_key(event, true)

	case yaml_PARSE_BLOCK_MAPPING_KEY_STATE:
		return parser.parse_block_mapping_key(event, false)

	case yaml_PARSE_BLOCK_MAPPING_VALUE_STATE:
		return parser.parse_block_mapping_value(event)

	case yaml_PARSE_FLOW_SEQUENCE_FIRST_ENTRY_STATE:
		return parser.parse_flow_sequence_entry(event, true)

	case yaml_PARSE_FLOW_SEQUENCE_ENTRY_STATE:
		return parser.parse_flow_sequence_entry(event, false)

	case yaml_PARSE_FLOW_SEQUENCE_ENTRY_MAPPING_KEY_STATE:
		return parser.parse_flow_sequence_entry_mapping_key(event)

	case yaml_PARSE_FLOW_SEQUENCE_ENTRY_MAPPING_VALUE_STATE:
		return parser.parse_flow_sequence_entry_mapping_value(event)

	case yaml_PARSE_FLOW_SEQUENCE_ENTRY_MAPPING_END_STATE:
		return parser.parse_flow_sequence_entry_mapping_end(event)

	case yaml_PARSE_FLOW_MAPPING_FIRST_KEY_STATE:
		return parser.parse_flow_mapping_key(event, true)

	case yaml_PARSE_FLOW_MAPPING_KEY_STATE:
		return parser.parse_flow_mapping_key(event, false)

	case yaml_PARSE_FLOW_MAPPING_VALUE_STATE:
		return parser.parse_flow_mapping_value(event, false)

	case yaml_PARSE_FLOW_MAPPING_EMPTY_VALUE_STATE:
		return parser.parse_flow_mapping_value(event, true)

	default:
		panic("invalid parser state")
	}
}

// Parse the production:
// stream   ::= STREAM-START implicit_document? explicit_document* STREAM-END
//
//	************
func (parser *yaml_parser_t) parse_stream_start(event *yaml_event_t) bool {
	token := parser.peek_token()
	if token == nil {
		return false
	}
	if token.typ != yaml_STREAM_START_TOKEN {
		return parser.set_parser_error("did not find expected <stream-start>", token.start_mark)
	}
	parser.state = yaml_PARSE_IMPLICIT_DOCUMENT_START_STATE
	*event = yaml_event_t{
		typ:        yaml_STREAM_START_EVENT,
		start_mark: token.start_mark,
		end_mark:   token.end_mark,
		encoding:   token.encoding,
	}
	parser.skip_token()
	return true
}

// Parse the productions:
// implicit_document    ::= block_node DOCUMENT-END*
//
//	*
//
// explicit_document    ::= DIRECTIVE* DOCUMENT-START block_node? DOCUMENT-END*
//
//	*************************
func (parser *yaml_parser_t) parse_document_start(event *yaml_event_t, implicit bool) bool {

	token := parser.peek_token()
	if token == nil {
		return false
	}

	// Parse extra document end indicators.
	if !implicit {
		for token.typ == yaml_DOCUMENT_END_TOKEN {
			parser.skip_token()
			token = parser.peek_token()
			if token == nil {
				return false
			}
		}
	}

	if implicit && token.typ != yaml_VERSION_DIRECTIVE_TOKEN &&
		token.typ != yaml_TAG_DIRECTIVE_TOKEN &&
		token.typ != yaml_DOCUMENT_START_TOKEN &&
		token.typ != yaml_STREAM_END_TOKEN {
		// Parse an implicit document.
		if !parser.process_directives(nil, nil) {
			return false
		}
		parser.states = append(parser.states, yaml_PARSE_DOCUMENT_END_STATE)
		parser.state = yaml_PARSE_BLOCK_NODE_STATE

		var head_comment []byte
		if len(parser.head_comment) > 0 {
			// [Go] Scan the header comment backwards, and if an empty line is found, break
			//      the header so the part before the last empty line goes into the
			//      document header, while the bottom of it goes into a follow up event.
			for i := len(parser.head_comment) - 1; i > 0; i-- {
				if parser.head_comment[i] == '\n' {
					if i == len(parser.head_comment)-1 {
						head_comment = parser.head_comment[:i]
						parser.head_comment = parser.head_comment[i+1:]
						break
					} else if parser.head_comment[i-1] == '\n' {
						head_comment = parser.head_comment[:i-1]
						parser.head_comment = parser.head_comment[i+1:]
						break
					}
				}
			}
		}

		*event = yaml_event_t{
			typ:        yaml_DOCUMENT_START_EVENT,
			start_mark: token.start_mark,
			end_mark:   token.end_mark,

			head_comment: head_comment,
		}

	} else if token.typ != yaml_STREAM_END_TOKEN {
		// Parse an explicit document.
		var version_directive *yaml_version_directive_t
		var tag_directives []yaml_tag_directive_t
		start_mark := token.start_mark
		if !parser.process_directives(&version_directive, &tag_directives) {
			return false
		}
		token = parser.peek_token()
		if token == nil {
			return false
		}
		if token.typ != yaml_DOCUMENT_START_TOKEN {
			parser.set_parser_error("did not find expected <document start>", token.start_mark)
			return false
		}
		parser.states = append(parser.states, yaml_PARSE_DOCUMENT_END_STATE)
		parser.state = yaml_PARSE_DOCUMENT_CONTENT_STATE
		end_mark := token.end_mark

		*event = yaml_event_t{
			typ:               yaml_DOCUMENT_START_EVENT,
			start_mark:        start_mark,
			end_mark:          end_mark,
			version_directive: version_directive,
			tag_directives:    tag_directives,
			implicit:          false,
		}
		parser.skip_token()

	} else {
		// Parse the stream end.
		parser.state = yaml_PARSE_END_STATE
		*event = yaml_event_t{
			typ:        yaml_STREAM_END_EVENT,
			start_mark: token.start_mark,
			end_mark:   token.end_mark,
		}
		parser.skip_token()
	}

	return true
}

// Parse the productions:
// explicit_document    ::= DIRECTIVE* DOCUMENT-START block_node? DOCUMENT-END*
//
//	***********
func (parser *yaml_parser_t) parse_document_content(event *yaml_event_t) bool {
	token := parser.peek_token()
	if token == nil {
		return false
	}

	if token.typ == yaml_VERSION_DIRECTIVE_TOKEN ||
		token.typ == yaml_TAG_DIRECTIVE_TOKEN ||
		token.typ == yaml_DOCUMENT_START_TOKEN ||
		token.typ == yaml_DOCUMENT_END_TOKEN ||
		token.typ == yaml_STREAM_END_TOKEN {
		parser.state = parser.states[len(parser.states)-1]
		parser.states = parser.states[:len(parser.states)-1]
		return parser.process_empty_scalar(event,
			token.start_mark)
	}
	return parser.parse_node(event, true, false)
}

// Parse the productions:
// implicit_document    ::= block_node DOCUMENT-END*
//
//	*************
//
// explicit_document    ::= DIRECTIVE* DOCUMENT-START block_node? DOCUMENT-END*
func (parser *yaml_parser_t) parse_document_end(event *yaml_event_t) bool {
	token := parser.peek_token()
	if token == nil {
		return false
	}

	start_mark := token.start_mark
	end_mark := token.start_mark

	implicit := true
	if token.typ == yaml_DOCUMENT_END_TOKEN {
		end_mark = token.end_mark
		parser.skip_token()
		implicit = false
	}

	parser.tag_directives = parser.tag_directives[:0]

	parser.state = yaml_PARSE_DOCUMENT_START_STATE
	*event = yaml_event_t{
		typ:        yaml_DOCUMENT_END_EVENT,
		start_mark: start_mark,
		end_mark:   end_mark,
		implicit:   implicit,
	}
	parser.set_event_comments(event)
	if len(event.head_comment) > 0 && len(event.foot_comment) == 0 {
		event.foot_comment = event.head_comment
		event.head_comment = nil
	}
	return true
}

func (parser *yaml_parser_t) set_event_comments(event *yaml_event_t) {
	event.head_comment = parser.head_comment
	event.line_comment = parser.line_comment
	event.foot_comment = parser.foot_comment
	parser.head_comment = nil
	parser.line_comment = nil
	parser.foot_comment = nil
	parser.tail_comment = nil
	parser.stem_comment = nil
}

// Parse the productions:
// block_node_or_indentless_sequence    ::=
//
//	ALIAS
//	*****
//	| properties (block_content | indentless_block_sequence)?
//	  **********  *
//	| block_content | indentless_block_sequence
//	  *
//
// block_node           ::= ALIAS
//
//	*****
//	| properties block_content?
//	  ********** *
//	| block_content
//	  *
//
// flow_node            ::= ALIAS
//
//	*****
//	| properties flow_content?
//	  ********** *
//	| flow_content
//	  *
//
// properties           ::= TAG ANCHOR? | ANCHOR TAG?
//
//	*************************
//
// block_content        ::= block_collection | flow_collection | SCALAR
//
//	******
//
// flow_content         ::= flow_collection | SCALAR
//
//	******
func (parser *yaml_parser_t) parse_node(event *yaml_event_t, block, indentless_sequence bool) bool {

	token := parser.peek_token()
	if token == nil {
		return false
	}

	if token.typ == yaml_ALIAS_TOKEN {
		parser.state = parser.states[len(parser.states)-1]
		parser.states = parser.states[:len(parser.states)-1]
		*event = yaml_event_t{
			typ:        yaml_ALIAS_EVENT,
			start_mark: token.start_mark,
			end_mark:   token.end_mark,
			anchor:     token.value,
		}
		parser.set_event_comments(event)
		parser.skip_token()
		return true
	}

	start_mark := token.start_mark
	end_mark := token.start_mark

	var tag_token bool
	var tag_handle, tag_suffix, anchor []byte
	var tag_mark yaml_mark_t
	switch token.typ {
	case yaml_ANCHOR_TOKEN:
		anchor = token.value
		start_mark = token.start_mark
		end_mark = token.end_mark
		parser.skip_token()
		token = parser.peek_token()
		if token == nil {
			return false
		}
		if token.typ == yaml_TAG_TOKEN {
			tag_token = true
			tag_handle = token.value
			tag_suffix = token.suffix
			tag_mark = token.start_mark
			end_mark = token.end_mark
			parser.skip_token()
			token = parser.peek_token()
			if token == nil {
				return false
			}
		}
	case yaml_TAG_TOKEN:
		tag_token = true
		tag_handle = token.value
		tag_suffix = token.suffix
		start_mark = token.start_mark
		tag_mark = token.start_mark
		end_mark = token.end_mark
		parser.skip_token()
		token = parser.peek_token()
		if token == nil {
			return false
		}
		if token.typ == yaml_ANCHOR_TOKEN {
			anchor = token.value
			end_mark = token.end_mark
			parser.skip_token()
			token = parser.peek_token()
			if token == nil {
				return false
			}
		}
	}

	var tag []byte
	if tag_token {
		if len(tag_handle) == 0 {
			tag = tag_suffix
		} else {
			for i := range parser.tag_directives {
				if bytes.Equal(parser.tag_directives[i].handle, tag_handle) {
					tag = append([]byte(nil), parser.tag_directives[i].prefix...)
					tag = append(tag, tag_suffix...)
					break
				}
			}
			if len(tag) == 0 {
				parser.set_parser_error_context("while parsing a node", start_mark,
					"found undefined tag handle", tag_mark)
				return false
			}
		}
	}

	implicit := len(tag) == 0
	if indentless_sequence && token.typ == yaml_BLOCK_ENTRY_TOKEN {
		end_mark = token.end_mark
		parser.state = yaml_PARSE_INDENTLESS_SEQUENCE_ENTRY_STATE
		*event = yaml_event_t{
			typ:        yaml_SEQUENCE_START_EVENT,
			start_mark: start_mark,
			end_mark:   end_mark,
			anchor:     anchor,
			tag:        tag,
			implicit:   implicit,
			style:      yaml_style_t(yaml_BLOCK_SEQUENCE_STYLE),
		}
		return true
	}
	if token.typ == yaml_SCALAR_TOKEN {
		var plain_implicit, quoted_implicit bool
		end_mark = token.end_mark
		if (len(tag) == 0 && token.style == yaml_PLAIN_SCALAR_STYLE) || (len(tag) == 1 && tag[0] == '!') {
			plain_implicit = true
		} else if len(tag) == 0 {
			quoted_implicit = true
		}
		parser.state = parser.states[len(parser.states)-1]
		parser.states = parser.states[:len(parser.states)-1]

		*event = yaml_event_t{
			typ:             yaml_SCALAR_EVENT,
			start_mark:      start_mark,
			end_mark:        end_mark,
			anchor:          anchor,
			tag:             tag,
			value:           token.value,
			implicit:        plain_implicit,
			quoted_implicit: quoted_implicit,
			style:           yaml_style_t(token.style),
		}
		parser.set_event_comments(event)
		parser.skip_token()
		return true
	}
	if token.typ == yaml_FLOW_SEQUENCE_START_TOKEN {
		// [Go] Some of the events below can be merged as they differ only on style.
		end_mark = token.end_mark
		parser.state = yaml_PARSE_FLOW_SEQUENCE_FIRST_ENTRY_STATE
		*event = yaml_event_t{
			typ:        yaml_SEQUENCE_START_EVENT,
			start_mark: start_mark,
			end_mark:   end_mark,
			anchor:     anchor,
			tag:        tag,
			implicit:   implicit,
			style:      yaml_style_t(yaml_FLOW_SEQUENCE_STYLE),
		}
		parser.set_event_comments(event)
		return true
	}
	if token.typ == yaml_FLOW_MAPPING_START_TOKEN {
		end_mark = token.end_mark
		parser.state = yaml_PARSE_FLOW_MAPPING_FIRST_KEY_STATE
		*event = yaml_event_t{
			typ:        yaml_MAPPING_START_EVENT,
			start_mark: start_mark,
			end_mark:   end_mark,
			anchor:     anchor,
			tag:        tag,
			implicit:   implicit,
			style:      yaml_style_t(yaml_FLOW_MAPPING_STYLE),
		}
		parser.set_event_comments(event)
		return true
	}
	if block && token.typ == yaml_BLOCK_SEQUENCE_START_TOKEN {
		end_mark = token.end_mark
		parser.state = yaml_PARSE_BLOCK_SEQUENCE_FIRST_ENTRY_STATE
		*event = yaml_event_t{
			typ:        yaml_SEQUENCE_START_EVENT,
			start_mark: start_mark,
			end_mark:   end_mark,
			anchor:     anchor,
			tag:        tag,
			implicit:   implicit,
			style:      yaml_style_t(yaml_BLOCK_SEQUENCE_STYLE),
		}
		if parser.stem_comment != nil {
			event.head_comment = parser.stem_comment
			parser.stem_comment = nil
		}
		return true
	}
	if block && token.typ == yaml_BLOCK_MAPPING_START_TOKEN {
		end_mark = token.end_mark
		parser.state = yaml_PARSE_BLOCK_MAPPING_FIRST_KEY_STATE
		*event = yaml_event_t{
			typ:        yaml_MAPPING_START_EVENT,
			start_mark: start_mark,
			end_mark:   end_mark,
			anchor:     anchor,
			tag:        tag,
			implicit:   implicit,
			style:      yaml_style_t(yaml_BLOCK_MAPPING_STYLE),
		}
		if parser.stem_comment != nil {
			event.head_comment = parser.stem_comment
			parser.stem_comment = nil
		}
		return true
	}
	if len(anchor) > 0 || len(tag) > 0 {
		parser.state = parser.states[len(parser.states)-1]
		parser.states = parser.states[:len(parser.states)-1]

		*event = yaml_event_t{
			typ:             yaml_SCALAR_EVENT,
			start_mark:      start_mark,
			end_mark:        end_mark,
			anchor:          anchor,
			tag:             tag,
			implicit:        implicit,
			quoted_implicit: false,
			style:           yaml_style_t(yaml_PLAIN_SCALAR_STYLE),
		}
		return true
	}

	context := "while parsing a flow node"
	if block {
		context = "while parsing a block node"
	}
	parser.set_parser_error_context(context, start_mark,
		"did not find expected node content", token.start_mark)
	return false
}

// Parse the productions:
// block_sequence ::= BLOCK-SEQUENCE-START (BLOCK-ENTRY block_node?)* BLOCK-END
//
//	********************  *********** *             *********
func (parser *yaml_parser_t) parse_block_sequence_entry(event *yaml_event_t, first bool) bool {
	if first {
		token := parser.peek_token()
		if token == nil {
			return false
		}
		parser.marks = append(parser.marks, token.start_mark)
		parser.skip_token()
	}

	token := parser.peek_token()
	if token == nil {
		return false
	}

	if token.typ == yaml_BLOCK_ENTRY_TOKEN {
		mark := token.end_mark
		prior_head_len := len(parser.head_comment)
		parser.skip_token()
		parser.split_stem_comment(prior_head_len)
		token = parser.peek_token()
		if token == nil {
			return false
		}
		if token.typ != yaml_BLOCK_ENTRY_TOKEN && token.typ != yaml_BLOCK_END_TOKEN {
			parser.states = append(parser.states, yaml_PARSE_BLOCK_SEQUENCE_ENTRY_STATE)
			return parser.parse_node(event, true, false)
		} else {
			parser.state = yaml_PARSE_BLOCK_SEQUENCE_ENTRY_STATE
			return parser.process_empty_scalar(event, mark)
		}
	}
	if token.typ == yaml_BLOCK_END_TOKEN {
		parser.state = parser.states[len(parser.states)-1]
		parser.states = parser.states[:len(parser.states)-1]
		parser.marks = parser.marks[:len(parser.marks)-1]

		*event = yaml_event_t{
			typ:        yaml_SEQUENCE_END_EVENT,
			start_mark: token.start_mark,
			end_mark:   token.end_mark,
		}

		parser.skip_token()
		return true
	}

	context_mark := parser.marks[len(parser.marks)-1]
	parser.marks = parser.marks[:len(parser.marks)-1]
	return parser.set_parser_error_context("while parsing a block collection", context_mark,
		"did not find expected '-' indicator", token.start_mark)
}

// Parse the productions:
// indentless_sequence  ::= (BLOCK-ENTRY block_node?)+
//
//	*********** *
func (parser *yaml_parser_t) parse_indentless_sequence_entry(event *yaml_event_t) bool {
	token := parser.peek_token()
	if token == nil {
		return false
	}

	if token.typ == yaml_BLOCK_ENTRY_TOKEN {
		mark := token.end_mark
		prior_head_len := len(parser.head_comment)
		parser.skip_token()
		parser.split_stem_comment(prior_head_len)
		token = parser.peek_token()
		if token == nil {
			return false
		}
		if token.typ != yaml_BLOCK_ENTRY_TOKEN &&
			token.typ != yaml_KEY_TOKEN &&
			token.typ != yaml_VALUE_TOKEN &&
			token.typ != yaml_BLOCK_END_TOKEN {
			parser.states = append(parser.states, yaml_PARSE_INDENTLESS_SEQUENCE_ENTRY_STATE)
			return parser.parse_node(event, true, false)
		}
		parser.state = yaml_PARSE_INDENTLESS_SEQUENCE_ENTRY_STATE
		return parser.process_empty_scalar(event, mark)
	}
	parser.state = parser.states[len(parser.states)-1]
	parser.states = parser.states[:len(parser.states)-1]

	*event = yaml_event_t{
		typ:        yaml_SEQUENCE_END_EVENT,
		start_mark: token.start_mark,
		end_mark:   token.start_mark, // [Go] Shouldn't this be token.end_mark?
	}
	return true
}

// Split stem comment from head comment.
//
// When a sequence or map is found under a sequence entry, the former head comment
// is assigned to the underlying sequence or map as a whole, not the individual
// sequence or map entry as would be expected otherwise. To handle this case the
// previous head comment is moved aside as the stem comment.
func (parser *yaml_parser_t) split_stem_comment(stem_len int) {
	if stem_len == 0 {
		return
	}

	token := parser.peek_token()
	if token == nil || token.typ != yaml_BLOCK_SEQUENCE_START_TOKEN && token.typ != yaml_BLOCK_MAPPING_START_TOKEN {
		return
	}

	parser.stem_comment = parser.head_comment[:stem_len]
	if len(parser.head_comment) == stem_len {
		parser.head_comment = nil
	} else {
		// Copy suffix to prevent very strange bugs if someone ever appends
		// further bytes to the prefix in the stem_comment slice above.
		parser.head_comment = append([]byte(nil), parser.head_comment[stem_len+1:]...)
	}
}

// Parse the productions:
// block_mapping        ::= BLOCK-MAPPING_START
//
//	*******************
//	((KEY block_node_or_indentless_sequence?)?
//	  *** *
//	(VALUE block_node_or_indentless_sequence?)?)*
//
//	BLOCK-END
//	*********
func (parser *yaml_parser_t) parse_block_mapping_key(event *yaml_event_t, first bool) bool {
	if first {
		token := parser.peek_token()
		if token == nil {
			return false
		}
		parser.marks = append(parser.marks, token.start_mark)
		parser.skip_token()
	}

	token := parser.peek_token()
	if token == nil {
		return false
	}

	// [Go] A tail comment was left from the prior mapping value processed. Emit an event
	//      as it needs to be processed with that value and not the following key.
	if len(parser.tail_comment) > 0 {
		*event = yaml_event_t{
			typ:          yaml_TAIL_COMMENT_EVENT,
			start_mark:   token.start_mark,
			end_mark:     token.end_mark,
			foot_comment: parser.tail_comment,
		}
		parser.tail_comment = nil
		return true
	}

	switch token.typ {
	case yaml_KEY_TOKEN:
		mark := token.end_mark
		parser.skip_token()
		token = parser.peek_token()
		if token == nil {
			return false
		}
		if token.typ != yaml_KEY_TOKEN &&
			token.typ != yaml_VALUE_TOKEN &&
			token.typ != yaml_BLOCK_END_TOKEN {
			parser.states = append(parser.states, yaml_PARSE_BLOCK_MAPPING_VALUE_STATE)
			return parser.parse_node(event, true, true)
		} else {
			parser.state = yaml_PARSE_BLOCK_MAPPING_VALUE_STATE
			return parser.process_empty_scalar(event, mark)
		}
	case yaml_BLOCK_END_TOKEN:
		parser.state = parser.states[len(parser.states)-1]
		parser.states = parser.states[:len(parser.states)-1]
		parser.marks = parser.marks[:len(parser.marks)-1]
		*event = yaml_event_t{
			typ:        yaml_MAPPING_END_EVENT,
			start_mark: token.start_mark,
			end_mark:   token.end_mark,
		}
		parser.set_event_comments(event)
		parser.skip_token()
		return true
	}

	context_mark := parser.marks[len(parser.marks)-1]
	parser.marks = parser.marks[:len(parser.marks)-1]
	return parser.set_parser_error_context("while parsing a block mapping", context_mark,
		"did not find expected key", token.start_mark)
}

// Parse the productions:
// block_mapping        ::= BLOCK-MAPPING_START
//
//	((KEY block_node_or_indentless_sequence?)?
//
//	(VALUE block_node_or_indentless_sequence?)?)*
//	 ***** *
//	BLOCK-END
func (parser *yaml_parser_t) parse_block_mapping_value(event *yaml_event_t) bool {
	token := parser.peek_token()
	if token == nil {
		return false
	}
	if token.typ == yaml_VALUE_TOKEN {
		mark := token.end_mark
		parser.skip_token()
		token = parser.peek_token()
		if token == nil {
			return false
		}
		if token.typ != yaml_KEY_TOKEN &&
			token.typ != yaml_VALUE_TOKEN &&
			token.typ != yaml_BLOCK_END_TOKEN {
			parser.states = append(parser.states, yaml_PARSE_BLOCK_MAPPING_KEY_STATE)
			return parser.parse_node(event, true, true)
		}
		parser.state = yaml_PARSE_BLOCK_MAPPING_KEY_STATE
		return parser.process_empty_scalar(event, mark)
	}
	parser.state = yaml_PARSE_BLOCK_MAPPING_KEY_STATE
	return parser.process_empty_scalar(event, token.start_mark)
}

// Parse the productions:
// flow_sequence        ::= FLOW-SEQUENCE-START
//
//	*******************
//	(flow_sequence_entry FLOW-ENTRY)*
//	 *                   **********
//	flow_sequence_entry?
//	*
//	FLOW-SEQUENCE-END
//	*****************
//
// flow_sequence_entry  ::= flow_node | KEY flow_node? (VALUE flow_node?)?
//
//	*
func (parser *yaml_parser_t) parse_flow_sequence_entry(event *yaml_event_t, first bool) bool {
	if first {
		token := parser.peek_token()
		if token == nil {
			return false
		}
		parser.marks = append(parser.marks, token.start_mark)
		parser.skip_token()
	}
	token := parser.peek_token()
	if token == nil {
		return false
	}
	if token.typ != yaml_FLOW_SEQUENCE_END_TOKEN {
		if !first {
			if token.typ == yaml_FLOW_ENTRY_TOKEN {
				parser.skip_token()
				token = parser.peek_token()
				if token == nil {
					return false
				}
			} else {
				context_mark := parser.marks[len(parser.marks)-1]
				parser.marks = parser.marks[:len(parser.marks)-1]
				return parser.set_parser_error_context("while parsing a flow sequence", context_mark,
					"did not find expected ',' or ']'", token.start_mark)
			}
		}

		if token.typ == yaml_KEY_TOKEN {
			parser.state = yaml_PARSE_FLOW_SEQUENCE_ENTRY_MAPPING_KEY_STATE
			*event = yaml_event_t{
				typ:        yaml_MAPPING_START_EVENT,
				start_mark: token.start_mark,
				end_mark:   token.end_mark,
				implicit:   true,
				style:      yaml_style_t(yaml_FLOW_MAPPING_STYLE),
			}
			parser.skip_token()
			return true
		} else if token.typ != yaml_FLOW_SEQUENCE_END_TOKEN {
			parser.states = append(parser.states, yaml_PARSE_FLOW_SEQUENCE_ENTRY_STATE)
			return parser.parse_node(event, false, false)
		}
	}

	parser.state = parser.states[len(parser.states)-1]
	parser.states = parser.states[:len(parser.states)-1]
	parser.marks = parser.marks[:len(parser.marks)-1]

	*event = yaml_event_t{
		typ:        yaml_SEQUENCE_END_EVENT,
		start_mark: token.start_mark,
		end_mark:   token.end_mark,
	}
	parser.set_event_comments(event)

	parser.skip_token()
	return true
}

// Parse the productions:
// flow_sequence_entry  ::= flow_node | KEY flow_node? (VALUE flow_node?)?
//
//	*** *
func (parser *yaml_parser_t) parse_flow_sequence_entry_mapping_key(event *yaml_event_t) bool {
	token := parser.peek_token()
	if token == nil {
		return false
	}
	if token.typ != yaml_VALUE_TOKEN &&
		token.typ != yaml_FLOW_ENTRY_TOKEN &&
		token.typ != yaml_FLOW_SEQUENCE_END_TOKEN {
		parser.states = append(parser.states, yaml_PARSE_FLOW_SEQUENCE_ENTRY_MAPPING_VALUE_STATE)
		return parser.parse_node(event, false, false)
	}
	mark := token.end_mark
	parser.skip_token()
	parser.state = yaml_PARSE_FLOW_SEQUENCE_ENTRY_MAPPING_VALUE_STATE
	return parser.process_empty_scalar(event, mark)
}

// Parse the productions:
// flow_sequence_entry  ::= flow_node | KEY flow_node? (VALUE flow_node?)?
//
//	***** *
func (parser *yaml_parser_t) parse_flow_sequence_entry_mapping_value(event *yaml_event_t) bool {
	token := parser.peek_token()
	if token == nil {
		return false
	}
	if token.typ == yaml_VALUE_TOKEN {
		parser.skip_token()
		token := parser.peek_token()
		if token == nil {
			return false
		}
		if token.typ != yaml_FLOW_ENTRY_TOKEN && token.typ != yaml_FLOW_SEQUENCE_END_TOKEN {
			parser.states = append(parser.states, yaml_PARSE_FLOW_SEQUENCE_ENTRY_MAPPING_END_STATE)
			return parser.parse_node(event, false, false)
		}
	}
	parser.state = yaml_PARSE_FLOW_SEQUENCE_ENTRY_MAPPING_END_STATE
	return parser.process_empty_scalar(event, token.start_mark)
}

// Parse the productions:
// flow_sequence_entry  ::= flow_node | KEY flow_node? (VALUE flow_node?)?
//
//	*
func (parser *yaml_parser_t) parse_flow_sequence_entry_mapping_end(event *yaml_event_t) bool {
	token := parser.peek_token()
	if token == nil {
		return false
	}
	parser.state = yaml_PARSE_FLOW_SEQUENCE_ENTRY_STATE
	*event = yaml_event_t{
		typ:        yaml_MAPPING_END_EVENT,
		start_mark: token.start_mark,
		end_mark:   token.start_mark, // [Go] Shouldn't this be end_mark?
	}
	return true
}

// Parse the productions:
// flow_mapping         ::= FLOW-MAPPING-START
//
//	******************
//	(flow_mapping_entry FLOW-ENTRY)*
//	 *                  **********
//	flow_mapping_entry?
//	******************
//	FLOW-MAPPING-END
//	****************
//
// flow_mapping_entry   ::= flow_node | KEY flow_node? (VALUE flow_node?)?
//   - *** *
func (parser *yaml_parser_t) parse_flow_mapping_key(event *yaml_event_t, first bool) bool {
	if first {
		token := parser.peek_token()
		parser.marks = append(parser.marks, token.start_mark)
		parser.skip_token()
	}

	token := parser.peek_token()
	if token == nil {
		return false
	}

	if token.typ != yaml_FLOW_MAPPING_END_TOKEN {
		if !first {
			if token.typ == yaml_FLOW_ENTRY_TOKEN {
				parser.skip_token()
				token = parser.peek_token()
				if token == nil {
					return false
				}
			} else {
				context_mark := parser.marks[len(parser.marks)-1]
				parser.marks = parser.marks[:len(parser.marks)-1]
				return parser.set_parser_error_context("while parsing a flow mapping", context_mark,
					"did not find expected ',' or '}'", token.start_mark)
			}
		}

		if token.typ == yaml_KEY_TOKEN {
			parser.skip_token()
			token = parser.peek_token()
			if token == nil {
				return false
			}
			if token.typ != yaml_VALUE_TOKEN &&
				token.typ != yaml_FLOW_ENTRY_TOKEN &&
				token.typ != yaml_FLOW_MAPPING_END_TOKEN {
				parser.states = append(parser.states, yaml_PARSE_FLOW_MAPPING_VALUE_STATE)
				return parser.parse_node(event, false, false)
			} else {
				parser.state = yaml_PARSE_FLOW_MAPPING_VALUE_STATE
				return parser.process_empty_scalar(event, token.start_mark)
			}
		} else if token.typ != yaml_FLOW_MAPPING_END_TOKEN {
			parser.states = append(parser.states, yaml_PARSE_FLOW_MAPPING_EMPTY_VALUE_STATE)
			return parser.parse_node(event, false, false)
		}
	}

	parser.state = parser.states[len(parser.states)-1]
	parser.states = parser.states[:len(parser.states)-1]
	parser.marks = parser.marks[:len(parser.marks)-1]
	*event = yaml_event_t{
		typ:        yaml_MAPPING_END_EVENT,
		start_mark: token.start_mark,
		end_mark:   token.end_mark,
	}
	parser.set_event_comments(event)
	parser.skip_token()
	return true
}

// Parse the productions:
// flow_mapping_entry   ::= flow_node | KEY flow_node? (VALUE flow_node?)?
//   - ***** *
func (parser *yaml_parser_t) parse_flow_mapping_value(event *yaml_event_t, empty bool) bool {
	token := parser.peek_token()
	if token == nil {
		return false
	}
	if empty {
		parser.state = yaml_PARSE_FLOW_MAPPING_KEY_STATE
		return parser.process_empty_scalar(event, token.start_mark)
	}
	if token.typ == yaml_VALUE_TOKEN {
		parser.skip_token()
		token = parser.peek_token()
		if token == nil {
			return false
		}
		if token.typ != yaml_FLOW_ENTRY_TOKEN && token.typ != yaml_FLOW_MAPPING_END_TOKEN {
			parser.states = append(parser.states, yaml_PARSE_FLOW_MAPPING_KEY_STATE)
			return parser.parse_node(event, false, false)
		}
	}
	parser.state = yaml_PARSE_FLOW_MAPPING_KEY_STATE
	return parser.process_empty_scalar(event, token.start_mark)
}

// Generate an empty scalar event.
func (parser *yaml_parser_t) process_empty_scalar(event *yaml_event_t, mark yaml_mark_t) bool {
	*event = yaml_event_t{
		typ:        yaml_SCALAR_EVENT,
		start_mark: mark,
		end_mark:   mark,
		value:      nil, // Empty
		implicit:   true,
		style:      yaml_style_t(yaml_PLAIN_SCALAR_STYLE),
	}
	return true
}

var default_tag_directives = []yaml_tag_directive_t{
	{[]byte("!"), []byte("!")},
	{[]byte("!!"), []byte("tag:yaml.org,2002:")},
}

// Parse directives.
func (parser *yaml_parser_t) process_directives(version_directive_ref **yaml_version_directive_t,
	tag_directives_ref *[]yaml_tag_directive_t) bool {

	var version_directive *yaml_version_directive_t
	var tag_directives []yaml_tag_directive_t

	token := parser.peek_token()
	if token == nil {
		return false
	}

	for token.typ == yaml_VERSION_DIRECTIVE_TOKEN || token.typ == yaml_TAG_DIRECTIVE_TOKEN {
		switch token.typ {
		case yaml_VERSION_DIRECTIVE_TOKEN:
			if version_directive != nil {
				parser.set_parser_error("found duplicate %YAML directive", token.start_mark)
				return false
			}
			if token.major != 1 || token.minor != 1 {
				parser.set_parser_error("found incompatible YAML document", token.start_mark)
				return false
			}
			version_directive = &yaml_version_directive_t{
				major: token.major,
				minor: token.minor,
			}
		case yaml_TAG_DIRECTIVE_TOKEN:
			value := yaml_tag_directive_t{
				handle: token.value,
				prefix: token.prefix,
			}
			if !parser.append_tag_directive(value, false, token.start_mark) {
				return false
			}
			tag_directives = append(tag_directives, value)
		}

		parser.skip_token()
		token = parser.peek_token()
		if token == nil {
			return false
		}
	}

	for i := range default_tag_directives {
		if !parser.append_tag_directive(default_tag_directives[i], true, token.start_mark) {
			return false
		}
	}

	if version_directive_ref != nil {
		*version_directive_ref = version_directive
	}
	if tag_directives_ref != nil {
		*tag_directives_ref = tag_directives
	}
	return true
}

// Append a tag directive to the directives stack.
func (parser *yaml_parser_t) append_tag_directive(value yaml_tag_directive_t, allow_duplicates bool, mark yaml_mark_t) bool {
	for i := range parser.tag_directives {
		if bytes.Equal(value.handle, parser.tag_directives[i].handle) {
			if allow_duplicates {
				return true
			}
			return parser.set_parser_error("found duplicate %TAG directive", mark)
		}
	}

	// [Go] I suspect the copy is unnecessary. This was likely done
	// because there was no way to track ownership of the data.
	value_copy := yaml_tag_directive_t{
		handle: make([]byte, len(value.handle)),
		prefix: make([]byte, len(value.prefix)),
	}
	copy(value_copy.handle, value.handle)
	copy(value_copy.prefix, value.prefix)
	parser.tag_directives = append(parser.tag_directives, value_copy)
	return true
}
