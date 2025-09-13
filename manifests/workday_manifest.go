package manifests

import (
	"os"

	"github.com/goccy/go-yaml"
)

type WorkdayCompany struct {
	Name                   string         `yaml:"name"`
	BaseUrl                string         `yaml:"baseURL"`
	ListAllJobsRequestBody map[string]any `yaml:"listAllJobsRequestBody"`
}

type WorkdayYAML struct {
	Companies map[string]WorkdayCompany `yaml:"company"`
}

func LoadWorkdayCompanies(filePath string) WorkdayYAML {
	var workdayCompanies WorkdayYAML
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(yamlFile, &workdayCompanies)
	if err != nil {
		panic(err)
	}

	return workdayCompanies
}
