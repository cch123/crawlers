package economist

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const economistBaseURL = "https://www.economist.com"

// Crawl the economist
func Crawl() {
	// step 1 : get latest weekly URL
	urlSuffix, date := getLatestWeeklyEditionURL()
	fmt.Println("[crawl] the latest edition is ", urlSuffix)

	// step 2 : get sections from weekly front page
	var sections, coverURL = getSectionsAndCoverByURL(economistBaseURL + urlSuffix)

	// step 3 : prepare markdown && images file directories
	err := os.RemoveAll(date)
	fmt.Println("[rmdir]", err)

	// step 3.1 : mkdir 2020-07-20
	err = os.MkdirAll(date, 0755)
	fmt.Println("[mkdir]", err)

	// step 3.2 : download cover image
	fmt.Println("[cover download]", coverURL)
	downloadImagesToDir(date, coverURL)

	// step 3.3 : prepare dirs for sections
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
			// step 4.1 : download article
			// economist.com + /2020-07-05/{title}
			fullURL := economistBaseURL + articleURL
			article := getArticleByURL(fullURL)

			// step 4.2 : download image
			// lead image
			downloadImagesToDir(getImageDir(date, sec.title), article.leadImageURL)
			// body images
			downloadImagesToDir(getImageDir(date, sec.title), article.imageURLs...)

			// step 4.3 : create markdown file
			f, err := os.Create(getMarkdownFilePath(date, sec.title, getFileNameFromURL(articleURL)))
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

func getFileNameFromURL(url string) string {
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

func getLatestWeeklyEditionURL() (url string, date string) {
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
	arr := strings.Split(latestURL, "/")
	latestDate := arr[len(arr)-1]

	return latestURL, latestDate
}
