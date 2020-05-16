package configuration

import (
	"io/ioutil"
	"os"

	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
)

// Config holds the configuration values of the mini pony
type Config struct {
	BaseURL string `yaml:"base_url"`
}

// GetConfig Gets the configuration off disk
// It also checks/sets some defaults if nothing is found in the config
func GetConfig() *Config {
	glog.Infoln("Fetching the config")

	config := Config{}
	if _, err := os.Stat("config.yml"); err == nil {
		glog.Infoln("Found a config on disk")
		yamlFile, err := ioutil.ReadFile("config.yml")
		if err != nil {
			glog.Warningf("Error opening your config file: %v\n", err)
			os.Exit(1)
		}

		err = yaml.Unmarshal(yamlFile, &config)
		if err != nil {
			glog.Fatalf("Unable to parse your configuration file: %v", err)
		}
	} else if os.IsNotExist(err) {
		glog.Infoln("Using config defaults")
		config.BaseURL = "https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_actions-resources-contextkeys.html"
	}

	return &config
}
