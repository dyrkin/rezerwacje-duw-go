package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
)

//Entity represents city or department details
type Entity struct {
	Name      string
	ShortName string
	Queue     string
	ID        string
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
	LpInfo                            string
	LpNameSurnameHeader               string
	LpDateOfBirthHeader               string
	LpPhoneHeader                     string
	LpReferenceNumberHeader           string
	LpSubmissionDateHeader            string
}

//ApplicationConfig - just it
type ApplicationConfig struct {
	Strings           Strings
	ParallelismFactor int
	Cities            []*Entity
	Departments       []*Entity
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
	ReferenceNumber        string
	SubmissionDate         string
}

type Row struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

//IsPermanentResidence - is application permanent
func (uc UserConfig) IsPermanentResidence() bool {
	return uc.ResidenceType != "temporary"
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
	return initializeUserConfig()
}

//ApplicationConf returns application config
func ApplicationConf() *ApplicationConfig {
	return initializeApplicationConfig()
}

//CollectApplicationSubmissionData returns user data related to
//application submission in "ready to convert to json" format
func CollectApplicationSubmissionData() []*Row {
	userConf := UserConf()
	applicationConf := ApplicationConf()
	data := []*Row{}
	strings := applicationConf.Strings
	if userConf.IsPermanentResidence() {
		data = append(data, &Row{strings.ResidenceTypeHeader, strings.ResidenceTypePermanent})
	} else {
		data = append(data, &Row{strings.ResidenceTypeHeader, strings.ResidenceTypeTemporary})
	}
	data = append(data, &Row{strings.NameSurnameHeader, fmt.Sprintf("%s %s", userConf.Surname, userConf.Name)})
	data = append(data, &Row{strings.CitizenshipHeader, userConf.Citizenship})
	data = append(data, &Row{strings.DateOfBirthHeader, userConf.DateOfBirth})
	data = append(data, &Row{strings.PhoneHeader, userConf.Phone})
	data = append(data, &Row{strings.PassportHeader, userConf.Passport})
	if userConf.ResidenceCard != "" {
		data = append(data, &Row{strings.ResidenceCardHeader, userConf.ResidenceCard})
	}
	data = append(data, &Row{strings.DataProcessingHeader, strings.DataProcessingValue})
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
		data = append(data, &Row{strings.AdditionalApplicationsHeader, applicant})
	}
	return data
}

//CollectHeadOfDepartmentData returns user data related to
//making reservation of a visit to head of department in "ready to convert to json" format
func CollectHeadOfDepartmentData() []*Row {
	userConf := UserConf()
	applicationConf := ApplicationConf()
	data := []*Row{}
	strings := applicationConf.Strings
	data = append(data, &Row{strings.LpInfo, ""})
	data = append(data, &Row{strings.LpNameSurnameHeader, fmt.Sprintf("%s %s", userConf.Surname, userConf.Name)})
	data = append(data, &Row{strings.LpDateOfBirthHeader, userConf.DateOfBirth})
	data = append(data, &Row{strings.LpPhoneHeader, userConf.Phone})
	data = append(data, &Row{strings.LpReferenceNumberHeader, userConf.ReferenceNumber})
	data = append(data, &Row{strings.LpSubmissionDateHeader, userConf.SubmissionDate})
	return data
}
