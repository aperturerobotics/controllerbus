package controller

import (
	"github.com/blang/semver"
)

// NewInfo constructs a new Info object.
func NewInfo(id string, version semver.Version, descrip string) Info {
	return Info{
		Id:          id,
		Version:     version.String(),
		Description: descrip,
	}
}
