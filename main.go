package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"rezerwacje-duw-go/captcha"
	"rezerwacje-duw-go/config"
	"rezerwacje-duw-go/log"
	"rezerwacje-duw-go/session"
)

var userConf = config.UserConf
var applicationConf = config.ApplicationConf

var dateEventsRegex = regexp.MustCompile("var dateEvents\\s+=\\s+(?P<Events>.*?);")
var slotsRegex = regexp.MustCompile("lock\\(.*?>([\\d:]+)<\\/a>")

var mutex = &sync.Mutex{}

func extractLatestDate(cityHTML string) string {
	groups := dateEventsRegex.FindStringSubmatch(cityHTML)
	data := []byte(groups[1])
	var values []map[string]string
	json.Unmarshal(data, &values)
	return values[len(values)-1]["date"]
}

func extractTerms(termsHTML string) []string {
	groups := slotsRegex.FindAllStringSubmatch(termsHTML, -1)
	terms := []string{}
	for _, group := range groups {
		terms = append(terms, group[1])
	}
	return terms
}

func acceptTerms(city *config.City) {
	url := fmt.Sprintf("http://rezerwacje.duw.pl/reservations/opmenus/terms/%s/%s?accepted=true", city.Queue, city.Id)
	acceptTermsRequest := session.Get(url, nil)
	acceptTermsRequest.SafeSend()
}

func latestDate(city config.City) string {
	acceptTerms(&city)
	url := fmt.Sprintf("http://rezerwacje.duw.pl/reservations/pol/queues/%s/%s", city.Queue, city.Id)
	cityRequest := session.Get(url, nil)
	cityHTML := cityRequest.SafeSend().AsString()
	return extractLatestDate(cityHTML)
}

func terms(city *config.City, date string) []string {
	url := fmt.Sprintf("http://rezerwacje.duw.pl/reservations/pol/queues/%s/%s/%s", city.Queue, city.Id, date)
	headers := session.Headers{"X-Requested-With": "XMLHttpRequest"}
	termsRequest := session.Get(url, headers)
	termsHTML := termsRequest.SafeSend().AsString()
	terms := extractTerms(termsHTML)
	log.Infof("Available terms for city %q: %q", city.Name, terms)
	return terms
}

func recognizeCaptcha() string {
	captchaRequest := session.Get("http://rezerwacje.duw.pl/reservations/captcha", nil)
	captchaImage := captchaRequest.SafeSend().AsBytes()
	return captcha.RecognizeCaptcha(&captchaImage)
}

func checkCaptcha(captcha string) bool {
	body := url.Values{"code": {captcha}}
	checkCaptchaRequest := session.Post("http://rezerwacje.duw.pl/reservations/captcha/check", session.Form(body), nil)
	result := checkCaptchaRequest.SafeSend().AsString()
	return result == "true"
}

func renderUserDataToJSON() string {
	userData := config.CollectUserData()
	jsonBytes, _ := json.Marshal(userData)
	return string(jsonBytes)
}

func postUserData(city *config.City, slot string) {
	body := renderUserDataToJSON()
	url := fmt.Sprintf("http://rezerwacje.duw.pl/reservations/reservations/updateFormData/%s/%s", slot, city.Id)
	headers := session.Headers{"Content-Type": "application/json; charset=utf-8"}
	postUserDataRequest := session.Post(url, session.Body(body), headers)
	postUserDataRequest.SafeSend()
}

func confirmTerm(city *config.City, slot string) {
	url := fmt.Sprintf("http://rezerwacje.duw.pl/reservations/reservations/reserv/%s/%s", slot, city.Id)
	confirmTermRequest := session.Get(url, nil)
	confirmTermRequest.SafeSend()
}

func reserve(city *config.City, time string, slot string) {
	log.Infof("Attempt to make reservation for %q, slot %q and time %q", city.Name, slot, time)
	recognizedCaptcha := recognizeCaptcha()
	log.Infof("Captcha value is %q", recognizedCaptcha)
	if checkCaptcha(recognizedCaptcha) {
		log.Infof("Captcha submitted successfully. Making reservation for city %q, slot %q and time %q", city.Name, slot, time)
		postUserData(city, slot)
		log.Infof("User data posted for city %q, slot %q and time %q", city.Name, slot, time)
		confirmTerm(city, slot)
		log.Infof("Reservation completed for city %q, slot %q and time %q. Check your email or DUW site", city.Name, slot, time)
	} else {
		mutex.Unlock()
	}
}

func tryLock(city *config.City, time string) string {
	lockResult := make(chan string)
	for i := 0; i < 5; i++ {
		go func() {
			body := url.Values{"time": {time}, "queue": {city.Queue}}
			lockRequest := session.Post("http://rezerwacje.duw.pl/reservations/reservations/lock", session.Form(body), nil)
			lockResult <- lockRequest.SafeSend().AsString()
		}()
	}
	return <-lockResult
}

func lock(city *config.City, time string) string {
	mutex.Lock()
	log.Infof("Locking term %s for city %q", time, city.Name)
	lockResult := tryLock(city, time)
	if strings.HasPrefix(lockResult, "OK") {
		slot := lockResult[3:]
		log.Infof("Term is locked. City %q, time %q, slot %q", city.Name, time, slot)
		return slot
	}
	log.Infof("Unable to lock term %q for city %q. Reason %q", time, city.Name, lockResult)
	mutex.Unlock()
	return ""
}

func makeReservation(city *config.City, date string, term string) {
	time := fmt.Sprintf("%s %s:00", date, term)
	slot := lock(city, time)
	if slot != "" {
		reserve(city, time, slot)
	}
}

func processCity(city config.City, date string) {
	log.Infof("Scanning terms for city %q and date %q", city.Name, date)
	terms := terms(&city, date)
	for _, term := range terms {
		makeReservation(&city, date, term)
	}
	go processCity(city, date)
}

func login() bool {
	body := url.Values{"data[User][email]": {userConf.Username}, "data[User][password]": {userConf.Password}}
	loginRequest := session.Post("http://rezerwacje.duw.pl/reservations/pol/login", session.Form(body), nil)
	loginResponse := loginRequest.SafeSend()
	return loginResponse.Response.StatusCode != 200
}

func parseDate(dateStr string) time.Time {
	layout := "2006-01-02"
	date, _ := time.Parse(layout, dateStr)
	return date
}

func validDate(cityDate string) bool {
	date := parseDate(cityDate)
	dayOfWeek := date.Weekday()
	return (dayOfWeek != time.Saturday) && (dayOfWeek != time.Sunday)
}

func await() {
	var input string
	fmt.Scanln(&input)
}

func main() {
	if login() {
		for _, city := range applicationConf.Cities {
			cityDate := latestDate(city)
			if validDate(cityDate) {
				for i := 0; i < applicationConf.ParallelismFactor; i++ {
					go processCity(city, cityDate)
				}
			} else {
				log.Infof("Ignoring city %q because there is weekend", city.Name)
			}
		}
	}
	await()
}
