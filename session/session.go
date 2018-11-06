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

type Headers map[string]string

type Cookies map[string]string

type RequestWrapper struct {
	Request *http.Request
}

type ResponseWrapper struct {
	Response *http.Response
}

func newSession() *csession.Session {
	jar, err := cookiejar.New(nil)
	if err != nil {
		jar = nil
	}
	dontFollowRedirects := func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	session := csession.NewSession(http.DefaultTransport, dontFollowRedirects, jar)
	session.HeadersFunc = func(req *http.Request) {
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

var client = newSession()

func setHeaders(request *http.Request, headers Headers) {
	if headers != nil {
		for name, value := range headers {
			request.Header.Set(name, value)
		}
	}
}

func Post(url string, body io.Reader, headers Headers) *RequestWrapper {
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	setHeaders(req, headers)
	return &RequestWrapper{req}
}

func Get(url string, headers Headers) *RequestWrapper {
	req, _ := http.NewRequest("GET", url, nil)
	setHeaders(req, headers)
	return &RequestWrapper{req}
}

func Form(body url.Values) io.Reader {
	return strings.NewReader(body.Encode())
}

func Body(body string) io.Reader {
	return strings.NewReader(body)
}

func (r *RequestWrapper) Send() (*http.Response, error) {
	return client.Do(r.Request)
}

func (r *RequestWrapper) SafeSend() *ResponseWrapper {
	resp, err := r.Send()
	if err != nil {
		log.Errorf("Error occurred while sending request. Try again\n%s", err)
		return r.SafeSend()
	}
	log.DebugHttp(httputil.DumpRequest(r.Request, true))
	log.DebugHttp(httputil.DumpResponse(resp, true))
	return &ResponseWrapper{resp}
}

func (r *ResponseWrapper) AsString() string {
	return string(r.AsBytes())
}

func (r *ResponseWrapper) AsBytes() []byte {
	defer r.Response.Body.Close()
	resp, _ := ioutil.ReadAll(r.Response.Body)
	return resp
}
