package model

import (
	"github.com/penguin-statistics/probe/densemver"
)

// Bonjour is a bonjour request in which the client initiates request with basic params
type Bonjour struct {
	ID string

	Version  *densemver.DenSemVer `query:"v" valid:"required"`
	Platform *Platform            `query:"p"`
	UID      string               `query:"u" valid:"stringlength(32|32),alphanum"`
	Legacy   uint8                `query:"l"`

	Referer    string `query:"r"`
	Reconnects int    `query:"i"`
}
