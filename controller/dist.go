//go:build !js && !tinygo

package controller

import "embed"

// DistSources contains the TypeScript controller messages for the web.
//
//go:embed controller.pb.ts
var DistSources embed.FS
