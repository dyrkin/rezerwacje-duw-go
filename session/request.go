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

//PostRequest wraps http.Request to add new functionality
type PostRequest struct {
	request func(body io.Reader, headers Headers, cookies Cookies) *http.Request
	body    io.Reader
	headers Headers
	cookies Cookies
}

//GetRequest wraps http.Request to add new functionality
type GetRequest struct {
	request func(headers Headers, cookies Cookies) *http.Request
	headers Headers
	cookies Cookies
}

//Get creates get request
func Get(url string) *GetRequest {
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
func Post(url string) *PostRequest {
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
func (pr *PostRequest) Headers(headers Headers) *PostRequest {
	pr.headers = headers
	return pr
}

//Headers adds headers to get request
func (pr *GetRequest) Headers(headers Headers) *GetRequest {
	pr.headers = headers
	return pr
}

//Cookies adds cookies to post request
func (pr *PostRequest) Cookies(cookies Cookies) *PostRequest {
	pr.cookies = cookies
	return pr
}

//Cookies adds cookies to get request
func (pr *GetRequest) Cookies(cookies Cookies) *GetRequest {
	pr.cookies = cookies
	return pr
}

//Form represents key/value post request body
func (pr *PostRequest) Form(body url.Values) *PostRequest {
	pr.body = strings.NewReader(body.Encode())
	return pr
}

//Body represents simple string post request body
func (pr *PostRequest) Body(body string) *PostRequest {
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
