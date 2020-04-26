package spec

import (
	"errors"
	"strings"

	"github.com/go-courier/semver"
)

func ParseRefOperator(s string) (*Ref, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return nil, errors.New("invalid operator ref")
	}

	v := &Ref{
		Name: parts[0],
	}

	version, err := semver.ParseVersion(parts[1])
	if err != nil {
		return nil, err
	}

	v.Version = *version

	return v, nil
}

func NewRefOperator(name string, version semver.Version) *Ref {
	return &Ref{Name: name, Version: version}
}

// openapi:strfmt ref
type Ref struct {
	Name    string
	Version semver.Version
}

func (v Ref) RefID() string {
	return v.Name + ":" + v.Version.String()
}

func (v Ref) String() string {
	return v.RefID()
}

func (v Ref) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *Ref) UnmarshalText(data []byte) error {
	refOperator, err := ParseRefOperator(string(data))
	if err != nil {
		return err
	}
	*v = *refOperator
	return nil
}
