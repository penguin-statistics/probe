package model

import (
	"github.com/penguin-statistics/probe/densemver"
	"time"
)

type Bonjour struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	Version  *densemver.DenSemVer `query:"v" valid:"required" gorm:"<-:create;type:integer;index"`
	Platform Platform             `query:"p" valid:"range(0|2)" gorm:"type:smallint;index"`
	UID      string               `query:"u" valid:"required,stringlength(32|32)" gorm:"type:char(32);index"`
}
