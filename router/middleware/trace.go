package middleware

import (
	// "github.com/gin-contrib/tracing"
	// "github.com/gin-gonic/gin"
	"os"

	stdopentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin/zipkin-go-opentracing"
)

var Tracer stdopentracing.Tracer

func InitTracer(url string) error {
	if url != "" {
		collector, err := zipkinot.NewHTTPCollector(url)
		if err != nil {
			return err
		}
		defer collector.Close()
		var (
			debug       = false
			hostPort    = "localhost:80"
			serviceName = "walm"
		)
		recorder := zipkinot.NewRecorder(collector, debug, hostPort, serviceName)
		Tracer, err = zipkinot.NewTracer(recorder)
		if err != nil {
			return err
		}
	} else {
		return error.New("zipin url is none")
	}
}
