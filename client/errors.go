package client

import (
	"fmt"
)

// BaseError is an error type that all other error types embed.
type BaseError struct {
	DefaultErrString string
	Info             string
}

func (e BaseError) Error() string {
	e.DefaultErrString = "An error occurred while executing a Gophercloud request."
	return e.choseErrString()
}

func (e BaseError) choseErrString() string {
	if e.Info != "" {
		return e.Info
	}
	return e.DefaultErrString
}

// ErrMissingInput is the error when input is required in a particular
// situation but not provided by the user
type ErrMissingInput struct {
	BaseError
	Argument string
}

func (e ErrMissingInput) Error() string {
	e.DefaultErrString = fmt.Sprintf("Missing input for argument [%s]", e.Argument)
	return e.choseErrString()
}

// ErrInvalidInput is an error type used for most non-HTTP Gophercloud errors.
type ErrInvalidInput struct {
	ErrMissingInput
	Value interface{}
}

func (e ErrInvalidInput) Error() string {
	e.DefaultErrString = fmt.Sprintf("Invalid input provided for argument [%s]: [%+v]", e.Argument, e.Value)
	return e.choseErrString()
}

// ErrUnexpectedResponseCode is returned by the Request method when a response code other than
// those listed in OkCodes is encountered.
type ErrUnexpectedResponseCode struct {
	BaseError
	URL      string
	Method   string
	Expected []int
	Actual   int
	Body     []byte
}

func (e ErrUnexpectedResponseCode) Error() string {
	e.DefaultErrString = fmt.Sprintf(
		"Expected HTTP response code %v when accessing [%s %s], but got %d instead\n%s",
		e.Expected, e.Method, e.URL, e.Actual, e.Body,
	)
	return e.choseErrString()
}

// ErrDefault400 is the default error type returned on a 400 HTTP response code.
type ErrDefault400 struct {
	ErrUnexpectedResponseCode
}

// ErrDefault401 is the default error type returned on a 401 HTTP response code.
type ErrDefault401 struct {
	ErrUnexpectedResponseCode
}

// ErrDefault403 is the default error type returned on a 403 HTTP response code.
type ErrDefault403 struct {
	ErrUnexpectedResponseCode
}

// ErrDefault404 is the default error type returned on a 404 HTTP response code.
type ErrDefault404 struct {
	ErrUnexpectedResponseCode
}

// ErrDefault405 is the default error type returned on a 405 HTTP response code.
type ErrDefault405 struct {
	ErrUnexpectedResponseCode
}

// ErrDefault408 is the default error type returned on a 408 HTTP response code.
type ErrDefault408 struct {
	ErrUnexpectedResponseCode
}

// ErrDefault429 is the default error type returned on a 429 HTTP response code.
type ErrDefault429 struct {
	ErrUnexpectedResponseCode
}

// ErrDefault500 is the default error type returned on a 500 HTTP response code.
type ErrDefault500 struct {
	ErrUnexpectedResponseCode
}

// ErrDefault503 is the default error type returned on a 503 HTTP response code.
type ErrDefault503 struct {
	ErrUnexpectedResponseCode
}

func (e ErrDefault400) Error() string {
	return "Invalid request due to incorrect syntax or missing required parameters."
}
func (e ErrDefault401) Error() string {
	return "Authentication failed"
}
func (e ErrDefault403) Error() string {
	e.DefaultErrString = fmt.Sprintf(
		"Request forbidden: [%s %s], error message: %s",
		e.Method, e.URL, e.Body,
	)
	return e.choseErrString()
}
func (e ErrDefault404) Error() string {
	return "Resource not found"
}
func (e ErrDefault405) Error() string {
	return "Method not allowed"
}
func (e ErrDefault408) Error() string {
	return "The server timed out waiting for the request"
}
func (e ErrDefault429) Error() string {
	return "Too many requests have been sent in a given amount of time. Pause" +
		" requests, wait up to one minute, and try again."
}
func (e ErrDefault500) Error() string {
	return "Internal Server Error"
}
func (e ErrDefault503) Error() string {
	return "The service is currently unable to handle the request due to a temporary" +
		" overloading or maintenance. This is a temporary condition. Try again later."
}

// Err400er is the interface resource error types implement to override the error message
// from a 400 error.
type Err400er interface {
	Error400(ErrUnexpectedResponseCode) error
}

// Err401er is the interface resource error types implement to override the error message
// from a 401 error.
type Err401er interface {
	Error401(ErrUnexpectedResponseCode) error
}

// Err403er is the interface resource error types implement to override the error message
// from a 403 error.
type Err403er interface {
	Error403(ErrUnexpectedResponseCode) error
}

// Err404er is the interface resource error types implement to override the error message
// from a 404 error.
type Err404er interface {
	Error404(ErrUnexpectedResponseCode) error
}

// Err405er is the interface resource error types implement to override the error message
// from a 405 error.
type Err405er interface {
	Error405(ErrUnexpectedResponseCode) error
}

// Err408er is the interface resource error types implement to override the error message
// from a 408 error.
type Err408er interface {
	Error408(ErrUnexpectedResponseCode) error
}

// Err429er is the interface resource error types implement to override the error message
// from a 429 error.
type Err429er interface {
	Error429(ErrUnexpectedResponseCode) error
}

// Err500er is the interface resource error types implement to override the error message
// from a 500 error.
type Err500er interface {
	Error500(ErrUnexpectedResponseCode) error
}

// Err503er is the interface resource error types implement to override the error message
// from a 503 error.
type Err503er interface {
	Error503(ErrUnexpectedResponseCode) error
}
