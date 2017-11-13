package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/valyala/fasthttp"
	"time"
)

const (
	CONTENTTYPE_JSON string = "application/json"
	METHOD_GET       string = "GET"
	METHOD_POST      string = "POST"
	METHOD_PUT       string = "PUT"
	METHOD_DELETE    string = "DELETE"
)

type HttpClient interface {
	Get(url string) (statusCode int, body []byte, err error)
	PostJSON(url string, reqJSON interface{}, respJSON interface{}) (int, error)
	PutJSON(url string, reqJSON interface{}, respJSON interface{}) (int, error)
	GetJSON(url string, respJSON interface{}) (int, error)
	DeleteJSON(url string, respJSON interface{}) (int, error)
}

type fastHttpClient struct {
	client *fasthttp.Client
}

func NewHttpClient(to ...time.Duration) HttpClient {
	timeout := time.Second * 5
	if len(to) > 0 {
		timeout = to[0]
	}
	return &fastHttpClient{
		client: &fasthttp.Client{
			ReadTimeout:  timeout,
			WriteTimeout: timeout,
		},
	}
}

var _ HttpClient = &fastHttpClient{}

func (http *fastHttpClient) Get(url string) (int, []byte, error) {
	status, body, err := http.client.Get(nil, url)
	return status, body, err
}

// JSON request helpers: reqJSON can be either object, string or []byte, respJSON should always be object, *[]byte or nil
func (http *fastHttpClient) PostJSON(url string, reqJSON interface{}, respJSON interface{}) (int, error) {
	status, err := http.doJSON(METHOD_POST, url, reqJSON, respJSON)
	return status, err
}

func (http *fastHttpClient) PutJSON(url string, reqJSON interface{}, respJSON interface{}) (int, error) {
	status, err := http.doJSON(METHOD_PUT, url, reqJSON, respJSON)
	return status, err
}

func (http *fastHttpClient) GetJSON(url string, respJSON interface{}) (int, error) {
	status, err := http.doJSON(METHOD_GET, url, nil, respJSON)
	return status, err
}

func (http *fastHttpClient) DeleteJSON(url string, respJSON interface{}) (int, error) {
	status, err := http.doJSON(METHOD_DELETE, url, nil, respJSON)
	return status, err
}

func (http *fastHttpClient) doJSON(method string, url string, reqJSON interface{}, respJSON interface{}) (int, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.Header.SetMethod(method)
	req.SetRequestURI(url)
	if reqJSON != nil {
		if sval, ok := reqJSON.(string); ok {
			req.AppendBodyString(sval)
		} else if bval, ok := reqJSON.([]byte); ok {
			req.AppendBody(bval)
		} else {
			jval, err := json.Marshal(reqJSON)
			if err != nil {
				return -1, errors.New(fmt.Sprintf("json marshal: %s", err.Error()))
			}
			req.AppendBody(jval)
		}
		req.Header.SetContentType(CONTENTTYPE_JSON)
	}
	err := http.client.Do(req, resp)
	if err != nil {
		return -1, errors.New(fmt.Sprintf("httpcli do: %s", err.Error()))
	}
	status := resp.StatusCode()
	if status >= 400 {
		return status, errors.New(string(resp.Body()))
	}
	if respJSON != nil {
		if bref, ok := respJSON.(*[]byte); ok {
			b := resp.Body()
			data := make([]byte, len(b))
			copy(data, b)
			*bref = data
		} else {
			err = json.Unmarshal(resp.Body(), respJSON)
			if err != nil {
				return status, err
			}
		}
	}
	return resp.StatusCode(), nil
}
