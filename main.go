package main

import (
	"fmt"

	"github.com/araddon/dateparse"
	"github.com/cch123/crawlers/economist"
	"github.com/docopt/docopt-go"
)

const usage = `Crawlers.
Usage:
	exe -d <date> <magazine>
	exe -y <year> <magazine>
	exe <magazine>
`

type options struct {
	Magazine string `docopt:"<magazine>"`
	ByYear   bool   `docopt:"-y"`
	Year     string `docopt:"<year>"`
	ByDay    bool   `docopt:"-d"`
	Day      string `docopt:"<date>"`
}

func main() {
	v, err := docopt.ParseDoc(usage)
	if err != nil {
		fmt.Println("[cmd parse] failed", err)
		return
	}

	var opts options
	err = v.Bind(&opts)
	if err != nil {
		fmt.Println("[cmd parse] failed", err)
		return
	}

	switch opts.Magazine {
	case "economist":
		switch {
		case opts.ByDay:
			t, err := dateparse.ParseAny(opts.Day)
			if err != nil {
				fmt.Println("[date parse] failed", err)
				return
			}
			economist.CrawlByDay(t.Format("2006-01-02"))
		case opts.ByYear:
			economist.CrawlByYear(opts.Year)
		default:
			economist.CrawlLatest()
		}
	default:
		fmt.Println("this magazine is not supported")
	}
}
