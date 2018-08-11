package kioto

import (
	"context"
	"net/http"

	"github.com/delicb/kioto/cliware"
	"github.com/delicb/kioto/middlewares/headers"
	"github.com/delicb/kioto/middlewares/retry"
)

// Doer defines that object is capable of executing HTTP request with provided
// middlewares.
type Doer interface {
	Do(ctx context.Context, middlewares ...cliware.Middleware) (*Response, error)
}

// HTTPDoer defines interface needed to send HTTP request over the wire. Most
// often this will be *http.Client instance, which implements this interface.
// It is useful to have this abstraction for testing purposes.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is main point of contact with this library. It is used to set up
// basic configuration that will be used for all requests (except if it is
// overridden on per-request basis). For better performance (reuse of connections)
// only one instance of doer should be created.
type Client struct {
	preMiddlewares  *cliware.Chain
	postMiddlewares *cliware.Chain
	doer            HTTPDoer
}

// New returns fresh instance of Client configured with provided options.
// Without any options client is usable. For better performance it is recommended
// not to create multiple versions of client.
func New(options ...ClientOption) *Client {
	opts := &clientOptions{}
	for _, opt := range options {
		opt(opts)
	}

	sender := opts.httpClient
	if sender == nil {
		// not using http.DefaultClient since changing its RoundTripper would change it globally
		sender = &http.Client{
			Timeout: opts.timeout,
		}
	}

	if client, ok := sender.(*http.Client); ok && !opts.disableRetry {
		retry.Enable(client)
	}

	preMiddlewares := cliware.NewChain(opts.middlewares...)

	if !opts.userAgent.IsEmpty() {
		preMiddlewares.Use(headers.Set("User-Agent", opts.userAgent.String()))
	}

	return &Client{
		doer:            sender,
		preMiddlewares:  preMiddlewares,
		postMiddlewares: cliware.NewChain(opts.potsMiddlewares...),
	}
}

// Use adds middleware that will be applied to every request made with this client.
func (c *Client) Use(m cliware.Middleware) *Client {
	c.preMiddlewares.Use(m)
	return c
}

// UseFunc adds middleware defined as function that will be applied to every
// request made with this client.
func (c *Client) UseFunc(m func(cliware.Handler) cliware.Handler) *Client {
	c.preMiddlewares.UseFunc(m)
	return c
}

// UsePost adds middleware that will be applied to every request make with
// this client AFTER all request specific middlewares are applied.
func (c *Client) UsePost(m cliware.Middleware) *Client {
	c.postMiddlewares.Use(m)
	return c
}

// UsePostFunc adds middleware defined as a function that will be applied to
// every request made with this clients AFTER all request specific middlewares
// are applied.
func (c *Client) UsePostFunc(m func(cliware.Handler) cliware.Handler) *Client {
	c.postMiddlewares.UseFunc(m)
	return c
}

// Request creates and returns new instance of *Request.
func (c *Client) Request() *Request {
	return NewRequest(c)
}

// Do executes all provided middlewares together with permanent middlewares
// (added with Use* methods) and sends request. Context can be used to pass
// information to middlewares or for cancelation.
func (c *Client) Do(ctx context.Context, middlewares ...cliware.Middleware) (*Response, error) {
	return c.Request().WithContext(ctx).Use(middlewares...).Send()
}

func (c *Client) sendRequest(req *http.Request) (*http.Response, error) {
	return c.doer.Do(req)
}
