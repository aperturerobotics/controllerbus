package hot_compiler

// VirtualModuleTree builds a custom temporary Go module tree with relative
// import paths, and runs the Go compiler against the resultant Go tree.
//
// It can transform import paths for a set of packages so that the Go plugin
// loader will recognize them as unique implementations.
type VirtualModuleTree struct {
}

//
