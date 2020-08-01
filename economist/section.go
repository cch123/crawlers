package economist

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

type articleEntry struct {
	link string
	name string
}

type section struct {
	title          string
	articleLinks   []string
	articleEntries []articleEntry
}

// rootPath like : www.economist.com/weeklyedition/2020-07-20
func getSectionsAndCoverByURL(rootPath string) (sections []section, coverURL string, title string) {

	sectionCollector := colly.NewCollector()
	sectionCollector.OnHTML(".weekly-edition-header__headline", func(e *colly.HTMLElement) {
		title = e.Text
	})
	sectionCollector.OnHTML(".layout-weekly-edition-section", func(e *colly.HTMLElement) {
		title := e.ChildText(".ds-section-headline")
		children := e.ChildAttrs("a", "href")
		sections = append(sections, section{title: strings.Trim(title, `'`), articleLinks: dedup(children)})
	})
	sectionCollector.OnHTML(".weekly-edition-header__image", func(e *colly.HTMLElement) {
		coverURL = e.ChildAttr("img", "src")
	})

	err := sectionCollector.Visit(rootPath)
	if err != nil {
		fmt.Println(err)
	}

	return sections, coverURL, title
}

func dedup(input []string) []string {
	var res = []string{}
	var rec = map[string]struct{}{}
	for _, c := range input {
		if _, ok := rec[c]; !ok {
			res = append(res, c)
			rec[c] = struct{}{}
		}
	}
	return res
}
