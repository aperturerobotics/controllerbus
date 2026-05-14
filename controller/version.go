package controller

import (
	"errors"
	"strconv"
	"strings"
)

// Version identifies a controller implementation version.
type Version struct {
	value string
	major uint64
	minor uint64
	patch uint64
}

// ParseVersion parses a major.minor.patch controller version.
func ParseVersion(value string) (Version, error) {
	base := value
	if idx := strings.IndexAny(base, "-+"); idx >= 0 {
		base = base[:idx]
	}

	parts := strings.Split(base, ".")
	if len(parts) != 3 {
		return Version{}, errors.New("version must be major.minor.patch")
	}

	major, err := parseVersionPart(parts[0])
	if err != nil {
		return Version{}, err
	}
	minor, err := parseVersionPart(parts[1])
	if err != nil {
		return Version{}, err
	}
	patch, err := parseVersionPart(parts[2])
	if err != nil {
		return Version{}, err
	}

	return Version{
		value: value,
		major: major,
		minor: minor,
		patch: patch,
	}, nil
}

// MustParseVersion parses a controller version and panics on invalid input.
func MustParseVersion(value string) Version {
	version, err := ParseVersion(value)
	if err != nil {
		panic(err)
	}
	return version
}

// String returns the original version string.
func (v Version) String() string {
	return v.value
}

// GT returns true if the version is greater than the other version.
func (v Version) GT(other Version) bool {
	return compareVersion(v, other) > 0
}

// GTE returns true if the version is greater than or equal to the other version.
func (v Version) GTE(other Version) bool {
	return compareVersion(v, other) >= 0
}

func parseVersionPart(part string) (uint64, error) {
	if part == "" {
		return 0, errors.New("version part is empty")
	}
	return strconv.ParseUint(part, 10, 64)
}

func compareVersion(a, b Version) int {
	if a.major != b.major {
		if a.major > b.major {
			return 1
		}
		return -1
	}
	if a.minor != b.minor {
		if a.minor > b.minor {
			return 1
		}
		return -1
	}
	if a.patch != b.patch {
		if a.patch > b.patch {
			return 1
		}
		return -1
	}
	return strings.Compare(a.value, b.value)
}
