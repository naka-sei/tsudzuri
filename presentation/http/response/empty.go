package response

// EmptyResponse is a reusable type for endpoints that intentionally return no body.
// Adapters can detect this and write only status code.
// Using a struct instead of nil makes JSON encoding decisions explicit and keeps handler signatures uniform.
type EmptyResponse struct{}
