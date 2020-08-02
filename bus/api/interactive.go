package bus_api

import (
	"bytes"
	"strings"
)

// PrintPrettyStatus prints bus info as a pretty status output.
func (b *GetBusInfoResponse) PrintPrettyStatus() []byte {
	var dat bytes.Buffer
	if b == nil {
		_, _ = dat.WriteString("● no data\n")
		return dat.Bytes()
	}

	_, _ = dat.WriteString("✓ controller-bus running\n")
	_, _ = dat.WriteString("Controllers:")
	runningCtrl := b.GetRunningControllers()
	if len(runningCtrl) == 0 {
		_, _ = dat.WriteString(" [none]")
	} else {
		for _, ctrl := range runningCtrl {
			ctrlID := ctrl.GetId()
			ctrlVer := ctrl.GetVersion()
			_, _ = dat.WriteString("\n\t")
			_, _ = dat.WriteString(ctrlID)
			_, _ = dat.WriteString(" ")
			_, _ = dat.WriteString(ctrlVer)
		}
	}
	_, _ = dat.WriteString("\n")
	_, _ = dat.WriteString("Directives:")
	runningDirs := b.GetRunningDirectives()
	if len(runningDirs) == 0 {
		_, _ = dat.WriteString(" [none]")
	} else {
		for _, dir := range runningDirs {
			dirInfo := dir.GetInfo()
			dirName := dirInfo.GetName()
			_, _ = dat.WriteString("\n\t")
			_, _ = dat.WriteString(dirName)
			debugVals := dirInfo.GetDebugVals()
			for _, val := range debugVals {
				_, _ = dat.WriteString("\n\t\t")
				_, _ = dat.WriteString(val.GetKey())
				if nvals := val.GetValues(); len(nvals) != 0 {
					_, _ = dat.WriteString(" = ")
					_, _ = dat.WriteString(strings.Join(nvals, ", "))
				}
			}
		}
	}
	_, _ = dat.WriteString("\n")
	return dat.Bytes()
}
