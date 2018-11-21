package session

import (
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"

	"github.com/dyrkin/rezerwacje-duw-go/log"

	"github.com/lunny/csession"
)

//Response wraps http.Response to add new functionality
type Response struct {
	*http.Response
}

//Session the same as csession.Session, but I wan't to add additional functionality to it
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

//Send simply sends http request
func (s *Session) Send(request *http.Request) (*http.Response, error) {
	return s.Do(request)
}

//SafeSend safely sends http request. In case of error it tries again
func (s *Session) SafeSend(request *http.Request) *Response {
	debugHTTP("Sending request:\n%s\n", request)
	response, err := s.Send(request)
	if err != nil {
		log.Errorf("Error occurred while sending request. Try again\n%s", err)
		return s.SafeSend(request)
	}
	debugHTTP("Received response:\n%s\n", response)
	return &Response{response}
}

//AsString converts http response to string
func (r *Response) AsString() string {
	return string(r.AsBytes())
}

//AsBytes converts http response to byte array
func (r *Response) AsBytes() []byte {
	defer r.Response.Body.Close()
	resp, _ := ioutil.ReadAll(r.Response.Body)
	return resp
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
		log.Debugf(format, string(bytes))
	} else {
		log.Errorf(format, err)
	}
}
