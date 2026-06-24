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

	"github.com/alkemist-17/vida/token"
	"github.com/alkemist-17/vida/verror"
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

func loadFoundationHttpClient() Value {
	m := &Object{Value: make(map[string]Value, 4)}
	m.Value["newClient"] = NativeFunction(httpNewClient)
	m.Value["statusText"] = NativeFunction(httpStatusCodeText)
	m.Value["urlEncode"] = NativeFunction(httpURLEncode)
	m.Value["detectContentType"] = NativeFunction(httpDetectContentType)
	return m
}

func httpNewClient(ctx *Context, args ...Value) (Value, error) {
	return buildClientObject(ctx, newVidaHttpClient()), nil
}

func buildClientObject(ctx *Context, client *vidaHttpClient) *Object {
	obj := &Object{Value: make(map[string]Value, 7)}
	obj.Value["get"] = NativeFunction(makeRequestFn(ctx, client, httpGET))
	obj.Value["post"] = NativeFunction(makeRequestFn(ctx, client, httpPOST))
	obj.Value["put"] = NativeFunction(makeRequestFn(ctx, client, httpPUT))
	obj.Value["delete"] = NativeFunction(makeRequestFn(ctx, client, httpDELETE))
	obj.Value["patch"] = NativeFunction(makeRequestFn(ctx, client, httpPATCH))
	obj.Value["head"] = NativeFunction(makeRequestFn(ctx, client, httpHEAD))
	obj.Value["options"] = NativeFunction(makeRequestFn(ctx, client, httpOPTIONS))
	return obj
}

func makeRequestFn(ctx *Context, client *vidaHttpClient, fixedMethod string) func(*Context, ...Value) (Value, error) {
	return func(ctx *Context, args ...Value) (Value, error) {
		if len(args) > 0 {
			if _, isSelf := args[0].(*Object); isSelf {
				args = args[1:]
			}
		}

		config, err := resolveConfig(fixedMethod, args...)
		if err != nil {
			return &VidaError{Message: &String{Value: err.Error(), VTable: ctx.initialVTables[stringVT]}}, nil
		}

		context, cancel := context.WithTimeout(context.Background(), config.Timeout)
		defer cancel()

		resp, body, err := client.executeRequest(context, config)
		if err != nil {
			return &VidaError{Message: &String{Value: err.Error(), VTable: ctx.initialVTables[stringVT]}}, nil
		}
		return httpResponseToObject(ctx, resp, body), nil
	}
}

func resolveConfig(fixedMethod string, args ...Value) (*requestConfig, error) {
	var cfg *requestConfig
	var err error

	switch len(args) {
	case 0:
		return nil, errors.New("http client: no arguments provided")

	case 1:
		switch v := args[0].(type) {
		case *String:
			cfg, err = resolveRequestConfig(v.Value)
		case *Object:
			cfg, err = httpParseUserConfig(v, nil)
		default:
			return nil, errors.New("http client: argument must be a url string or config object")
		}

	default: // 2+ args: first is url string, second is config object
		if urlStr, ok := args[0].(*String); ok {
			if userCfg, ok := args[1].(*Object); ok {
				cfg, err = httpParseUserConfig(userCfg, &urlStr.Value)
			} else {
				cfg, err = resolveRequestConfig(urlStr.Value)
			}
		} else {
			return nil, errors.New("http client: first argument must be a url string")
		}
	}

	if err != nil {
		return nil, err
	}

	// Override method only when the factory was created with a fixed verb.
	if fixedMethod != "" {
		cfg.Method = fixedMethod
	}
	return cfg, nil
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

	if parsedURL.Scheme == EmptyString {
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

func httpResponseToObject(ctx *Context, resp *http.Response, body []byte) *Object {
	headersObj := &Object{Value: make(map[string]Value, len(resp.Header))}
	for k, v := range resp.Header {
		if len(v) > 0 {
			headersObj.Value[k] = &String{Value: v[0], VTable: ctx.initialVTables[stringVT]}
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
		reqConfig.Body = Nil
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
	return EmptyString, errors.New("http client must have an url")
}

func httpParseBase(userConfig *Object) string {
	if b, ok := userConfig.Value[httpBaseField].(*String); ok {
		return b.Value
	}
	return EmptyString
}

func httpResolveURL(rawURL, base string) (*url.URL, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	if parsed.IsAbs() {
		return parsed, nil
	}

	if base != EmptyString {
		parsedBase, err := url.Parse(base)
		if err != nil {
			return nil, err
		}
		if parsedBase.Scheme == EmptyString {
			parsedBase.Scheme = httpDefaultSchema
		}
		return parsedBase.JoinPath(parsed.String()), nil
	}

	if parsed.Scheme == EmptyString {
		parsed.Scheme = httpDefaultSchema
	}
	return parsed, nil
}

func httpBuildBodyReader(body Value) (io.Reader, string, error) {
	if body == nil {
		return http.NoBody, EmptyString, nil
	}
	switch v := body.(type) {
	case *String:
		return strings.NewReader(v.Value), httpContentTypeText, nil
	case *Bytes:
		return bytes.NewReader(v.Value), httpContentTypeBinary, nil
	case *Object:
		jsonBody, err := json.Marshal(v)
		if err != nil {
			return nil, EmptyString, err
		}
		return bytes.NewBuffer(jsonBody), httpContentTypeAppJSON, nil
	default:
		return nil, EmptyString, nil
	}
}

func httpSetHeaders(req *http.Request, headers map[string]string, contentType string) {
	if contentType != EmptyString {
		if _, exists := headers[httpContentType]; !exists {
			headers[httpContentType] = contentType
		}
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}

func httpCalculateDelayWithServerHint(res *http.Response, internalDelay time.Duration) time.Duration {
	serverDelay := parseRetryAfter(res)
	if serverDelay > internalDelay {
		return serverDelay
	}
	return internalDelay
}

func parseRetryAfter(res *http.Response) time.Duration {
	if ra := res.Header.Get(httpRetryAfterHeader); ra != EmptyString {
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

	if reset := res.Header.Get(httpXRateLimitResetHeader); reset != EmptyString {
		if unix, err := strconv.ParseInt(reset, 10, 64); err == nil {
			d := time.Until(time.Unix(unix, 0))
			if d > 0 {
				return d
			}
		}
	}

	return 0
}

func httpStatusCodeText(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if code, ok := args[0].(Integer); ok {
			return &String{Value: http.StatusText(int(code)), VTable: ctx.initialVTables[stringVT]}, nil
		}
	}
	return Nil, nil
}

func httpURLEncode(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if data, ok := args[0].(*Object); ok && len(data.Value) > 0 {
			values := url.Values{}
			for k, v := range data.Value {
				values.Add(k, v.ObjectKey())
			}
			return &String{Value: values.Encode(), VTable: ctx.initialVTables[stringVT]}, nil
		}
	}
	return Nil, nil
}

func httpDetectContentType(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if data, ok := args[0].(*Bytes); ok && len(data.Value) > 0 {
			return &String{Value: http.DetectContentType(data.Value), VTable: ctx.initialVTables[stringVT]}, nil
		}
	}
	return Nil, nil
}

type vidaHttpClient struct {
	ReferenceSemanticsImpl
	httpClient *http.Client
}

func (client *vidaHttpClient) Boolean() Bool {
	return True
}

func (client *vidaHttpClient) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return False, nil
	default:
		return Nil, verror.ErrPrefixOpNotDefined
	}
}

func (client *vidaHttpClient) Equals(other Value) Bool {
	x, ok := other.(*vidaHttpClient)
	return Bool(ok && x == client)
}

func (client *vidaHttpClient) String() string {
	return fmt.Sprintf("HttpClient(%p)", client)
}

func (client *vidaHttpClient) Type() string {
	return "httpClient"
}

func (client *vidaHttpClient) Clone() Value {
	return Nil
}

func (client *vidaHttpClient) ObjectKey() string {
	return fmt.Sprintf("HttpClient(%p)", client)
}

func (c *vidaHttpClient) executeRequest(ctx context.Context, cfg *requestConfig) (*http.Response, []byte, error) {
	bodyReader, contentType, err := httpBuildBodyReader(cfg.Body)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequestWithContext(ctx, cfg.Method, cfg.Url.String(), bodyReader)
	if err != nil {
		return nil, nil, err
	}

	httpSetHeaders(req, cfg.Headers, contentType)
	return c.doRequest(req, cfg)
}

func (c *vidaHttpClient) doRequest(req *http.Request, cfg *requestConfig) (*http.Response, []byte, error) {
	if cfg.Retry == nil {
		return c.doSimpleRequest(req, cfg)
	}
	return c.doRequestWithRetry(req, cfg)
}

func (c *vidaHttpClient) doSimpleRequest(req *http.Request, cfg *requestConfig) (*http.Response, []byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	defer io.Copy(io.Discard, resp.Body)

	body, err := io.ReadAll(io.LimitReader(resp.Body, cfg.MaxBodySize+1))
	if err != nil {
		return nil, nil, err
	}
	if int64(len(body)) > cfg.MaxBodySize {
		return nil, nil, fmt.Errorf("response exceeds %d bytes", cfg.MaxBodySize)
	}
	return resp, body, nil
}

func (c *vidaHttpClient) doRequestWithRetry(req *http.Request, cfg *requestConfig) (*http.Response, []byte, error) {
	for attempt := 0; attempt < cfg.Retry.MaxAttempts; attempt++ {
		res, body, err := c.doSimpleRequest(req, cfg)
		if err != nil {
			return nil, nil, err
		}
		if err := req.Context().Err(); err != nil {
			return nil, nil, err
		}
		if !cfg.Retry.shouldRetry(res.StatusCode) {
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

		delay := httpCalculateDelayWithServerHint(res, cfg.Retry.calculateBackoff(attempt))
		select {
		case <-req.Context().Done():
			return nil, nil, req.Context().Err()
		case <-time.After(delay):
		}
	}
	return nil, nil, fmt.Errorf("max retries exceeded: max attempts: %v", cfg.Retry.MaxAttempts)
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
