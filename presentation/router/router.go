package router

// Router is an abstract routing interface the presentation layer depends on.
// Concrete transports (HTTP, gRPC, etc.) implement this.
// Route allows grouping with a common path prefix.
type Router interface {
	Post(pattern string, handler any, opts ...Option)
	Get(pattern string, handler any, opts ...Option)
	Put(pattern string, handler any, opts ...Option)
	Delete(pattern string, handler any, opts ...Option)

	// Route groups routes under the given pattern.
	Route(pattern string, fn func(r Router))
}

// Option modifies per-route configuration.
type Option func(*RouteConfig)

type RouteConfig struct {
	SuccessStatus int
}

// WithStatus sets a custom success status code.
func WithStatus(code int) Option { return func(c *RouteConfig) { c.SuccessStatus = code } }

// Common helpers for semantic clarity.
func WithStatusCreated() Option   { return WithStatus(201) }
func WithStatusOK() Option        { return WithStatus(200) }
func WithStatusNoContent() Option { return WithStatus(204) }
