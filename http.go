package vida

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
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
	httpNetworkErr                 = "network"
	httpTimeoutErr                 = "timeout"
	httpTemporaryErr               = "temporary"
	httpURLField                   = "url"
	httpBaseField                  = "base"
	httpDefaultSchema              = "https"
	httpMethodField                = "method"
	httpTimeoutField               = "timeout"
	httpHeadersField               = "headers"
	httpBodyField                  = "body"
	httpStatusCodeField            = "statusCode"
	httpRetryField                 = "retry"
	httpCacheField                 = "cache"
	httpMaxField                   = "max"
	httpBackoffField               = "backoff"
	httpJitterField                = "jitter"
	httpEnabledField               = "enabled"
	httpTTLField                   = "ttl"
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
	httpDefaultTTL                 = 5 * time.Minute
	httpMaxCacheEntries            = 1000
	httpMaxIdleConnections         = 100
	httpMaxIdleConnectionsPerHost  = 10
	httpDefaultIdleConnTimeout     = 90 * time.Second
	httpDefaultTLSHandshakeTimeout = 10 * time.Second
	httpDefaultJitter              = true
)

var httpDefaultClient *localHttpClient

func loadFoundationHttpClient() Value {
	httpDefaultClient = newLocalHttpClient()
	m := &Object{Value: make(map[string]Value, 9)}
	m.Value["request"] = GFn(httpRequest)
	m.Value["get"] = GFn(httpRequest)
	m.Value["post"] = GFn(httpPost)
	m.Value["put"] = GFn(httpPut)
	m.Value["delete"] = GFn(httpDelete)
	m.Value["patch"] = GFn(httpPatch)
	m.Value["head"] = GFn(httpHead)
	m.Value["options"] = GFn(httpOptions)
	m.Value["statusText"] = GFn(httpStatusCodeText)
	return m
}

func httpRequest(args ...Value) (Value, error) {
	if len(args) > 0 {
		switch v := args[0].(type) {
		case *String:
			options, err := resolveRequestOptions(v.Value, args...)
			if err != nil {
				return VidaError{Message: &String{Value: err.Error()}}, nil
			}
			ctx, cancel := context.WithTimeout(context.Background(), options.Timeout)
			defer cancel()
			resp, body, err := httpExecuteRequest(ctx, options)
			if err != nil {
				return VidaError{Message: &String{Value: err.Error()}}, nil
			}
			return httpResponseToObject(resp, body), nil
		case *Object:
			options, err := httpParseUserOptions(v, nil)
			if err != nil {
				return VidaError{Message: &String{Value: err.Error()}}, nil
			}
			ctx, cancel := context.WithTimeout(context.Background(), options.Timeout)
			defer cancel()
			resp, body, err := httpExecuteRequest(ctx, options)
			if err != nil {
				return VidaError{Message: &String{Value: err.Error()}}, nil
			}
			return httpResponseToObject(resp, body), nil
		}
	}
	return NilValue, nil
}

func resolveRequestOptions(rawURL string, args ...Value) (*requestOptions, error) {
	if len(args) > 1 {
		if userOptions, ok := args[1].(*Object); ok {
			return httpParseUserOptions(userOptions, &rawURL)
		}
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	return &requestOptions{
		Url:     parsedURL,
		Method:  httpGET,
		Timeout: httpDefaultTimeout,
		Headers: make(map[string]string),
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

type requestOptions struct {
	Method  string
	Body    Value
	Url     *url.URL
	Base    *url.URL
	Timeout time.Duration
	Headers map[string]string
}

func httpParseUserOptions(userOptions *Object, userRawURL *string) (*requestOptions, error) {
	options := &requestOptions{
		Method:  httpGET,
		Timeout: httpDefaultTimeout,
		Headers: make(map[string]string),
	}

	httpParseMethod(userOptions, options)
	httpParseTimeout(userOptions, options)
	httpParseHeaders(userOptions, options)
	httpParseBody(userOptions, options)

	rawURL, err := httpResolveRawURL(userOptions, userRawURL)
	if err != nil {
		return nil, err
	}

	options.Url, err = httpResolveURL(rawURL, httpParseBase(userOptions))
	if err != nil {
		return nil, err
	}

	return options, nil
}

func httpParseMethod(userOptions *Object, options *requestOptions) {
	if m, ok := userOptions.Value[httpMethodField].(*String); ok {
		options.Method = strings.ToUpper(m.Value)
	}
}

func httpParseTimeout(userOptions *Object, options *requestOptions) {
	if t, ok := userOptions.Value[httpTimeoutField].(Integer); ok {
		options.Timeout = time.Duration(t) * time.Millisecond
	}
}

func httpParseHeaders(userOptions *Object, options *requestOptions) {
	if headers, ok := userOptions.Value[httpHeadersField].(*Object); ok {
		for k, v := range headers.Value {
			if s, ok := v.(*String); ok {
				options.Headers[k] = s.Value
			}
		}
	}
}

func httpParseBody(userOptions *Object, options *requestOptions) {
	body, exists := userOptions.Value[httpBodyField]
	if !exists {
		return
	}
	switch v := body.(type) {
	case *String, *Object, *Bytes:
		options.Body = v
	default:
		options.Body = NilValue
	}
}

func httpResolveRawURL(userOptions *Object, userRawURL *string) (string, error) {
	if userRawURL != nil {
		return *userRawURL, nil
	}
	if u, ok := userOptions.Value[httpURLField].(*String); ok {
		return u.Value, nil
	}
	if b, ok := userOptions.Value[httpBaseField].(*String); ok {
		return b.Value, nil
	}
	return "", errors.New("http client must have an url")
}

func httpParseBase(userOptions *Object) string {
	if b, ok := userOptions.Value[httpBaseField].(*String); ok {
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

func httpExecuteRequest(ctx context.Context, userOptions *requestOptions) (*http.Response, []byte, error) {
	bodyReader, contentType, err := httpBuildBodyReader(userOptions.Body)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequestWithContext(ctx, userOptions.Method, userOptions.Url.String(), bodyReader)
	if err != nil {
		return nil, nil, err
	}

	httpSetHeaders(req, userOptions.Headers, contentType)
	return httpDoRequest(req)
}

func httpBuildBodyReader(body Value) (io.Reader, string, error) {
	if body == nil {
		return nil, "", nil
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

func httpDoRequest(req *http.Request) (*http.Response, []byte, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, httpMaxBodySize+1))
	if err != nil {
		return nil, nil, err
	}

	if int64(len(body)) > httpMaxBodySize {
		return nil, nil, fmt.Errorf("response exceeds %d bytes", httpMaxBodySize)
	}

	return resp, body, nil
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

// Interceptors, retry logic and cache.
type requestInterceptor func(*http.Request) (*http.Request, error)

type responseInterceptor func(*http.Response, []byte) (*http.Response, []byte, error)

type interceptorChain struct {
	requestInterceptors  []requestInterceptor
	responseInterceptors []responseInterceptor
	mu                   sync.RWMutex
}

func (c *interceptorChain) addRequest(fn requestInterceptor) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requestInterceptors = append(c.requestInterceptors, fn)
}

func (c *interceptorChain) addResponse(fn responseInterceptor) {
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

func httpGenerateCacheKey(method, rawURL string, headers map[string]string, body []byte) string {
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
	httpClient     *http.Client
	interceptors   *interceptorChain
	retryConfig    *retryConfig
	cacheConfig    *cacheConfig
	baseURL        string
	defaultHeaders map[string]string
}

func newLocalHttpClient() *localHttpClient {
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

func (c *localHttpClient) executeRequestWithRetryLogic(ctx context.Context, rawURL string, opts *requestOptions, retryCfg *retryConfig) (*http.Response, []byte, error) {
	var lastErr error

	for attempt := 1; attempt <= retryCfg.MaxAttempts; attempt++ {
		resp, body, err := httpExecuteRequest(ctx, opts)
		if err == nil {
			if retryCfg.shouldRetry(nil, resp.StatusCode) {
				lastErr = fmt.Errorf("retryable_status: %d", resp.StatusCode)
				resp.Body.Close()
			} else {
				return resp, body, nil
			}
		} else {
			if retryCfg.shouldRetry(err, 0) {
				lastErr = err
			} else {
				return nil, nil, err
			}
		}

		// Wait before retry (if not last attempt)
		if attempt < retryCfg.MaxAttempts {
			delay := retryCfg.calculateBackoff(attempt)
			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return nil, nil, ctx.Err()
			}
		}
	}

	return nil, nil, fmt.Errorf("max_retries_exceeded: %v", lastErr)
}

func (c *localHttpClient) request(args ...Value) (Value, error) {
	if len(args) < 1 {
		return NilValue, fmt.Errorf("http.request: expected at least 1 argument (url string)")
	}

	urlVal, ok := args[0].(*String)
	if !ok {
		return NilValue, fmt.Errorf("http.request: first argument must be *String, got %T", args[0])
	}

	var opts *requestOptions = &requestOptions{}
	var retryCfg *retryConfig = c.retryConfig
	var cacheCfg *cacheConfig = c.cacheConfig

	if len(args) > 1 && args[1] != NilValue {
		if optsObj, ok := args[1].(*Object); ok {
			parsed, rCfg, cCfg := httpParseRetryAndCacheOptions(optsObj)
			opts = parsed
			if rCfg != nil {
				retryCfg = rCfg
			}
			if cCfg != nil {
				cacheCfg = cCfg
			}
		}
	}

	// Generate cache key if caching enabled
	var cacheKey string
	if cacheCfg.Enabled {
		cacheKey = httpGenerateCacheKey(opts.Method, urlVal.Value, opts.Headers, nil)
		if entry := cacheCfg.get(cacheKey); entry != nil {
			return httpResponseToObject(&http.Response{
				StatusCode: entry.StatusCode,
				Header:     entry.Headers,
				Request:    &http.Request{URL: &url.URL{Path: urlVal.Value}},
			}, entry.Body), nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	resp, body, err := c.executeRequestWithRetryLogic(ctx, urlVal.Value, opts, retryCfg)
	if err != nil {
		errMsg := err.Error()
		var kind, message string
		if _, err := fmt.Sscanf(errMsg, "%s: %s", &kind, &message); err != nil {
			kind = "unknown"
			message = errMsg
		}
		errObj := &Object{
			Value: map[string]Value{
				"kind":    &String{Value: kind},
				"message": &String{Value: message},
			},
		}
		return VidaError{Message: errObj}, nil
	}

	// Apply response interceptors
	resp, body, err = c.interceptors.executeResponse(resp, body)
	if err != nil {
		return NilValue, err
	}

	// Cache successful GET responses
	if cacheCfg.Enabled && opts.Method == httpGET && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		ttl := cacheCfg.DefaultTTL
		// Allow per-request TTL override
		if cacheCfg.DefaultTTL > 0 {
			ttl = cacheCfg.DefaultTTL
		}
		cacheCfg.set(cacheKey, &cacheEntry{
			Body:       body,
			StatusCode: resp.StatusCode,
			Headers:    resp.Header.Clone(),
			CreatedAt:  time.Now(),
			TTL:        ttl,
		})
	}

	return httpResponseToObject(resp, body), nil
}

func httpParseRetryAndCacheOptions(opts *Object) (*requestOptions, *retryConfig, *cacheConfig) {
	reqOpts, _ := httpParseUserOptions(opts, nil)

	var retryCfg *retryConfig
	var cacheCfg *cacheConfig

	if retryVal, exists := opts.Value[httpRetryField]; exists {
		if retryObj, ok := retryVal.(*Object); ok {
			retryCfg = httpParseRetryConfig(retryObj)
		}
	}

	if cacheVal, exists := opts.Value[httpCacheField]; exists {
		if cacheObj, ok := cacheVal.(*Object); ok {
			cacheCfg = httpParseCacheConfig(cacheObj)
		}
	}

	return reqOpts, retryCfg, cacheCfg
}

func httpParseRetryConfig(obj *Object) *retryConfig {
	cfg := defaultRetryConfig()

	if maxVal, exists := obj.Value[httpMaxField]; exists {
		if intVal, ok := maxVal.(Integer); ok {
			cfg.MaxAttempts = int(intVal)
		}
	}

	if backoffVal, exists := obj.Value[httpBackoffField]; exists {
		if intVal, ok := backoffVal.(Integer); ok {
			cfg.InitialDelay = time.Duration(intVal) * time.Millisecond
		}
	}

	if jitterVal, exists := obj.Value[httpJitterField]; exists {
		if boolVal, ok := jitterVal.(Bool); ok {
			cfg.Jitter = bool(boolVal)
		}
	}

	return cfg
}

func httpParseCacheConfig(obj *Object) *cacheConfig {
	cfg := newCacheConfig()

	if enabledVal, exists := obj.Value[httpEnabledField]; exists {
		if boolVal, ok := enabledVal.(Bool); ok {
			cfg.Enabled = bool(boolVal)
		}
	}

	if ttlVal, exists := obj.Value[httpTTLField]; exists {
		if intVal, ok := ttlVal.(Integer); ok {
			cfg.DefaultTTL = time.Duration(intVal) * time.Second
		}
	}

	return cfg
}

func httpRegisterInterceptor(args ...Value) (Value, error) {
	if len(args) == 0 {
		return NilValue, fmt.Errorf("http.use: expected at least 1 interceptor function")
	}

	if args[0] != NilValue {
		if reqFn, ok := args[0].(Value); Bool(ok) && reqFn.IsCallable() {
			httpDefaultClient.interceptors.addRequest(func(req *http.Request) (*http.Request, error) {
				reqObj := httpRequestToObject(req)
				_, err := reqFn.Call(reqObj)
				if err != nil {
					return nil, err
				}
				// Convert back to http.Request (simplified: assume same request)
				return req, nil
			})
		}
	}

	if len(args) > 1 && args[1] != NilValue {
		if resFn, ok := args[1].(Value); Bool(ok) && resFn.IsCallable() {
			httpDefaultClient.interceptors.addResponse(func(resp *http.Response, body []byte) (*http.Response, []byte, error) {
				respObj := httpResponseToObject(resp, body)
				_, err := resFn.Call(respObj)
				if err != nil {
					return nil, nil, err
				}
				// Convert back (simplified)
				return resp, body, nil
			})
		}
	}

	return NilValue, nil
}

func httpRequestToObject(req *http.Request) *Object {
	headers := &Object{Value: make(map[string]Value)}
	for k, v := range req.Header {
		if len(v) > 0 {
			headers.Value[k] = &String{Value: v[0]}
		}
	}

	return &Object{
		Value: map[string]Value{
			"method":  &String{Value: req.Method},
			"url":     &String{Value: req.URL.String()},
			"headers": headers,
		},
	}
}

func httpCacheControl(args ...Value) (Value, error) {
	if len(args) == 0 {
		return NilValue, nil
	}

	action, ok := args[0].(*String)
	if !ok {
		return NilValue, fmt.Errorf("http.cache: first argument must be action string")
	}

	switch action.Value {
	case "clear":
		httpDefaultClient.cacheConfig.clear()
		return NilValue, nil
	case "stats":
		// Return simple stats object
		stats := &Object{
			Value: map[string]Value{
				"enabled": Bool(httpDefaultClient.cacheConfig.Enabled),
				// Add more stats as needed
			},
		}
		return stats, nil
	default:
		return NilValue, fmt.Errorf("http.cache: unknown action %q", action.Value)
	}
}

func httpCreateClient(args ...Value) (Value, error) {
	client := newLocalHttpClient()

	if len(args) > 0 && args[0] != NilValue {
		if config, ok := args[0].(*Object); ok {
			if maxConnsVal, exists := config.Value["maxConns"]; exists {
				if intVal, ok := maxConnsVal.(Integer); ok {
					if transport, ok := client.httpClient.Transport.(*http.Transport); ok {
						transport.MaxIdleConns = int(intVal)
					}
				}
			}
			// Add more config options as needed
		}
	}

	clientObj := &Object{Value: make(map[string]Value)}
	clientObj.Value["request"] = GFn(client.request)
	clientObj.Value["use"] = GFn(func(args ...Value) (Value, error) {
		return httpRegisterInterceptor(args...)
	})

	return clientObj, nil
}
