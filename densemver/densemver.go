package densemver

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/Masterminds/semver"
	"reflect"
)

const (
	MaxSemVerString = "65535.255.255"
)

type DenSemVer struct {
	semver *semver.Version
}

func (v *DenSemVer) Scan(src interface{}) error {
	switch src.(type) {
	case uint32:
		dsv, err := FromInt(src.(uint32))
		if err != nil {
			return err
		}
		*v = *dsv
		return nil
	default:
		return errors.New("unknown scanning type" + reflect.TypeOf(src).String())
	}
}

func (v DenSemVer) Value() (driver.Value, error) {
	if v.semver == nil {
		return nil, errors.New("invalid DenSemVer passed")
	}
	return v.Int(), nil
}

func (v *DenSemVer) UnmarshalParam(param string) error {
	dsv, err := FromString(param)
	if err != nil {
		return err
	}
	*v = *dsv
	return nil
}

// initialize a DenSemVer instance from a semver string
func FromString(v string) (dsv *DenSemVer, err error) {
	semv, err := semver.NewVersion(v)
	if err != nil {
		return nil, err
	}
	if semv.Metadata() != "" {
		return nil, errors.New("unexpected semver with metadata field")
	}
	largest := semver.MustParse(MaxSemVerString)
	if semv.Major() > largest.Major() || semv.Minor() > largest.Minor() || semv.Patch() > largest.Patch() {
		return nil, errors.New("unexpected semver segment greater than " + MaxSemVerString)
	}
	//fmt.Println("parsed string as", semv.String())
	return &DenSemVer{semver: semv}, nil
}

// initialize a DenSemVer instance from a DenSemVer integer representation
func FromInt(i uint32) (dsv *DenSemVer, err error) {
	if i > (1<<32)-1 || i < 0 {
		return nil, errors.New(fmt.Sprintf("semver version out of range: %v", i))
	}
	major := i >> 16
	minor := (i >> 8) & 0xFF
	patch := i & 0xFF

	//fmt.Println("parsed semver from int", strconv.FormatUint(uint64(i>>16), 2))
	ver := fmt.Sprintf("%d.%d.%d", major, minor, patch)
	semv, err := semver.NewVersion(ver)
	if err != nil {
		return nil, err
	}
	return &DenSemVer{semver: semv}, nil
}

// Int returns a integer representation of the DenSemVer instance
func (v *DenSemVer) Int() (r uint32) {
	major, minor, patch := uint32(v.semver.Major()), uint32(v.semver.Minor()), uint32(v.semver.Patch())
	r = major
	r = (r << 8) + minor
	r = (r << 8) + patch
	//fmt.Println("converted as", strconv.FormatUint(uint64(r), 2))
	return r
}

// String returns a string representation of the DenSemVer instance
func (v *DenSemVer) String() string {
	return v.semver.String()
}
