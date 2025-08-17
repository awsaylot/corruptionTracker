package sites

// SiteSpecificRules contains selectors and rules for specific news sites
var SiteSpecificRules = map[string]struct {
	Name      string
	Domain    string
	Selectors struct {
		Article []string
		Title   []string
		Content []string
		Author  []string
		Date    []string
		Paywall []string
		Cookies []string
		Cleanup []string
	}
}{
	"theguardian.com": {
		Name:   "The Guardian",
		Domain: "theguardian.com",
		Selectors: struct {
			Article []string
			Title   []string
			Content []string
			Author  []string
			Date    []string
			Paywall []string
			Cookies []string
			Cleanup []string
		}{
			Article: []string{".article"},
			Title:   []string{".dcr-y70er3"},
			Content: []string{".article-body-commercial-selector"},
			Author:  []string{".dcr-676fqo"},
			Date:    []string{"[itemprop='datePublished']"},
			Cookies: []string{".css-v8k49"},
			Cleanup: []string{".submeta", ".content-footer"},
		},
	},
	"reuters.com": {
		Name:   "Reuters",
		Domain: "reuters.com",
		Selectors: struct {
			Article []string
			Title   []string
			Content []string
			Author  []string
			Date    []string
			Paywall []string
			Cookies []string
			Cleanup []string
		}{
			Article: []string{"article"},
			Title:   []string{".article-header__title__text"},
			Content: []string{".article-body__content__17Yit"},
			Author:  []string{".author-name"},
			Date:    []string{"time"},
			Paywall: []string{".paywall-article"},
			Cookies: []string{"#onetrust-consent-sdk"},
			Cleanup: []string{".article-body__tags__17Yit"},
		},
	},
	// Add more news sites as needed
}

// GetSiteRules returns the rules for a specific news site
func GetSiteRules(domain string) (struct {
	Name      string
	Domain    string
	Selectors struct {
		Article []string
		Title   []string
		Content []string
		Author  []string
		Date    []string
		Paywall []string
		Cookies []string
		Cleanup []string
	}
}, bool) {
	rules, found := SiteSpecificRules[domain]
	return rules, found
}
