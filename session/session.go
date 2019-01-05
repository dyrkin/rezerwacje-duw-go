package session

import (
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"time"

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
	transport := http.DefaultTransport.(*http.Transport)
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	transport.MaxIdleConnsPerHost = 30
	transport.TLSHandshakeTimeout = 10 * time.Second
	session := &Session{}
	session.Session = csession.NewSession(transport, dontFollowRedirects, jar)
	session.Session.HeadersFunc = func(req *http.Request) {
		csession.DefaultHeadersFunc(req)
		userAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36"
		encoding := "gzip, deflate"
		acceptLanguage := "ru,en-US;q=0.9,en;q=0.8"
		lang := "pol"
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Accept-Encoding", encoding)
		req.Header.Set("Accept-Language", acceptLanguage)
		req.AddCookie(&http.Cookie{Name: "config[lang]", Value: lang})
	}
	return session
}

//Send simply sends http request
func (s *Session) Send(request *http.Request) (*http.Response, error) {
	debugHTTP("Sending request:\n%s\n", request)
	resp, err := s.Do(request)
	if err == nil {
		debugHTTP("Received response:\n%s\n", resp)
	} else {
		log.Errorf("Received error:\n%s", err)
	}
	return resp, err
}

//SafeSend safely sends http request. In case of error it tries again
func (s *Session) SafeSend(requestBuilder Builder) *Response {
	response, err := s.Send(requestBuilder.Build())
	if err != nil {
		log.Errorf("Error occurred while sending request. Try again\n%s", err)
		return s.SafeSend(requestBuilder)
	}
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

func (r *Response) Drain() *Response {
	defer r.Response.Body.Close()
	io.Copy(ioutil.Discard, r.Response.Body)
	return r
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
