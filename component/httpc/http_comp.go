package httpc

import (
	"bytes"
	"context"
	"flag"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/getsentry/sentry-go/attribute"
	sctx "github.com/taimaifika/service-context"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type config struct {
	httpProxy  string
	httpsProxy string

	timeout         time.Duration
	maxIdleConn     int
	idleConnTimeout time.Duration
}

type HTTPComponent struct {
	id string
	*config

	client          *http.Client
	clientWithProxy *http.Client
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
	flag.StringVar(&h.httpProxy, h.id+"-proxy-http", "", "http proxy url")
	flag.StringVar(&h.httpsProxy, h.id+"-proxy-https", "", "https proxy url")
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

	h.clientWithProxy = &http.Client{
		Timeout: h.timeout, // Set a timeout for HTTP requests
		Transport: otelhttp.NewTransport(
			&http.Transport{
				MaxIdleConns:       h.config.maxIdleConn,
				IdleConnTimeout:    h.config.idleConnTimeout,
				DisableCompression: true,
				Proxy: func(req *http.Request) (*url.URL, error) {
					switch req.URL.Scheme {
					case "http":
						httpProxy, err := url.Parse(h.httpProxy)
						if err != nil {
							return nil, err
						}
						return httpProxy, nil
					case "https":
						httpsProxy, err := url.Parse(h.httpsProxy)
						if err != nil {
							return nil, err
						}
						return httpsProxy, nil
					default:
						return nil, nil
					}
				},
			}),
	}

	return nil
}

func (h *HTTPComponent) Stop() error {
	return nil
}

func (h *HTTPComponent) GetHttpClient() *http.Client {
	return h.client
}

func (h *HTTPComponent) GetHttpClientWithProxy() *http.Client {
	return h.clientWithProxy
}

type ReqOption struct {
	Method     string // GET, POST, PUT, DELETE
	Body       io.Reader
	Header     http.Header
	UseProxy   bool
	Timeout    time.Duration
	TraceLevel string // e.g., "info", "debug"
}

type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

func (h *HTTPComponent) MakeRequest(ctx context.Context, url string, opt *ReqOption) (*Response, error) {
	ctx, span := otel.Tracer("httpClient").Start(ctx, "MakeRequest")
	defer span.End()

	if opt.Body != nil {
		bodyBytes, err := io.ReadAll(opt.Body)
		if err != nil {
			return nil, err
		}
		// Create a new reader for the request since we consumed the original
		opt.Body = bytes.NewReader(bodyBytes)
		if opt.TraceLevel == "debug" {
			span.SetAttributes(
				attribute.String("http.request.body", string(bodyBytes)),
			)
		}
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, opt.Method, url, opt.Body)
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

	var client *http.Client
	if opt.UseProxy {
		client = h.clientWithProxy
	} else {
		client = h.client
	}
	resp, err := client.Do(req)
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

	if opt.TraceLevel == "debug" {
		span.SetAttributes(
			attribute.Int("http.response.status_code", resp.StatusCode),
			attribute.String("http.response.body", string(buffer)),
		)
	}
	// Return response
	return &Response{
		StatusCode: resp.StatusCode,
		Body:       buffer,
		Header:     resp.Header.Clone(),
	}, err
}
