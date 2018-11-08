package session

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"rezerwacje-duw-go/log"
	"strings"

	"github.com/lunny/csession"
)

//Headers kay/value storage
type Headers map[string]string

//Cookies kay/value storage
type Cookies map[string]string

//RequestWrapper wraps http.Request to add new functionality
type PartialRequest struct {
	request func(body io.Reader, headers Headers, cookies Cookies) *http.Request
	headers Headers
	cookies Cookies
	body    io.Reader
}

//ResponseWrapper wraps http.Response to add new functionality
type ResponseWrapper struct {
	Response *http.Response
}

type Session struct {
	*csession.Session
}

//New creates new session
func New() *Session {
	jar, err := cookiejar.New(nil)
	if err != nil {
		jar = nil
	}
	dontFollowRedirects := func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	session := &Session{}
	session.Session = csession.NewSession(http.DefaultTransport, dontFollowRedirects, jar)
	session.Session.HeadersFunc = func(req *http.Request) {
		csession.DefaultHeadersFunc(req)
		userAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36"
		encoding := "gzip, deflate"
		acceptLanguage := "ru,en-US;q=0.9,en;q=0.8"
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Accept-Encoding", encoding)
		req.Header.Set("Accept-Language", acceptLanguage)
		lang := "pol"
		req.AddCookie(&http.Cookie{Name: "config[lang]", Value: lang})
	}
	return session
}

func setHeaders(request *http.Request, headers Headers) {
	if headers != nil {
		for name, value := range headers {
			request.Header.Set(name, value)
		}
	}
}

func setCookies(request *http.Request, cookies Cookies) {
	if cookies != nil {
		for name, value := range cookies {
			request.AddCookie(&http.Cookie{Name: name, Value: value})
		}
	}
}

//Post creates post request
func Post(url string) *PartialRequest {
	pr := &PartialRequest{}
	pr.request = func(body io.Reader, headers Headers, cookies Cookies) *http.Request {
		req, _ := http.NewRequest("POST", url, body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		setHeaders(req, headers)
		setCookies(req, cookies)
		return req
	}
	return pr
}

//Headers adds headers to request
func (pr *PartialRequest) Headers(headers Headers) *PartialRequest {
	pr.headers = headers
	return pr
}

//Cookies adds cookies to request
func (pr *PartialRequest) Cookies(cookies Cookies) *PartialRequest {
	pr.cookies = cookies
	return pr
}

//Make builds http.Request from PartialRequest
func (pr *PartialRequest) Make() *http.Request {
	return pr.request(pr.body, pr.headers, pr.cookies)
}

//Get creates get request
func Get(url string) *PartialRequest {
	pr := &PartialRequest{}
	pr.request = func(body io.Reader, headers Headers, cookies Cookies) *http.Request {
		req, _ := http.NewRequest("GET", url, nil)
		setHeaders(req, headers)
		setCookies(req, cookies)
		return req
	}
	return pr
}

//Form represents key/value post request body
func (pr *PartialRequest) Form(body url.Values) *PartialRequest {
	pr.body = strings.NewReader(body.Encode())
	return pr
}

//Body represents simple string post request body
func (pr *PartialRequest) Body(body string) *PartialRequest {
	pr.body = strings.NewReader(body)
	return pr
}

//Send simply sends http request
func (s *Session) Send(request *http.Request) (*http.Response, error) {
	return s.Do(request)
}

//SafeSend safely sends http request. In case of error it tries again
func (s *Session) SafeSend(request *http.Request) *ResponseWrapper {
	response, err := s.Send(request)
	if err != nil {
		log.Errorf("Error occurred while sending request. Try again\n%s", err)
		return s.SafeSend(request)
	}
	debugHTTP("Sending request:\n%s", request)
	debugHTTP("Received response:\n%s", response)
	return &ResponseWrapper{response}
}

func debugHTTP(format string, r interface{}) {
	var bytes []byte
	var err error
	switch r := r.(type) {
	case *http.Request:
		bytes, err = httputil.DumpRequest(r, true)
	case *http.Response:
		bytes, err = httputil.DumpResponse(r, true)
	}
	if err == nil {
		log.Debugf(format, bytes)
	} else {
		log.Errorf(format, err)
	}
}

//AsString converts http response to string
func (r *ResponseWrapper) AsString() string {
	return string(r.AsBytes())
}

//AsBytes converts http response to byte array
func (r *ResponseWrapper) AsBytes() []byte {
	defer r.Response.Body.Close()
	resp, _ := ioutil.ReadAll(r.Response.Body)
	return resp
}
