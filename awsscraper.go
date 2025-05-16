package main

import (
	"flag"

	"github.com/golang/glog"
	"github.com/iammadeeasy/awsscraper/configuration"
	"github.com/iammadeeasy/awsscraper/pages"
	"github.com/iammadeeasy/awsscraper/utilities"
)

func init() {
	// Supress errors about flags from glog
	flag.Parse()
}

func main() {
	glog.Infoln("Firing up the collector")
	debug := false 

	if debug == false {
		config := configuration.GetConfig()
		visitPages := pages.GetPages(*config)

		for i := 0; i < len(visitPages); i++ {
			url := utilities.URI(*config, visitPages[i])
			pages.GetPage(url)
		}
	} else {
		pages.GetPage("https://docs.aws.amazon.com/IAM/latest/UserGuide/./list_awsssodirectory.html")
	}
}
