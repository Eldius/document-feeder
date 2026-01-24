package model

import "time"

type Feed struct {
	Title       string     `json:"title" storm:"unique"`
	Description string     `json:"description"`
	Link        string     `json:"link" storm:"unique"`
	FeedLink    string     `json:"feedLink" storm:"id"`
	Links       []string   `json:"links"`
	Language    string     `json:"language"`
	Image       Image      `json:"image"`
	Extensions  Extensions `json:"extensions"`
	Items       []Article  `json:"items"`
	FeedType    string     `json:"feedType"`
	FeedVersion string     `json:"feedVersion"`
}

type Image struct {
	Url   string `json:"url"`
	Title string `json:"title"`
}

type Extensions map[string]ExtensionMap

type ExtensionMap map[string][]Extension

type Extension struct {
	Name     string                 `json:"name"`
	Value    string                 `json:"value"`
	Attrs    map[string]string      `json:"attrs"`
	Children map[string][]Extension `json:"children"`
}

type Atom struct {
	Link []Link `json:"link"`
}

type Link struct {
	Name     string         `json:"name"`
	Value    string         `json:"value"`
	Attrs    LinkAttr       `json:"attrs"`
	Children map[string]any `json:"children"`
}

type LinkAttr struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
	Type string `json:"type"`
}

type Article struct {
	Title           string           `json:"title"`
	Description     string           `json:"description"`
	Content         string           `json:"content"`
	Link            string           `json:"link"`
	Links           []string         `json:"links"`
	Published       string           `json:"published"`
	PublishedParsed *time.Time       `json:"publishedParsed"`
	Author          Author           `json:"author"`
	Authors         []Author         `json:"authors"`
	Guid            string           `json:"guid"`
	DcExt           DcExt            `json:"dcExt"`
	Extensions      ArticleExtension `json:"extensions"`
	Image           ArticleImage     `json:"image,omitempty"`
	Categories      []string         `json:"categories,omitempty"`
}

type Author struct {
	Name string `json:"name"`
}

type DcExt struct {
	Creator []string `json:"creator"`
}

type ArticleExtension struct {
	Dc ArticleDc `json:"dc"`
}

type ArticleDc struct {
	Creator []ArticleCreator `json:"creator"`
}

type ArticleCreator struct {
	Name     string         `json:"name"`
	Value    string         `json:"value"`
	Attrs    map[string]any `json:"attrs"`
	Children map[string]any `json:"children"`
}

type ArticleImage struct {
	Url string `json:"url"`
}

func (f *Feed) AddItems(a ...Article) {
	hash := map[string]struct{}{}
	for _, item := range f.Items {
		hash[item.Link] = struct{}{}
	}
	for _, item := range a {
		if _, ok := hash[item.Link]; ok {
			continue
		}
		f.Links = append(f.Links, item.Link)
	}
}
