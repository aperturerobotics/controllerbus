//go:build !js && !tinygo

package controller_exec

import "embed"

// DistSources contains the TypeScript exec messages for the web.
//
//go:embed exec.pb.ts
var DistSources embed.FS
