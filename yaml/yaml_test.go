package yaml

import (
	"encoding/json"
	"testing"
)

func TestYAMLToJSONScalars(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		json string
	}{
		{name: "null", yaml: "null\n", json: "null"},
		{name: "bool", yaml: "true\n", json: "true"},
		{name: "old bool", yaml: "ON\n", json: "true"},
		{name: "int", yaml: "42\n", json: "42"},
		{name: "float", yaml: "1.25\n", json: "1.25"},
		{name: "quoted bool", yaml: "\"true\"\n", json: `"true"`},
		{name: "quoted null", yaml: "'null'\n", json: `"null"`},
		{name: "empty", yaml: "", json: "null"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := YAMLToJSON([]byte(test.yaml))
			if err != nil {
				t.Fatal(err.Error())
			}
			assertJSONEqual(t, got, []byte(test.json))
		})
	}
}

func TestYAMLToJSONNestedMapsSequencesAndAliases(t *testing.T) {
	got, err := YAMLToJSON([]byte(`
root:
  z:
    - 3
    - "true"
    - null
  a:
    nested: false
first: &item
  name: test 123
second: *item
`))
	if err != nil {
		t.Fatal(err.Error())
	}
	assertJSONEqual(t, got, []byte(`{
		"first": {"name": "test 123"},
		"root": {"a": {"nested": false}, "z": [3, "true", null]},
		"second": {"name": "test 123"}
	}`))
}

func TestJSONToYAMLSortedKeys(t *testing.T) {
	got, err := JSONToYAML([]byte(`{"z":1,"a":2,"m":{"b":1,"a":2}}`))
	if err != nil {
		t.Fatal(err.Error())
	}
	want := "a: 2\nm:\n  a: 2\n  b: 1\nz: 1\n"
	if string(got) != want {
		t.Fatalf("unexpected YAML:\n%s", string(got))
	}
}

func TestJSONToYAMLQuotedStrings(t *testing.T) {
	got, err := JSONToYAML([]byte(`{"normal":"test 123","bool":"true","null":"null","int":"1"}`))
	if err != nil {
		t.Fatal(err.Error())
	}
	want := "bool: \"true\"\nint: \"1\"\nnormal: test 123\n\"null\": \"null\"\n"
	if string(got) != want {
		t.Fatalf("unexpected YAML:\n%s", string(got))
	}

	jdat, err := YAMLToJSON(got)
	if err != nil {
		t.Fatal(err.Error())
	}
	assertJSONEqual(t, jdat, []byte(`{"normal":"test 123","bool":"true","null":"null","int":"1"}`))
}

func TestJSONToYAMLStability(t *testing.T) {
	first, err := JSONToYAML([]byte(`{"b":[{"z":null,"a":true},1.5],"a":"test 123","c":1}`))
	if err != nil {
		t.Fatal(err.Error())
	}
	jdat, err := YAMLToJSON(first)
	if err != nil {
		t.Fatal(err.Error())
	}
	second, err := JSONToYAML(jdat)
	if err != nil {
		t.Fatal(err.Error())
	}
	if string(first) != string(second) {
		t.Fatalf("unstable YAML:\nfirst:\n%s\nsecond:\n%s", string(first), string(second))
	}
}

func assertJSONEqual(t *testing.T, got, want []byte) {
	t.Helper()

	var gotVal any
	if err := json.Unmarshal(got, &gotVal); err != nil {
		t.Fatalf("unmarshal got JSON: %s\n%s", err.Error(), string(got))
	}
	var wantVal any
	if err := json.Unmarshal(want, &wantVal); err != nil {
		t.Fatalf("unmarshal want JSON: %s\n%s", err.Error(), string(want))
	}

	gotCanon, err := json.Marshal(gotVal)
	if err != nil {
		t.Fatal(err.Error())
	}
	wantCanon, err := json.Marshal(wantVal)
	if err != nil {
		t.Fatal(err.Error())
	}
	if string(gotCanon) != string(wantCanon) {
		t.Fatalf("JSON mismatch\ngot:  %s\nwant: %s", string(gotCanon), string(wantCanon))
	}
}
