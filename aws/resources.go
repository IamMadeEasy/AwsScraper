package aws

import (
	"github.com/gocolly/colly"
)

// Resource - This handles the rows in the Resources table, or the second/middle (when three are showing) table, on an IAM page
type Resource struct {
	ResourceType  string   `selector:"td:nth-child(1)"`
	Arn           string   `selector:"td:nth-child(2)"`
	ConditionKeys []string `selector:"td:nth-child(3) > p"`
	Conditions    []Condition
	Required      bool
}

// ProcessResourceTypes - Take table data and turn it into a map of Resource structs where the XXX
func ProcessResourceTypes(table *colly.HTMLElement, conditions map[string]Condition) map[string]Resource {
	var resources = map[string]Resource{}

	table.ForEach("tr", func(_ int, tr *colly.HTMLElement) {
		resource := Resource{}
		tr.Unmarshal(&resource)

		// Given the condition listed in a resource row, look it up from the conditions passed in, and assign it's structure here.
		// Conditions can also be assigned to a non resource element, so we need to handle that as well.
		for i := 0; i < len(resource.ConditionKeys); i++ {
			for k := range conditions {
				if k == resource.ConditionKeys[i] {
					resource.Conditions = append(resource.Conditions, conditions[k])
					break
				}
			}
		}

		// Only need these at load and processing, not writing
		resource.ConditionKeys = nil

		resources[resource.ResourceType] = resource
	})

	return resources
}
