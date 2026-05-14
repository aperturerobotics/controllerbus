package controller

// NewInfo constructs a new Info object.
func NewInfo(id string, version Version, descrip string) *Info {
	return &Info{
		Id:          id,
		Version:     version.String(),
		Description: descrip,
	}
}

// Clone copies the Info object.
func (i *Info) Clone() *Info {
	if i == nil {
		return nil
	}

	return &Info{
		Id:          i.Id,
		Version:     i.Version,
		Description: i.Description,
	}
}
