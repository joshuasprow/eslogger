# ElasticSearch Logger

This is a golang package that sends logs to an ElasticSearch instance through
a custom [zerolog]("https://github.com/rs/zerolog") transport.

### Basic Usage (with zerolog integration)

```golang
package main

import (
	"github.com/joshuasprow/eslogger"
	"github.com/rs/zerolog/log"
)

func main() {
	conf := eslogger.Config{
		Index: "test-index", // Will be appended with "-year-month-date"
		Addresses: []string{
			"http://localhost:9200",
			"http://localhost:9201",
		},
		Username: "me",
		Password: "supersecretpassword",
		Headers: map[string]string{
			"Auth-Client-Id":     "stringofnastycharacters",
			"Auth-Client-Secret": "stringofnastycharacters",
		},
	}

	logger, err := eslogger.New(conf)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize logger")
	}

	log.Logger = logger

	log.Info().Msg("I'm logging to ElasticSearch!")
}
```
