package session

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

//Headers kay/value storage
type Headers map[string]string

//Cookies kay/value storage
type Cookies map[string]string

//PostRequestBuilder is a post request builder interface
type PostRequestBuilder interface {
	Headers(headers Headers) PostRequestBuilder
	Cookies(cookies Cookies) PostRequestBuilder
	Form(body url.Values) PostRequestBuilder
	Body(body string) PostRequestBuilder
	Make() *http.Request
}

//PostRequest holds collected post request specific data
type PostRequest struct {
	request func(body io.Reader, headers Headers, cookies Cookies) *http.Request
	body    io.Reader
	headers Headers
	cookies Cookies
}

//GetRequestBuilder is a get request builder interface
type GetRequestBuilder interface {
	Headers(headers Headers) GetRequestBuilder
	Cookies(cookies Cookies) GetRequestBuilder
	Make() *http.Request
}

//GetRequest holds collected get request specific data
type GetRequest struct {
	request func(headers Headers, cookies Cookies) *http.Request
	headers Headers
	cookies Cookies
}

//Get creates get request
func Get(url string) GetRequestBuilder {
	pr := &GetRequest{}
	pr.request = func(headers Headers, cookies Cookies) *http.Request {
		req, _ := http.NewRequest("GET", url, nil)
		setHeaders(req, headers)
		setCookies(req, cookies)
		return req
	}
	return pr
}

//Post creates post request
func Post(url string) PostRequestBuilder {
	pr := &PostRequest{}
	pr.request = func(body io.Reader, headers Headers, cookies Cookies) *http.Request {
		req, _ := http.NewRequest("POST", url, body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		setHeaders(req, headers)
		setCookies(req, cookies)
		return req
	}
	return pr
}

//Headers adds headers to post request
func (pr *PostRequest) Headers(headers Headers) PostRequestBuilder {
	pr.headers = headers
	return pr
}

//Headers adds headers to get request
func (pr *GetRequest) Headers(headers Headers) GetRequestBuilder {
	pr.headers = headers
	return pr
}

//Cookies adds cookies to post request
func (pr *PostRequest) Cookies(cookies Cookies) PostRequestBuilder {
	pr.cookies = cookies
	return pr
}

//Cookies adds cookies to get request
func (pr *GetRequest) Cookies(cookies Cookies) GetRequestBuilder {
	pr.cookies = cookies
	return pr
}

//Form represents key/value post request body
func (pr *PostRequest) Form(body url.Values) PostRequestBuilder {
	pr.body = strings.NewReader(body.Encode())
	return pr
}

//Body represents simple string post request body
func (pr *PostRequest) Body(body string) PostRequestBuilder {
	pr.body = strings.NewReader(body)
	return pr
}

//Make builds http.Request from PartialRequest
func (pr *PostRequest) Make() *http.Request {
	return pr.request(pr.body, pr.headers, pr.cookies)
}

//Make builds http.Request from PartialRequest
func (pr *GetRequest) Make() *http.Request {
	return pr.request(pr.headers, pr.cookies)
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
