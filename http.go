package vida

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	httpGET                = "GET"
	httpPOST               = "POST"
	httpPUT                = "PUT"
	httpDELETE             = "DELETE"
	httpPATCH              = "PATCH"
	httpHEAD               = "HEAD"
	httpOPTIONS            = "OPTIONS"
	httpInvalidURLErr      = "invalid-url"
	httpNetworkErr         = "network"
	httpTimeoutErr         = "timeout"
	httpTemporary          = "temporary"
	httpBodyReadErr        = "body-read"
	httpLargeBodyErr       = "body-size"
	httpInvalidReqErr      = "invalid-request"
	httpJsonEncodeErr      = "json-encode"
	httpDefaultSchema      = "https"
	httpMethodField        = "method"
	httpTimeoutField       = "timeout"
	httpHeadersField       = "headers"
	httpBodyField          = "body"
	httpURLField           = "url"
	httpStatusCodeField    = "statusCode"
	httpElapsedField       = "elapsed"
	httpContentTypeText    = "text/plain"
	httpContentTypeBinary  = "application/octet-stream"
	httpContentTypeAppJSON = "application/json"
	httpContentType        = "Content-Type"
	httpKind               = "kind"
	httpCauseMessage       = "cause"
	httpMaxBodySize        = 10 << 20
	httpDefaultTimeout     = 30 * time.Second
)

func loadFoundationHttpClient() Value {
	m := &Object{Value: make(map[string]Value, 9)}
	m.Value["request"] = GFn(httpRequest)
	m.Value["get"] = GFn(httpRequest)
	m.Value["post"] = GFn(httpPost)
	m.Value["put"] = GFn(httpPut)
	m.Value["delete"] = GFn(httpDelete)
	m.Value["patch"] = GFn(httpPatch)
	m.Value["head"] = GFn(httpHead)
	m.Value["options"] = GFn(httpOptions)
	m.Value["getHeader"] = GFn(httpResponseGetHeader)
	m.Value["statusText"] = GFn(httpStatusCodeText)
	return m
}

func httpRequest(args ...Value) (Value, error) {
	if len(args) > 0 {
		if rawURL, ok := args[0].(*String); ok {
			parsedURL, err := url.Parse(rawURL.Value)
			if err != nil {
				errObject := &Object{
					Value: map[string]Value{
						httpKind:         &String{Value: httpInvalidURLErr},
						httpCauseMessage: &String{Value: fmt.Sprintf("failed to parse URL %q: %v", rawURL, err)},
					},
				}
				return VidaError{Message: errObject}, nil
			}
			if parsedURL.Scheme == "" {
				parsedURL.Scheme = httpDefaultSchema
			}
			var options *requestOptions
			if len(args) > 1 {
				if optsObj, ok := args[1].(*Object); ok {
					options = httpParseOptions(optsObj)
				}
			}
			if options == nil {
				options = &requestOptions{
					Method:  httpGET,
					Timeout: httpDefaultTimeout,
					Headers: make(map[string]string),
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), options.Timeout)
			defer cancel()
			resp, body, elapsed, err := httpExecuteRequest(ctx, parsedURL.String(), options)
			if err != nil {
				if reqErr, ok := err.(*RequestError); ok {
					errObj := &Object{
						Value: map[string]Value{
							httpKind:         &String{Value: reqErr.Kind},
							httpCauseMessage: &String{Value: reqErr.Cause},
						},
					}
					return VidaError{Message: errObj}, nil
				}
				return VidaError{Message: &String{Value: err.Error()}}, nil
			}
			return httpResponseToObject(resp, body, elapsed), nil
		}
	}
	return NilValue, nil
}

func httpResponseToObject(resp *http.Response, body []byte, elapsed int64) *Object {
	headersObj := &Object{Value: make(map[string]Value)}
	for k, v := range resp.Header {
		if len(v) > 0 {
			headersObj.Value[k] = &String{Value: v[0]}
		}
	}
	return &Object{
		Value: map[string]Value{
			httpStatusCodeField: Integer(resp.StatusCode),
			httpHeadersField:    headersObj,
			httpBodyField:       &Bytes{Value: body},
			httpURLField:        &String{Value: resp.Request.URL.String()},
			httpElapsedField:    Integer(elapsed),
		},
	}
}

func httpResponseGetHeader(args ...Value) (Value, error) {
	if len(args) > 1 {
		respObj, okresp := args[0].(*Object)
		name, okname := args[1].(*String)
		if okresp && okname {
			if headers, ok := respObj.Value[httpHeadersField].(*Object); ok {
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
	Kind  string
	Cause string
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Kind, e.Cause)
}

type requestOptions struct {
	Method  string
	Headers map[string]string
	Body    Value
	Timeout time.Duration
}

func httpParseOptions(opts *Object) *requestOptions {
	options := &requestOptions{
		Method:  httpGET,
		Timeout: httpDefaultTimeout,
		Headers: make(map[string]string),
	}

	if methodVal, exists := opts.Value[httpMethodField]; exists {
		if methodStr, ok := methodVal.(*String); ok {
			options.Method = strings.ToUpper(methodStr.Value)
		}
	}

	if timeoutVal, exists := opts.Value[httpTimeoutField]; exists {
		if timeoutInt, ok := timeoutVal.(Integer); ok {
			options.Timeout = time.Duration(timeoutInt) * time.Millisecond
		}
	}

	if headersVal, exists := opts.Value[httpHeadersField]; exists {
		if headersObj, ok := headersVal.(*Object); ok {
			for k, v := range headersObj.Value {
				if strVal, ok := v.(*String); ok {
					options.Headers[k] = strVal.Value
				}
			}
		}
	}

	if bodyVal, exists := opts.Value[httpBodyField]; exists {
		switch v := bodyVal.(type) {
		case *String:
			options.Body = v
		case *Bytes:
			options.Body = v
		case *Object:
			options.Body = v
		default:
			options.Body = NilValue
		}
	}

	return options
}

func httpExecuteRequest(ctx context.Context, rawURL string, opts *requestOptions) (*http.Response, []byte, int64, error) {
	var bodyReader io.Reader
	var contentType string

	if opts.Body != nil {
		switch v := opts.Body.(type) {
		case *String:
			bodyReader = strings.NewReader(v.Value)
			contentType = httpContentTypeText
		case *Bytes:
			bodyReader = bytes.NewReader(v.Value)
			contentType = httpContentTypeBinary
		case *Object:
			jsonBody, err := json.Marshal(v)
			if err != nil {
				return nil, nil, 0, &RequestError{Kind: httpJsonEncodeErr, Cause: fmt.Sprintf("failed to json encode body: %v", err)}
			}
			bodyReader = bytes.NewBuffer(jsonBody)
			contentType = httpContentTypeAppJSON
		}
	}

	req, err := http.NewRequestWithContext(ctx, opts.Method, rawURL, bodyReader)

	if err != nil {
		return nil, nil, 0, &RequestError{Kind: httpInvalidReqErr, Cause: fmt.Sprintf("failed to create request: %v", err)}
	}

	if opts.Body != nil {
		if _, exists := opts.Headers[httpContentType]; !exists && contentType != "" {
			opts.Headers[httpContentType] = contentType
		}
	}

	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		kind := httpNetworkErr
		if urlErr, ok := err.(*url.Error); ok {
			if urlErr.Timeout() {
				kind = httpTimeoutErr
			} else if urlErr.Temporary() {
				kind = httpTemporary
			}
		}
		return nil, nil, 0, &RequestError{Kind: kind, Cause: fmt.Sprintf("request to %q failed: %v", rawURL, err)}
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, httpMaxBodySize+1))

	if err != nil {
		return nil, nil, 0, &RequestError{Kind: httpBodyReadErr, Cause: fmt.Sprintf("failed to read response: %v", err)}
	}

	if int64(len(body)) > httpMaxBodySize {
		return nil, nil, 0, &RequestError{Kind: httpLargeBodyErr, Cause: fmt.Sprintf("response exceeds %d bytes", httpMaxBodySize)}
	}

	return resp, body, elapsed, nil
}

func httpRequestWithMethod(method string, args ...Value) (Value, error) {
	switch len(args) {
	case 1:
		optsObj := &Object{
			Value: map[string]Value{
				httpMethodField: &String{Value: method},
			},
		}
		return httpRequest(args[0], optsObj)
	case 2:
		if optsObj, ok := args[1].(*Object); ok {
			optsObj.Value[httpMethodField] = &String{Value: method}
			return httpRequest(args[0], optsObj)
		}
		return httpRequest(args[0])
	}
	return NilValue, nil
}

func httpPost(args ...Value) (Value, error) {
	return httpRequestWithMethod(httpPOST, args...)
}

func httpPut(args ...Value) (Value, error) {
	return httpRequestWithMethod(httpPUT, args...)
}

func httpDelete(args ...Value) (Value, error) {
	return httpRequestWithMethod(httpDELETE, args...)
}

func httpPatch(args ...Value) (Value, error) {
	return httpRequestWithMethod(httpPATCH, args...)
}

func httpHead(args ...Value) (Value, error) {
	return httpRequestWithMethod(httpHEAD, args...)
}

func httpOptions(args ...Value) (Value, error) {
	return httpRequestWithMethod(httpOPTIONS, args...)
}

func httpStatusCodeText(args ...Value) (Value, error) {
	if len(args) > 0 {
		if code, ok := args[0].(Integer); ok {
			return &String{Value: http.StatusText(int(code))}, nil
		}
	}
	return NilValue, nil
}
