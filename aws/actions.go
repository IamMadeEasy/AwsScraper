package aws

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
	"gopkg.in/yaml.v2"
)

// ResourceRequired - This handles if a resource is required or not
type ResourceRequired struct {
	Resource string
	Required bool
}

// ActionTableDataRow -  This handles primary row data in the table
type ActionTableDataRow struct {
	Service          string
	Action           string
	Description      string
	AccessLevel      string
	Resources        []Resource
	ConditionKeys    [][]Condition
	DependentActions []string
}

// ProcessActions - Processes the Actions table into a somewhat useable format
func ProcessActions(table *colly.HTMLElement, rawResources map[string]Resource, rawConditions map[string]Condition, service string, iamPrefix string) []ActionTableDataRow {
	// Need to gather up all rows we encounter
	var actionTableDataRows []ActionTableDataRow
	actionTableDataRow := ActionTableDataRow{}

	var rows []colly.HTMLElement

	table.ForEach("tr", func(_ int, tr *colly.HTMLElement) {
		rows = append(rows, *tr)
	})

	// We start at 1, instead of 0, to skip the row representing the column names
	for x := 1; x < len(rows); x++ {
		columns := rows[x].DOM.ChildrenFiltered("td").Length()

		rows[x].ForEach("td", func(_ int, td *colly.HTMLElement) {
			text := strings.TrimSpace(td.Text)

			if columns == 6 {
				// Primary row representing an action
				if td.Index == 0 {
					// Service
					actionTableDataRow.Service = service

					// Action
					re := regexp.MustCompile(`^(\w+)\s+.*`)
					text = re.ReplaceAllString(text, "$1")
					actionTableDataRow.Action = iamPrefix + ":" + text
				} else if td.Index == 1 {
					space := regexp.MustCompile(`\s+`)
					actionTableDataRow.Description = space.ReplaceAllString(text, " ")
				} else if td.Index == 2 {
					actionTableDataRow.AccessLevel = text
				} else if td.Index == 3 {
					children := td.DOM.Children()

					if children.Length() > 0 {
						td.ForEach("p", func(_ int, p *colly.HTMLElement) {
							text := strings.TrimSpace(p.Text)
							if text != "" {
								requiredResource := getResource(text)
								if requiredResource.Required == true {
									var rawResource = rawResources[requiredResource.Resource]
									rawResource.Required = true
									actionTableDataRow.Resources = append(actionTableDataRow.Resources, rawResource)
								} else {
									actionTableDataRow.Resources = append(actionTableDataRow.Resources, rawResources[text])
								}
							}
						})
					} else {
						actionTableDataRow.Resources = append(actionTableDataRow.Resources, rawResources[text])
					}
				} else if td.Index == 4 {
					var conditions []Condition

					td.ForEach("p", func(_ int, p *colly.HTMLElement) {
						text := strings.TrimSpace(p.Text)
						conditions = append(conditions, rawConditions[text])
					})

					length := len(actionTableDataRow.Resources)
					if len(conditions) > 0 {
						// We have a condition when there is no resource specified
						actionTableDataRow.Resources[length-1].Conditions = append(actionTableDataRow.Resources[length-1].Conditions, conditions...)
					} else {
						// Dirty hack, we don't have conditions, it's possible we also didn't have a resource, so let's pop the last row off the actionTableDataRow if that's the case. We need to re-write this whole file.
						resource := actionTableDataRow.Resources[length-1]

						if resource.ResourceType == "" {
							actionTableDataRow.Resources = actionTableDataRow.Resources[:len(actionTableDataRow.Resources)-1]
						}
					}
				} else if td.Index == 5 {
					td.ForEach("p", func(_ int, p *colly.HTMLElement) {
						text := strings.TrimSpace(p.Text)
						actionTableDataRow.DependentActions = append(actionTableDataRow.DependentActions, text)
					})
				}
			} else if columns == 3 {
				// Secondary row representing additional resources for an action
				var resourceType string

				if td.Index == 0 {
					children := td.DOM.Children()

					// We don't need to make a resource, one should already exist
					if children.Length() > 0 {

						td.ForEach("p", func(_ int, p *colly.HTMLElement) {
							resourceType = strings.TrimSpace(p.Text)

							requiredResource := getResource(resourceType)
							if requiredResource.Required == true {
								var rawResource = rawResources[requiredResource.Resource]
								rawResource.Required = true
								actionTableDataRow.Resources = append(actionTableDataRow.Resources, rawResource)
							} else {
								actionTableDataRow.Resources = append(actionTableDataRow.Resources, rawResources[resourceType])
							}
						})
					} else {
						actionTableDataRow.Resources = append(actionTableDataRow.Resources, rawResources[text])
					}
				} else if td.Index == 1 {
					var conditions []Condition

					td.ForEach("p", func(_ int, p *colly.HTMLElement) {
						text := strings.TrimSpace(p.Text)
						conditions = append(conditions, rawConditions[text])
					})

					if len(conditions) > 0 {

						length := len(actionTableDataRow.Resources)

						// We have a conditions when there is no resource specified
						actionTableDataRow.Resources[length-1].Conditions = append(actionTableDataRow.Resources[length-1].Conditions, conditions...)
					}
				} else if td.Index == 2 {
					td.ForEach("p", func(_ int, p *colly.HTMLElement) {
						text := strings.TrimSpace(p.Text)
						actionTableDataRow.DependentActions = append(actionTableDataRow.DependentActions, text)
					})
				}
			}
		})

		// Find the length of the next row so we know what to do with this one
		if x+1 >= len(rows) {
			actionTableDataRows = append(actionTableDataRows, actionTableDataRow)
		} else {
			nextRowColumns := rows[x+1].DOM.ChildrenFiltered("td").Length()
			if nextRowColumns == 6 {
				actionTableDataRows = append(actionTableDataRows, actionTableDataRow)
				actionTableDataRow = ActionTableDataRow{}
			}
		}
	}

	return actionTableDataRows
}

// WriteActions - Writes the actual full action to disk
func WriteActions(actionTableDataRows []ActionTableDataRow) {
	var yamlAction []byte
	var err error
        yaml.FutureLineWrap()
	yamlAction, err = yaml.Marshal(&actionTableDataRows)
	if err != nil {
		fmt.Println("error: ", err)
	}

	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile("iam.yml", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write(yamlAction); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

// getResource - Takes a string of either "thing", or "thing*", and returns a resource struct
func getResource(resourceType string) ResourceRequired {
	strippedResourceType := strings.Replace(resourceType, "*", "", -1)

	resource := ResourceRequired{}
	resource.Resource = strippedResourceType

	if resourceType == strippedResourceType {
		resource.Required = false
	} else {
		resource.Required = true
	}

	return resource
}
