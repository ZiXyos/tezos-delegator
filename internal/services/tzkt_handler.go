package services

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type HTTPHandler struct {
	logger  *slog.Logger
	client  *http.Client
	baseURL string
}

type HandlerOptions func(*HTTPHandler)

func HandlerWithLogger(logger *slog.Logger) HandlerOptions {
	return func(h *HTTPHandler) {
		h.logger = logger
	}
}

func HandlerWithClient(client *http.Client) HandlerOptions {
	return func(h *HTTPHandler) {
		h.client = client
	}
}

func HandlerWithBaseURL(baseURL string) HandlerOptions {
	return func(h *HTTPHandler) {
		h.baseURL = baseURL
	}
}

func (h *HTTPHandler) GetDelegations() (any, error) {
	// https://api.tzkt.io/#operation/Operations_GetDelegations
	res, err := h.client.Get(h.baseURL + "operations/delegations?limit=1000&sort.desc=id")

	if err != nil {
		h.logger.Warn("error getting delegations: ", err)
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			h.logger.Warn("error closing body: ", err)
			return
		}
	}(res.Body)

	if res.StatusCode != http.StatusOK {
		h.logger.Warn("bad status code", "status", res.StatusCode)
		return nil, fmt.Errorf("API returned status %d", res.StatusCode)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		h.logger.Warn("error reading delegations body: ", err)
		return nil, err
	}

	return data, nil

}

func NewHTTPHandler(opts ...HandlerOptions) *HTTPHandler {
	h := &HTTPHandler{}
	for _, opt := range opts {
		opt(h)
	}

	return h
}
