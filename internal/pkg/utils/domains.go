package utils

import (
	"github.com/elliotchance/pie/pie"
	"github.com/jpillora/go-tld"
	"net/url"
	"strings"
)

var (
	PenguinDomains = pie.Strings{"penguin-stats.io", "penguin-stats.cn", "penguin-stats.com", "exusi.ai"}
)

func IsValidDomain(u *url.URL) bool {
	t, err := tld.Parse(u.Hostname())
	if err != nil {
		return false
	}
	d := strings.Join([]string{t.Domain, t.TLD}, ".")

	return PenguinDomains.Contains(d)
}
