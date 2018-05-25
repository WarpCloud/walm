package middleware

import (
	// "github.com/gin-contrib/tracing"
	// "github.com/gin-gonic/gin"
	"errors"

	stdopentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin/zipkin-go-opentracing"
)

var Tracer stdopentracing.Tracer
var collector zipkinot.Collector

func InitTracer(url string) error {
	var err error
	if url != "" {
		collector, err = zipkinot.NewHTTPCollector(url)
		if err != nil {
			return err
		}

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
		return errors.New("zipin url is none")
	}
	return nil
}

/*
func EnableTrace() gin.HandlerFunc {
	return trace.SpanFromHeaders(Tracer, "Walm", stdopentracing.ChildOf, false)
}
*/

func EndTrace() {
	defer collector.Close()
}
