package graupel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	Version = "v0.0.1"

	// Snowflake Cortex API base URL pattern
	defaultUserAgent = "graupel" + "/" + Version
)

var errNonNilContext = errors.New("context must be non-nil")

// NewClient returns a new Snowflake Cortex API client. If a nil httpClient is
// provided, a new http.Client will be used. o use API methods which require
// authentication, either use Client.WithAuthToken or provide NewClient with
// an http.Client that will perform the authentication for you (such as that
// provided by the golang.org/x/oauth2 library).
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		// No client provided, so create a new default http.Client with zero values
		// This client will use default settings (no timeout, default transport, etc.)
		httpClient = &http.Client{}
	}
	// Dereference the httpClient pointer to get the actual http.Client struct value,
	// then create a NEW copy of that struct (shallow copy).
	// This ensures we have our own copy that we can safely modify without affecting
	// the caller's original httpClient. This is important because the library may
	// need to modify CheckRedirect or other client settings for specific operations.
	httpClient2 := *httpClient

	c := &Client{client: &httpClient2}
	// Call the initialize method to set up default values and initialize
	// all the service objects (Agents, Threads, etc.).
	c.initialize()

	return c
}

// Client manages communication with the Snowflake Cortex API.
type Client struct {
	// The clientMu mutex is used to ensure thread-safety when the CheckRedirect function
	// of the HTTP client is being modified. This is necessary because:
	//   Concurrent Access: If multiple goroutines are using the same Client instance, they
	//     might try to modify the CheckRedirect function at the same time.
	//   CheckRedirect Modifications: Some API endpoints need to handle redirects differently.
	//     When making requests that require custom redirect behavior, the library temporarily
	//     modifies the http.Client.CheckRedirect function.
	//   Race Condition Prevention: Without the mutex, concurrent requests that need different
	//     redirect behaviors would create a race condition, potentially causing one request
	//     to use another's redirect logic.
	clientMu              sync.Mutex   // clientMu protects the client during calls that modify the CheckRedirect func.
	client                *http.Client // HTTP client used to communicate with the API
	clientIgnoreRedirects *http.Client // HTTP client used to communicate with the API on endpoints where we don't want to follow redirects.

	BaseURL   *url.URL
	UserAgent string

	// The common service field is a small shared container that holds the pointer back
	// to the Client so all service types (Agents, Threads, etc.) can access the client
	// without each declaring their own client *Client field. Saves allocations on the
	// heap as well as reduces boilerplate code.
	common service

	// Services used for talking to different parts of the Snowflake Cortex API
	Agents *AgentObjectService
}

type service struct {
	client *Client
}

// Client returns the http.Client used by this client.
func (c *Client) Client() *http.Client {
	c.clientMu.Lock()
	defer c.clientMu.Unlock()
	clientCopy := *c.client
	return &clientCopy
}

// ClientOption is a functional option for configuring the Client.
type ClientOption func(*Client) error

// WithProgrammaticAccessToken (PAT) returns a copy of the client configured to use the
// provided PAT for the Authorization header.
func (c *Client) WithProgrammaticAccessToken(pat string) *Client {
	// Create a copy of the current client. This ensures we don't modify the original
	// client that this method was called on. This allows for immutable-style chaining
	// where each WithX method returns a new client without affecting the original.
	c2 := c.copy()

	// This sets up default values and initializes services on the copied client after
	//  we've made our modifications.
	defer c2.initialize()

	// Get the HTTP transport from the copied client.
	// The transport is responsible for actually making the HTTP request.
	transport := c2.client.Transport

	// If the client doesn't have custom transport configured (it's nil),
	// use Go's http.DefaultTransport which provides sensible defaults
	// (connection pooling, timeouts, etc.)
	if transport == nil {
		transport = http.DefaultTransport
	}

	// Wrap the existing transport with a custom RoundTripper that adds
	// the Authorization header to every request. roundTripperFunc is a helper
	// that converts a function into a type that implements http.RoundTripper.
	c2.client.Transport = roundTripperFunc(
		// This anonymous function is called for every HTTP request
		func(req *http.Request) (*http.Response, error) {
			// Clone the request to create a new copy. This is required because
			// we're going to modify headers, and http.RoundTripper spec says
			// we must not modify the original request.
			req = req.Clone(req.Context())

			// If a token was provided (not empty string), add it as a Bearer token in
			// the Authorization header. This authenticates the request with Snowflake
			// using the Programmatic Access Token.
			if pat != "" {
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", pat))
			}

			// Pass the modified request to the original transport to actually
			// make the HTTP call and return the response (or error) back up the chain.
			return transport.RoundTrip(req)
		},
	)

	// Return the new client copy with the authentication transport configured
	return c2
}

// WithSnowflakeBaseURL returns a copy of the client configured to use the provided base
// URL. If the base URL does not have the suffix "/api/v2/", it will be added
// automatically.
//
// Note that WithSnowflakeBaseURL is a convenience helper only; its behavior is equivalent to
// setting the BaseURL field.
func (c *Client) WithSnowflakeBaseURL(snowflakeBaseURL string) (*Client, error) {
	c2 := c.copy()
	defer c2.initialize()
	var err error
	c2.BaseURL, err = url.Parse(snowflakeBaseURL)
	if err != nil {
		return nil, err
	}

	// Enforce that the hostname ends with snowflakecomputing.com
	if !strings.HasSuffix(c2.BaseURL.Hostname(), "snowflakecomputing.com") {
		return nil, fmt.Errorf("snowflake base URL must end with snowflakecomputing.com, got: %s", c2.BaseURL.Hostname())
	}

	// Ensure the BaseURL path ends with a slash
	if !strings.HasSuffix(c2.BaseURL.Path, "/") {
		c2.BaseURL.Path += "/"
	}

	return c2, nil
}

// initialize sets default values and initializes services.
func (c *Client) initialize() {
	if c.client == nil {
		c.client = &http.Client{}
	}
	// Copy the main http client into the IgnoreRedirects one, overriding the `CheckRedirect` func
	c.clientIgnoreRedirects = &http.Client{}
	c.clientIgnoreRedirects.Transport = c.client.Transport
	c.clientIgnoreRedirects.Timeout = c.client.Timeout
	c.clientIgnoreRedirects.Jar = c.client.Jar
	c.clientIgnoreRedirects.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	if c.UserAgent == "" {
		c.UserAgent = defaultUserAgent
	}
	c.common.client = c
	c.Agents = (*AgentObjectService)(&c.common)
}

// copy returns a copy of the current client. It must be initialized before use.
func (c *Client) copy() *Client {
	c.clientMu.Lock()
	// can't use *c here because that would copy mutexes by value.
	clone := Client{
		client:    &http.Client{},
		UserAgent: c.UserAgent,
		BaseURL:   c.BaseURL,
	}
	c.clientMu.Unlock()
	if c.client != nil {
		clone.client.Transport = c.client.Transport
		clone.client.CheckRedirect = c.client.CheckRedirect
		clone.client.Jar = c.client.Jar
		clone.client.Timeout = c.client.Timeout
	}
	return &clone
}

// RequestOption represents an option that can modify an http.Request.
type RequestOption func(req *http.Request)

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash. If
// specified, the value pointed to by body is JSON encoded and included as the
// request body.
func (c *Client) NewRequest(method, urlStr string, body any, opts ...RequestOption) (*http.Request, error) {
	// Ensure BaseURL path includes /api/v2/
	baseURL := *c.BaseURL
	if !strings.Contains(baseURL.Path, "/api/v2/") {
		// Add /api/v2/ to the path
		if strings.HasSuffix(baseURL.Path, "/") {
			baseURL.Path += "api/v2/"
		} else {
			baseURL.Path += "/api/v2/"
		}
	}

	if !strings.HasSuffix(baseURL.Path, "/") {
		return nil, fmt.Errorf("baseURL must have a trailing slash, but %q does not", baseURL.Host)
	}

	u, err := baseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err = enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	var req *http.Request
	req, err = http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	for _, opt := range opts {
		opt(req)
	}

	return req, nil
}

// newResponse creates a new Response for the provided http.Response.
// r must not be nil.
func newResponse(r *http.Response) *Response {
	response := &Response{Response: r}
	return response
}

// Response is a Snowflake Cortex API response. This wraps the standard
// http.Response returned from Snowflake and provides convenient access
// for future items (pagination, etc.).
type Response struct {
	*http.Response
}

// bareDo sends an API request using `caller` http.Client passed in the parameters
// and lets you handle the api response. If an error or API Error occurs, the error
// will contain more information. Otherwise, you are supposed to read and close the
// response's Body.
//
// The provided ctx must be non-nil, if it is nil an error is returned. If it is
// canceled or times out, ctx.Err() will be returned.
func (c *Client) bareDo(ctx context.Context, caller *http.Client, req *http.Request) (*Response, error) {
	if ctx == nil {
		return nil, errNonNilContext
	}

	req = req.WithContext(ctx)

	resp, err := caller.Do(req)
	var response *Response
	if resp != nil {
		response = newResponse(resp)
	}

	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return response, ctx.Err()
		default:
		}

		// If the error type is *url.Error, sanitize its URL before returning.
		var e *url.Error
		if errors.As(err, &e) {
			if u, err := url.Parse(e.URL); err == nil {
				e.URL = sanitizeURL(u).String()
				return response, e
			}
		}

		return response, err
	}

	err = CheckResponse(resp)
	if err != nil {
		defer resp.Body.Close()
		return nil, err

	}

	return response, err
}

// BareDo sends an API request and lets you handle the api response. If an error
// or API Error occurs, the error will contain more information. Otherwise, you
// are supposed to read and close the response's Body. If rate limit is exceeded
// and reset time is in the future, BareDo returns *RateLimitError immediately
// without making a network API call.
//
// The provided ctx must be non-nil, if it is nil an error is returned. If it is
// canceled or times out, ctx.Err() will be returned.
func (c *Client) BareDo(ctx context.Context, req *http.Request) (*Response, error) {
	return c.bareDo(ctx, c.client, req)
}

// Do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred. If v implements the io.Writer interface,
// the raw response body will be written to v, without attempting to first
// decode it. If v is nil, and no error happens, the response is returned as is.
//
// The provided ctx must be non-nil, if it is nil an error is returned. If it
// is canceled or times out, ctx.Err() will be returned.
func (c *Client) Do(ctx context.Context, req *http.Request, v any) (*Response, error) {
	resp, err := c.BareDo(ctx, req)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()

	switch v := v.(type) {
	case nil:
	case io.Writer:
		// TODO - check if v2 JSON package allows streaming decode to io.Writer directly

		_, err = io.Copy(v, resp.Body)
	default:
		decErr := json.NewDecoder(resp.Body).Decode(v)
		if decErr == io.EOF {
			decErr = nil // ignore EOF errors caused by empty response body
		}
		if decErr != nil {
			err = decErr
		}
	}
	return resp, err
}

// ErrorResponse represents an error response from the API.
type ErrorResponse struct {
	Response *http.Response // HTTP response that caused this error
	Message  string         `json:"message"` // error message
	Code     string         `json:"code,omitempty"`
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.Message)
}

// CheckResponse checks the API response for errors, and returns them if
// present. A response is considered an error if it has a status code outside
// the 200 range.
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}

	// Initialize an ErrorResponse with the HTTP response. This serves as a fallback
	// if we cannot parse the response body, ensuring we always have at least the
	// HTTP response details available for error reporting.
	errorResponse := &ErrorResponse{Response: r}

	// Read the entire response body into memory. We need to consume the body here
	// so we can attempt to parse it as JSON, and also preserve it for later use.
	// Note: The body can only be read once, so we'll need to restore it afterward.
	data, err := io.ReadAll(r.Body)

	// Only attempt to unmarshal if the read was successful, and we actually got data.
	// This avoids trying to parse nil or empty data which would fail unnecessarily.
	if err == nil && data != nil {
		// Attempt to decode the JSON error response from the API. If successful,
		// this will populate the Message and Code fields of errorResponse.
		err = json.Unmarshal(data, errorResponse)
		if err != nil {
			// If JSON unmarshaling fails (e.g., response body is not valid JSON,
			// or doesn't match the ErrorResponse structure), reset the errorResponse
			// to its original state with only the HTTP response. This ensures we don't
			// return a partially-populated or corrupted error structure. Better to have
			// just the HTTP details than misleading or incomplete error information.
			errorResponse = &ErrorResponse{Response: r}
		}
	}

	// Restore the original response body so it can be read again later if needed.
	r.Body = io.NopCloser(bytes.NewBuffer(data))

	return errorResponse
}

// Ptr is a helper routine that allocates a new T value
// to store v and returns a pointer to it.
func Ptr[T any](v T) *T {
	return &v
}

// roundTripperFunc creates a RoundTripper (transport).
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

// sanitizeURL redacts the client_secret parameter from the URL which may be exposed to
// the user.
func sanitizeURL(uri *url.URL) *url.URL {
	if uri == nil {
		return nil
	}
	params := uri.Query()
	if len(params.Get("client_secret")) > 0 {
		params.Set("client_secret", "REDACTED")
		uri.RawQuery = params.Encode()
	}
	return uri
}
