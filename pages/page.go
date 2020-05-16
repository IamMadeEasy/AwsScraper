package pages

import (
	"fmt"
	"regexp"

	"github.com/iammadeeasy/awsscraper/aws"

	"github.com/gocolly/colly"
)

var pageCollector = *colly.NewCollector()
var service = ""
var iamPrefix = ""

// GetPage - Gets data off a page and returns it
func GetPage(url string) {
	pageCollector = *colly.NewCollector()
	fmt.Println("Getting page", url)

	findService()
	findIamPrefix()
	processTables()

	pageCollector.Visit(url)
}

func findService() {
	pageCollector.OnHTML("#main-col-body > :nth-child(4)", func(e *colly.HTMLElement) {
		re := regexp.MustCompile(`(?s)\s+\(.*`)
		service = re.ReplaceAllString(e.Text, "")
	})
}

func findIamPrefix() {
	pageCollector.OnHTML("#main-col-body > :nth-child(4) > code", func(e *colly.HTMLElement) {
		iamPrefix = e.Text
	})
}

func processTables() {
	// Find all table rows
	var tables []*colly.HTMLElement

	pageCollector.OnHTML(".table-contents", func(table *colly.HTMLElement) {
		tables = append(tables, table)
	})

	pageCollector.OnScraped(func(_ *colly.Response) {
		conditions := map[string]aws.Condition{}
		resources := map[string]aws.Resource{}

		// Intentionally parse backwards. We want conditions first, as conditions are used in both resources and actions and resources are used in actions as well
		for i := len(tables) - 1; i >= 0; i-- {
			table := tables[i]
			tableHeader := table.ChildText("table > tbody > tr > th:nth-child(1)")

			if tableHeader == "Actions" {
				actionRows := aws.ProcessActions(table, resources, conditions, service, iamPrefix)
				aws.WriteActions(actionRows)
			} else if tableHeader == "Resource Types" {
				resources = aws.ProcessResourceTypes(table, conditions)
			} else if tableHeader == "Condition Keys" {
				conditions = aws.ProcessConditionKeys(table)
			}
		}
	})
}
