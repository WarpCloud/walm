package helper

import (
	"WarpCloud/walm/pkg/util/compression"
	"errors"
	"strings"
)

const (
	GzipPrefix = "{gzip}"
)

var C = func(s string) (string, error) {
	res, err := compression.GzipCompress(s)
	if err != nil {
		return "", err
	}
	return strings.Join([]string{GzipPrefix, res}, ""), nil
}

var D = func(s string, err error) (string, error) {
	if err != nil {
		return "", err
	}
	if s == GzipPrefix {
		return "", errors.New("invalid format: a prefix {gzip} is given but no valid content is present")
	}
	// TODO: support extracting prefix and dispatching to different implementations
	if len(s) < len(GzipPrefix) {
		return s, nil
	} else if s[:len(GzipPrefix)] == GzipPrefix {
		return compression.GzipDecompress(s[len(GzipPrefix):])
	}
	// for backward compatibility
	return s, nil
}
