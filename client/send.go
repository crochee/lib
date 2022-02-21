package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/crochee/lirity/e"
	jsoniter "github.com/json-iterator/go"
)

type clientHandler struct {
	client  Client
	api     jsoniter.API
	handler URLHandler
}

func (h clientHandler) Send(ctx context.Context, method, uri string, value url.Values, header http.Header,
	request, response interface{}, wantStatusCode int, responseCode e.ErrorCode) error {
	var (
		body []byte
		err  error
	)
	if request != nil {
		if body, err = h.api.Marshal(request); err != nil {
			return err
		}
	}
	var req *http.Request
	if req, err = NewRequest(ctx, method, h.handler.URLWithQuery(ctx, uri, value), body, h.handler.Header(ctx, header)); err != nil {
		return err
	}
	var httpResponse *http.Response
	if httpResponse, err = h.client.Do(req); err != nil {
		return err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode != wantStatusCode {
		if err = h.api.NewDecoder(httpResponse.Body).Decode(responseCode); err != nil {
			return fmt.Errorf("http code %d,but not %d,%w", httpResponse.StatusCode, wantStatusCode, err)
		}
		return responseCode
	}
	if httpResponse.StatusCode == http.StatusNoContent {
		return nil
	}
	return h.api.NewDecoder(httpResponse.Body).Decode(response)
}

func (h clientHandler) Get(ctx context.Context, uri string, value url.Values, header http.Header,
	response interface{}, wantStatusCode int, responseCode e.ErrorCode) error {
	return h.Send(ctx, http.MethodGet, uri, value, header, nil, response, wantStatusCode, responseCode)
}

func (h clientHandler) DefaultGet(ctx context.Context, uri string, value url.Values, header http.Header, response interface{}) error {
	return h.Get(ctx, uri, value, header, response, http.StatusOK, &e.ErrCode{})
}
