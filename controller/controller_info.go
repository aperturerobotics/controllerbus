package controller

import (
	"github.com/blang/semver"
)

// NewInfo constructs a new Info object.
func NewInfo(id string, version semver.Version, descrip string) *Info {
	return &Info{
		Id:          id,
		Version:     version.String(),
		Description: descrip,
	}
}

// Clone copies the Info object.
func (i *Info) Clone() *Info {
	return &Info{
		Id:          i.Id,
		Version:     i.Version,
		Description: i.Description,
	}
}
