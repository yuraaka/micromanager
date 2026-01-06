// Package api defines request and response types for the __SERVICE_NAME_CAMEL__ service API.
package api

type HelloRequest struct {
	Friend User `json:"friend"`
}
