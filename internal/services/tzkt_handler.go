package services

import (
	"delegator/pkg/domain"
	"encoding/json"
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

func (h *HTTPHandler) GetDelegations() ([]domain.TzktApiDelegationsResponse, error) {
	return h.GetDelegationsFromLevel(0, 100)
}

func (h *HTTPHandler) GetDelegationsFromLevel(lastLevel int64, limit int) ([]domain.TzktApiDelegationsResponse, error) {
	var url string
	if lastLevel > 0 {
		url = fmt.Sprintf("%soperations/delegations?level.gt=%d&limit=%d&sort.asc=level", h.baseURL, lastLevel, limit)
	} else {
		url = fmt.Sprintf("%soperations/delegations?limit=%d&sort.desc=level", h.baseURL, limit)
	}

	h.logger.Info("fetching delegations", "url", url, "lastLevel", lastLevel, "limit", limit)

	res, err := h.client.Get(url)
	if err != nil {
		h.logger.Warn("error getting delegations", "error", err)
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			h.logger.Warn("error closing body", "error", err)
			return
		}
	}(res.Body)

	if res.StatusCode != http.StatusOK {
		h.logger.Warn("bad status code", "status", res.StatusCode)
		return nil, fmt.Errorf("API returned status %d", res.StatusCode)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		h.logger.Warn("error reading delegations body", "error", err)
		return nil, err
	}

	var response []domain.TzktApiDelegationsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		h.logger.Warn("error unmarshaling delegations", "error", err)
		return nil, fmt.Errorf("failed to unmarshal delegations: %w", err)
	}

	h.logger.Info("fetched delegations", "count", len(response))
	return response, nil
}

func NewHTTPHandler(opts ...HandlerOptions) *HTTPHandler {
	h := &HTTPHandler{}
	for _, opt := range opts {
		opt(h)
	}

	return h
}
