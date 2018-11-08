package config

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/tkanos/gonfig"
)

type City struct {
	Name  string
	Queue string
	Id    string
}

type Strings struct {
	ResidenceTypeHeader               string
	ResidenceTypeTemporary            string
	ResidenceTypePermanent            string
	NameSurnameHeader                 string
	CitizenshipHeader                 string
	DateOfBirthHeader                 string
	PhoneHeader                       string
	PassportHeader                    string
	ResidenceCardHeader               string
	DataProcessingHeader              string
	DataProcessingValue               string
	AdditionalApplicationsHeader      string
	AdditionalApplicationTypeChild    string
	AdditionalApplicationTypeSpouse   string
	AdditionalApplicationTypeChildren string
}

type ApplicationConfig struct {
	Strings           Strings
	ParallelismFactor int
	Cities            []*City
}

type UserConfig struct {
	Username               string
	Password               string
	Name                   string
	Surname                string
	DateOfBirth            string
	Citizenship            string
	Phone                  string
	Passport               string
	ResidenceCard          string
	ResidenceType          string
	AdditionalApplications []string
}

type row struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (uc UserConfig) IsPermanentResidence() bool {
	return uc.ResidenceType != "temporary"
}

var userConf *UserConfig
var applicationConf *ApplicationConfig

func init() {
	userConf = initializeUserConfig()
	applicationConf = initializeApplicationConfig()
}

func initializeConfig(name string, configuration interface{}) {
	path, _ := filepath.Abs(name)
	err := gonfig.GetConf(path, configuration)
	if err != nil {
		log.Fatalf("Can not read config\n%s\n\n", err)
	}
}

func initializeUserConfig() *UserConfig {
	configuration := UserConfig{}
	initializeConfig("user.yml", &configuration)
	return &configuration
}

func initializeApplicationConfig() *ApplicationConfig {
	configuration := ApplicationConfig{}
	initializeConfig("application.yml", &configuration)
	return &configuration
}

func UserConf() *UserConfig {
	return userConf
}

func ApplicationConf() *ApplicationConfig {
	return applicationConf
}

func CollectUserData() []*row {
	data := []*row{}
	strings := applicationConf.Strings
	if userConf.IsPermanentResidence() {
		data = append(data, &row{strings.ResidenceTypeHeader, strings.ResidenceTypePermanent})
	} else {
		data = append(data, &row{strings.ResidenceTypeHeader, strings.ResidenceTypeTemporary})
	}
	data = append(data, &row{strings.NameSurnameHeader, fmt.Sprintf("%s %s", userConf.Surname, userConf.Name)})
	data = append(data, &row{strings.CitizenshipHeader, userConf.Citizenship})
	data = append(data, &row{strings.DateOfBirthHeader, userConf.DateOfBirth})
	data = append(data, &row{strings.PhoneHeader, userConf.Phone})
	data = append(data, &row{strings.PassportHeader, userConf.Passport})
	if userConf.ResidenceCard != "" {
		data = append(data, &row{strings.ResidenceCardHeader, userConf.ResidenceCard})
	}
	data = append(data, &row{strings.DataProcessingHeader, strings.DataProcessingValue})
	for _, additionalApplication := range userConf.AdditionalApplications {
		var applicant string
		switch additionalApplication {
		case "child":
			applicant = strings.AdditionalApplicationTypeChild
		case "spouse":
			applicant = strings.AdditionalApplicationTypeSpouse
		case "children":
			applicant = strings.AdditionalApplicationTypeChildren
		}
		data = append(data, &row{strings.AdditionalApplicationsHeader, applicant})
	}
	return data
}
