package sites

// Common selectors that work across most news sites
var CommonSelectors = struct {
	Article    []string
	Title      []string
	Content    []string
	Author     []string
	Date       []string
	SocialTags []string
}{
	Article: []string{
		"article",
		"[role='article']",
		".article",
		".post",
		".story",
		"#article-body",
		".article-body",
	},
	Title: []string{
		"h1",
		".article-title",
		".entry-title",
		".post-title",
		"[itemprop='headline']",
	},
	Content: []string{
		".article-content",
		".post-content",
		".entry-content",
		".story-content",
		"[itemprop='articleBody']",
		".story-body",
	},
	Author: []string{
		"[rel='author']",
		".author",
		".byline",
		"[itemprop='author']",
		".writer",
		".article-author",
	},
	Date: []string{
		"[itemprop='datePublished']",
		"time",
		".date",
		".published",
		"meta[property='article:published_time']",
		".article-date",
		".post-date",
	},
	SocialTags: []string{
		"meta[property^='og:']",
		"meta[name^='twitter:']",
		"meta[property^='article:']",
	},
}

// ContentCleanupRules defines rules for cleaning article content
var ContentCleanupRules = struct {
	RemoveSelectors []string
	UnwrapSelectors []string
}{
	RemoveSelectors: []string{
		".advertisement",
		".social-share",
		".related-articles",
		".newsletter-signup",
		".comments",
		"script",
		"style",
		"iframe",
	},
	UnwrapSelectors: []string{
		".article-text",
		".article-paragraph",
		"p",
	},
}
