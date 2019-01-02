package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dyrkin/rezerwacje-duw-go/queue"

	"github.com/dyrkin/rezerwacje-duw-go/captcha"
	"github.com/dyrkin/rezerwacje-duw-go/cmd"
	"github.com/dyrkin/rezerwacje-duw-go/config"
	"github.com/dyrkin/rezerwacje-duw-go/log"
	"github.com/dyrkin/rezerwacje-duw-go/session"
)

var dateEventsRegex = regexp.MustCompile("var dateEvents\\s+=\\s+(?P<Events>.*?);")
var slotsRegex = regexp.MustCompile("lock\\(.*?>([\\d:]+)<\\/a>")

var mutex = &sync.Mutex{}

var client = session.New()

var reservationQueue = queue.NewWithLimit(5)

func extractLatestDate(entityHTML string) string {
	groups := dateEventsRegex.FindStringSubmatch(entityHTML)
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

func acceptTerms(entity *config.Entity) {
	url := fmt.Sprintf("http://rezerwacje.duw.pl/reservations/opmenus/terms/%s/%s?accepted=true", entity.Queue, entity.ID)
	acceptTermsRequest := session.Get(url)
	client.SafeSend(acceptTermsRequest).DiscardBody()
}

func latestDate(entity *config.Entity) string {
	acceptTerms(entity)
	url := fmt.Sprintf("http://rezerwacje.duw.pl/reservations/pol/queues/%s/%s", entity.Queue, entity.ID)
	entityRequest := session.Get(url)
	entityHTML := client.SafeSend(entityRequest).AsString()
	return extractLatestDate(entityHTML)
}

func terms(entity *config.Entity, date string) []string {
	url := fmt.Sprintf("http://rezerwacje.duw.pl/reservations/pol/queues/%s/%s/%s", entity.Queue, entity.ID, date)
	headers := session.Headers{"X-Requested-With": "XMLHttpRequest"}
	termsRequest := session.Get(url).Headers(headers)
	termsHTML := client.SafeSend(termsRequest).AsString()
	terms := extractTerms(termsHTML)
	log.Infof("Available terms for %q: %q", entity.Name, terms)
	return terms
}

func recognizeCaptcha() string {
	captchaRequest := session.Get("http://rezerwacje.duw.pl/reservations/captcha")
	captchaImage := client.SafeSend(captchaRequest).AsBytes()
	return captcha.RecognizeCaptcha(&captchaImage)
}

func checkCaptcha(captcha string) bool {
	body := url.Values{"code": {captcha}}
	checkCaptchaRequest := session.Post("http://rezerwacje.duw.pl/reservations/captcha/check").Form(body)
	result := client.SafeSend(checkCaptchaRequest).AsString()
	return result == "true"
}

func renderUserDataToJSON(userData []*config.Row) string {
	jsonBytes, _ := json.Marshal(userData)
	return string(jsonBytes)
}

func postUserData(entity *config.Entity, slot string, userData *[]*config.Row) {
	body := renderUserDataToJSON(*userData)
	url := fmt.Sprintf("http://rezerwacje.duw.pl/reservations/reservations/updateFormData/%s/%s", slot, entity.ID)
	headers := session.Headers{"Content-Type": "application/json; charset=utf-8"}
	postUserDataRequest := session.Post(url).Body(body).Headers(headers)
	client.SafeSend(postUserDataRequest).DiscardBody()
}

func confirmTerm(entity *config.Entity, slot string) {
	url := fmt.Sprintf("http://rezerwacje.duw.pl/reservations/reservations/reserv/%s/%s", slot, entity.ID)
	confirmTermRequest := session.Get(url)
	client.SafeSend(confirmTermRequest).DiscardBody()
}

func reserve(entity *config.Entity, time string, slot string, userData *[]*config.Row) {
	log.Infof("Attempt to make reservation for %q, slot %q and time %q", entity.Name, slot, time)
	recognizedCaptcha := recognizeCaptcha()
	log.Infof("Captcha value is %q", recognizedCaptcha)
	if checkCaptcha(recognizedCaptcha) {
		log.Infof("Captcha submitted successfully. Making reservation for %q, slot %q and time %q", entity.Name, slot, time)
		postUserData(entity, slot, userData)
		log.Infof("User data posted for %q, slot %q and time %q", entity.Name, slot, time)
		confirmTerm(entity, slot)
		log.Infof("Reservation completed for %q, slot %q and time %q. Check your email or DUW site", entity.Name, slot, time)
	} else {
		mutex.Unlock()
	}
}

func tryLock(entity *config.Entity, time string) string {
	lockResult := make(chan string)
	for i := 0; i < 5; i++ {
		go func() {
			body := url.Values{"time": {time}, "queue": {entity.Queue}}
			lockRequest := session.Post("http://rezerwacje.duw.pl/reservations/reservations/lock").Form(body)
			lockResult <- client.SafeSend(lockRequest).AsString()
		}()
	}
	return <-lockResult
}

func lock(entity *config.Entity, time string) (slot string, locked bool) {
	mutex.Lock()
	log.Infof("Locking term %s for %q", time, entity.Name)
	lockResult := tryLock(entity, time)
	if strings.HasPrefix(lockResult, "OK") {
		slot := lockResult[3:]
		log.Infof("Term is locked. %q, time %q, slot %q", entity.Name, time, slot)
		return slot, true
	}
	log.Infof("Unable to lock term %q for %q. Reason %q", time, entity.Name, lockResult)
	mutex.Unlock()
	return "", false
}

func initQueueProcessor() {
	go func() {
		for {
			reservation := reservationQueue.Take()
			time := fmt.Sprintf("%s %s:00", reservation.Date, reservation.Term)
			if slot, ok := lock(reservation.Entity, time); ok {
				reserve(reservation.Entity, time, slot, reservation.UserData)
			}
		}
	}()
}

func scheduleReservation(entity *config.Entity, date string, term string, userData *[]*config.Row) {
	reservation := &queue.Reservation{Entity: entity, Date: date, Term: term, UserData: userData}
	reservationQueue.Push(reservation)
}

func process(entity config.Entity, date string, userData *[]*config.Row) {
	log.Infof("Scanning terms for %q and date %q", entity.Name, date)
	terms := terms(&entity, date)
	for _, term := range terms {
		scheduleReservation(&entity, date, term, userData)
	}
	go process(entity, date, userData)
}

func login() bool {
	body := url.Values{"data[User][email]": {config.UserConf().Login}, "data[User][password]": {config.UserConf().Password}}
	loginRequest := session.Post("http://rezerwacje.duw.pl/reservations/pol/login").Form(body)
	loginResponse := client.SafeSend(loginRequest).DiscardBody()
	return loginResponse.Response.StatusCode != 200
}

func parseDate(dateStr string) time.Time {
	layout := "2006-01-02"
	date, _ := time.Parse(layout, dateStr)
	return date
}

func convertDateToDayOfWeek(stringDate string) time.Weekday {
	date := parseDate(stringDate)
	return date.Weekday()
}

func validCityDate(date string) (time.Weekday, bool) {
	dayOfWeek := convertDateToDayOfWeek(date)
	return dayOfWeek, (dayOfWeek != time.Saturday) && (dayOfWeek != time.Sunday)
}

func validDepartmentDate(date string) (time.Weekday, bool) {
	dayOfWeek := convertDateToDayOfWeek(date)
	return dayOfWeek, (dayOfWeek == time.Tuesday) || (dayOfWeek == time.Thursday)
}

func findEntity(entities []*config.Entity, shortName string) (*config.Entity, bool) {
	for _, entity := range entities {
		if entity.ShortName == shortName {
			return entity, true
		}
	}
	return nil, false
}

func collectActiveEntities(entities []*config.Entity, validation func(date string) (time.Weekday, bool), failMessage string) map[*config.Entity]string {
	entitiesToProcess := map[*config.Entity]string{}
	for _, entity := range entities {
		entityDate := latestDate(entity)
		log.Infof("Validating current latest date %q for %q", entityDate, entity.Name)
		if weekday, ok := validation(entityDate); ok {
			log.Infof("Going to process %q for date %q", entity.Name, entityDate)
			entitiesToProcess[entity] = entityDate
		} else {
			log.Infof(failMessage, entityDate, weekday, entity.Name)
		}
	}
	return entitiesToProcess
}

func collectActiveDepartments(enabledDepartment string) map[*config.Entity]string {
	if department, ok := findEntity(config.ApplicationConf().Departments, enabledDepartment); ok {
		departments := []*config.Entity{department}
		return collectActiveEntities(departments, validDepartmentDate,
			"Date %q(%q) is wrong for %q because it is not Tuesday or Thursday")
	}
	panic(fmt.Sprintf("Unsupported department [%s]", enabledDepartment))
}

func collectActiveCities(enabledCities []string) map[*config.Entity]string {
	cities := []*config.Entity{}
	if enabledCities != nil {
		for _, enabledCity := range enabledCities {
			if city, ok := findEntity(config.ApplicationConf().Cities, enabledCity); ok {
				cities = append(cities, city)
			} else {
				panic(fmt.Sprintf("Unsupported city [%s]", enabledCity))
			}
		}
	} else {
		cities = config.ApplicationConf().Cities
	}
	return collectActiveEntities(cities, validCityDate,
		"Date %q(%q) is wrong for %q because it is weekend")
}

func await() {
	var input string
	fmt.Scanln(&input)
}

func processEntities(entities map[*config.Entity]string, userData []*config.Row) {
	for i := 0; i < config.ApplicationConf().ParallelismFactor; i++ {
		for entity, date := range entities {
			go process(*entity, date, &userData)
		}
	}
	await()
}

func processCommand(command string, args []string) {
	if command == cmd.HelpCommand {
		fmt.Println("Help")
		cmd.PrintHelp()
	} else {
		log.Infof("Logging in...")
		if login() {
			log.Infof("Successfully logged in")
			var userData []*config.Row
			var entities map[*config.Entity]string
			switch command {
			case cmd.ApplicationCommand:
				userData = config.CollectApplicationSubmissionData()
				enabledCities := args
				entities = collectActiveCities(enabledCities)
			case cmd.HeadofCommand:
				userData = config.CollectHeadOfDepartmentData()
				enabledDepartment := args[0]
				entities = collectActiveDepartments(enabledDepartment)
			}
			initQueueProcessor()
			processEntities(entities, userData)
		} else {
			log.Infoln("Invalid login or password")
		}
	}
}

func main() {
	command, args, err := cmd.ParseArgs()
	if err == nil {
		processCommand(command, args)
	} else {
		fmt.Printf("%s\n", err)
		cmd.PrintHelp()
	}
}
