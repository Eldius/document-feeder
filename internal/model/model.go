package model

import "time"

type SearchResult struct {
	FeedTitle string `json:"feedTitle"`

	Article Article

	Similarity       float32
	SanitizedContent string
	Embeddings       []float32
}

type AnswerCache struct {
	ID       string `json:"id" storm:"id"`
	Question string `json:"feedTitle" storm:"index"`
	Answer   string `json:"answer"`
}

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
	Title           string           `json:"title" storm:"index"`
	Description     string           `json:"description"`
	Content         string           `json:"content"`
	Link            string           `json:"link" storm:"index"`
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
	Name string `json:"name" storm:"index"`
}

type DcExt struct {
	Creator []string `json:"creator" storm:"index"`
}

type ArticleExtension struct {
	Dc ArticleDc `json:"dc"`
}

type ArticleDc struct {
	Creator []ArticleCreator `json:"creator" storm:"index"`
}

type ArticleCreator struct {
	Name     string         `json:"name" storm:"index"`
	Value    string         `json:"value"`
	Attrs    map[string]any `json:"attrs"`
	Children map[string]any `json:"children"`
}

type ArticleImage struct {
	Url string `json:"url"`
}

func (f *Feed) AddItems(a ...Article) {
	hash := make(map[string]int)
	for i, item := range f.Items {
		hash[item.Link] = i
	}
	for _, item := range a {
		if _, ok := hash[item.Link]; ok {
			f.Items[hash[item.Link]] = item
		}
		f.Items = append(f.Items, item)
	}
}
