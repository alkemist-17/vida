package vida

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

const (
	httpGET                        = "GET"
	httpPOST                       = "POST"
	httpPUT                        = "PUT"
	httpDELETE                     = "DELETE"
	httpPATCH                      = "PATCH"
	httpHEAD                       = "HEAD"
	httpOPTIONS                    = "OPTIONS"
	httpURLField                   = "url"
	httpBaseField                  = "base"
	httpDefaultSchema              = "https"
	httpMethodField                = "method"
	httpTimeoutField               = "timeout"
	httpHeadersField               = "headers"
	httpBodyField                  = "body"
	httpQueryParamsField           = "params"
	httpStatusCodeField            = "statusCode"
	httpMaxBodySizeField           = "maxBodySize"
	httpRetryField                 = "retry"
	httpMaxAttemptsField           = "max"
	httpInitialBackoffField        = "backoff"
	httpMaxBackoffField            = "maxBackoff"
	httpRetryableCodesField        = "statusCodes"
	httpRetryAfterHeader           = "Retry-After"
	httpXRateLimitResetHeader      = "X-RateLimit-Reset"
	httpContentTypeText            = "text/plain"
	httpContentTypeBinary          = "application/octet-stream"
	httpContentTypeAppJSON         = "application/json"
	httpContentType                = "Content-Type"
	httpMaxBodySize                = 10 << 20
	httpDefaultTimeout             = 30 * time.Second
	httpMaxRetryAttempts           = 3
	httpInitialDelay               = 100 * time.Millisecond
	httpMaxDelay                   = 10 * time.Second
	httpDelayMultiplier            = 2.0
	httpMaxIdleConnections         = 200
	httpMaxConnsPerHost            = 0
	httpMaxIdleConnectionsPerHost  = 100
	httpDefaultIdleConnTimeout     = 90 * time.Second
	httpDefaultTLSHandshakeTimeout = 10 * time.Second
	httpResponseHeaderTimeout      = 15 * time.Second
	httpExpectContinueTimeout      = 1 * time.Second
	httpDefaultJitter              = true
)

var httpDefaultVidaHttpClient *vidaHttpClient

func loadFoundationHttpClient() Value {
	httpDefaultVidaHttpClient = newVidaHttpClient()
	m := &Object{Value: make(map[string]Value, 11)}
	m.Value["request"] = GFn(httpRequest)
	m.Value["get"] = GFn(httpRequest)
	m.Value["post"] = GFn(httpPost)
	m.Value["put"] = GFn(httpPut)
	m.Value["delete"] = GFn(httpDelete)
	m.Value["patch"] = GFn(httpPatch)
	m.Value["head"] = GFn(httpHead)
	m.Value["options"] = GFn(httpOptions)
	m.Value["statusText"] = GFn(httpStatusCodeText)
	m.Value["urlEncode"] = GFn(httpURLEncode)
	m.Value["detectContentType"] = GFn(httpDetectContentType)
	return m
}

func httpRequest(args ...Value) (Value, error) {
	if len(args) > 0 {
		switch v := args[0].(type) {
		case *String:
			config, err := resolveRequestConfig(v.Value, args...)
			if err != nil {
				return VidaError{Message: &String{Value: err.Error()}}, nil
			}
			ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
			defer cancel()
			resp, body, err := httpExecuteRequest(ctx, config)
			if err != nil {
				return VidaError{Message: &String{Value: err.Error()}}, nil
			}
			return httpResponseToObject(resp, body), nil
		case *Object:
			config, err := httpParseUserConfig(v, nil)
			if err != nil {
				return VidaError{Message: &String{Value: err.Error()}}, nil
			}
			ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
			defer cancel()
			resp, body, err := httpExecuteRequest(ctx, config)
			if err != nil {
				return VidaError{Message: &String{Value: err.Error()}}, nil
			}
			return httpResponseToObject(resp, body), nil
		}
	}
	return NilValue, nil
}

func resolveRequestConfig(userRawURL string, args ...Value) (*requestConfig, error) {
	if len(args) > 1 {
		if userConfig, ok := args[1].(*Object); ok {
			return httpParseUserConfig(userConfig, &userRawURL)
		}
	}

	parsedURL, err := url.Parse(userRawURL)
	if err != nil {
		return nil, err
	}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = httpDefaultSchema
	}

	return &requestConfig{
		Url:         parsedURL,
		Method:      httpGET,
		Timeout:     httpDefaultTimeout,
		MaxBodySize: httpMaxBodySize,
		Headers:     make(map[string]string),
	}, nil
}

func httpResponseToObject(resp *http.Response, body []byte) *Object {
	headersObj := &Object{Value: make(map[string]Value, len(resp.Header))}
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
		},
	}
}

type requestConfig struct {
	Method      string
	Body        Value
	Url         *url.URL
	Base        *url.URL
	Timeout     time.Duration
	MaxBodySize int64
	Headers     map[string]string
	Retry       *retryConfig
}

func httpParseUserConfig(userConfig *Object, userRawURL *string) (*requestConfig, error) {
	reqConfig := &requestConfig{
		Method:      httpGET,
		Timeout:     httpDefaultTimeout,
		MaxBodySize: httpMaxBodySize,
		Headers:     make(map[string]string),
	}

	httpParseMethod(userConfig, reqConfig)
	httpParseTimeout(userConfig, reqConfig)
	httpParseHeaders(userConfig, reqConfig)
	httpParseBody(userConfig, reqConfig)
	httpParseBodySize(userConfig, reqConfig)
	httpParseRetry(userConfig, reqConfig)

	rawURL, err := httpResolveRawURL(userConfig, userRawURL)
	if err != nil {
		return nil, err
	}

	reqConfig.Url, err = httpResolveURL(rawURL, httpParseBase(userConfig))
	if err != nil {
		return nil, err
	}

	httpParseQueryParams(userConfig, reqConfig.Url)

	return reqConfig, nil
}

func httpParseMethod(userConfig *Object, reqConfig *requestConfig) {
	if m, ok := userConfig.Value[httpMethodField].(*String); ok {
		reqConfig.Method = strings.ToUpper(m.Value)
	}
}

func httpParseTimeout(userConfig *Object, reqConfig *requestConfig) {
	if t, ok := userConfig.Value[httpTimeoutField].(Integer); ok && t >= 0 {
		reqConfig.Timeout = time.Duration(t) * time.Millisecond
	}
}

func httpParseHeaders(userConfig *Object, reqConfig *requestConfig) {
	if headers, ok := userConfig.Value[httpHeadersField].(*Object); ok {
		for k, v := range headers.Value {
			if s, ok := v.(*String); ok {
				reqConfig.Headers[k] = s.Value
			}
		}
	}
}

func httpParseBody(userConfig *Object, reqConfig *requestConfig) {
	body, exists := userConfig.Value[httpBodyField]
	if !exists {
		return
	}
	switch v := body.(type) {
	case *String, *Object, *Bytes:
		reqConfig.Body = v
	default:
		reqConfig.Body = NilValue
	}
}

func httpParseBodySize(userConfig *Object, reqConfig *requestConfig) {
	if mbs, ok := userConfig.Value[httpMaxBodySizeField].(Integer); ok && mbs >= 0 {
		reqConfig.MaxBodySize = int64(mbs)
	}
}

func httpParseRetry(userConfig *Object, reqConfig *requestConfig) {
	retry, ok := userConfig.Value[httpRetryField].(*Object)
	if !ok || len(retry.Value) == 0 {
		return
	}
	httpParseRetryFields(retry, reqConfig)
}

func httpParseRetryFields(userRetryConfig *Object, reqConfig *requestConfig) {
	retry := defaultRetryConfig()
	if maxAttempts, ok := userRetryConfig.Value[httpMaxAttemptsField].(Integer); ok && maxAttempts > 0 {
		retry.MaxAttempts = int(maxAttempts)
	}
	if backoff, ok := userRetryConfig.Value[httpInitialBackoffField].(Integer); ok && backoff > 0 {
		retry.InitialDelay = time.Duration(backoff) * time.Millisecond
	}
	if maxDelay, ok := userRetryConfig.Value[httpMaxBackoffField].(Integer); ok && maxDelay > 0 {
		retry.MaxDelay = time.Duration(maxDelay) * time.Millisecond
	}
	if codes, ok := userRetryConfig.Value[httpRetryableCodesField].(*Array); ok && len(codes.Value) > 0 {
		var c []int
		for _, v := range codes.Value {
			if code, ok := v.(Integer); ok && 100 <= code && code <= 599 {
				c = append(c, int(code))
			}
		}
		if len(c) > 0 {
			retry.RetryableCodes = c
		}
	}
	reqConfig.Retry = retry
}

func httpParseQueryParams(userConfig *Object, reqConfigURL *url.URL) {
	if queryParams, ok := userConfig.Value[httpQueryParamsField].(*Object); ok {
		q := url.Values{}
		for k, v := range queryParams.Value {
			q.Add(k, v.String())
		}
		reqConfigURL.RawQuery = q.Encode()
	}
}

func httpResolveRawURL(userConfig *Object, userRawURL *string) (string, error) {
	if userRawURL != nil {
		return *userRawURL, nil
	}
	if u, ok := userConfig.Value[httpURLField].(*String); ok {
		return u.Value, nil
	}
	if b, ok := userConfig.Value[httpBaseField].(*String); ok {
		return b.Value, nil
	}
	return "", errors.New("http client must have an url")
}

func httpParseBase(userConfig *Object) string {
	if b, ok := userConfig.Value[httpBaseField].(*String); ok {
		return b.Value
	}
	return ""
}

func httpResolveURL(rawURL, base string) (*url.URL, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	if parsed.IsAbs() {
		return parsed, nil
	}

	if base != "" {
		parsedBase, err := url.Parse(base)
		if err != nil {
			return nil, err
		}
		if parsedBase.Scheme == "" {
			parsedBase.Scheme = httpDefaultSchema
		}
		return parsedBase.JoinPath(parsed.String()), nil
	}

	if parsed.Scheme == "" {
		parsed.Scheme = httpDefaultSchema
	}
	return parsed, nil
}

func httpExecuteRequest(ctx context.Context, requestConfig *requestConfig) (*http.Response, []byte, error) {
	bodyReader, contentType, err := httpBuildBodyReader(requestConfig.Body)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequestWithContext(ctx, requestConfig.Method, requestConfig.Url.String(), bodyReader)
	if err != nil {
		return nil, nil, err
	}

	httpSetHeaders(req, requestConfig.Headers, contentType)
	return httpDoRequest(req, requestConfig)
}

func httpBuildBodyReader(body Value) (io.Reader, string, error) {
	if body == nil {
		return http.NoBody, "", nil
	}
	switch v := body.(type) {
	case *String:
		return strings.NewReader(v.Value), httpContentTypeText, nil
	case *Bytes:
		return bytes.NewReader(v.Value), httpContentTypeBinary, nil
	case *Object:
		jsonBody, err := json.Marshal(v)
		if err != nil {
			return nil, "", err
		}
		return bytes.NewBuffer(jsonBody), httpContentTypeAppJSON, nil
	default:
		return nil, "", nil
	}
}

func httpSetHeaders(req *http.Request, headers map[string]string, contentType string) {
	if contentType != "" {
		if _, exists := headers[httpContentType]; !exists {
			headers[httpContentType] = contentType
		}
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}

func httpDoRequest(req *http.Request, requestConfig *requestConfig) (*http.Response, []byte, error) {
	if requestConfig.Retry == nil {
		return httpDoSimpleRequest(req, requestConfig)
	}
	return httpDoRequestWithRetry(req, requestConfig)
}

func httpDoSimpleRequest(req *http.Request, requestConfig *requestConfig) (*http.Response, []byte, error) {
	resp, err := httpDefaultVidaHttpClient.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	defer io.Copy(io.Discard, resp.Body)

	body, err := io.ReadAll(io.LimitReader(resp.Body, requestConfig.MaxBodySize+1))
	if err != nil {
		return nil, nil, err
	}

	if int64(len(body)) > requestConfig.MaxBodySize {
		return nil, nil, fmt.Errorf("response exceeds %d bytes", requestConfig.MaxBodySize)
	}

	return resp, body, nil
}

func httpDoRequestWithRetry(req *http.Request, requestConfig *requestConfig) (*http.Response, []byte, error) {
	// Should handle body, context and idempotency
	for attemp := 0; attemp < requestConfig.Retry.MaxAttempts; attemp++ {
		res, body, err := httpDoSimpleRequest(req, requestConfig)
		if err != nil {
			return nil, nil, err
		}
		if err := req.Context().Err(); err != nil {
			return nil, nil, err
		}
		if !requestConfig.Retry.shouldRetry(res.StatusCode) {
			return res, body, nil
		}

		clonedReq := req.Clone(req.Context())
		if body != nil {
			clonedReq.Body = io.NopCloser(bytes.NewReader(body))
			clonedReq.ContentLength = int64(len(body))
			clonedReq.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(body)), nil
			}
		}
		req = clonedReq

		delay := httpCalculateDelayWithServerHint(res, requestConfig.Retry.calculateBackoff(attemp))
		select {
		case <-req.Context().Done():
			return nil, nil, req.Context().Err()
		case <-time.After(delay):
		}
	}
	return nil, nil, fmt.Errorf("max retries exceeded: max attempts: %v", requestConfig.Retry.MaxAttempts)
}

func httpCalculateDelayWithServerHint(res *http.Response, internalDelay time.Duration) time.Duration {
	serverDelay := parseRetryAfter(res)
	if serverDelay > internalDelay {
		return serverDelay
	}
	return internalDelay
}

func parseRetryAfter(res *http.Response) time.Duration {
	if ra := res.Header.Get(httpRetryAfterHeader); ra != "" {
		if secs, err := strconv.Atoi(ra); err == nil {
			return time.Duration(secs) * time.Second
		}
		if t, err := http.ParseTime(ra); err == nil {
			d := time.Until(t)
			if d > 0 {
				return d
			}
		}
	}

	if reset := res.Header.Get(httpXRateLimitResetHeader); reset != "" {
		if unix, err := strconv.ParseInt(reset, 10, 64); err == nil {
			d := time.Until(time.Unix(unix, 0))
			if d > 0 {
				return d
			}
		}
	}

	return 0
}

func httpRequestWithMethod(method string, args ...Value) (Value, error) {
	switch len(args) {
	case 1:
		switch v := args[0].(type) {
		case *String:
			userOptions := &Object{
				Value: map[string]Value{
					httpMethodField: &String{Value: method},
				},
			}
			return httpRequest(v, userOptions)
		case *Object:
			newUO := v.Clone()
			newUO.(*Object).Value[httpMethodField] = &String{Value: method}
			return httpRequest(newUO)
		}
	case 2:
		if userOptions, ok := args[1].(*Object); ok {
			newUO := userOptions.Clone()
			newUO.(*Object).Value[httpMethodField] = &String{Value: method}
			return httpRequest(args[0], newUO)
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

func httpURLEncode(args ...Value) (Value, error) {
	if len(args) > 0 {
		if data, ok := args[0].(*Object); ok && len(data.Value) > 0 {
			values := url.Values{}
			for k, v := range data.Value {
				values.Add(k, v.ObjectKey())
			}
			return &String{Value: values.Encode()}, nil
		}
	}
	return NilValue, nil
}

func httpDetectContentType(args ...Value) (Value, error) {
	if len(args) > 0 {
		if data, ok := args[0].(*Bytes); ok && len(data.Value) > 0 {
			return &String{Value: http.DetectContentType(data.Value)}, nil
		}
	}
	return NilValue, nil
}

type vidaHttpClient struct {
	httpClient *http.Client
}

func newVidaHttpClient() *vidaHttpClient {
	return &vidaHttpClient{
		httpClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:          httpMaxIdleConnections,
				MaxIdleConnsPerHost:   httpMaxIdleConnectionsPerHost,
				MaxConnsPerHost:       httpMaxConnsPerHost,
				IdleConnTimeout:       httpDefaultIdleConnTimeout,
				TLSHandshakeTimeout:   httpDefaultTLSHandshakeTimeout,
				ResponseHeaderTimeout: httpResponseHeaderTimeout,
				ExpectContinueTimeout: httpExpectContinueTimeout,
			},
			Timeout: httpDefaultTimeout,
			Jar:     httpNewCookieJar(),
		},
	}
}

func httpNewCookieJar() *cookiejar.Jar {
	if jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}); err == nil {
		return jar
	}
	return nil
}

type retryConfig struct {
	MaxAttempts    int           // Max retry attempts
	InitialDelay   time.Duration // Initial backoff delay in milliseconds
	MaxDelay       time.Duration // Max backoff delay cap in milliseconds
	Multiplier     float64       // Backoff multiplier
	Jitter         bool          // Add randomness to backoff
	RetryableCodes []int         // HTTP status codes to retry
}

func defaultRetryConfig() *retryConfig {
	return &retryConfig{
		MaxAttempts:    httpMaxRetryAttempts,
		InitialDelay:   httpInitialDelay,
		MaxDelay:       httpMaxDelay,
		Multiplier:     httpDelayMultiplier,
		Jitter:         httpDefaultJitter,
		RetryableCodes: []int{429, 500, 502, 503, 504},
	}
}

func (rc *retryConfig) shouldRetry(statusCode int) bool {
	return slices.Contains(rc.RetryableCodes, statusCode)
}

func (rc *retryConfig) calculateBackoff(attempt int) time.Duration {
	delay := float64(rc.InitialDelay) * math.Pow(rc.Multiplier, float64(attempt))
	if delay > float64(rc.MaxDelay) {
		delay = float64(rc.MaxDelay)
	}
	// Add random value in range [0.5*delay, 1.5*delay]
	if rc.Jitter {
		jitter := 0.5 + (0.5 * float64(time.Now().UnixNano()%1000) / 1000.0)
		delay *= jitter
	}
	return time.Duration(delay)
}
