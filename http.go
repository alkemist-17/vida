package vida

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	httpInvalidURLErr = "invalid url"
	httpNetworkErr    = "network"
	httpTimeoutErr    = "timeout"
	httpBodyReadErr   = "body read"
	httpLargeBodyErr  = "large body"
	httpMaxBodySize   = 10 << 20
)

func loadFoundationHttpClient() Value {
	m := &Object{Value: make(map[string]Value)}
	m.Value["request"] = GFn(httpRequest)
	m.Value["get"] = GFn(httpRequest)
	m.Value["header"] = GFn(responseHeaderHelper)
	m.Value["text"] = GFn(responseBodyToString)
	return m
}

type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
	URL        string
	Elapsed    time.Duration
}

func request(rawURL string) (*http.Response, []byte, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, nil, &RequestError{
			Kind:    httpInvalidURLErr,
			Message: fmt.Sprintf("failed to parse URL %q", rawURL),
			Err:     err,
		}
	}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
		rawURL = parsedURL.String()
	}

	resp, err := http.Get(rawURL)

	if err != nil {
		kind := httpNetworkErr
		if urlErr, ok := err.(*url.Error); ok {
			if urlErr.Timeout() {
				kind = httpTimeoutErr
			}
		}
		return nil, nil, &RequestError{
			Kind:    kind,
			Message: fmt.Sprintf("request to %q failed", rawURL),
			Err:     err,
		}
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, httpMaxBodySize+1))

	if err != nil {
		return nil, nil, &RequestError{
			Kind:    httpBodyReadErr,
			Message: "failed to read response body",
			Err:     err,
		}
	}

	if int64(len(body)) > httpMaxBodySize {
		return nil, nil, &RequestError{
			Kind:    httpLargeBodyErr,
			Message: fmt.Sprintf("response body exceeds %d bytes", httpMaxBodySize),
		}
	}

	headers := make(map[string]string, len(resp.Header))

	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return resp, body, nil
}

func httpRequest(args ...Value) (Value, error) {
	if len(args) == 1 {
		urlVal, ok := args[0].(*String)
		if !ok {
			return VidaError{Message: &String{Value: fmt.Sprintf("http.request: first argument must be string, got %T", args[0])}}, nil
		}
		resp, body, err := request(urlVal.Value)
		if err != nil {
			if reqErr, ok := err.(*RequestError); ok {
				errObj := &Object{
					Value: map[string]Value{
						"kind":    &String{Value: reqErr.Kind},
						"message": &String{Value: reqErr.Message},
						"url":     urlVal,
					},
				}
				return VidaError{Message: errObj}, nil
			}
			return VidaError{Message: &String{Value: err.Error()}}, nil
		}
		return responseToObject(resp, body), nil
	}
	return NilValue, nil
}

func responseToObject(resp *http.Response, body []byte) *Object {
	headersObj := &Object{Value: make(map[string]Value)}
	for k, v := range resp.Header {
		if len(v) > 0 {
			headersObj.Value[k] = &String{Value: v[0]}
		}
	}
	dl, _ := resp.Request.Context().Deadline()
	return &Object{
		Value: map[string]Value{
			"statusCode": Integer(resp.StatusCode),
			"headers":    headersObj,
			"body":       &Bytes{Value: body},
			"url":        &String{Value: resp.Request.URL.String()},
			"elapsed":    Integer(time.Since(dl).Milliseconds()),
		},
	}
}

func responseBodyToString(args ...Value) (Value, error) {
	if len(args) > 0 {
		if respObj, ok := args[0].(*Object); ok {
			if body, ok := respObj.Value["body"].(*Bytes); ok {
				return &String{Value: string(body.Value)}, nil
			}
		}
	}
	return NilValue, nil
}

func responseHeaderHelper(args ...Value) (Value, error) {
	if len(args) > 1 {
		respObj, okresp := args[0].(*Object)
		name, okname := args[1].(*String)
		if okresp && okname {
			if headers, ok := respObj.Value["headers"].(*Object); ok {
				for k, v := range headers.Value {
					if strings.EqualFold(k, name.Value) {
						return v, nil
					}
				}
			}
		}
	}
	return NilValue, nil
}

type RequestError struct {
	Kind    string
	Message string
	Err     error
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("[%s] %s: %v", e.Kind, e.Message, e.Err)
}
