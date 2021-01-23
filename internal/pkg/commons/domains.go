package commons

import (
	"github.com/elliotchance/pie/pie"
	"github.com/jpillora/go-tld"
	"github.com/spf13/viper"
	"net/http"
	"net/url"
	"strings"
)

var (
	// PenguinDomains are domains which Penguin Statistics hold
	PenguinDomains = pie.Strings{"penguin-stats.io", "penguin-stats.cn", "penguin-stats.com", "exusi.ai"}
)

// IsValidDomain validates if a domain in url.URL provided is a Penguin Statistics official domain
func IsValidDomain(u *url.URL) bool {
	t, err := tld.Parse(u.Hostname())
	if err != nil {
		return false
	}
	d := strings.Join([]string{t.Domain, t.TLD}, ".")

	return PenguinDomains.Contains(d)
}

func GenOriginChecker() func(r *http.Request) bool {
	if viper.GetBool("app.debug") {
		return func(r *http.Request) bool {
			return true
		}
	} else {
		return func(r *http.Request) bool {
			return IsValidDomain(r.URL)
		}
	}
}

func PenguinDomainsOrigin() (origins []string) {
	for _, domain := range PenguinDomains {
		origins = append(origins, "https://"+domain, "https://*."+domain)
	}
	return origins
}
