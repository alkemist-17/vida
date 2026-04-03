package vida

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	httpGET                        = "GET"
	httpPOST                       = "POST"
	httpPUT                        = "PUT"
	httpDELETE                     = "DELETE"
	httpPATCH                      = "PATCH"
	httpHEAD                       = "HEAD"
	httpOPTIONS                    = "OPTIONS"
	httpInvalidURLErr              = "invalid-url"
	httpNetworkErr                 = "network"
	httpTimeoutErr                 = "timeout"
	httpTemporaryErr               = "temporary"
	httpBodyReadErr                = "body-read"
	httpLargeBodyErr               = "body-size"
	httpInvalidReqErr              = "invalid-request"
	httpJsonEncodeErr              = "json-encode"
	httpDefaultSchema              = "https"
	httpMethodField                = "method"
	httpTimeoutField               = "timeout"
	httpHeadersField               = "headers"
	httpBodyField                  = "body"
	httpURLField                   = "url"
	httpStatusCodeField            = "statusCode"
	httpElapsedField               = "elapsed"
	httpContentTypeText            = "text/plain"
	httpContentTypeBinary          = "application/octet-stream"
	httpContentTypeAppJSON         = "application/json"
	httpContentType                = "Content-Type"
	httpKind                       = "kind"
	httpCauseMessage               = "cause"
	httpMaxBodySize                = 10 << 20
	httpDefaultTimeout             = 30 * time.Second
	httpMaxRetryAttempts           = 3
	httpInitialDelay               = 100 * time.Millisecond
	httpMaxDelay                   = 10 * time.Second
	httpDelayMultiplier            = 2.0
	httpDefaultTTL                 = 5 * time.Minute
	httpMaxCacheEntries            = 1000
	httpMaxIdleConnections         = 100
	httpMaxIdleConnectionsPerHost  = 10
	httpDefaultIdleConnTimeout     = 90 * time.Second
	httpDefaultTLSHandshakeTimeout = 10 * time.Second
	httpDefaultJitter              = true
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
				kind = httpTemporaryErr
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

type requestInterceptor func(*http.Request) (*http.Request, error)

type responseInterceptor func(*http.Response, []byte) (*http.Response, []byte, error)

type interceptorChain struct {
	requestInterceptors  []requestInterceptor
	responseInterceptors []responseInterceptor
	mu                   sync.RWMutex
}

func (c *interceptorChain) dddRequest(fn requestInterceptor) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requestInterceptors = append(c.requestInterceptors, fn)
}

func (c *interceptorChain) dddResponse(fn responseInterceptor) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.responseInterceptors = append(c.responseInterceptors, fn)
}

func (c *interceptorChain) executeRequest(req *http.Request) (*http.Request, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var err error
	for _, fn := range c.requestInterceptors {
		req, err = fn(req)
		if err != nil {
			return nil, err
		}
	}
	return req, nil
}

func (c *interceptorChain) executeResponse(resp *http.Response, body []byte) (*http.Response, []byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var err error
	for i := len(c.responseInterceptors) - 1; i >= 0; i-- {
		resp, body, err = c.responseInterceptors[i](resp, body)
		if err != nil {
			return nil, nil, err
		}
	}
	return resp, body, nil
}

type retryConfig struct {
	MaxAttempts     int           // Max retry attempts (default: 3)
	InitialDelay    time.Duration // Initial backoff delay (default: 100ms)
	MaxDelay        time.Duration // Max backoff delay cap (default: 10s)
	Multiplier      float64       // Backoff multiplier (default: 2.0 for exponential)
	Jitter          bool          // Add randomness to backoff (default: true)
	RetryableCodes  []int         // HTTP status codes to retry (default: [429, 500, 502, 503, 504])
	RetryableErrors []string      // Error kinds to retry (default: ["network", "timeout", "temporary"])
}

func defaultRetryConfig() *retryConfig {
	return &retryConfig{
		MaxAttempts:     httpMaxRetryAttempts,
		InitialDelay:    httpInitialDelay,
		MaxDelay:        httpMaxDelay,
		Multiplier:      httpDelayMultiplier,
		Jitter:          httpDefaultJitter,
		RetryableCodes:  []int{429, 500, 502, 503, 504},
		RetryableErrors: []string{httpNetworkErr, httpTimeoutErr, httpTemporaryErr},
	}
}

func (rc *retryConfig) shouldRetry(err error, statusCode int) bool {
	if slices.Contains(rc.RetryableCodes, statusCode) {
		return true
	}

	if err != nil {
		if reqErr, ok := err.(*RequestError); ok {
			if strings.Contains(err.Error(), reqErr.Kind) {
				return true
			}
		}

		errMsg := err.Error()
		for _, kind := range rc.RetryableErrors {
			if strings.Contains(errMsg, kind+":") {
				return true
			}
		}
	}

	return false
}

func (rc *retryConfig) calculateBackoff(attempt int) time.Duration {
	delay := float64(rc.InitialDelay) * math.Pow(rc.Multiplier, float64(attempt-1))
	if delay > float64(rc.MaxDelay) {
		delay = float64(rc.MaxDelay)
	}
	// Add random value in [0.5*delay, 1.5*delay]
	if rc.Jitter {
		jitter := 0.5 + (0.5 * float64(time.Now().UnixNano()%1000) / 1000.0)
		delay *= jitter
	}
	return time.Duration(delay)
}

type cacheEntry struct {
	Body       []byte
	StatusCode int
	Headers    http.Header
	CreatedAt  time.Time
	TTL        time.Duration
}

func (ce *cacheEntry) isExpired() bool {
	if ce.TTL <= 0 {
		return false // Infinite TTL
	}
	return time.Since(ce.CreatedAt) > ce.TTL
}

type cacheConfig struct {
	Enabled    bool
	DefaultTTL time.Duration
	MaxEntries int // 0 = unlimited
	cache      map[string]*cacheEntry
	mu         sync.RWMutex
}

func newCacheConfig() *cacheConfig {
	return &cacheConfig{
		Enabled:    true,
		DefaultTTL: httpDefaultTTL,
		MaxEntries: httpMaxCacheEntries,
		cache:      make(map[string]*cacheEntry),
	}
}

func (cc *cacheConfig) get(key string) *cacheEntry {
	if !cc.Enabled {
		return nil
	}

	cc.mu.RLock()
	defer cc.mu.RUnlock()

	entry, exists := cc.cache[key]
	if !exists || entry.isExpired() {
		if exists {
			delete(cc.cache, key)
		}
		return nil
	}

	return entry
}

func (cc *cacheConfig) set(key string, entry *cacheEntry) {
	if !cc.Enabled {
		return
	}

	cc.mu.Lock()
	defer cc.mu.Unlock()

	if cc.MaxEntries > 0 && len(cc.cache) >= cc.MaxEntries {
		for k := range cc.cache {
			delete(cc.cache, k)
			break
		}
	}

	cc.cache[key] = entry
}

func (cc *cacheConfig) clear() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.cache = nil
	cc.cache = make(map[string]*cacheEntry)
}

func generateCacheKey(method, rawURL string, headers map[string]string, body []byte) string {
	hash := sha256.New()
	hash.Write([]byte(method))
	hash.Write([]byte(rawURL))

	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		hash.Write([]byte(k))
		hash.Write([]byte(headers[k]))
	}

	if len(body) > 0 {
		hash.Write(body)
	}

	return hex.EncodeToString(hash.Sum(nil))
}

type localHttpClient struct {
	httpClient     *http.Client // Reused for connection pooling
	interceptors   *interceptorChain
	retryConfig    *retryConfig
	cacheConfig    *cacheConfig
	baseURL        string
	defaultHeaders map[string]string
}

func newClient() *localHttpClient {
	return &localHttpClient{
		httpClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        httpMaxIdleConnections,
				MaxIdleConnsPerHost: httpMaxIdleConnectionsPerHost,
				IdleConnTimeout:     httpDefaultIdleConnTimeout,
				TLSHandshakeTimeout: httpDefaultTLSHandshakeTimeout,
			},
			Timeout: httpDefaultTimeout,
		},
		interceptors:   &interceptorChain{},
		retryConfig:    defaultRetryConfig(),
		cacheConfig:    newCacheConfig(),
		defaultHeaders: make(map[string]string),
	}
}
