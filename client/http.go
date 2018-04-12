package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

var applicationJSON = "application/json"

// RequestOpts customizes the behavior of the http.Request() method.
type RequestOpts struct {
	// JSONBody, if provided, will be encoded as JSON and used as the body of the HTTP request. The
	// content type of the request will default to "application/json".
	// It's an error to specify both a JSONBody and a RawBody.
	JSONBody interface{}
	// RawBody contains an io.Reader that will be consumed by the request directly.
	RawBody io.Reader
	// OkCodes contains a list of numeric HTTP status codes that should be interpreted as success.
	// If the response has a different code, an error will be returned.
	OkCodes []int
	// ErrorContext specifies the resource error type to return if an error is encountered.
	// This lets resources override default error messages based on the response status code.
	ErrorContext error
}

func defaultOkCodes(method string) []int {
	switch {
	case method == "GET":
		return []int{200}
	case method == "POST":
		return []int{201, 202}
	case method == "PUT":
		return []int{201, 202}
	case method == "PATCH":
		return []int{200, 204}
	case method == "DELETE":
		return []int{202, 204}
	}

	return []int{}
}

type HTTPClient struct {
	httpClient http.Client
}

func (c *HTTPClient) initReqOpts(url string, JSONBody interface{}, opts *RequestOpts) {
	if v, ok := (JSONBody).(io.Reader); ok {
		opts.RawBody = v
	} else if JSONBody != nil {
		opts.JSONBody = JSONBody
	}
}

func (c *HTTPClient) Request(method, url string, options *RequestOpts) (*http.Response, error) {
	var body io.Reader
	var contentType *string

	// Derive the content body by either encoding an arbitrary object as JSON, or by taking a provided
	// io.ReadSeeker as-is. Default the content-type to application/json.
	if options.JSONBody != nil {
		if options.RawBody != nil {
			panic("Please provide only one of JSONBody or RawBody to Request().")
		}

		rendered, err := json.Marshal(options.JSONBody)
		if err != nil {
			return nil, err
		}

		body = bytes.NewReader(rendered)
		contentType = &applicationJSON
	}

	if options.RawBody != nil {
		body = options.RawBody
	}
	// Construct the http.Request.
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// Populate the request headers. Apply options.MoreHeaders last, to give the caller the chance to
	// modify or omit any header.
	if contentType != nil {
		req.Header.Set("Content-Type", *contentType)
	}
	req.Header.Set("Accept", applicationJSON)

	// Set connection parameter to close the connection immediately when we've got the response
	req.Close = true

	// Issue the request.
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Allow default OkCodes if none explicitly set
	if options.OkCodes == nil {
		options.OkCodes = defaultOkCodes(method)
	}

	// Validate the HTTP response status.
	var ok bool
	for _, code := range options.OkCodes {
		if resp.StatusCode == code {
			ok = true
			break
		}
	}

	if !ok {
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		respErr := ErrUnexpectedResponseCode{
			URL:      url,
			Method:   method,
			Expected: options.OkCodes,
			Actual:   resp.StatusCode,
			Body:     body,
		}

		errType := options.ErrorContext
		switch resp.StatusCode {
		case http.StatusBadRequest:
			err = ErrDefault400{respErr}
			if error400er, ok := errType.(Err400er); ok {
				err = error400er.Error400(respErr)
			}
		case http.StatusForbidden:
			err = ErrDefault403{respErr}
			if error403er, ok := errType.(Err403er); ok {
				err = error403er.Error403(respErr)
			}
		case http.StatusNotFound:
			err = ErrDefault404{respErr}
			if error404er, ok := errType.(Err404er); ok {
				err = error404er.Error404(respErr)
			}
		case http.StatusMethodNotAllowed:
			err = ErrDefault405{respErr}
			if error405er, ok := errType.(Err405er); ok {
				err = error405er.Error405(respErr)
			}
		case http.StatusRequestTimeout:
			err = ErrDefault408{respErr}
			if error408er, ok := errType.(Err408er); ok {
				err = error408er.Error408(respErr)
			}
		case 429:
			err = ErrDefault429{respErr}
			if error429er, ok := errType.(Err429er); ok {
				err = error429er.Error429(respErr)
			}
		case http.StatusInternalServerError:
			err = ErrDefault500{respErr}
			if error500er, ok := errType.(Err500er); ok {
				err = error500er.Error500(respErr)
			}
		case http.StatusServiceUnavailable:
			err = ErrDefault503{respErr}
			if error503er, ok := errType.(Err503er); ok {
				err = error503er.Error503(respErr)
			}
		}

		if err == nil {
			err = respErr
		}

		return resp, err
	}

	return resp, nil
}

func (c *HTTPClient) Get(url string, opts *RequestOpts) Result {
	if opts == nil {
		opts = new(RequestOpts)
	}
	c.initReqOpts(url, nil, opts)
	return ResultFrom(c.Request("GET", url, opts))
}

func (c *HTTPClient) Put(url string, JSONBody interface{}, opts *RequestOpts) Result {
	if opts == nil {
		opts = new(RequestOpts)
	}
	c.initReqOpts(url, JSONBody, opts)
	return ResultFrom(c.Request("PUT", url, opts))
}

func (c *HTTPClient) Post(url string, JSONBody interface{}, opts *RequestOpts) Result {
	if opts == nil {
		opts = new(RequestOpts)
	}
	c.initReqOpts(url, JSONBody, opts)
	return ResultFrom(c.Request("POST", url, opts))
}

func (c *HTTPClient) Delete(url string, opts *RequestOpts) error {
	if opts == nil {
		opts = new(RequestOpts)
	}
	c.initReqOpts(url, nil, opts)
	resp, err := c.Request("DELETE", url, opts)
	defer resp.Body.Close()
	return err
}

/*
Result is an internal type to be used by individual resource packages, but its
methods will be available on a wide variety of user-facing embedding types.

It acts as a base struct that other Result types, returned from request
functions, can embed for convenience. All Results capture basic information
from the HTTP transaction that was performed, including the response body,
HTTP headers, and any errors that happened.

Generally, each Result type will have an Extract method that can be used to
further interpret the result's payload in a specific context. Extensions or
providers can then provide additional extraction functions to pull out
provider- or extension-specific information as well.
*/
type Result struct {
	// Body is the payload of the HTTP response from the server. In most cases,
	// this will be the deserialized JSON structure.
	Body interface{}

	// Header contains the HTTP header structure from the original response.
	Header http.Header

	// Err is an error that occurred during the operation. It's deferred until
	// extraction to make it easier to chain the Extract call.
	Err error
}

// ExtractInto allows users to provide an object into which `Extract` will extract
// the `Result.Body`. This would be useful for OpenStack providers that have
// different fields in the response object than OpenStack proper.
func (r Result) ExtractInto(to interface{}) error {
	if r.Err != nil {
		return r.Err
	}

	if reader, ok := r.Body.(io.Reader); ok {
		if readCloser, ok := reader.(io.Closer); ok {
			defer readCloser.Close()
		}
		return json.NewDecoder(reader).Decode(to)
	}

	b, err := json.Marshal(r.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, to)

	return err
}

func (r Result) extractIntoPtr(to interface{}, label string) error {
	if label == "" {
		return r.ExtractInto(&to)
	}

	var m map[string]interface{}
	err := r.ExtractInto(&m)
	if err != nil {
		return err
	}

	b, err := json.Marshal(m[label])
	if err != nil {
		return err
	}

	toValue := reflect.ValueOf(to)
	if toValue.Kind() == reflect.Ptr {
		toValue = toValue.Elem()
	}

	switch toValue.Kind() {
	case reflect.Slice:
		typeOfV := toValue.Type().Elem()
		if typeOfV.Kind() == reflect.Struct {
			if typeOfV.NumField() > 0 && typeOfV.Field(0).Anonymous {
				newSlice := reflect.MakeSlice(reflect.SliceOf(typeOfV), 0, 0)
				newType := reflect.New(typeOfV).Elem()

				for _, v := range m[label].([]interface{}) {
					b, err := json.Marshal(v)
					if err != nil {
						return err
					}

					for i := 0; i < newType.NumField(); i++ {
						s := newType.Field(i).Addr().Interface()
						err = json.NewDecoder(bytes.NewReader(b)).Decode(s)
						if err != nil {
							return err
						}
					}
					newSlice = reflect.Append(newSlice, newType)
				}
				toValue.Set(newSlice)
			}
		}
	case reflect.Struct:
		typeOfV := toValue.Type()
		if typeOfV.NumField() > 0 && typeOfV.Field(0).Anonymous {
			for i := 0; i < toValue.NumField(); i++ {
				toField := toValue.Field(i)
				if toField.Kind() == reflect.Struct {
					s := toField.Addr().Interface()
					err = json.NewDecoder(bytes.NewReader(b)).Decode(s)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	err = json.Unmarshal(b, &to)
	return err
}

// ExtractIntoStructPtr will unmarshal the Result (r) into the provided
// interface{} (to).
//
// NOTE: For internal use only
//
// `to` must be a pointer to an underlying struct type
//
// If provided, `label` will be filtered out of the response
// body prior to `r` being unmarshalled into `to`.
func (r Result) ExtractIntoStructPtr(to interface{}, label string) error {
	if r.Err != nil {
		return r.Err
	}

	t := reflect.TypeOf(to)
	if k := t.Kind(); k != reflect.Ptr {
		return fmt.Errorf("Expected pointer, got %v", k)
	}
	switch t.Elem().Kind() {
	case reflect.Struct:
		return r.extractIntoPtr(to, label)
	default:
		return fmt.Errorf("Expected pointer to struct, got: %v", t)
	}
}

// ExtractIntoSlicePtr will unmarshal the Result (r) into the provided
// interface{} (to).
//
// NOTE: For internal use only
//
// `to` must be a pointer to an underlying slice type
//
// If provided, `label` will be filtered out of the response
// body prior to `r` being unmarshalled into `to`.
func (r Result) ExtractIntoSlicePtr(to interface{}, label string) error {
	if r.Err != nil {
		return r.Err
	}

	t := reflect.TypeOf(to)
	if k := t.Kind(); k != reflect.Ptr {
		return fmt.Errorf("Expected pointer, got %v", k)
	}
	switch t.Elem().Kind() {
	case reflect.Slice:
		return r.extractIntoPtr(to, label)
	default:
		return fmt.Errorf("Expected pointer to slice, got: %v", t)
	}
}

func ResultFrom(resp *http.Response, err error) Result {
	defer resp.Body.Close()

	if err != nil {
		return Result{Body: nil, Err: err}
	}

	var parsedBody interface{}

	rawBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Result{Body: nil, Err: err}
	}

	if strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		err = json.Unmarshal(rawBody, &parsedBody)
		if err != nil {
			return Result{Body: nil, Err: err}
		}
	} else {
		parsedBody = rawBody
	}

	return Result{
		Body:   parsedBody,
		Header: resp.Header,
	}
}
