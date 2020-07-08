package eslogger

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config includes the elasticsearch connection information and index name
type Config struct {
	Addresses []string
	Headers   map[string]string
	Index     string
	Password  string
	Username  string
}

// This allows elasticsearch to add custom headers to each request
// https://github.com/elastic/go-elasticsearch/blob/8413c97f30112984737796ca9db1e93a11fe7e5a/_examples/customization.go#L20
// https://stackoverflow.com/questions/41229694/how-to-add-headers-info-using-transport-in-golang-net-http
type transport struct {
	http.RoundTripper
	headers map[string]string
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, value := range t.headers {
		req.Header.Add(key, value)
	}

	return http.DefaultTransport.RoundTrip(req)
}

type writer struct {
	es    *elasticsearch.Client
	index string
}

func (w *writer) Write(reqBody []byte) (n int, err error) {
	// CODE_SMELL: Setting returned number of bytes to always be equal to the
	// request's. Was consistently getting a "sort write" error when using
	// MultiWriter.
	n = len(reqBody)

	index := w.index
	if index == "" {
		return n, errors.New("you must provide an index")
	}

	if w.es == nil {
		return n, errors.New("elasticsearch.Client is nil")
	}

	year, month, day := time.Now().Date()
	idx := fmt.Sprintf("%s-%d-%d-%d", index, year, int(month), day)

	res, err := w.es.Index(idx, strings.NewReader(string(reqBody)))
	if err != nil {
		return n, errors.Wrap(err, "failed to index log entry")
	}

	if res.HasWarnings() {
		for _, warning := range res.Warnings() {
			log.Warn().Msg(warning)
		}
	}

	if res.IsError() {
		return n, fmt.Errorf("ERROR: %v", res.Status())
	}

	return n, err
}

// New returns a zerolog.Logger, which transports to an elasticsearch instance
// using the passed-in config struct.
func New(conf Config) (logger zerolog.Logger, err error) {
	c := elasticsearch.Config{
		Addresses: conf.Addresses,
		Username:  conf.Username,
		Password:  conf.Password,
		Transport: &transport{headers: conf.Headers},
	}

	es, err := elasticsearch.NewClient(c)
	if err != nil {
		return logger, errors.Wrap(err, "failed to initialize elasticsearch client")
	}

	w := &writer{es: es, index: conf.Index}

	return zerolog.New(w).With().Timestamp().Logger(), nil
}
