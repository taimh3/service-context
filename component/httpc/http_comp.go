package httpc

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	sctx "github.com/taimaifika/service-context"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type config struct {
	timeout         time.Duration
	maxIdleConn     int
	idleConnTimeout time.Duration
}

type HTTPComponent struct {
	id string
	*config

	client *http.Client
}

func NewHTTPComponent(id string) *HTTPComponent {
	return &HTTPComponent{
		id:     id,
		config: new(config),
	}
}

func (h *HTTPComponent) ID() string {
	return h.id
}

func (h *HTTPComponent) InitFlags() {
	flag.DurationVar(&h.timeout, h.id+"-timeout", time.Second*5, "req http timeout")
	flag.IntVar(&h.maxIdleConn, h.id+"-max-idle-conn", 100, "max idle connections for http client")
	flag.DurationVar(&h.idleConnTimeout, h.id+"-idle-conn-timeout", 90*time.Second, "idle connection timeout for http client")
}

func (h *HTTPComponent) Activate(_ sctx.ServiceContext) error {
	h.client = &http.Client{
		Timeout: h.timeout,
		Transport: otelhttp.NewTransport(
			&http.Transport{
				MaxIdleConns:       h.config.maxIdleConn,
				IdleConnTimeout:    h.config.idleConnTimeout,
				DisableCompression: true,
			},
		),
	}

	return nil
}

func (h *HTTPComponent) Stop() error {
	return nil
}

func (h *HTTPComponent) GetHttpClient() *http.Client {
	return h.client
}

type ReqOption struct {
	Body         io.Reader
	Header       http.Header
	Timeout      time.Duration
	TraceReqBody bool
}

type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

func (h *HTTPComponent) MakeRequest(ctx context.Context, method string, reqUrl string, opt *ReqOption, output interface{}) (*Response, error) {
	ctx, span := otel.Tracer("httpClient").Start(ctx, "MakeRequest")
	defer span.End()

	if opt.Body != nil {
		bodyBytes, err := io.ReadAll(opt.Body)
		if err != nil {
			return nil, err
		}
		// Create a new reader for the request since we consumed the original
		opt.Body = bytes.NewReader(bodyBytes)
		if opt.TraceReqBody {
			span.SetAttributes(
				attribute.String("http.request.body", string(bodyBytes)),
				attribute.String("http.request.body.size", strconv.Itoa(len(bodyBytes))),
			)
		}
	}

	// Apply timeout if specified in options
	if opt.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opt.Timeout)
		defer cancel()
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, method, reqUrl, opt.Body)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	// Add headers to the request
	for key, arr := range opt.Header {
		for _, value := range arr {
			req.Header.Add(key, value)
		}
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	// Read response body
	buffer, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// Close response body
	defer func() {
		if clsErr := resp.Body.Close(); clsErr != nil {
			err = clsErr
		}
	}()

	// If output is provided, unmarshal the response body into it
	if output != nil {
		if err = json.Unmarshal(buffer, &output); err != nil {
			return nil, err
		}
	}
	if opt.TraceReqBody {
		span.SetAttributes(
			attribute.Int("http.response.status_code", resp.StatusCode),
			attribute.String("http.response.body", string(buffer)),
			attribute.String("http.response.body.size", strconv.Itoa(len(buffer))),
		)
	}
	// Return response
	return &Response{
		StatusCode: resp.StatusCode,
		Body:       buffer,
		Header:     resp.Header.Clone(),
	}, err
}

func (h *HTTPComponent) MakeRequestWithProxy(ctx context.Context, method string, reqUrl string, proxy string, opt *ReqOption, output interface{}) (*Response, error) {
	ctx, span := otel.Tracer("httpClient").Start(ctx, "MakeRequestWithProxy")
	defer span.End()

	if opt.Body != nil {
		bodyBytes, err := io.ReadAll(opt.Body)
		if err != nil {
			return nil, err
		}
		// Create a new reader for the request since we consumed the original
		opt.Body = bytes.NewReader(bodyBytes)
		if opt.TraceReqBody {
			span.SetAttributes(
				attribute.String("http.request.body", string(bodyBytes)),
				attribute.String("http.request.body.size", strconv.Itoa(len(bodyBytes))),
			)
		}
	}

	// Apply timeout if specified in options
	if opt.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opt.Timeout)
		defer cancel()
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, method, reqUrl, opt.Body)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	// Add headers to the request
	for key, arr := range opt.Header {
		for _, value := range arr {
			req.Header.Add(key, value)
		}
	}

	// Create a new HTTP client with the proxy settings
	clientWithProxy := &http.Client{
		Timeout: h.timeout,
		Transport: otelhttp.NewTransport(
			&http.Transport{
				MaxIdleConns:       h.config.maxIdleConn,
				IdleConnTimeout:    h.config.idleConnTimeout,
				DisableCompression: true,
				Proxy: func(_ *http.Request) (*url.URL, error) {
					proxy, err := url.Parse(proxy)
					if err != nil {
						span.SetStatus(codes.Error, err.Error())
						span.RecordError(err)
						return nil, err
					}
					return proxy, nil
				},
			}),
	}

	// Make the HTTP request using the client with proxy
	resp, err := clientWithProxy.Do(req)
	if err != nil {
		return nil, err
	}
	// Read response body
	buffer, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// Close response body
	defer func() {
		if clsErr := resp.Body.Close(); clsErr != nil {
			err = clsErr
		}
	}()

	// If output is provided, unmarshal the response body into it
	if output != nil {
		if err = json.Unmarshal(buffer, &output); err != nil {
			return nil, err
		}
	}
	if opt.TraceReqBody {
		span.SetAttributes(
			attribute.Int("http.response.status_code", resp.StatusCode),
			attribute.String("http.response.body", string(buffer)),
			attribute.String("http.response.body.size", strconv.Itoa(len(buffer))),
		)
	}
	// Return response
	return &Response{
		StatusCode: resp.StatusCode,
		Body:       buffer,
		Header:     resp.Header.Clone(),
	}, err
}
