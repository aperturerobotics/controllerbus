//go:build !js && !tinygo

package configset_proto

import "embed"

// DistSources contains the TypeScript configset messages for the web.
//
//go:embed configset.pb.ts
var DistSources embed.FS
