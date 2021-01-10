package model

import (
	"database/sql/driver"
	"errors"
)

const (
	PlatformWeb Platform = iota
	PlatformAppIOS
	PlatformAppAndroid
)

type Platform int

func (p Platform) Value() (driver.Value, error) {
	return int(p), nil
}

func (p *Platform) Scan(src interface{}) error {
	*p = src.(Platform)
	return nil
}

func (p *Platform) UnmarshalParam(param string) error {
	switch param {
	case "web":
		*p = PlatformWeb
		return nil
	case "app:ios":
		*p = PlatformAppIOS
		return nil
	case "app:android":
		*p = PlatformAppAndroid
		return nil
	default:
		return errors.New("unknown platform")
	}
}
