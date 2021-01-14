package model

import (
	"github.com/penguin-statistics/probe/densemver"
	"time"
)

// Bonjour is a bonjour request in which the client initiates request with basic params
type Bonjour struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	Version  *densemver.DenSemVer `query:"v" valid:"required" gorm:"<-:create;type:integer;index"`
	Platform *Platform            `query:"p" gorm:"type:smallint;index"`
	UID      string               `query:"u" valid:"required,stringlength(32|32),alphanum" gorm:"type:char(32);index"`

	Reconnects int `query:"i" gorm:"-"`
}
