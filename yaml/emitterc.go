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
	"fmt"
)

// Set an emitter error and return false.
func yaml_emitter_set_emitter_error(emitter *yaml_emitter_t, problem string) bool {
	emitter.error = yaml_EMITTER_ERROR
	emitter.problem = problem
	return false
}

// Emit an event.
func yaml_emitter_emit(emitter *yaml_emitter_t, event *yaml_event_t) bool {
	emitter.events = append(emitter.events, *event)
	for !yaml_emitter_need_more_events(emitter) {
		event := &emitter.events[emitter.events_head]
		if !yaml_emitter_analyze_event(emitter, event) {
			return false
		}
		if !yaml_emitter_state_machine(emitter, event) {
			return false
		}
		yaml_event_delete(event)
		emitter.events_head++
	}
	return true
}

// Check if we need to accumulate more events before emitting.
//
// We accumulate extra
//   - 1 event for DOCUMENT-START
//   - 2 events for SEQUENCE-START
//   - 3 events for MAPPING-START
func yaml_emitter_need_more_events(emitter *yaml_emitter_t) bool {
	if emitter.events_head == len(emitter.events) {
		return true
	}
	var accumulate int
	switch emitter.events[emitter.events_head].typ {
	case yaml_DOCUMENT_START_EVENT:
		accumulate = 1
	case yaml_SEQUENCE_START_EVENT:
		accumulate = 2
	case yaml_MAPPING_START_EVENT:
		accumulate = 3
	default:
		return false
	}
	if len(emitter.events)-emitter.events_head > accumulate {
		return false
	}
	var level int
	for i := emitter.events_head; i < len(emitter.events); i++ {
		switch emitter.events[i].typ {
		case yaml_STREAM_START_EVENT, yaml_DOCUMENT_START_EVENT, yaml_SEQUENCE_START_EVENT, yaml_MAPPING_START_EVENT:
			level++
		case yaml_STREAM_END_EVENT, yaml_DOCUMENT_END_EVENT, yaml_SEQUENCE_END_EVENT, yaml_MAPPING_END_EVENT:
			level--
		}
		if level == 0 {
			return false
		}
	}
	return true
}

// Append a directive to the directives stack.
func yaml_emitter_append_tag_directive(emitter *yaml_emitter_t, value *yaml_tag_directive_t, allow_duplicates bool) bool {
	for i := 0; i < len(emitter.tag_directives); i++ {
		if bytes.Equal(value.handle, emitter.tag_directives[i].handle) {
			if allow_duplicates {
				return true
			}
			return yaml_emitter_set_emitter_error(emitter, "duplicate %TAG directive")
		}
	}

	// [Go] Do we actually need to copy this given garbage collection
	// and the lack of deallocating destructors?
	tag_copy := yaml_tag_directive_t{
		handle: make([]byte, len(value.handle)),
		prefix: make([]byte, len(value.prefix)),
	}
	copy(tag_copy.handle, value.handle)
	copy(tag_copy.prefix, value.prefix)
	emitter.tag_directives = append(emitter.tag_directives, tag_copy)
	return true
}

// Increase the indentation level.
func yaml_emitter_increase_indent(emitter *yaml_emitter_t, flow, indentless bool) {
	emitter.indents = append(emitter.indents, emitter.indent)
	if emitter.indent < 0 {
		if flow {
			emitter.indent = emitter.best_indent
		} else {
			emitter.indent = 0
		}
	} else if !indentless {
		// [Go] This was changed so that indentations are more regular.
		if emitter.states[len(emitter.states)-1] == yaml_EMIT_BLOCK_SEQUENCE_ITEM_STATE {
			// The first indent inside a sequence will just skip the "- " indicator.
			emitter.indent += 2
		} else {
			// Everything else aligns to the chosen indentation.
			emitter.indent = emitter.best_indent * ((emitter.indent + emitter.best_indent) / emitter.best_indent)
		}
	}
}

// State dispatcher.
func yaml_emitter_state_machine(emitter *yaml_emitter_t, event *yaml_event_t) bool {
	switch emitter.state {
	default:
	case yaml_EMIT_STREAM_START_STATE:
		return yaml_emitter_emit_stream_start(emitter, event)

	case yaml_EMIT_FIRST_DOCUMENT_START_STATE:
		return yaml_emitter_emit_document_start(emitter, event, true)

	case yaml_EMIT_DOCUMENT_START_STATE:
		return yaml_emitter_emit_document_start(emitter, event, false)

	case yaml_EMIT_DOCUMENT_CONTENT_STATE:
		return yaml_emitter_emit_document_content(emitter, event)

	case yaml_EMIT_DOCUMENT_END_STATE:
		return yaml_emitter_emit_document_end(emitter, event)

	case yaml_EMIT_FLOW_SEQUENCE_FIRST_ITEM_STATE:
		return yaml_emitter_emit_flow_sequence_item(emitter, event, true, false)

	case yaml_EMIT_FLOW_SEQUENCE_TRAIL_ITEM_STATE:
		return yaml_emitter_emit_flow_sequence_item(emitter, event, false, true)

	case yaml_EMIT_FLOW_SEQUENCE_ITEM_STATE:
		return yaml_emitter_emit_flow_sequence_item(emitter, event, false, false)

	case yaml_EMIT_FLOW_MAPPING_FIRST_KEY_STATE:
		return yaml_emitter_emit_flow_mapping_key(emitter, event, true, false)

	case yaml_EMIT_FLOW_MAPPING_TRAIL_KEY_STATE:
		return yaml_emitter_emit_flow_mapping_key(emitter, event, false, true)

	case yaml_EMIT_FLOW_MAPPING_KEY_STATE:
		return yaml_emitter_emit_flow_mapping_key(emitter, event, false, false)

	case yaml_EMIT_FLOW_MAPPING_SIMPLE_VALUE_STATE:
		return yaml_emitter_emit_flow_mapping_value(emitter, event, true)

	case yaml_EMIT_FLOW_MAPPING_VALUE_STATE:
		return yaml_emitter_emit_flow_mapping_value(emitter, event, false)

	case yaml_EMIT_BLOCK_SEQUENCE_FIRST_ITEM_STATE:
		return yaml_emitter_emit_block_sequence_item(emitter, event, true)

	case yaml_EMIT_BLOCK_SEQUENCE_ITEM_STATE:
		return yaml_emitter_emit_block_sequence_item(emitter, event, false)

	case yaml_EMIT_BLOCK_MAPPING_FIRST_KEY_STATE:
		return yaml_emitter_emit_block_mapping_key(emitter, event, true)

	case yaml_EMIT_BLOCK_MAPPING_KEY_STATE:
		return yaml_emitter_emit_block_mapping_key(emitter, event, false)

	case yaml_EMIT_BLOCK_MAPPING_SIMPLE_VALUE_STATE:
		return yaml_emitter_emit_block_mapping_value(emitter, event, true)

	case yaml_EMIT_BLOCK_MAPPING_VALUE_STATE:
		return yaml_emitter_emit_block_mapping_value(emitter, event, false)

	case yaml_EMIT_END_STATE:
		return yaml_emitter_set_emitter_error(emitter, "expected nothing after STREAM-END")
	}
	panic("invalid emitter state")
}

// Expect STREAM-START.
func yaml_emitter_emit_stream_start(emitter *yaml_emitter_t, event *yaml_event_t) bool {
	if event.typ != yaml_STREAM_START_EVENT {
		return yaml_emitter_set_emitter_error(emitter, "expected STREAM-START")
	}
	if emitter.encoding == yaml_ANY_ENCODING {
		emitter.encoding = event.encoding
		if emitter.encoding == yaml_ANY_ENCODING {
			emitter.encoding = yaml_UTF8_ENCODING
		}
	}
	if emitter.best_indent < 2 || emitter.best_indent > 9 {
		emitter.best_indent = 2
	}
	if emitter.best_width >= 0 && emitter.best_width <= emitter.best_indent*2 {
		emitter.best_width = 80
	}
	if emitter.best_width < 0 {
		emitter.best_width = 1<<31 - 1
	}
	if emitter.writer.line_break == yaml_ANY_BREAK {
		emitter.writer.line_break = yaml_LN_BREAK
	}

	emitter.indent = -1
	emitter.writer.line = 0
	emitter.writer.column = 0
	emitter.writer.whitespace = true
	emitter.writer.indention = true
	emitter.writer.space_above = true
	emitter.foot_indent = -1

	if emitter.encoding != yaml_UTF8_ENCODING {
		write_bom(&emitter.writer)
	}
	emitter.state = yaml_EMIT_FIRST_DOCUMENT_START_STATE
	return true
}

// Expect DOCUMENT-START or STREAM-END.
func yaml_emitter_emit_document_start(emitter *yaml_emitter_t, event *yaml_event_t, first bool) bool {

	if event.typ == yaml_DOCUMENT_START_EVENT {

		if event.version_directive != nil {
			if !yaml_emitter_analyze_version_directive(emitter, event.version_directive) {
				return false
			}
		}

		for i := 0; i < len(event.tag_directives); i++ {
			tag_directive := &event.tag_directives[i]
			if !yaml_emitter_analyze_tag_directive(emitter, tag_directive) {
				return false
			}
			if !yaml_emitter_append_tag_directive(emitter, tag_directive, false) {
				return false
			}
		}

		for i := 0; i < len(default_tag_directives); i++ {
			tag_directive := &default_tag_directives[i]
			if !yaml_emitter_append_tag_directive(emitter, tag_directive, true) {
				return false
			}
		}

		implicit := event.implicit
		if !first || emitter.canonical {
			implicit = false
		}

		if emitter.open_ended && (event.version_directive != nil || len(event.tag_directives) > 0) {
			yaml_emitter_write_indicator(emitter, []byte("..."), true, false, false)
			yaml_emitter_write_indent(emitter)
		}

		if event.version_directive != nil {
			implicit = false
			yaml_emitter_write_indicator(emitter, []byte("%YAML"), true, false, false)
			yaml_emitter_write_indicator(emitter, []byte("1.1"), true, false, false)
			yaml_emitter_write_indent(emitter)
		}

		if len(event.tag_directives) > 0 {
			implicit = false
			for i := 0; i < len(event.tag_directives); i++ {
				tag_directive := &event.tag_directives[i]
				yaml_emitter_write_indicator(emitter, []byte("%TAG"), true, false, false)
				yaml_emitter_write_tag_handle(emitter, tag_directive.handle)
				yaml_emitter_write_tag_content(emitter, tag_directive.prefix, true)
				yaml_emitter_write_indent(emitter)
			}
		}

		if yaml_emitter_check_empty_document(emitter) {
			implicit = false
		}
		if !implicit {
			yaml_emitter_write_indent(emitter)
			yaml_emitter_write_indicator(emitter, []byte("---"), true, false, false)
			if emitter.canonical || true {
				yaml_emitter_write_indent(emitter)
			}
		}

		if len(emitter.head_comment) > 0 {
			yaml_emitter_process_head_comment(emitter)
			put_break(&emitter.writer)
		}

		emitter.state = yaml_EMIT_DOCUMENT_CONTENT_STATE
		return true
	}

	if event.typ == yaml_STREAM_END_EVENT {
		if emitter.open_ended {
			yaml_emitter_write_indicator(emitter, []byte("..."), true, false, false)
			yaml_emitter_write_indent(emitter)
		}
		emitter.state = yaml_EMIT_END_STATE
		return true
	}

	return yaml_emitter_set_emitter_error(emitter, "expected DOCUMENT-START or STREAM-END")
}

// Expect the root node.
func yaml_emitter_emit_document_content(emitter *yaml_emitter_t, event *yaml_event_t) bool {
	emitter.states = append(emitter.states, yaml_EMIT_DOCUMENT_END_STATE)

	yaml_emitter_process_head_comment(emitter)
	if !yaml_emitter_emit_node(emitter, event, true, false, false, false) {
		return false
	}
	yaml_emitter_process_line_comment(emitter)
	yaml_emitter_process_foot_comment(emitter)
	return true
}

// Expect DOCUMENT-END.
func yaml_emitter_emit_document_end(emitter *yaml_emitter_t, event *yaml_event_t) bool {
	if event.typ != yaml_DOCUMENT_END_EVENT {
		return yaml_emitter_set_emitter_error(emitter, "expected DOCUMENT-END")
	}
	// [Go] Force document foot separation.
	emitter.foot_indent = 0
	yaml_emitter_process_foot_comment(emitter)
	emitter.foot_indent = -1
	yaml_emitter_write_indent(emitter)
	if !event.implicit {
		// [Go] Allocate the slice elsewhere.
		yaml_emitter_write_indicator(emitter, []byte("..."), true, false, false)
		yaml_emitter_write_indent(emitter)
	}
	emitter.state = yaml_EMIT_DOCUMENT_START_STATE
	emitter.tag_directives = emitter.tag_directives[:0]
	return true
}

// Expect a flow item node.
func yaml_emitter_emit_flow_sequence_item(emitter *yaml_emitter_t, event *yaml_event_t, first, trail bool) bool {
	if first {
		yaml_emitter_write_indicator(emitter, []byte{'['}, true, true, false)
		yaml_emitter_increase_indent(emitter, true, false)
		emitter.flow_level++
	}

	if event.typ == yaml_SEQUENCE_END_EVENT {
		if emitter.canonical && !first && !trail {
			yaml_emitter_write_indicator(emitter, []byte{','}, false, false, false)
		}
		emitter.flow_level--
		emitter.indent = emitter.indents[len(emitter.indents)-1]
		emitter.indents = emitter.indents[:len(emitter.indents)-1]
		if emitter.writer.column == 0 || emitter.canonical && !first {
			yaml_emitter_write_indent(emitter)
		}
		yaml_emitter_write_indicator(emitter, []byte{']'}, false, false, false)
		yaml_emitter_process_line_comment(emitter)
		yaml_emitter_process_foot_comment(emitter)
		emitter.state = emitter.states[len(emitter.states)-1]
		emitter.states = emitter.states[:len(emitter.states)-1]

		return true
	}

	if !first && !trail {
		yaml_emitter_write_indicator(emitter, []byte{','}, false, false, false)
	}

	yaml_emitter_process_head_comment(emitter)
	if emitter.writer.column == 0 {
		yaml_emitter_write_indent(emitter)
	}

	if emitter.canonical || emitter.writer.column > emitter.best_width {
		yaml_emitter_write_indent(emitter)
	}
	if len(emitter.line_comment)+len(emitter.foot_comment)+len(emitter.tail_comment) > 0 {
		emitter.states = append(emitter.states, yaml_EMIT_FLOW_SEQUENCE_TRAIL_ITEM_STATE)
	} else {
		emitter.states = append(emitter.states, yaml_EMIT_FLOW_SEQUENCE_ITEM_STATE)
	}
	if !yaml_emitter_emit_node(emitter, event, false, true, false, false) {
		return false
	}
	if len(emitter.line_comment)+len(emitter.foot_comment)+len(emitter.tail_comment) > 0 {
		yaml_emitter_write_indicator(emitter, []byte{','}, false, false, false)
	}
	yaml_emitter_process_line_comment(emitter)
	yaml_emitter_process_foot_comment(emitter)
	return true
}

// Expect a flow key node.
func yaml_emitter_emit_flow_mapping_key(emitter *yaml_emitter_t, event *yaml_event_t, first, trail bool) bool {
	if first {
		yaml_emitter_write_indicator(emitter, []byte{'{'}, true, true, false)
		yaml_emitter_increase_indent(emitter, true, false)
		emitter.flow_level++
	}

	if event.typ == yaml_MAPPING_END_EVENT {
		if (emitter.canonical || len(emitter.head_comment)+len(emitter.foot_comment)+len(emitter.tail_comment) > 0) && !first && !trail {
			yaml_emitter_write_indicator(emitter, []byte{','}, false, false, false)
		}
		yaml_emitter_process_head_comment(emitter)
		emitter.flow_level--
		emitter.indent = emitter.indents[len(emitter.indents)-1]
		emitter.indents = emitter.indents[:len(emitter.indents)-1]
		if emitter.canonical && !first {
			yaml_emitter_write_indent(emitter)
		}
		yaml_emitter_write_indicator(emitter, []byte{'}'}, false, false, false)
		yaml_emitter_process_line_comment(emitter)
		yaml_emitter_process_foot_comment(emitter)
		emitter.state = emitter.states[len(emitter.states)-1]
		emitter.states = emitter.states[:len(emitter.states)-1]
		return true
	}

	if !first && !trail {
		yaml_emitter_write_indicator(emitter, []byte{','}, false, false, false)
	}

	yaml_emitter_process_head_comment(emitter)

	if emitter.writer.column == 0 {
		yaml_emitter_write_indent(emitter)
	}

	if emitter.canonical || emitter.writer.column > emitter.best_width {
		yaml_emitter_write_indent(emitter)
	}

	if !emitter.canonical && yaml_emitter_check_simple_key(emitter) {
		emitter.states = append(emitter.states, yaml_EMIT_FLOW_MAPPING_SIMPLE_VALUE_STATE)
		return yaml_emitter_emit_node(emitter, event, false, false, true, true)
	}
	yaml_emitter_write_indicator(emitter, []byte{'?'}, true, false, false)
	emitter.states = append(emitter.states, yaml_EMIT_FLOW_MAPPING_VALUE_STATE)
	return yaml_emitter_emit_node(emitter, event, false, false, true, false)
}

// Expect a flow value node.
func yaml_emitter_emit_flow_mapping_value(emitter *yaml_emitter_t, event *yaml_event_t, simple bool) bool {
	if simple {
		yaml_emitter_write_indicator(emitter, []byte{':'}, false, false, false)
	} else {
		if emitter.canonical || emitter.writer.column > emitter.best_width {
			yaml_emitter_write_indent(emitter)
		}
		yaml_emitter_write_indicator(emitter, []byte{':'}, true, false, false)
	}
	if len(emitter.line_comment)+len(emitter.foot_comment)+len(emitter.tail_comment) > 0 {
		emitter.states = append(emitter.states, yaml_EMIT_FLOW_MAPPING_TRAIL_KEY_STATE)
	} else {
		emitter.states = append(emitter.states, yaml_EMIT_FLOW_MAPPING_KEY_STATE)
	}
	if !yaml_emitter_emit_node(emitter, event, false, false, true, false) {
		return false
	}
	if len(emitter.line_comment)+len(emitter.foot_comment)+len(emitter.tail_comment) > 0 {
		yaml_emitter_write_indicator(emitter, []byte{','}, false, false, false)
	}
	yaml_emitter_process_line_comment(emitter)
	yaml_emitter_process_foot_comment(emitter)
	return true
}

// Expect a block item node.
func yaml_emitter_emit_block_sequence_item(emitter *yaml_emitter_t, event *yaml_event_t, first bool) bool {
	if first {
		yaml_emitter_increase_indent(emitter, false, false)
	}
	if event.typ == yaml_SEQUENCE_END_EVENT {
		emitter.indent = emitter.indents[len(emitter.indents)-1]
		emitter.indents = emitter.indents[:len(emitter.indents)-1]
		emitter.state = emitter.states[len(emitter.states)-1]
		emitter.states = emitter.states[:len(emitter.states)-1]
		return true
	}
	yaml_emitter_process_head_comment(emitter)
	yaml_emitter_write_indent(emitter)
	yaml_emitter_write_indicator(emitter, []byte{'-'}, true, false, true)
	emitter.states = append(emitter.states, yaml_EMIT_BLOCK_SEQUENCE_ITEM_STATE)
	if !yaml_emitter_emit_node(emitter, event, false, true, false, false) {
		return false
	}
	yaml_emitter_process_line_comment(emitter)
	yaml_emitter_process_foot_comment(emitter)
	return true
}

// Expect a block key node.
func yaml_emitter_emit_block_mapping_key(emitter *yaml_emitter_t, event *yaml_event_t, first bool) bool {
	if first {
		yaml_emitter_increase_indent(emitter, false, false)
	}
	yaml_emitter_process_head_comment(emitter)
	if event.typ == yaml_MAPPING_END_EVENT {
		emitter.indent = emitter.indents[len(emitter.indents)-1]
		emitter.indents = emitter.indents[:len(emitter.indents)-1]
		emitter.state = emitter.states[len(emitter.states)-1]
		emitter.states = emitter.states[:len(emitter.states)-1]
		return true
	}
	yaml_emitter_write_indent(emitter)
	if len(emitter.line_comment) > 0 {
		// [Go] A line comment was provided for the key. That's unusual as the
		//      scanner associates line comments with the value. Either way,
		//      save the line comment and render it appropriately later.
		emitter.key_line_comment = emitter.line_comment
		emitter.line_comment = nil
	}
	if yaml_emitter_check_simple_key(emitter) {
		emitter.states = append(emitter.states, yaml_EMIT_BLOCK_MAPPING_SIMPLE_VALUE_STATE)
		return yaml_emitter_emit_node(emitter, event, false, false, true, true)
	}
	yaml_emitter_write_indicator(emitter, []byte{'?'}, true, false, true)
	emitter.states = append(emitter.states, yaml_EMIT_BLOCK_MAPPING_VALUE_STATE)
	return yaml_emitter_emit_node(emitter, event, false, false, true, false)
}

// Expect a block value node.
func yaml_emitter_emit_block_mapping_value(emitter *yaml_emitter_t, event *yaml_event_t, simple bool) bool {
	if simple {
		yaml_emitter_write_indicator(emitter, []byte{':'}, false, false, false)
	} else {
		yaml_emitter_write_indent(emitter)
		yaml_emitter_write_indicator(emitter, []byte{':'}, true, false, true)
	}
	if len(emitter.key_line_comment) > 0 {
		// [Go] Line comments are generally associated with the value, but when there's
		//      no value on the same line as a mapping key they end up attached to the
		//      key itself.
		if event.typ == yaml_SCALAR_EVENT {
			if len(emitter.line_comment) == 0 {
				// A scalar is coming and it has no line comments by itself yet,
				// so just let it handle the line comment as usual. If it has a
				// line comment, we can't have both so the one from the key is lost.
				emitter.line_comment = emitter.key_line_comment
				emitter.key_line_comment = nil
			}
		} else if event.sequence_style() != yaml_FLOW_SEQUENCE_STYLE && (event.typ == yaml_MAPPING_START_EVENT || event.typ == yaml_SEQUENCE_START_EVENT) {
			// An indented block follows, so write the comment right now.
			emitter.line_comment, emitter.key_line_comment = emitter.key_line_comment, emitter.line_comment
			yaml_emitter_process_line_comment(emitter)
			emitter.line_comment, emitter.key_line_comment = emitter.key_line_comment, emitter.line_comment
		}
	}
	emitter.states = append(emitter.states, yaml_EMIT_BLOCK_MAPPING_KEY_STATE)
	if !yaml_emitter_emit_node(emitter, event, false, false, true, false) {
		return false
	}
	yaml_emitter_process_line_comment(emitter)
	yaml_emitter_process_foot_comment(emitter)
	return true
}

// Expect a node.
func yaml_emitter_emit_node(emitter *yaml_emitter_t, event *yaml_event_t,
	root bool, sequence bool, mapping bool, simple_key bool) bool {

	emitter.root_context = root
	emitter.sequence_context = sequence
	emitter.mapping_context = mapping
	emitter.simple_key_context = simple_key

	switch event.typ {
	case yaml_ALIAS_EVENT:
		yaml_emitter_emit_alias(emitter, event)
		return true
	case yaml_SCALAR_EVENT:
		return yaml_emitter_emit_scalar(emitter, event)
	case yaml_SEQUENCE_START_EVENT:
		yaml_emitter_emit_sequence_start(emitter, event)
		return true
	case yaml_MAPPING_START_EVENT:
		yaml_emitter_emit_mapping_start(emitter, event)
		return true
	default:
		return yaml_emitter_set_emitter_error(emitter,
			fmt.Sprintf("expected SCALAR, SEQUENCE-START, MAPPING-START, or ALIAS, but got %v", event.typ))
	}
}

// Expect ALIAS.
func yaml_emitter_emit_alias(emitter *yaml_emitter_t, event *yaml_event_t) {
	yaml_emitter_process_anchor(emitter)
	emitter.state = emitter.states[len(emitter.states)-1]
	emitter.states = emitter.states[:len(emitter.states)-1]
}

// Expect SCALAR.
func yaml_emitter_emit_scalar(emitter *yaml_emitter_t, event *yaml_event_t) bool {
	if !yaml_emitter_select_scalar_style(emitter, event) {
		return false
	}
	yaml_emitter_process_anchor(emitter)
	yaml_emitter_process_tag(emitter)
	yaml_emitter_increase_indent(emitter, true, false)
	if !yaml_emitter_process_scalar(emitter) {
		return false
	}
	emitter.indent = emitter.indents[len(emitter.indents)-1]
	emitter.indents = emitter.indents[:len(emitter.indents)-1]
	emitter.state = emitter.states[len(emitter.states)-1]
	emitter.states = emitter.states[:len(emitter.states)-1]
	return true
}

// Expect SEQUENCE-START.
func yaml_emitter_emit_sequence_start(emitter *yaml_emitter_t, event *yaml_event_t) {
	yaml_emitter_process_anchor(emitter)
	yaml_emitter_process_tag(emitter)
	if emitter.flow_level > 0 || emitter.canonical || event.sequence_style() == yaml_FLOW_SEQUENCE_STYLE ||
		yaml_emitter_check_empty_sequence(emitter) {
		emitter.state = yaml_EMIT_FLOW_SEQUENCE_FIRST_ITEM_STATE
	} else {
		emitter.state = yaml_EMIT_BLOCK_SEQUENCE_FIRST_ITEM_STATE
	}
}

// Expect MAPPING-START.
func yaml_emitter_emit_mapping_start(emitter *yaml_emitter_t, event *yaml_event_t) {
	yaml_emitter_process_anchor(emitter)
	yaml_emitter_process_tag(emitter)
	if emitter.flow_level > 0 || emitter.canonical || event.mapping_style() == yaml_FLOW_MAPPING_STYLE ||
		yaml_emitter_check_empty_mapping(emitter) {
		emitter.state = yaml_EMIT_FLOW_MAPPING_FIRST_KEY_STATE
	} else {
		emitter.state = yaml_EMIT_BLOCK_MAPPING_FIRST_KEY_STATE
	}
}

// Check if the document content is an empty scalar.
func yaml_emitter_check_empty_document(emitter *yaml_emitter_t) bool {
	return false // [Go] Huh?
}

// Check if the next events represent an empty sequence.
func yaml_emitter_check_empty_sequence(emitter *yaml_emitter_t) bool {
	if len(emitter.events)-emitter.events_head < 2 {
		return false
	}
	return emitter.events[emitter.events_head].typ == yaml_SEQUENCE_START_EVENT &&
		emitter.events[emitter.events_head+1].typ == yaml_SEQUENCE_END_EVENT
}

// Check if the next events represent an empty mapping.
func yaml_emitter_check_empty_mapping(emitter *yaml_emitter_t) bool {
	if len(emitter.events)-emitter.events_head < 2 {
		return false
	}
	return emitter.events[emitter.events_head].typ == yaml_MAPPING_START_EVENT &&
		emitter.events[emitter.events_head+1].typ == yaml_MAPPING_END_EVENT
}

// Check if the next node can be expressed as a simple key.
func yaml_emitter_check_simple_key(emitter *yaml_emitter_t) bool {
	length := 0
	switch emitter.events[emitter.events_head].typ {
	case yaml_ALIAS_EVENT:
		length += len(emitter.anchor_data.anchor)
	case yaml_SCALAR_EVENT:
		if emitter.scalar_data.multiline {
			return false
		}
		length += len(emitter.anchor_data.anchor) +
			len(emitter.tag_data.handle) +
			len(emitter.tag_data.suffix) +
			len(emitter.scalar_data.value)
	case yaml_SEQUENCE_START_EVENT:
		if !yaml_emitter_check_empty_sequence(emitter) {
			return false
		}
		length += len(emitter.anchor_data.anchor) +
			len(emitter.tag_data.handle) +
			len(emitter.tag_data.suffix)
	case yaml_MAPPING_START_EVENT:
		if !yaml_emitter_check_empty_mapping(emitter) {
			return false
		}
		length += len(emitter.anchor_data.anchor) +
			len(emitter.tag_data.handle) +
			len(emitter.tag_data.suffix)
	default:
		return false
	}
	return length <= 128
}

// Determine an acceptable scalar style.
func yaml_emitter_select_scalar_style(emitter *yaml_emitter_t, event *yaml_event_t) bool {

	no_tag := len(emitter.tag_data.handle) == 0 && len(emitter.tag_data.suffix) == 0
	if no_tag && !event.implicit && !event.quoted_implicit {
		return yaml_emitter_set_emitter_error(emitter, "neither tag nor implicit flags are specified")
	}

	style := event.scalar_style()
	if style == yaml_ANY_SCALAR_STYLE {
		style = yaml_PLAIN_SCALAR_STYLE
	}
	if emitter.canonical {
		style = yaml_DOUBLE_QUOTED_SCALAR_STYLE
	}
	if emitter.simple_key_context && emitter.scalar_data.multiline {
		style = yaml_DOUBLE_QUOTED_SCALAR_STYLE
	}

	if style == yaml_PLAIN_SCALAR_STYLE {
		if emitter.flow_level > 0 && !emitter.scalar_data.flow_plain_allowed ||
			emitter.flow_level == 0 && !emitter.scalar_data.block_plain_allowed {
			style = yaml_SINGLE_QUOTED_SCALAR_STYLE
		}
		if len(emitter.scalar_data.value) == 0 && (emitter.flow_level > 0 || emitter.simple_key_context) {
			style = yaml_SINGLE_QUOTED_SCALAR_STYLE
		}
		if no_tag && !event.implicit {
			style = yaml_SINGLE_QUOTED_SCALAR_STYLE
		}
	}
	if style == yaml_SINGLE_QUOTED_SCALAR_STYLE {
		if !emitter.scalar_data.single_quoted_allowed {
			style = yaml_DOUBLE_QUOTED_SCALAR_STYLE
		}
	}
	if style == yaml_LITERAL_SCALAR_STYLE || style == yaml_FOLDED_SCALAR_STYLE {
		if !emitter.scalar_data.block_allowed || emitter.flow_level > 0 || emitter.simple_key_context {
			style = yaml_DOUBLE_QUOTED_SCALAR_STYLE
		}
	}

	if no_tag && !event.quoted_implicit && style != yaml_PLAIN_SCALAR_STYLE {
		emitter.tag_data.handle = []byte{'!'}
	}
	emitter.scalar_data.style = style
	return true
}

// Write an anchor.
func yaml_emitter_process_anchor(emitter *yaml_emitter_t) {
	if emitter.anchor_data.anchor == nil {
		return
	}
	c := []byte{'&'}
	if emitter.anchor_data.alias {
		c[0] = '*'
	}
	yaml_emitter_write_indicator(emitter, c, true, false, false)
	yaml_emitter_write_anchor(emitter, emitter.anchor_data.anchor)
}

// Write a tag.
func yaml_emitter_process_tag(emitter *yaml_emitter_t) {
	if len(emitter.tag_data.handle) == 0 && len(emitter.tag_data.suffix) == 0 {
		return
	}
	if len(emitter.tag_data.handle) > 0 {
		yaml_emitter_write_tag_handle(emitter, emitter.tag_data.handle)
		if len(emitter.tag_data.suffix) > 0 {
			yaml_emitter_write_tag_content(emitter, emitter.tag_data.suffix, false)
		}
	} else {
		// [Go] Allocate these slices elsewhere.
		yaml_emitter_write_indicator(emitter, []byte("!<"), true, false, false)
		yaml_emitter_write_tag_content(emitter, emitter.tag_data.suffix, false)
		yaml_emitter_write_indicator(emitter, []byte{'>'}, false, false, false)
	}
}

// Write a scalar.
func yaml_emitter_process_scalar(emitter *yaml_emitter_t) bool {
	switch emitter.scalar_data.style {
	case yaml_PLAIN_SCALAR_STYLE:
		yaml_emitter_write_plain_scalar(emitter, emitter.scalar_data.value, !emitter.simple_key_context)
		return true

	case yaml_SINGLE_QUOTED_SCALAR_STYLE:
		yaml_emitter_write_single_quoted_scalar(emitter, emitter.scalar_data.value, !emitter.simple_key_context)
		return true

	case yaml_DOUBLE_QUOTED_SCALAR_STYLE:
		yaml_emitter_write_double_quoted_scalar(emitter, emitter.scalar_data.value, !emitter.simple_key_context)
		return true

	case yaml_LITERAL_SCALAR_STYLE:
		return yaml_emitter_write_literal_scalar(emitter, emitter.scalar_data.value)

	case yaml_FOLDED_SCALAR_STYLE:
		return yaml_emitter_write_folded_scalar(emitter, emitter.scalar_data.value)
	}
	panic("unknown scalar style")
}

// Write a head comment.
func yaml_emitter_process_head_comment(emitter *yaml_emitter_t) {
	if len(emitter.tail_comment) > 0 {
		yaml_emitter_write_indent(emitter)
		yaml_emitter_write_comment(emitter, emitter.tail_comment)
		emitter.tail_comment = emitter.tail_comment[:0]
		emitter.foot_indent = emitter.indent
		if emitter.foot_indent < 0 {
			emitter.foot_indent = 0
		}
	}

	if len(emitter.head_comment) == 0 {
		return
	}
	yaml_emitter_write_indent(emitter)
	yaml_emitter_write_comment(emitter, emitter.head_comment)
	emitter.head_comment = emitter.head_comment[:0]
}

// Write an line comment.
func yaml_emitter_process_line_comment(emitter *yaml_emitter_t) {
	if len(emitter.line_comment) == 0 {
		return
	}
	if !emitter.writer.whitespace {
		put(&emitter.writer, ' ')
	}
	yaml_emitter_write_comment(emitter, emitter.line_comment)
	emitter.line_comment = emitter.line_comment[:0]
}

// Write a foot comment.
func yaml_emitter_process_foot_comment(emitter *yaml_emitter_t) {
	if len(emitter.foot_comment) == 0 {
		return
	}
	yaml_emitter_write_indent(emitter)
	yaml_emitter_write_comment(emitter, emitter.foot_comment)
	emitter.foot_comment = emitter.foot_comment[:0]
	emitter.foot_indent = emitter.indent
	if emitter.foot_indent < 0 {
		emitter.foot_indent = 0
	}
}

// Check if a %YAML directive is valid.
func yaml_emitter_analyze_version_directive(emitter *yaml_emitter_t, version_directive *yaml_version_directive_t) bool {
	if version_directive.major != 1 || version_directive.minor != 1 {
		return yaml_emitter_set_emitter_error(emitter, "incompatible %YAML directive")
	}
	return true
}

// Check if a %TAG directive is valid.
func yaml_emitter_analyze_tag_directive(emitter *yaml_emitter_t, tag_directive *yaml_tag_directive_t) bool {
	handle := tag_directive.handle
	prefix := tag_directive.prefix
	if len(handle) == 0 {
		return yaml_emitter_set_emitter_error(emitter, "tag handle must not be empty")
	}
	if handle[0] != '!' {
		return yaml_emitter_set_emitter_error(emitter, "tag handle must start with '!'")
	}
	if handle[len(handle)-1] != '!' {
		return yaml_emitter_set_emitter_error(emitter, "tag handle must end with '!'")
	}
	for i := 1; i < len(handle)-1; i += width(handle[i]) {
		if !is_alpha(handle, i) {
			return yaml_emitter_set_emitter_error(emitter, "tag handle must contain alphanumerical characters only")
		}
	}
	if len(prefix) == 0 {
		return yaml_emitter_set_emitter_error(emitter, "tag prefix must not be empty")
	}
	return true
}

// Check if an anchor is valid.
func yaml_emitter_analyze_anchor(emitter *yaml_emitter_t, anchor []byte, alias bool) bool {
	if len(anchor) == 0 {
		problem := "anchor value must not be empty"
		if alias {
			problem = "alias value must not be empty"
		}
		return yaml_emitter_set_emitter_error(emitter, problem)
	}
	for i := 0; i < len(anchor); i += width(anchor[i]) {
		if !is_alpha(anchor, i) {
			problem := "anchor value must contain alphanumerical characters only"
			if alias {
				problem = "alias value must contain alphanumerical characters only"
			}
			return yaml_emitter_set_emitter_error(emitter, problem)
		}
	}
	emitter.anchor_data.anchor = anchor
	emitter.anchor_data.alias = alias
	return true
}

// Check if a tag is valid.
func yaml_emitter_analyze_tag(emitter *yaml_emitter_t, tag []byte) bool {
	if len(tag) == 0 {
		return yaml_emitter_set_emitter_error(emitter, "tag value must not be empty")
	}
	for i := 0; i < len(emitter.tag_directives); i++ {
		tag_directive := &emitter.tag_directives[i]
		if bytes.HasPrefix(tag, tag_directive.prefix) {
			emitter.tag_data.handle = tag_directive.handle
			emitter.tag_data.suffix = tag[len(tag_directive.prefix):]
			return true
		}
	}
	emitter.tag_data.suffix = tag
	return true
}

// Check if a scalar is valid.
func yaml_emitter_analyze_scalar(emitter *yaml_emitter_t, value []byte) {
	var (
		block_indicators   = false
		flow_indicators    = false
		line_breaks        = false
		special_characters = false
		tab_characters     = false

		leading_space  = false
		leading_break  = false
		trailing_space = false
		trailing_break = false
		break_space    = false
		space_break    = false

		preceded_by_whitespace bool
		followed_by_whitespace bool
		previous_space         = false
		previous_break         = false
	)

	emitter.scalar_data.value = value

	if len(value) == 0 {
		emitter.scalar_data.multiline = false
		emitter.scalar_data.flow_plain_allowed = false
		emitter.scalar_data.block_plain_allowed = true
		emitter.scalar_data.single_quoted_allowed = true
		emitter.scalar_data.block_allowed = false
		return
	}

	if len(value) >= 3 && ((value[0] == '-' && value[1] == '-' && value[2] == '-') || (value[0] == '.' && value[1] == '.' && value[2] == '.')) {
		block_indicators = true
		flow_indicators = true
	}

	preceded_by_whitespace = true
	for i, w := 0, 0; i < len(value); i += w {
		w = width(value[i])
		followed_by_whitespace = i+w >= len(value) || is_blank(value, i+w)

		if i == 0 {
			switch value[i] {
			case '#', ',', '[', ']', '{', '}', '&', '*', '!', '|', '>', '\'', '"', '%', '@', '`':
				flow_indicators = true
				block_indicators = true
			case '?', ':':
				flow_indicators = true
				if followed_by_whitespace {
					block_indicators = true
				}
			case '-':
				if followed_by_whitespace {
					flow_indicators = true
					block_indicators = true
				}
			}
		} else {
			switch value[i] {
			case ',', '?', '[', ']', '{', '}':
				flow_indicators = true
			case ':':
				flow_indicators = true
				if followed_by_whitespace {
					block_indicators = true
				}
			case '#':
				if preceded_by_whitespace {
					flow_indicators = true
					block_indicators = true
				}
			}
		}

		if value[i] == '\t' {
			tab_characters = true
		} else if !is_printable(value, i) || !is_ascii(value, i) && !emitter.unicode {
			special_characters = true
		}
		if is_space(value, i) {
			if i == 0 {
				leading_space = true
			}
			if i+width(value[i]) == len(value) {
				trailing_space = true
			}
			if previous_break {
				break_space = true
			}
			previous_space = true
			previous_break = false
		} else if is_break(value, i) {
			line_breaks = true
			if i == 0 {
				leading_break = true
			}
			if i+width(value[i]) == len(value) {
				trailing_break = true
			}
			if previous_space {
				space_break = true
			}
			previous_space = false
			previous_break = true
		} else {
			previous_space = false
			previous_break = false
		}

		// [Go]: Why 'z'? Couldn't be the end of the string as that's the loop condition.
		preceded_by_whitespace = is_blankz(value, i)
	}

	emitter.scalar_data.multiline = line_breaks
	emitter.scalar_data.flow_plain_allowed = true
	emitter.scalar_data.block_plain_allowed = true
	emitter.scalar_data.single_quoted_allowed = true
	emitter.scalar_data.block_allowed = true

	if leading_space || leading_break || trailing_space || trailing_break {
		emitter.scalar_data.flow_plain_allowed = false
		emitter.scalar_data.block_plain_allowed = false
	}
	if trailing_space {
		emitter.scalar_data.block_allowed = false
	}
	if break_space {
		emitter.scalar_data.flow_plain_allowed = false
		emitter.scalar_data.block_plain_allowed = false
		emitter.scalar_data.single_quoted_allowed = false
	}
	if space_break || tab_characters || special_characters {
		emitter.scalar_data.flow_plain_allowed = false
		emitter.scalar_data.block_plain_allowed = false
		emitter.scalar_data.single_quoted_allowed = false
	}
	if space_break || special_characters {
		emitter.scalar_data.block_allowed = false
	}
	if line_breaks {
		emitter.scalar_data.flow_plain_allowed = false
		emitter.scalar_data.block_plain_allowed = false
	}
	if flow_indicators {
		emitter.scalar_data.flow_plain_allowed = false
	}
	if block_indicators {
		emitter.scalar_data.block_plain_allowed = false
	}
}

// Check if the event data is valid.
func yaml_emitter_analyze_event(emitter *yaml_emitter_t, event *yaml_event_t) bool {

	emitter.anchor_data.anchor = nil
	emitter.tag_data.handle = nil
	emitter.tag_data.suffix = nil
	emitter.scalar_data.value = nil

	if len(event.head_comment) > 0 {
		emitter.head_comment = event.head_comment
	}
	if len(event.line_comment) > 0 {
		emitter.line_comment = event.line_comment
	}
	if len(event.foot_comment) > 0 {
		emitter.foot_comment = event.foot_comment
	}
	if len(event.tail_comment) > 0 {
		emitter.tail_comment = event.tail_comment
	}

	switch event.typ {
	case yaml_ALIAS_EVENT:
		if !yaml_emitter_analyze_anchor(emitter, event.anchor, true) {
			return false
		}

	case yaml_SCALAR_EVENT:
		if len(event.anchor) > 0 {
			if !yaml_emitter_analyze_anchor(emitter, event.anchor, false) {
				return false
			}
		}
		if len(event.tag) > 0 && (emitter.canonical || (!event.implicit && !event.quoted_implicit)) {
			if !yaml_emitter_analyze_tag(emitter, event.tag) {
				return false
			}
		}
		yaml_emitter_analyze_scalar(emitter, event.value)

	case yaml_SEQUENCE_START_EVENT:
		if len(event.anchor) > 0 {
			if !yaml_emitter_analyze_anchor(emitter, event.anchor, false) {
				return false
			}
		}
		if len(event.tag) > 0 && (emitter.canonical || !event.implicit) {
			if !yaml_emitter_analyze_tag(emitter, event.tag) {
				return false
			}
		}

	case yaml_MAPPING_START_EVENT:
		if len(event.anchor) > 0 {
			if !yaml_emitter_analyze_anchor(emitter, event.anchor, false) {
				return false
			}
		}
		if len(event.tag) > 0 && (emitter.canonical || !event.implicit) {
			if !yaml_emitter_analyze_tag(emitter, event.tag) {
				return false
			}
		}
	}
	return true
}

func yaml_emitter_write_indent(emitter *yaml_emitter_t) {
	indent := emitter.indent
	if indent < 0 {
		indent = 0
	}
	if !emitter.writer.indention || emitter.writer.column > indent || (emitter.writer.column == indent && !emitter.writer.whitespace) {
		put_break(&emitter.writer)
	}
	if emitter.foot_indent == indent {
		put_break(&emitter.writer)
	}
	for emitter.writer.column < indent {
		put(&emitter.writer, ' ')
	}
	emitter.writer.whitespace = true
	//emitter.writer.indention = true
	emitter.writer.space_above = false
	emitter.foot_indent = -1
}

func yaml_emitter_write_indicator(emitter *yaml_emitter_t, indicator []byte, need_whitespace, is_whitespace, is_indention bool) {
	if need_whitespace && !emitter.writer.whitespace {
		put(&emitter.writer, ' ')
	}
	write_all(&emitter.writer, indicator)
	emitter.writer.whitespace = is_whitespace
	emitter.writer.indention = (emitter.writer.indention && is_indention)
	emitter.open_ended = false
}

func yaml_emitter_write_anchor(emitter *yaml_emitter_t, value []byte) {
	write_all(&emitter.writer, value)
	emitter.writer.whitespace = false
	emitter.writer.indention = false
}

func yaml_emitter_write_tag_handle(emitter *yaml_emitter_t, value []byte) {
	if !emitter.writer.whitespace {
		put(&emitter.writer, ' ')
	}
	write_all(&emitter.writer, value)
	emitter.writer.whitespace = false
	emitter.writer.indention = false
}

func yaml_emitter_write_tag_content(emitter *yaml_emitter_t, value []byte, need_whitespace bool) {
	if need_whitespace && !emitter.writer.whitespace {
		put(&emitter.writer, ' ')
	}
	for i := 0; i < len(value); {
		var must_write bool
		switch value[i] {
		case ';', '/', '?', ':', '@', '&', '=', '+', '$', ',', '_', '.', '~', '*', '\'', '(', ')', '[', ']':
			must_write = true
		default:
			must_write = is_alpha(value, i)
		}
		if must_write {
			write(&emitter.writer, value, &i)
		} else {
			w := width(value[i])
			for k := 0; k < w; k++ {
				octet := value[i]
				i++
				put(&emitter.writer, '%')

				c := octet >> 4
				if c < 10 {
					c += '0'
				} else {
					c += 'A' - 10
				}
				put(&emitter.writer, c)

				c = octet & 0x0f
				if c < 10 {
					c += '0'
				} else {
					c += 'A' - 10
				}
				put(&emitter.writer, c)
			}
		}
	}
	emitter.writer.whitespace = false
	emitter.writer.indention = false
}

func yaml_emitter_write_plain_scalar(emitter *yaml_emitter_t, value []byte, allow_breaks bool) {
	if len(value) > 0 && !emitter.writer.whitespace {
		put(&emitter.writer, ' ')
	}

	spaces := false
	breaks := false
	for i := 0; i < len(value); {
		if is_space(value, i) {
			if allow_breaks && !spaces && emitter.writer.column > emitter.best_width && !is_space(value, i+1) {
				yaml_emitter_write_indent(emitter)
				i += width(value[i])
			} else {
				write(&emitter.writer, value, &i)
			}
			spaces = true
		} else if is_break(value, i) {
			if !breaks && value[i] == '\n' {
				put_break(&emitter.writer)
			}
			write_break(&emitter.writer, value, &i)
			//emitter.writer.indention = true
			breaks = true
		} else {
			if breaks {
				yaml_emitter_write_indent(emitter)
			}
			write(&emitter.writer, value, &i)
			emitter.writer.indention = false
			spaces = false
			breaks = false
		}
	}

	if len(value) > 0 {
		emitter.writer.whitespace = false
	}
	emitter.writer.indention = false
	if emitter.root_context {
		emitter.open_ended = true
	}
}

func yaml_emitter_write_single_quoted_scalar(emitter *yaml_emitter_t, value []byte, allow_breaks bool) {

	yaml_emitter_write_indicator(emitter, []byte{'\''}, true, false, false)

	spaces := false
	breaks := false
	for i := 0; i < len(value); {
		if is_space(value, i) {
			if allow_breaks && !spaces && emitter.writer.column > emitter.best_width && i > 0 && i < len(value)-1 && !is_space(value, i+1) {
				yaml_emitter_write_indent(emitter)
				i += width(value[i])
			} else {
				write(&emitter.writer, value, &i)
			}
			spaces = true
		} else if is_break(value, i) {
			if !breaks && value[i] == '\n' {
				put_break(&emitter.writer)
			}
			write_break(&emitter.writer, value, &i)
			//emitter.writer.indention = true
			breaks = true
		} else {
			if breaks {
				yaml_emitter_write_indent(emitter)
			}
			if value[i] == '\'' {
				put(&emitter.writer, '\'')
			}
			write(&emitter.writer, value, &i)
			emitter.writer.indention = false
			spaces = false
			breaks = false
		}
	}
	yaml_emitter_write_indicator(emitter, []byte{'\''}, false, false, false)
	emitter.writer.whitespace = false
	emitter.writer.indention = false
}

func yaml_emitter_write_double_quoted_scalar(emitter *yaml_emitter_t, value []byte, allow_breaks bool) {
	spaces := false
	yaml_emitter_write_indicator(emitter, []byte{'"'}, true, false, false)

	for i := 0; i < len(value); {
		if !is_printable(value, i) || (!emitter.unicode && !is_ascii(value, i)) ||
			is_bom(value, i) || is_break(value, i) ||
			value[i] == '"' || value[i] == '\\' {

			octet := value[i]

			var w int
			var v rune
			switch {
			case octet&0x80 == 0x00:
				w, v = 1, rune(octet&0x7F)
			case octet&0xE0 == 0xC0:
				w, v = 2, rune(octet&0x1F)
			case octet&0xF0 == 0xE0:
				w, v = 3, rune(octet&0x0F)
			case octet&0xF8 == 0xF0:
				w, v = 4, rune(octet&0x07)
			}
			for k := 1; k < w; k++ {
				octet = value[i+k]
				v = (v << 6) + (rune(octet) & 0x3F)
			}
			i += w

			put(&emitter.writer, '\\')

			switch v {
			case 0x00:
				put(&emitter.writer, '0')
			case 0x07:
				put(&emitter.writer, 'a')
			case 0x08:
				put(&emitter.writer, 'b')
			case 0x09:
				put(&emitter.writer, 't')
			case 0x0A:
				put(&emitter.writer, 'n')
			case 0x0b:
				put(&emitter.writer, 'v')
			case 0x0c:
				put(&emitter.writer, 'f')
			case 0x0d:
				put(&emitter.writer, 'r')
			case 0x1b:
				put(&emitter.writer, 'e')
			case 0x22:
				put(&emitter.writer, '"')
			case 0x5c:
				put(&emitter.writer, '\\')
			case 0x85:
				put(&emitter.writer, 'N')
			case 0xA0:
				put(&emitter.writer, '_')
			case 0x2028:
				put(&emitter.writer, 'L')
			case 0x2029:
				put(&emitter.writer, 'P')
			default:
				if v <= 0xFF {
					put(&emitter.writer, 'x')
					w = 2
				} else if v <= 0xFFFF {
					put(&emitter.writer, 'u')
					w = 4
				} else {
					put(&emitter.writer, 'U')
					w = 8
				}
				for k := (w - 1) * 4; k >= 0; k -= 4 {
					digit := byte((v >> uint(k)) & 0x0F)
					if digit < 10 {
						put(&emitter.writer, digit+'0')
					} else {
						put(&emitter.writer, digit+'A'-10)
					}
				}
			}
			spaces = false
		} else if is_space(value, i) {
			if allow_breaks && !spaces && emitter.writer.column > emitter.best_width && i > 0 && i < len(value)-1 {
				yaml_emitter_write_indent(emitter)
				if is_space(value, i+1) {
					put(&emitter.writer, '\\')
				}
				i += width(value[i])
			} else {
				write(&emitter.writer, value, &i)
			}
			spaces = true
		} else {
			write(&emitter.writer, value, &i)
			spaces = false
		}
	}
	yaml_emitter_write_indicator(emitter, []byte{'"'}, false, false, false)
	emitter.writer.whitespace = false
	emitter.writer.indention = false
}

func yaml_emitter_write_block_scalar_hints(emitter *yaml_emitter_t, value []byte) bool {
	if is_space(value, 0) || is_break(value, 0) {
		if emitter.best_indent < 2 || emitter.best_indent > 9 {
			return yaml_emitter_set_emitter_error(emitter, "indent must be between 2 and 9")
		}
		indent_hint := []byte{'0' + byte(emitter.best_indent)}
		yaml_emitter_write_indicator(emitter, indent_hint, false, false, false)
	}

	emitter.open_ended = false

	var chomp_hint [1]byte
	if len(value) == 0 {
		chomp_hint[0] = '-'
	} else {
		i := len(value) - 1
		for value[i]&0xC0 == 0x80 {
			i--
		}
		if !is_break(value, i) {
			chomp_hint[0] = '-'
		} else if i == 0 {
			chomp_hint[0] = '+'
			emitter.open_ended = true
		} else {
			i--
			for value[i]&0xC0 == 0x80 {
				i--
			}
			if is_break(value, i) {
				chomp_hint[0] = '+'
				emitter.open_ended = true
			}
		}
	}
	if chomp_hint[0] != 0 {
		yaml_emitter_write_indicator(emitter, chomp_hint[:], false, false, false)
	}
	return true
}

func yaml_emitter_write_literal_scalar(emitter *yaml_emitter_t, value []byte) bool {
	yaml_emitter_write_indicator(emitter, []byte{'|'}, true, false, false)
	if !yaml_emitter_write_block_scalar_hints(emitter, value) {
		return false
	}
	yaml_emitter_process_line_comment(emitter)
	//emitter.writer.indention = true
	emitter.writer.whitespace = true
	breaks := true
	for i := 0; i < len(value); {
		if is_break(value, i) {
			write_break(&emitter.writer, value, &i)
			//emitter.writer.indention = true
			breaks = true
		} else {
			if breaks {
				yaml_emitter_write_indent(emitter)
			}
			write(&emitter.writer, value, &i)
			emitter.writer.indention = false
			breaks = false
		}
	}

	return true
}

func yaml_emitter_write_folded_scalar(emitter *yaml_emitter_t, value []byte) bool {
	yaml_emitter_write_indicator(emitter, []byte{'>'}, true, false, false)
	if !yaml_emitter_write_block_scalar_hints(emitter, value) {
		return false
	}
	yaml_emitter_process_line_comment(emitter)

	//emitter.writer.indention = true
	emitter.writer.whitespace = true

	breaks := true
	leading_spaces := true
	for i := 0; i < len(value); {
		if is_break(value, i) {
			if !breaks && !leading_spaces && value[i] == '\n' {
				k := 0
				for is_break(value, k) {
					k += width(value[k])
				}
				if !is_blankz(value, k) {
					put_break(&emitter.writer)
				}
			}
			write_break(&emitter.writer, value, &i)
			//emitter.writer.indention = true
			breaks = true
		} else {
			if breaks {
				yaml_emitter_write_indent(emitter)
				leading_spaces = is_blank(value, i)
			}
			if !breaks && is_space(value, i) && !is_space(value, i+1) && emitter.writer.column > emitter.best_width {
				yaml_emitter_write_indent(emitter)
				i += width(value[i])
			} else {
				write(&emitter.writer, value, &i)
			}
			emitter.writer.indention = false
			breaks = false
		}
	}
	return true
}

func yaml_emitter_write_comment(emitter *yaml_emitter_t, comment []byte) {
	breaks := false
	pound := false
	for i := 0; i < len(comment); {
		if is_break(comment, i) {
			write_break(&emitter.writer, comment, &i)
			//emitter.writer.indention = true
			breaks = true
			pound = false
		} else {
			if breaks {
				yaml_emitter_write_indent(emitter)
			}
			if !pound {
				if comment[i] != '#' {
					put(&emitter.writer, '#')
					put(&emitter.writer, ' ')
				}
				pound = true
			}
			write(&emitter.writer, comment, &i)
			emitter.writer.indention = false
			breaks = false
		}
	}
	if !breaks {
		put_break(&emitter.writer)
	}

	emitter.writer.whitespace = true
	//emitter.writer.indention = true
}
