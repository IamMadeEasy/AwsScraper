package aws

import (
	"regexp"

	"github.com/gocolly/colly"
)

// Condition - This handles the rows in the Condition Keys table, or the third/bottom table, on an IAM page
type Condition struct {
	Key         string `selector:"td:nth-child(1)"`
	Description string `selector:"td:nth-child(2)"`
	Type        string `selector:"td:nth-child(3)"`
}

// ProcessConditionKeys - Take table data and turn it into a map of condition structs where the key is the key in the map
func ProcessConditionKeys(table *colly.HTMLElement) map[string]Condition {
	var conditions = map[string]Condition{}

	table.ForEach("tr", func(_ int, tr *colly.HTMLElement) {
		condition := Condition{}
		tr.Unmarshal(&condition)

		// Get rid of TH's in the table - I can only work so much query selector magic
		if condition.Key != "" {
			space := regexp.MustCompile(`\s+`)
			condition.Description = space.ReplaceAllString(condition.Description, " ")
			conditions[condition.Key] = condition
		}
	})

	return conditions
}
