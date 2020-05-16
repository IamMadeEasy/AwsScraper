package pages

import (
	"fmt"

	"github.com/iammadeeasy/awsscraper/configuration"

	"github.com/gocolly/colly"
)

var linkCollector = *colly.NewCollector()

// GetPages - Get a list of pages to scrape for IAM data
func GetPages(config configuration.Config) (urls []string) {
	fmt.Println("Getting pages...")

	// Find and visit all links
	linkCollector.OnHTML(".highlights a", func(element *colly.HTMLElement) {
		urls = append(urls, element.Attr("href"))
	})

	linkCollector.Visit(config.BaseURL)

	return
}
