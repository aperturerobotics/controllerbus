package boilerplate_v1

import (
	"testing"

	"github.com/sirupsen/logrus"
)

// TestNegativeUint tests unmarshaling a negative uint.
// This should never be allowed.
func TestNegativeUint(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	jsonDat := `{"printedLen": -100}`
	msg := &BoilerplateResult{}
	err := msg.UnmarshalJSON([]byte(jsonDat))
	if err != nil {
		le.Infof("protojson unmarshal returned expected result: %v", err.Error())
	} else {
		t.FailNow()
	}
}
