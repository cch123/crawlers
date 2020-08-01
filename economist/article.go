package economist

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

type article struct {
	// header part
	leadImageURL string
	headline     string
	subHeadline  string
	description  string

	// body part
	meta       string
	paragraphs []string

	// paragraph images
	imageURLs []string

	// full url
	fullURL string
}

func getArticleByURL(url string) article {
	articleCollector := colly.NewCollector()
	var (
		// header
		leadImgURL  string
		headline    string
		subHeadline string
		description string

		// body
		meta       string
		paragraphs []string

		// images
		imageURLs []string // image url in this article
	)

	// header part
	// ds-layout-grid ds-layout-grid--edged layout-article-header
	articleCollector.OnHTML(".layout-article-header", func(e *colly.HTMLElement) {
		headline = e.ChildText(".article__headline")
		subHeadline = e.ChildText(".article__subheadline")
		leadImgURL = e.ChildAttr("img", "src")
		description = e.ChildText(".article__description")
	})

	// body part
	// ds-layout-grid ds-layout-grid--edged layout-article-body
	articleCollector.OnHTML(".layout-article-body", func(e *colly.HTMLElement) {
		meta = e.ChildText(".layout-article-meta")
		e.ForEach(".article__body-text, img", func(idx int, internal *colly.HTMLElement) {
			if internal.Name == "img" {
				// xxxx.jpg 2048
				imageRawURL := internal.Attr("src")
				arr := strings.Split(imageRawURL, " ")

				var imageURL = arr[0]
				imageURLs = append(imageURLs, imageURL)

				// insert this image as a img element to markdown paragraph
				imageContent := fmt.Sprintf("![](./images/%v)", getLastSegmentFromURL(imageURL))

				paragraphs = append(paragraphs, imageContent)
			} else {
				// convert links to markdown format
				internal.ForEach("a", func(_ int, e *colly.HTMLElement) {
					e.DOM.SetText(fmt.Sprintf("[%v](https://www.economist.com/%v)", e.Text, e.Attr("href")))
				})
				paragraphs = append(paragraphs, internal.DOM.Text())
			}

		})
	})

	err := articleCollector.Visit(url)
	if err != nil {
		fmt.Println("[crawl] failed to crawl article", url, err)
		return article{}
	}

	fmt.Println("[crawl]visit url", url, headline, subHeadline, leadImgURL)

	return article{
		// header
		leadImageURL: leadImgURL,
		headline:     headline,
		subHeadline:  subHeadline,
		description:  description,

		// body
		meta:       meta,
		paragraphs: paragraphs,

		// images
		imageURLs: imageURLs,

		fullURL: url,
	}
}

func (a article) generateMarkdown() string {
	var content string
	if a.leadImageURL != "" {
		content += fmt.Sprintf("![](./images/%v)", getLastSegmentFromURL(a.leadImageURL))
		content += "\n\n"
	}

	if a.subHeadline != "" {
		content += "## " + a.subHeadline + "\n\n"
	}

	if a.headline != "" {
		content += "# " + a.headline + "\n\n"
	}

	if a.description != "" {
		content += "> " + a.description + "\n\n"
	}

	if a.meta != "" {
		content += "> " + a.meta + "\n\n"
	}

	if len(a.paragraphs) > 0 {
		content += strings.Join(a.paragraphs, "\n\n")
	}

	content += "\n\n## URL\n\n" + a.fullURL + "\n"

	return content
}
