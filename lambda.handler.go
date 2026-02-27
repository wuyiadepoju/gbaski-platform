package main

import (
	"context"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// Handler function for Lambda with HTTP API support
func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	// Convert headers to multi-value headers format
	multiValueHeaders := make(map[string][]string)

	for key, value := range req.Headers {
		// Convert header keys to canonical format
		canonicalKey := strings.ToLower(key)
		if _, exists := multiValueHeaders[canonicalKey]; !exists {
			multiValueHeaders[canonicalKey] = []string{}
		}
		multiValueHeaders[canonicalKey] = append(multiValueHeaders[canonicalKey], value)
	}

	// multiValueHeaders["content-type"] = []string{"application/json"}

	// Convert query parameters to multi-value format
	multiValueQueryParams := make(map[string][]string)
	for key, value := range req.QueryStringParameters {
		multiValueQueryParams[key] = []string{value}
	}

	proxyReq := events.APIGatewayProxyRequest{
		Resource:                        req.RawPath,
		Path:                            req.RawPath,
		HTTPMethod:                      req.RequestContext.HTTP.Method,
		Headers:                         req.Headers,
		MultiValueHeaders:               multiValueHeaders,
		QueryStringParameters:           req.QueryStringParameters,
		MultiValueQueryStringParameters: multiValueQueryParams,
		PathParameters:                  req.PathParameters,
		StageVariables:                  req.StageVariables,
		Body:                            req.Body,
		IsBase64Encoded:                 req.IsBase64Encoded,
		RequestContext: events.APIGatewayProxyRequestContext{
			AccountID:  req.RequestContext.AccountID,
			RequestID:  req.RequestContext.RequestID,
			Stage:      req.RequestContext.Stage,
			HTTPMethod: req.RequestContext.HTTP.Method,
		},
	}

	// Add cookies to the headers if present
	if len(req.Cookies) > 0 {
		cookies := strings.Join(req.Cookies, "; ")
		proxyReq.Headers["cookie"] = cookies
		proxyReq.MultiValueHeaders["cookie"] = []string{cookies}
	}

	// Use the Fiber adapter
	resp, err := fiberLambda.ProxyWithContext(ctx, proxyReq)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       `{"error": "Internal Server Error"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, err
	}

	// Ensure response headers are properly set
	if resp.Headers == nil {
		resp.Headers = make(map[string]string)
	}

	// Convert multi-value headers to single value headers for V2 response
	headers := make(map[string]string)
	for k, v := range resp.MultiValueHeaders {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}
	for k, v := range resp.Headers {
		headers[k] = v
	}

	// Extract cookies from the response if present
	cookies := []string{}
	if setCookieValues, exists := resp.MultiValueHeaders["set-cookie"]; exists && len(setCookieValues) > 0 {
		cookies = setCookieValues
	} else if setCookie, exists := resp.Headers["Set-Cookie"]; exists {
		cookies = append(cookies, setCookie)
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode:      resp.StatusCode,
		Headers:         headers,
		Body:            resp.Body,
		IsBase64Encoded: resp.IsBase64Encoded,
		Cookies:         cookies,
	}, nil
}
