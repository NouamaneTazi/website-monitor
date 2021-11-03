package inspect

import (
	"context"
	"crypto/tls"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"
)

// Report collects useful metrics from a single HTTP request made by an Inspector
type Report struct {
	url              string
	StatusCode       int
	DNSLookup        time.Duration
	TCPConnection    time.Duration
	TLSHandshake     time.Duration
	ServerProcessing time.Duration
	ContentTransfer  time.Duration
	NameLookup       time.Duration
	Connect          time.Duration
	PreTransfer      time.Duration
	StartTransfer    time.Duration
	Total            time.Duration
}

// Inspector provides the instance for monitoring a single url at inspectionInterval
type Inspector struct {
	// UserAgent is the User-Agent string used by HTTP requests
	UserAgent          string
	wg                 *sync.WaitGroup
	lock               *sync.RWMutex
	url                string
	reports            []*Report
	inspectionInterval time.Duration
}

// A InspectorOption sets an option on a Inspector.
type InspectorOption func(*Inspector)

// NewInspector creates a new Inspector instance with default configuration
func NewInspector(options ...InspectorOption) *Inspector {
	inspector := &Inspector{}
	inspector.Init()

	for _, f := range options {
		f(inspector)
	}

	return inspector
}

// URL sets the url to be monitored by the Inspector.
func URL(url string) InspectorOption {
	return func(inspector *Inspector) {
		inspector.url = url
	}
}

// intervalInspection sets the interval at which the url will be monitored.
func intervalInspection(interval time.Duration) InspectorOption {
	return func(inspector *Inspector) {
		inspector.inspectionInterval = interval
		maxNumOfReports := int(maxHistoryPerURL / interval) // TODO: this is only an estimation
		inspector.reports = make([]*Report, 0, maxNumOfReports)
		for i := 0; i < cap(inspector.reports); i++ {
			inspector.reports = append(inspector.reports, &Report{})
		}
	}
}

// Init initializes the Inspector's private variables and sets default
// configuration for the Inspector
func (inspector *Inspector) Init() {
	inspector.UserAgent = "Go Website Monitor - https://github.com/NouamaneTazi/website-monitor"
	inspector.wg = &sync.WaitGroup{}
	inspector.lock = &sync.RWMutex{}
}

// startLoop starts monitoring loop
func (inspector *Inspector) startLoop() {
	for {
		inspector.visit(inspector.url)
		time.Sleep(inspector.inspectionInterval)
	}
}

// visit visits a url and times the interaction.
// If the response is a 30x, visit follows the redirect.
func (inspector *Inspector) visit(url string) {
	// println("Visiting", url)

	// Creates request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Panicf("failed to create http request: %v", err)
	}
	// Add http tracing
	httpTrace := &HTTPTrace{}
	req = req.WithContext(httptrace.WithClientTrace(context.Background(), httpTrace.trace()))

	// Sends http request
	resp, err := inspector.Do(req)
	if err != nil {
		log.Panicf("failed to read response: %v", err)
	}
	// Reads and discard body and get timing
	inspector.readResponseBody(req, resp)
	httpTrace.GotResponseBody = time.Now()

	resp.Request = req
	resp.Trace = httpTrace

	// Update url reports
	inspector.updateURLReports(url, resp.generateReport())
}

// updateURLReports updates URL reports with useful metrics about website
// a single http request generates a single report
// we drop reports older than maxHistoryPerURL
func (inspector *Inspector) updateURLReports(url string, report *Report) {
	queue := inspector.reports
	queue = queue[1:] // TODO: make sure we reallocate memory
	inspector.reports = append(queue, report)
}

// Do sends an HTTP request and returns an HTTP response, following policy (such as redirects, cookies, auth) as configured on the client.
func (inspector *Inspector) Do(request *http.Request) (*Response, error) {

	//TODO: transport configuration (add timeout)
	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	client := &http.Client{Transport: tr}

	res, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode: res.StatusCode,
		Headers:    &res.Header,
		Body:       res.Body,
	}, nil
}

// readResponseBody consumes the body of the response.
// readResponseBody returns an informational message about the
// disposition of the response body's contents.
func (inspector *Inspector) readResponseBody(req *http.Request, resp *Response) string {
	if isRedirect(resp) || req.Method == http.MethodHead {
		return ""
	}

	w := ioutil.Discard
	msg := "Body was replaced with this text"

	if _, err := io.Copy(w, resp.Body); err != nil && w != ioutil.Discard {
		log.Panicf("failed to read response body: %v", err)
	}
	defer resp.Body.Close()
	return msg
}

func isRedirect(resp *Response) bool {
	return resp.StatusCode > 299 && resp.StatusCode < 400
}
