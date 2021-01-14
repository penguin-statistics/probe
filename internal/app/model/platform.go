package model

import (
	"database/sql/driver"
	"errors"
	"fmt"
)

const (
	// PlatformWeb is the web platform
	PlatformWeb Platform = iota
	// PlatformAppIOS is the iOS App platform
	PlatformAppIOS
	// PlatformAppAndroid is the Android App platform
	PlatformAppAndroid
)

// Platform describes the device user initiates request from
type Platform int

// Value implements driver.Valuer
func (p Platform) Value() (driver.Value, error) {
	return int(p), nil
}

// Scan implements sql.Scanner
func (p *Platform) Scan(src interface{}) error {
	i := src.(int64)
	switch i {
	case int64(PlatformWeb):
		*p = PlatformWeb
		return nil
	case int64(PlatformAppIOS):
		*p = PlatformAppIOS
		return nil
	case int64(PlatformAppAndroid):
		*p = PlatformAppAndroid
		return nil
	}
	return fmt.Errorf("unknown src value of %v", src)
}

func (p *Platform) Marshal() string {
	switch *p {
	case PlatformWeb:
		return "web"
	case PlatformAppIOS:
		return "app:ios"
	case PlatformAppAndroid:
		return "app:android"
	}
	return "unknown"
}

// UnmarshalParam implements echo.BindUnmarshaler
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
