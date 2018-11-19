package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
)

//City represents city
type City struct {
	Name  string
	Queue string
	Id    string
}

//Strings represents booking specific strings
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

//ApplicationConfig - just it
type ApplicationConfig struct {
	Strings           Strings
	ParallelismFactor int
	Cities            []*City
}

//UserConfig - just it
type UserConfig struct {
	Login                  string
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

//IsPermanentResidence - is application permanent
func (uc UserConfig) IsPermanentResidence() bool {
	return uc.ResidenceType != "temporary"
}

var userConf *UserConfig
var applicationConf *ApplicationConfig

func init() {
	userConf = initializeUserConfig()
	applicationConf = initializeApplicationConfig()
}

func unmarshalConfig(path string, configuration interface{}) (err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(data, &configuration)
	if err != nil {
		return
	}
	return
}

func initializeConfig(name string, configuration interface{}) {
	path, _ := filepath.Abs(name)
	err := unmarshalConfig(path, configuration)
	if err != nil {
		log.Fatalf("Can not read config\n%s\n\n", err)
	}
}

func initializeUserConfig() *UserConfig {
	configuration := &UserConfig{}
	initializeConfig("user.yml", configuration)
	return configuration
}

func initializeApplicationConfig() *ApplicationConfig {
	configuration := &ApplicationConfig{}
	initializeConfig("application.yml", configuration)
	return configuration
}

//UserConf returns config with user specific details
func UserConf() *UserConfig {
	return userConf
}

//ApplicationConf returns application config
func ApplicationConf() *ApplicationConfig {
	return applicationConf
}

//CollectUserData returns used data in "ready to convert to json" format
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
