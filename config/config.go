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
	Cities            []City
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

type Row struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (uc UserConfig) IsPermanentResidence() bool {
	return uc.ResidenceType != "temporary"
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

var UserConf = initializeUserConfig()
var ApplicationConf = initializeApplicationConfig()

func CollectUserData() []*Row {
	data := []*Row{}
	strings := ApplicationConf.Strings
	if UserConf.IsPermanentResidence() {
		data = append(data, &Row{strings.ResidenceTypeHeader, strings.ResidenceTypePermanent})
	} else {
		data = append(data, &Row{strings.ResidenceTypeHeader, strings.ResidenceTypeTemporary})
	}
	data = append(data, &Row{strings.NameSurnameHeader, fmt.Sprintf("%s %s", UserConf.Surname, UserConf.Name)})
	data = append(data, &Row{strings.CitizenshipHeader, UserConf.Citizenship})
	data = append(data, &Row{strings.DateOfBirthHeader, UserConf.DateOfBirth})
	data = append(data, &Row{strings.PhoneHeader, UserConf.Phone})
	data = append(data, &Row{strings.PassportHeader, UserConf.Passport})
	if UserConf.ResidenceCard != "" {
		data = append(data, &Row{strings.ResidenceCardHeader, UserConf.ResidenceCard})
	}
	data = append(data, &Row{strings.DataProcessingHeader, strings.DataProcessingValue})
	for _, additionalApplication := range UserConf.AdditionalApplications {
		var applicant string
		switch additionalApplication {
		case "child":
			applicant = strings.AdditionalApplicationTypeChild
		case "spouse":
			applicant = strings.AdditionalApplicationTypeSpouse
		case "children":
			applicant = strings.AdditionalApplicationTypeChildren
		}
		data = append(data, &Row{strings.AdditionalApplicationsHeader, applicant})
	}
	return data
}
