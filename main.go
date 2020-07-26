package main

import (
	"fmt"

	"github.com/araddon/dateparse"
	"github.com/cch123/crawlers/economist"
	"github.com/docopt/docopt-go"
)

const usage = `Crawlers.
Usage:
	exe -d <date> <magazine> [--seconds=<value>]
	exe -y <year> <magazine>
	exe <magazine>
`

type options struct {
	Magazine     string `docopt:"<magazine>"`
	ByYear       bool   `docopt:"-y"`
	Year         string `docopt:"<year>"`
	ByDay        bool   `docopt:"-d"`
	Day          string `docopt:"<date>"`
	GrabInterval int    `docopt:"--seconds"`
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
		cr := economist.NewCrawler(opts.GrabInterval)
		switch {
		case opts.ByDay:
			t, err := dateparse.ParseAny(opts.Day)
			if err != nil {
				fmt.Println("[date parse] failed", err)
				return
			}
			cr.CrawlByDay(t.Format("2006-01-02"))
		case opts.ByYear:
			cr.CrawlByYear(opts.Year)
		default:
			cr.CrawlLatest()
		}
	default:
		fmt.Println("this magazine is not supported")
	}
}
