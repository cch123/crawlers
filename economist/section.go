package economist

import (
	"fmt"

	"github.com/gocolly/colly"
)

type section struct {
	title        string
	articleLinks []string
}

// rootPath like : www.economist.com/weeklyedition/2020-07-20
func getSectionsAndCoverByURL(rootPath string) ([]section, string) {
	var (
		sections []section
		coverURL string
	)

	sectionCollector := colly.NewCollector()
	sectionCollector.OnHTML(".layout-weekly-edition-section", func(e *colly.HTMLElement) {
		title := e.ChildText(".ds-section-headline")
		children := e.ChildAttrs("a", "href")
		sections = append(sections, section{title: title, articleLinks: children})
	})
	sectionCollector.OnHTML(".weekly-edition-header__image", func(e *colly.HTMLElement) {
		coverURL = e.ChildAttr("img", "src")
	})

	err := sectionCollector.Visit(rootPath)
	if err != nil {
		fmt.Println(err)
	}

	return sections, coverURL
}
