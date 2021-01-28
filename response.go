package main

import (
	"io"
	"net/http"
)

// Response is the representation of a HTTP response made by a Collector
type Response struct {
	// StatusCode is the status code of the Response
	StatusCode int
	// Body is the content of the Response
	Body io.ReadCloser
	// Request is the Request object of the response
	Request *http.Request
	// Headers contains the Response's HTTP headers
	Headers *http.Header
	// Trace contains the HTTPTrace for the request.
	Trace *HTTPTrace
}

func (response *Response) generateReport() *Report {
	report := &Report{
		url:              response.Request.URL.String(),
		StatusCode:       response.StatusCode,
		DNSLookup:        response.Trace.DNSLookup(),
		TCPConnection:    response.Trace.TCPConnection(),
		TLSHandshake:     response.Trace.TLSHandshake(),
		ServerProcessing: response.Trace.ServerProcessing(),
		ContentTransfer:  response.Trace.ContentTransfer(),
		NameLookup:       response.Trace.NameLookup(),
		Connect:          response.Trace.Connect(),
		PreTransfer:      response.Trace.PreTransfer(),
		StartTransfer:    response.Trace.StartTransfer(),
		Total:            response.Trace.Total()}

	return report
}
