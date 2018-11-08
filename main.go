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

var dateEventsRegex = regexp.MustCompile("var dateEvents\\s+=\\s+(?P<Events>.*?);")
var slotsRegex = regexp.MustCompile("lock\\(.*?>([\\d:]+)<\\/a>")

var mutex = &sync.Mutex{}

var client = session.New()

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
	acceptTermsRequest := session.Get(url).Make()
	client.SafeSend(acceptTermsRequest)
}

func latestDate(city *config.City) string {
	acceptTerms(city)
	url := fmt.Sprintf("http://rezerwacje.duw.pl/reservations/pol/queues/%s/%s", city.Queue, city.Id)
	cityRequest := session.Get(url).Make()
	cityHTML := client.SafeSend(cityRequest).AsString()
	return extractLatestDate(cityHTML)
}

func terms(city *config.City, date string) []string {
	url := fmt.Sprintf("http://rezerwacje.duw.pl/reservations/pol/queues/%s/%s/%s", city.Queue, city.Id, date)
	headers := session.Headers{"X-Requested-With": "XMLHttpRequest"}
	termsRequest := session.Get(url).Headers(headers).Make()
	termsHTML := client.SafeSend(termsRequest).AsString()
	terms := extractTerms(termsHTML)
	log.Infof("Available terms for city %q: %q", city.Name, terms)
	return terms
}

func recognizeCaptcha() string {
	captchaRequest := session.Get("http://rezerwacje.duw.pl/reservations/captcha").Make()
	captchaImage := client.SafeSend(captchaRequest).AsBytes()
	return captcha.RecognizeCaptcha(&captchaImage)
}

func checkCaptcha(captcha string) bool {
	body := url.Values{"code": {captcha}}
	checkCaptchaRequest := session.Post("http://rezerwacje.duw.pl/reservations/captcha/check").Form(body).Make()
	result := client.SafeSend(checkCaptchaRequest).AsString()
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
	postUserDataRequest := session.Post(url).Body(body).Headers(headers).Make()
	client.SafeSend(postUserDataRequest)
}

func confirmTerm(city *config.City, slot string) {
	url := fmt.Sprintf("http://rezerwacje.duw.pl/reservations/reservations/reserv/%s/%s", slot, city.Id)
	confirmTermRequest := session.Get(url).Make()
	client.SafeSend(confirmTermRequest)
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
			lockRequest := session.Post("http://rezerwacje.duw.pl/reservations/reservations/lock").Form(body).Make()
			lockResult <- client.SafeSend(lockRequest).AsString()
		}()
	}
	return <-lockResult
}

func lock(city *config.City, time string) (slot string, locked bool) {
	mutex.Lock()
	log.Infof("Locking term %s for city %q", time, city.Name)
	lockResult := tryLock(city, time)
	if strings.HasPrefix(lockResult, "OK") {
		slot := lockResult[3:]
		log.Infof("Term is locked. City %q, time %q, slot %q", city.Name, time, slot)
		return slot, true
	}
	log.Infof("Unable to lock term %q for city %q. Reason %q", time, city.Name, lockResult)
	mutex.Unlock()
	return "", false
}

func makeReservation(city *config.City, date string, term string) {
	time := fmt.Sprintf("%s %s:00", date, term)
	if slot, ok := lock(city, time); ok {
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
	body := url.Values{"data[User][email]": {config.UserConf().Username}, "data[User][password]": {config.UserConf().Password}}
	loginRequest := session.Post("http://rezerwacje.duw.pl/reservations/pol/login").Form(body).Make()
	loginResponse := client.SafeSend(loginRequest)
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

func collectActiveCities() map[*config.City]string {
	citiesToProcess := map[*config.City]string{}
	for _, city := range config.ApplicationConf().Cities {
		cityDate := latestDate(city)
		if validDate(cityDate) {
			log.Infof("Going to process city %q for date %q", city.Name, cityDate)
			citiesToProcess[city] = cityDate
		} else {
			log.Infof("Ignoring city %q because there is weekend", city.Name)
		}
	}
	return citiesToProcess
}

func main() {
	if login() {
		activeCities := collectActiveCities()
		for i := 0; i < config.ApplicationConf().ParallelismFactor; i++ {
			for city, date := range activeCities {
				go processCity(*city, date)
			}
		}
	}
	await()
}
