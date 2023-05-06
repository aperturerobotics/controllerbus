package boilerplate_v1

import (
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
)

// TestNegativeUint tests unmarshaling a negative uint.
// This should never be allowed.
func TestNegativeUint(t *testing.T) {
	jsonDat := `{"printedLen": -100}`
	msg := &BoilerplateResult{}
	if err := protojson.Unmarshal([]byte(jsonDat), msg); err != nil {
		t.Logf("protojson unmarshal returned expected result: %v", err.Error())
	} else {
		t.FailNow()
	}
}
