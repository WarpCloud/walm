package compression

import (
	"bytes"
	"compress/gzip"
	"io"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	GaugeCompressionRate = "compression_rate"
)

var (
	TotalPlainSize prometheus.Gauge

	TotalCompressed prometheus.Gauge

	Compressions prometheus.Counter
)

func GzipCompress(s string) (string, error) {
	buf := bytes.NewBuffer(make([]byte, 0, len(s)))
	writer := gzip.NewWriter(buf)
	_, err := writer.Write([]byte(s))
	if err != nil {
		return "", err
	}
	err = writer.Close()
	if err != nil {
		return "", err
	}
	res, err := io.ReadAll(buf)
	if err != nil {
		return "", err
	}

	TotalPlainSize.Add(float64(len(s)) / 1024)
	TotalCompressed.Add(float64(len(res)) / 1024)
	Compressions.Inc()

	return string(res), nil
}

func GzipDecompress(compressed string) (string, error) {
	buf := bytes.NewBuffer([]byte(compressed))
	r, err := gzip.NewReader(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	res, err := io.ReadAll(r)
	if err != nil && err != io.EOF {
		return "", err
	}
	return string(res), nil
}

func init() {
	TotalPlainSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "walm",
		Subsystem: "gzip",
		Name:      "plain_size_ki",
	})

	TotalCompressed = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "walm",
		Subsystem: "gzip",
		Name:      "compressed_size_ki",
	})

	Compressions = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "walm",
		Subsystem: "gzip",
		Name:      "runs",
	})

	prometheus.Register(TotalPlainSize)
	prometheus.Register(TotalCompressed)
	prometheus.Register(Compressions)
}
