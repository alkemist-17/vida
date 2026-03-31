package vida

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var defaultClient *HttpClient
var validMethods map[string]bool

type HttpClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func loadFoundationHttpClient() Value {
	defaultClient = &HttpClient{HTTPClient: &http.Client{}}
	validMethods = map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"PATCH":   true,
		"HEAD":    true,
		"OPTIONS": true,
	}
	m := &Object{Value: make(map[string]Value)}
	m.Value["request"] = GFn(httpRequest)
	return m
}

func httpRequest(args ...Value) (Value, error) {
	l := len(args)
	if l == 1 {
		if url, ok := args[0].(*String); ok {
			options := httpCreateDefaultRequestOptions(url)
			response, err := defaultClient.RequestWithDefaultOptions(options)
			if err != nil {
				return &Error{Message: &String{Value: err.Error()}}, nil
			}
			return response, nil
		}
	}
	return NilValue, nil
}

func httpCreateResponseObject(statusCode Integer, headers *Object, body []byte) *Object {
	response := &Object{Value: make(map[string]Value)}
	response.Value["statusCode"] = statusCode
	response.Value["headers"] = headers
	response.Value["body"] = &Bytes{Value: body}
	return response
}

func httpCreateDefaultRequestOptions(url *String) *Object {
	reqOps := &Object{Value: make(map[string]Value)}
	reqOps.Value["method"] = &String{Value: "GET"}
	reqOps.Value["url"] = url
	reqOps.Value["baseURL"] = &String{Value: ""}
	reqOps.Value["params"] = &Object{Value: make(map[string]Value)}
	reqOps.Value["body"] = &Object{Value: make(map[string]Value)}
	reqOps.Value["headers"] = &Object{Value: make(map[string]Value)}
	reqOps.Value["timeout"] = Integer(5000)
	reqOps.Value["auth"] = &Object{}
	reqOps.Value["responseType"] = &String{Value: "json"}
	reqOps.Value["responseEncoding"] = &String{Value: "utf8"}
	reqOps.Value["maxRedirects"] = Integer(21)
	reqOps.Value["maxContentLength"] = Integer(192 * 1024)
	reqOps.Value["maxBodyLength"] = Integer(192 * 1024)
	reqOps.Value["decompress"] = Bool(true)
	return reqOps
}

func (c *HttpClient) RequestWithDefaultOptions(options *Object) (Value, error) {
	upperMethod := strings.ToUpper(options.Value["method"].(*String).Value)
	if !validMethods[upperMethod] {
		return nil, fmt.Errorf("invalid HTTP method: %q", options.Value["method"])
	}
	var fullURL string
	if c.BaseURL != "" {
		var err error
		fullURL, err = url.JoinPath(c.BaseURL, options.Value["url"].(*String).Value)
		if err != nil {
			return nil, err
		}
	} else if options.Value["baseURL"].(*String).Value != "" {
		var err error
		fullURL, err = url.JoinPath(options.Value["baseURL"].(*String).Value, options.Value["url"].(*String).Value)
		if err != nil {
			return nil, err
		}
	} else {
		fullURL = options.Value["url"].(*String).Value
	}

	if len(options.Value["params"].(*Object).Value) > 0 {
		parsedURL, err := url.Parse(fullURL)
		if err != nil {
			return nil, err
		}
		q := parsedURL.Query()
		for k, v := range options.Value["params"].(*Object).Value {
			if paramValue, ok := v.(*String); ok {
				q.Add(k, paramValue.Value)
			}
		}
		parsedURL.RawQuery = q.Encode()
		fullURL = parsedURL.String()
	}

	var bodyReader io.Reader
	var bodyLength int64

	switch v := options.Value["body"].(type) {
	case *String:
		bodyReader = strings.NewReader(v.Value)
		bodyLength = int64(len(v.Value))
	case *Bytes:
		bodyReader = bytes.NewReader(v.Value)
		bodyLength = int64(len(v.Value))
	default:
		jsonBody, err := json.Marshal(options.Value["body"])
		if err != nil {
			return NilValue, err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
		bodyLength = int64(len(jsonBody))
	}
	if options.Value["maxBodyLength"].(Integer) > 0 && Integer(bodyLength) > options.Value["maxBodyLength"].(Integer) {
		return NilValue, errors.New("request body length exceeded maxBodyLength")
	}

	req, err := http.NewRequest(options.Value["method"].(*String).Value, fullURL, bodyReader)

	if err != nil {
		return NilValue, err
	}

	headers, _ := options.Value["headers"].(*Object)
	if _, exists := headers.Value["Content-Type"]; !exists {
		headers.Value["Content-Type"] = &String{Value: "application/json"}
	}

	for key, v := range headers.Value {
		if str, ok := v.(*String); ok {
			req.Header.Set(key, str.Value)
		}
	}

	httpClient := &http.Client{
		Timeout: time.Duration(options.Value["timeout"].(Integer)) * time.Millisecond,
	}

	if options.Value["maxRedirects"].(Integer) > 0 {
		httpClient.CheckRedirect = func(_ *http.Request, via []*http.Request) error {
			if len(via) >= int(options.Value["maxRedirects"].(Integer)) {
				return fmt.Errorf("too many redirects (max: %d)", options.Value["maxRedirects"].(Integer))
			}
			return nil
		}
	}

	resp, err := httpClient.Do(req)

	if err != nil {
		return NilValue, err
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			if err != nil {
				err = fmt.Errorf("%w; failed to close response body: %v", err, cerr)
			} else {
				err = fmt.Errorf("failed to close response body: %v", cerr)
			}
		}
	}()

	var responseBody []byte
	limitedReader := io.LimitReader(resp.Body, int64(options.Value["maxContentLength"].(Integer))+1)
	responseBody, err = io.ReadAll(limitedReader)
	if err != nil {
		return NilValue, err
	}

	if Integer(len(responseBody)) > options.Value["maxContentLength"].(Integer) {
		return NilValue, errors.New("response content length exceeded maxContentLength")
	}

	return httpCreateResponseObject(Integer(resp.StatusCode), httpCreateObjectFromHeader(resp.Header), responseBody), err
}

func httpCreateObjectFromHeader(header http.Header) *Object {
	obj := &Object{Value: map[string]Value{}}
	for k, vals := range header {
		obj.Value[k] = &String{Value: vals[0]}
	}
	return obj
}
