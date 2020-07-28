package economist

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/araddon/dateparse"
	"github.com/gocolly/colly"
)

const economistBaseURL = "https://www.economist.com"

// Crawler is an economist crawler
type Crawler struct {
	sleepInterval time.Duration
}

// NewCrawler get an economist crawler
func NewCrawler(grabInterval int) Crawler {
	// default, sleep 2s for each page
	var c = Crawler{time.Second * 2}
	if grabInterval > 0 {
		c.sleepInterval = time.Duration(grabInterval) * time.Second
	}

	return c
}

// CrawlByDay crawl by day
func (c Crawler) CrawlByDay(date string) {
	urlSuffix := "/weeklyedition/" + date
	c.crawl(urlSuffix, date)
	t, _ := dateparse.ParseAny(date)
	c.refreshEditionPage(fmt.Sprint(t.Year()))
}

// CrawlLatest crawl the latest
func (c Crawler) CrawlLatest() {
	// step 1 : get latest weekly URL
	date := getLatestWeeklyEditionURL()
	c.CrawlByDay(date)
}

type edition struct {
	url      string
	title    string
	date     string
	coverURL string
}

// original date is Oct 18, 2020
// we need to convert it
func generateEditionListPage(year string, editionList []edition) {
	// https://www.economist.com/weeklyedition/archive
	var pageContent string
	var tr = []string{}
	for idx, e := range editionList {
		t, _ := dateparse.ParseAny(e.date)
		date := t.Format("2006-01-02")
		tdContent := ""
		tdContent += "<p><img src = './" + date + "/cover.jpg'/></p>"
		tdContent += "<p>" + date + "</p>"
		tdContent += fmt.Sprintf("<p><a href='./%v'>%v</a></p>", date, e.title)
		tdContent = "<td>" + tdContent + "</td>"
		tr = append(tr, tdContent)
		if idx%4 == 3 || idx == len(editionList)-1 {
			// generate tr for table
			pageContent += "<tr>" + strings.Join(tr, "") + "</tr>"
			tr = tr[:0]
			continue
		}
	}

	pageContent = fmt.Sprintf("# Economist %v\n", year) + "<table>" + pageContent + "</table>"

	f, err := os.Create("readme.md")
	if err != nil {
		fmt.Println("[generate edition page]failed", err)
		return
	}
	defer f.Close()

	_, err = f.WriteString(pageContent)
	if err != nil {
		fmt.Println("[generate edition page]failed", err)
		return
	}
}

// CrawlByYear crawl economist by year
func (c Crawler) CrawlByYear(year string) {
	editionList := c.refreshEditionPage(year)

	for _, e := range editionList {
		date := getLastSegmentFromURL(e.url)
		c.CrawlByDay(date)
	}
}

func (c Crawler) refreshEditionPage(year string) []edition {
	// https://www.economist.com/weeklyedition/archive?year=2019
	co := colly.NewCollector()
	var editionList []edition
	co.OnHTML(".edition-teaser", func(elem *colly.HTMLElement) {
		var e edition
		e.coverURL = elem.ChildAttr("img", "src")
		e.title = elem.ChildText(".edition-teaser__headline")
		e.url = elem.ChildAttr(".headline-link", "href")
		e.date = getLastSegmentFromURL(e.url)
		// e.date = elem.ChildText(".edition-teaser__subheadline"), this date is not correct
		editionList = append(editionList, e)
	})
	co.Visit(fmt.Sprintf("https://www.economist.com/weeklyedition/archive?year=%v", year))
	generateEditionListPage(year, editionList)
	return editionList
}

// crawl the economist
func (c Crawler) crawl(urlSuffix, date string) {
	// step 2 : get sections from weekly front page
	var sections, coverURL, pageTitle = getSectionsAndCoverByURL(economistBaseURL + urlSuffix)

	// step 3 : prepare markdown && images file directories
	err := os.RemoveAll(date)
	fmt.Println("[rmdir]", err)

	// step 3.1 : mkdir 2020-07-20
	err = os.MkdirAll(date, 0755)
	fmt.Println("[mkdir]", err)

	// step 3.2 : download cover image
	fmt.Println("[cover download]", coverURL)
	downloadImagesToDir("cover.jpg", date, coverURL)

	// step 3.3 : create section list page
	createSectionArticleListPage(date, pageTitle, sections)

	// step 3.4 : prepare dirs for sections
	for _, sec := range sections {
		// dir for markdown files
		err = os.MkdirAll(getMarkdownFileDir(date, sec.title), 0755)
		if err != nil {
			fmt.Println("[mkdir markdown]", err)
		}

		// dir for image files
		err = os.MkdirAll(getImageDir(date, sec.title), 0755)
		if err != nil {
			fmt.Println("[mkdir img]", err)
		}
	}

	// step 4 : download articles && images
	for _, sec := range sections {
		for _, articleURL := range sec.articleLinks {
			if c.sleepInterval > 0 {
				time.Sleep(c.sleepInterval)
			}
			// step 4.1 : download article
			// economist.com + /2020-07-05/{title}
			fullURL := economistBaseURL + articleURL
			article := getArticleByURL(fullURL)

			// step 4.2 : download image
			// lead image
			downloadImagesToDir("", getImageDir(date, sec.title), article.leadImageURL)
			// body images
			downloadImagesToDir("", getImageDir(date, sec.title), article.imageURLs...)

			// step 4.3 : create markdown file
			f, err := os.Create(getMarkdownFilePath(date, sec.title, getLastSegmentFromURL(articleURL)))
			if err != nil {
				fmt.Println(err)
				continue
			}
			defer f.Close()

			// step 4.4 : write content to files
			_, err = f.WriteString(article.generateMarkdown())
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	}
}

func createSectionArticleListPage(date string, pageTitle string, sections []section) {
	fmt.Println("[create section list page]")
	f, err := os.Create(date + "/readme.md")
	if err != nil {
		fmt.Println("[create section list] failed", err)
		return
	}
	defer f.Close()

	sectionListPageContent := "## " + pageTitle + "\n"
	sectionListPageContent += "![](./cover.jpg)\n"
	for _, sec := range sections {
		sectionListPageContent += "### " + sec.title + "\n"
		for _, articleURL := range sec.articleLinks {
			line := fmt.Sprintf("[%v](%v)\n", getTileFromDashed(getLastSegmentFromURL(articleURL)), strings.ReplaceAll("./"+sec.title+"/"+getLastSegmentFromURL(articleURL)+".md", " ", "%20"))
			// line = strings.ReplaceAll(line, " ", "%20")
			sectionListPageContent += "#### " + line
		}
	}

	f.WriteString(sectionListPageContent)
	fmt.Println("[create section list] done")
}

func getTileFromDashed(str string) string {
	str = strings.ReplaceAll(str, "-", " ")
	result := []rune{unicode.ToUpper(rune(str[0]))}
	result = append(result, []rune(str[1:])...)
	return string(result)
}

func getLastSegmentFromURL(url string) string {
	var arr = strings.Split(url, "/")
	var lastIdx = len(arr) - 1
	return arr[lastIdx]
}

func getMarkdownFilePath(date, sectionTitle, articleTitle string) string {
	return date + "/" + sectionTitle + "/" + articleTitle + ".md"
}

func getMarkdownFileDir(date, sectionTitle string) string {
	return date + "/" + sectionTitle
}

func getImageDir(date, sectionTitle string) string {
	return date + "/" + sectionTitle + "/images"
}

func getLatestWeeklyEditionURL() (date string) {
	client := http.Client{
		Timeout: time.Second * 5,
		// just tell me the redirect target
		// and don't redirect
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Get("https://www.economist.com/weeklyedition")
	if err != nil {
		fmt.Println("[getLatest] failed", err)
		os.Exit(1)
	}

	latestURL := resp.Header.Get("Location")
	return getLastSegmentFromURL(latestURL)
}
