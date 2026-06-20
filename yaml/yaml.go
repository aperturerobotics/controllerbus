package yaml

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"sort"
	"strconv"

	"github.com/pkg/errors"
)

const (
	shortNullTag  = "!!null"
	shortBoolTag  = "!!bool"
	shortStrTag   = "!!str"
	shortIntTag   = "!!int"
	shortFloatTag = "!!float"
	longTagPrefix = "tag:yaml.org,2002:"
)

type eventParser struct {
	parser  yaml_parser_t
	anchors map[string]any
}

// YAMLToJSON converts YAML bytes to JSON bytes.
func YAMLToJSON(y []byte) ([]byte, error) {
	if len(bytes.TrimSpace(y)) == 0 {
		return []byte("null"), nil
	}

	p := newEventParser(y)
	ev, err := p.next()
	if err != nil {
		return nil, err
	}
	if ev.typ != yaml_STREAM_START_EVENT {
		return nil, errors.Errorf("yaml: expected stream start, got %s", ev.typ.String())
	}

	for {
		ev, err = p.next()
		if err != nil {
			return nil, err
		}
		switch ev.typ {
		case yaml_DOCUMENT_START_EVENT:
			ev, err = p.next()
			if err != nil {
				return nil, err
			}
			if ev.typ == yaml_DOCUMENT_END_EVENT {
				return json.Marshal(nil)
			}
			val, err := p.parseValue(ev)
			if err != nil {
				return nil, err
			}
			if err := p.expectDocumentEnd(); err != nil {
				return nil, err
			}
			return json.Marshal(val)
		case yaml_STREAM_END_EVENT:
			return json.Marshal(nil)
		default:
			return nil, errors.Errorf("yaml: expected document start, got %s", ev.typ.String())
		}
	}
}

// JSONToYAML converts JSON bytes to YAML bytes.
func JSONToYAML(j []byte) ([]byte, error) {
	dec := json.NewDecoder(bytes.NewReader(j))
	dec.UseNumber()

	var tree any
	if err := dec.Decode(&tree); err != nil {
		return nil, err
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("yaml: JSON input contains multiple values")
	}

	var out []byte
	emitter := newEmitter(&out)
	if err := emit(&emitter, yaml_event_t{
		typ:      yaml_STREAM_START_EVENT,
		encoding: yaml_UTF8_ENCODING,
	}); err != nil {
		return nil, err
	}
	if err := emit(&emitter, yaml_event_t{
		typ:      yaml_DOCUMENT_START_EVENT,
		implicit: true,
	}); err != nil {
		return nil, err
	}
	if err := emitValue(&emitter, tree); err != nil {
		return nil, err
	}
	if err := emit(&emitter, yaml_event_t{
		typ:      yaml_DOCUMENT_END_EVENT,
		implicit: true,
	}); err != nil {
		return nil, err
	}
	if err := emit(&emitter, yaml_event_t{
		typ: yaml_STREAM_END_EVENT,
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func newEventParser(dat []byte) *eventParser {
	p := &eventParser{
		anchors: make(map[string]any),
	}
	p.parser = yaml_parser_t{
		raw_buffer:   make([]byte, 0, input_raw_buffer_size),
		buffer:       make([]byte, 0, input_buffer_size),
		read_handler: yamlInputReadHandler,
		input:        dat,
	}
	return p
}

func yamlInputReadHandler(parser *yaml_parser_t, buffer []byte) (int, error) {
	if parser.input_pos == len(parser.input) {
		return 0, io.EOF
	}
	n := copy(buffer, parser.input[parser.input_pos:])
	parser.input_pos += n
	return n, nil
}

func yaml_insert_token(parser *yaml_parser_t, pos int, token *yaml_token_t) {
	if parser.tokens_head > 0 && len(parser.tokens) == cap(parser.tokens) {
		if parser.tokens_head != len(parser.tokens) {
			copy(parser.tokens, parser.tokens[parser.tokens_head:])
		}
		parser.tokens = parser.tokens[:len(parser.tokens)-parser.tokens_head]
		parser.tokens_head = 0
	}
	parser.tokens = append(parser.tokens, *token)
	if pos < 0 {
		return
	}
	copy(parser.tokens[parser.tokens_head+pos+1:], parser.tokens[parser.tokens_head+pos:])
	parser.tokens[parser.tokens_head+pos] = *token
}

func (p *eventParser) next() (yaml_event_t, error) {
	var ev yaml_event_t
	if yaml_parser_parse(&p.parser, &ev) {
		return ev, nil
	}
	return ev, parserError(&p.parser)
}

func parserError(parser *yaml_parser_t) error {
	if parser.problem == "" {
		return errors.New("yaml: parse failed")
	}
	if parser.context != "" {
		return errors.Errorf(
			"yaml: %s: %s at line %d column %d",
			parser.context,
			parser.problem,
			parser.problem_mark.line+1,
			parser.problem_mark.column+1,
		)
	}
	return errors.Errorf(
		"yaml: %s at line %d column %d",
		parser.problem,
		parser.problem_mark.line+1,
		parser.problem_mark.column+1,
	)
}

func (p *eventParser) expectDocumentEnd() error {
	ev, err := p.next()
	if err != nil {
		return err
	}
	if ev.typ == yaml_DOCUMENT_END_EVENT {
		return nil
	}
	return errors.Errorf("yaml: expected document end, got %s", ev.typ.String())
}

func (p *eventParser) parseValue(ev yaml_event_t) (any, error) {
	switch ev.typ {
	case yaml_ALIAS_EVENT:
		val, ok := p.anchors[string(ev.anchor)]
		if !ok {
			return nil, errors.Errorf("yaml: unknown alias %q", string(ev.anchor))
		}
		return val, nil
	case yaml_SCALAR_EVENT:
		val, err := scalarValue(ev)
		if err != nil {
			return nil, err
		}
		if len(ev.anchor) != 0 {
			p.anchors[string(ev.anchor)] = val
		}
		return val, nil
	case yaml_SEQUENCE_START_EVENT:
		val, err := p.parseSequence()
		if err != nil {
			return nil, err
		}
		if len(ev.anchor) != 0 {
			p.anchors[string(ev.anchor)] = val
		}
		return val, nil
	case yaml_MAPPING_START_EVENT:
		val, err := p.parseMapping()
		if err != nil {
			return nil, err
		}
		if len(ev.anchor) != 0 {
			p.anchors[string(ev.anchor)] = val
		}
		return val, nil
	default:
		return nil, errors.Errorf("yaml: unexpected %s", ev.typ.String())
	}
}

func (p *eventParser) parseSequence() ([]any, error) {
	var vals []any
	for {
		ev, err := p.next()
		if err != nil {
			return nil, err
		}
		if ev.typ == yaml_SEQUENCE_END_EVENT {
			return vals, nil
		}
		val, err := p.parseValue(ev)
		if err != nil {
			return nil, err
		}
		vals = append(vals, val)
	}
}

func (p *eventParser) parseMapping() (map[string]any, error) {
	vals := make(map[string]any)
	for {
		keyEv, err := p.next()
		if err != nil {
			return nil, err
		}
		if keyEv.typ == yaml_MAPPING_END_EVENT {
			return vals, nil
		}
		keyVal, err := p.parseValue(keyEv)
		if err != nil {
			return nil, err
		}
		key, err := keyString(keyVal)
		if err != nil {
			return nil, err
		}

		valEv, err := p.next()
		if err != nil {
			return nil, err
		}
		val, err := p.parseValue(valEv)
		if err != nil {
			return nil, err
		}
		vals[key] = val
	}
}

func scalarValue(ev yaml_event_t) (any, error) {
	tag := shortTag(ev.tag)
	if !ev.implicit && tag == "" {
		return string(ev.value), nil
	}

	switch tag {
	case shortStrTag:
		return string(ev.value), nil
	case shortNullTag:
		return nil, nil
	case shortBoolTag:
		val, ok := boolScalar(ev.value)
		if ok {
			return val, nil
		}
		return nil, errors.Errorf("yaml: cannot decode %q as bool", string(ev.value))
	case shortIntTag:
		val, ok := intScalar(ev.value)
		if ok {
			return val, nil
		}
		return nil, errors.Errorf("yaml: cannot decode %q as int", string(ev.value))
	case shortFloatTag:
		val, ok, err := floatScalar(ev.value, true)
		if err != nil {
			return nil, err
		}
		if ok {
			return val, nil
		}
		return nil, errors.Errorf("yaml: cannot decode %q as float", string(ev.value))
	case "":
		if ev.implicit {
			return resolvePlainScalar(ev.value)
		}
	}
	return string(ev.value), nil
}

func resolvePlainScalar(dat []byte) (any, error) {
	if nullScalar(dat) {
		return nil, nil
	}
	if val, ok := boolScalar(dat); ok {
		return val, nil
	}
	if val, ok := intScalar(dat); ok {
		return val, nil
	}
	if integerScalar(dat) {
		val, ok, err := floatScalar(dat, true)
		if ok || err != nil {
			return val, err
		}
	}
	if val, ok, err := floatScalar(dat, false); ok || err != nil {
		return val, err
	}
	return string(dat), nil
}

func shortTag(tag []byte) string {
	if len(tag) == 0 {
		return ""
	}
	if bytes.HasPrefix(tag, []byte(longTagPrefix)) {
		return "!!" + string(tag[len(longTagPrefix):])
	}
	return string(tag)
}

func nullScalar(dat []byte) bool {
	return len(dat) == 0 ||
		bytes.Equal(dat, []byte("~")) ||
		bytes.EqualFold(dat, []byte("null"))
}

func boolScalar(dat []byte) (bool, bool) {
	switch string(dat) {
	case "y", "Y", "yes", "Yes", "YES", "true", "True", "TRUE", "on", "On", "ON":
		return true, true
	case "n", "N", "no", "No", "NO", "false", "False", "FALSE", "off", "Off", "OFF":
		return false, true
	default:
		return false, false
	}
}

func intScalar(dat []byte) (int64, bool) {
	clean := cleanNumber(dat)
	if len(clean) == 0 {
		return 0, false
	}
	val, err := strconv.ParseInt(string(clean), 0, 64)
	if err == nil {
		return val, true
	}
	return 0, false
}

func floatScalar(dat []byte, allowInt bool) (float64, bool, error) {
	if nonFiniteFloat(dat) {
		return 0, false, errors.Errorf("yaml: non-finite float %q cannot convert to JSON", string(dat))
	}
	if !allowInt && !floatLikeScalar(dat) {
		return 0, false, nil
	}
	clean := cleanNumber(dat)
	if len(clean) == 0 {
		return 0, false, nil
	}
	val, err := strconv.ParseFloat(string(clean), 64)
	if err != nil {
		return 0, false, nil
	}
	if math.IsInf(val, 0) || math.IsNaN(val) {
		return 0, false, errors.Errorf("yaml: non-finite float %q cannot convert to JSON", string(dat))
	}
	return val, true, nil
}

func cleanNumber(dat []byte) []byte {
	if bytes.IndexByte(dat, '_') < 0 {
		return dat
	}
	return bytes.ReplaceAll(dat, []byte("_"), nil)
}

func nonFiniteFloat(dat []byte) bool {
	return bytes.EqualFold(dat, []byte(".nan")) ||
		bytes.EqualFold(dat, []byte(".inf")) ||
		bytes.EqualFold(dat, []byte("+.inf")) ||
		bytes.EqualFold(dat, []byte("-.inf"))
}

func floatLikeScalar(dat []byte) bool {
	for _, c := range dat {
		switch c {
		case '.', 'e', 'E':
			return true
		}
	}
	return false
}

func integerScalar(dat []byte) bool {
	if len(dat) == 0 {
		return false
	}
	start := 0
	if dat[0] == '+' || dat[0] == '-' {
		start = 1
	}
	if start == len(dat) {
		return false
	}
	for _, c := range dat[start:] {
		if c == '_' {
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func keyString(val any) (string, error) {
	switch typed := val.(type) {
	case nil:
		return "null", nil
	case string:
		return typed, nil
	case bool:
		if typed {
			return "true", nil
		}
		return "false", nil
	case int64:
		return strconv.FormatInt(typed, 10), nil
	case float64:
		return strconv.FormatFloat(typed, 'g', -1, 64), nil
	default:
		dat, err := json.Marshal(typed)
		if err != nil {
			return "", errors.Wrap(err, "marshal YAML map key")
		}
		return string(dat), nil
	}
}

func newEmitter(out *[]byte) yaml_emitter_t {
	return yaml_emitter_t{
		write_handler: yamlOutputWriteHandler,
		output_buffer: out,
		buffer:        make([]byte, output_buffer_size),
		raw_buffer:    make([]byte, 0, output_raw_buffer_size),
		states:        make([]yaml_emitter_state_t, 0, initial_stack_size),
		events:        make([]yaml_event_t, 0, initial_queue_size),
		best_indent:   2,
		best_width:    -1,
		unicode:       true,
	}
}

func yamlOutputWriteHandler(emitter *yaml_emitter_t, buffer []byte) error {
	*emitter.output_buffer = append(*emitter.output_buffer, buffer...)
	return nil
}

func emit(emitter *yaml_emitter_t, ev yaml_event_t) error {
	if yaml_emitter_emit(emitter, &ev) {
		return nil
	}
	if emitter.problem == "" {
		return errors.New("yaml: emit failed")
	}
	return errors.Errorf("yaml: %s", emitter.problem)
}

func emitValue(emitter *yaml_emitter_t, val any) error {
	switch typed := val.(type) {
	case nil:
		return emitScalar(emitter, "null", yaml_PLAIN_SCALAR_STYLE)
	case bool:
		if typed {
			return emitScalar(emitter, "true", yaml_PLAIN_SCALAR_STYLE)
		}
		return emitScalar(emitter, "false", yaml_PLAIN_SCALAR_STYLE)
	case json.Number:
		if err := validateJSONNumber(typed); err != nil {
			return err
		}
		return emitScalar(emitter, typed.String(), yaml_PLAIN_SCALAR_STYLE)
	case float64:
		if math.IsInf(typed, 0) || math.IsNaN(typed) {
			return errors.Errorf("yaml: non-finite float %q cannot convert to YAML", strconv.FormatFloat(typed, 'g', -1, 64))
		}
		return emitScalar(emitter, strconv.FormatFloat(typed, 'g', -1, 64), yaml_PLAIN_SCALAR_STYLE)
	case string:
		return emitScalar(emitter, typed, stringScalarStyle(typed))
	case []any:
		return emitSequence(emitter, typed)
	case map[string]any:
		return emitMapping(emitter, typed)
	default:
		return errors.Errorf("yaml: unsupported JSON value %T", val)
	}
}

func validateJSONNumber(num json.Number) error {
	dat := []byte(num.String())
	if integerScalar(dat) {
		if _, err := strconv.ParseInt(num.String(), 10, 64); err == nil {
			return nil
		}
	}
	val, err := strconv.ParseFloat(num.String(), 64)
	if err != nil {
		return errors.Wrapf(err, "parse JSON number %q", num.String())
	}
	if math.IsInf(val, 0) || math.IsNaN(val) {
		return errors.Errorf("yaml: non-finite JSON number %q cannot convert to YAML", num.String())
	}
	return nil
}

func emitSequence(emitter *yaml_emitter_t, vals []any) error {
	if err := emit(emitter, yaml_event_t{
		typ:      yaml_SEQUENCE_START_EVENT,
		implicit: true,
		style:    yaml_style_t(yaml_BLOCK_SEQUENCE_STYLE),
	}); err != nil {
		return err
	}
	for _, val := range vals {
		if err := emitValue(emitter, val); err != nil {
			return err
		}
	}
	return emit(emitter, yaml_event_t{
		typ: yaml_SEQUENCE_END_EVENT,
	})
}

func emitMapping(emitter *yaml_emitter_t, vals map[string]any) error {
	if err := emit(emitter, yaml_event_t{
		typ:      yaml_MAPPING_START_EVENT,
		implicit: true,
		style:    yaml_style_t(yaml_BLOCK_MAPPING_STYLE),
	}); err != nil {
		return err
	}

	keys := make([]string, 0, len(vals))
	for key := range vals {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if err := emitScalar(emitter, key, stringScalarStyle(key)); err != nil {
			return err
		}
		if err := emitValue(emitter, vals[key]); err != nil {
			return err
		}
	}
	return emit(emitter, yaml_event_t{
		typ: yaml_MAPPING_END_EVENT,
	})
}

func emitScalar(emitter *yaml_emitter_t, val string, style yaml_scalar_style_t) error {
	return emit(emitter, yaml_event_t{
		typ:             yaml_SCALAR_EVENT,
		value:           []byte(val),
		implicit:        true,
		quoted_implicit: true,
		style:           yaml_style_t(style),
	})
}

func stringScalarStyle(val string) yaml_scalar_style_t {
	if val == "" {
		return yaml_DOUBLE_QUOTED_SCALAR_STYLE
	}
	resolved, err := resolvePlainScalar([]byte(val))
	if err != nil {
		return yaml_DOUBLE_QUOTED_SCALAR_STYLE
	}
	if _, ok := resolved.(string); !ok {
		return yaml_DOUBLE_QUOTED_SCALAR_STYLE
	}
	return yaml_PLAIN_SCALAR_STYLE
}

func yaml_event_delete(event *yaml_event_t) {
	*event = yaml_event_t{}
}
