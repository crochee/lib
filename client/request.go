package client

import (
	"bytes"
	"context"
	"net/http"

	"moul.io/http2curl"

	"github.com/crochee/lirity/log"
)

var DefaultClient Client = NewStandardClient()

// Send request with default client
func Send(ctx context.Context, method string, uri string,
	body []byte, headers http.Header) (*http.Response, error) {
	req, err := NewRequest(ctx, method, uri, body, headers)
	if err != nil {
		return nil, err
	}
	return Do(req)
}

// Do request with default client
func Do(req *http.Request) (*http.Response, error) {
	return DefaultClient.Do(req)
}

// NewRequest create request
func NewRequest(ctx context.Context, method string, uri string,
	body []byte, headers http.Header) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, uri, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	for key, header := range headers {
		for _, value := range header {
			req.Header.Add(key, value)
		}
	}
	// 打印curl语句，便于问题分析和定位
	var curl *http2curl.CurlCommand
	if curl, err = http2curl.GetCurlCommand(req); err == nil {
		log.FromContext(ctx).Debug(curl.String())
	} else {
		log.FromContext(ctx).Error(err.Error())
	}
	return req, nil
}
