package utilities

import (
	"strings"

	"github.com/iammadeeasy/awsscraper/configuration"
)

// URI - Retrieve the base URI from the configuration, strip off the page, and append it to any URI's which don't have a base
func URI(config configuration.Config, uri string) string {
	if !strings.HasPrefix(uri, "http") {
		uri = config.BaseURL[0:strings.LastIndex(config.BaseURL, "/")+1] + uri
	}
	return uri
}
