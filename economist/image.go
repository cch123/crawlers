package economist

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func downloadImagesToDir(imageFileName, imageDir string, imageURLs ...string) {
	// extract image urls from article content
	var downloadFunc = func(url string) {
		// www.economist.com/sites/default/files/images/print-edition/20200725_WWC588.png
		var fileName string
		if imageFileName != "" {
			fileName = imageFileName
		} else {
			fileName = getFileNameFromURL(url)
			if fileName == "" {
				return
			}
		}

		f, err := os.Create(imageDir + "/" + fileName)
		if err != nil {
			fmt.Println("[create image file] failed to create img file : ", url, err)
			return
		}
		defer f.Close()

		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("[download image] failed to download img : ", url, err)
			return
		}
		defer resp.Body.Close()

		imgBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("[download image] read img resp failed: ", url, err)
			return
		}

		_, err = f.Write(imgBytes)
		if err != nil {
			fmt.Println("[write image] write img to file failed: ", url, err)
			return
		}
	}

	for _, url := range imageURLs {
		downloadFunc(url)
	}
}
