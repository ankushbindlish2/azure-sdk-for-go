// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package shared

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/stretchr/testify/require"
)

// mockTransport implements policy.Transporter for testing.
// It captures the request headers as seen by the transport layer.
type mockTransport struct {
	statusCode    int
	lastExpectHdr string
}

func (m *mockTransport) Do(req *http.Request) (*http.Response, error) {
	m.lastExpectHdr = req.Header.Get("Expect")
	return &http.Response{
		StatusCode: m.statusCode,
		Header:     http.Header{},
		Body:       http.NoBody,
		Request:    req,
	}, nil
}

func newTestPipeline(ecPolicy *ExpectContinuePolicy, transport *mockTransport) runtime.Pipeline {
	return runtime.NewPipeline("test", "v1.0.0",
		runtime.PipelineOptions{
			PerRetry: []policy.Policy{ecPolicy},
		},
		&policy.ClientOptions{
			Transport: transport,
			// MaxRetries -1 disables retries so tests isolate the policy behavior.
			Retry: policy.RetryOptions{MaxRetries: -1},
		},
	)
}

func newPutRequest(contentLength int64) (*policy.Request, error) {
	req, err := runtime.NewRequest(context.Background(), http.MethodPut, "https://example.blob.core.windows.net/container/blob")
	if err != nil {
		return nil, err
	}
	if contentLength > 0 {
		body := strings.NewReader(strings.Repeat("x", int(contentLength)))
		err = req.SetBody(readSeekCloser{body}, "application/octet-stream")
		if err != nil {
			return nil, err
		}
		req.Raw().ContentLength = contentLength
	}
	return req, nil
}

// readSeekCloser wraps a strings.Reader to implement io.ReadSeekCloser.
type readSeekCloser struct {
	*strings.Reader
}

func (r readSeekCloser) Close() error { return nil }

// ---------------------------------------------------------------------------
// Auto mode tests (default) — threshold 0 means all PUT requests with body
// ---------------------------------------------------------------------------

func TestAutoMode_InactiveByDefault(t *testing.T) {
	transport := &mockTransport{statusCode: http.StatusCreated}
	ecPolicy := NewExpectContinuePolicy(nil)
	pl := newTestPipeline(ecPolicy, transport)

	req, err := newPutRequest(100)
	require.NoError(t, err)

	resp, err := pl.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Empty(t, transport.lastExpectHdr)
}

func TestAutoMode_ActivatesOn429(t *testing.T) {
	transport := &mockTransport{statusCode: http.StatusTooManyRequests}
	ecPolicy := NewExpectContinuePolicy(nil)
	pl := newTestPipeline(ecPolicy, transport)

	req1, err := newPutRequest(100)
	require.NoError(t, err)
	_, err = pl.Do(req1)
	require.NoError(t, err)
	require.Empty(t, transport.lastExpectHdr)

	transport.statusCode = http.StatusCreated
	req2, err := newPutRequest(100)
	require.NoError(t, err)
	resp, err := pl.Do(req2)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, "100-continue", transport.lastExpectHdr)
}

func TestAutoMode_ActivatesOn500(t *testing.T) {
	transport := &mockTransport{statusCode: http.StatusInternalServerError}
	ecPolicy := NewExpectContinuePolicy(nil)
	pl := newTestPipeline(ecPolicy, transport)

	req1, err := newPutRequest(100)
	require.NoError(t, err)
	_, err = pl.Do(req1)
	require.NoError(t, err)
	require.Empty(t, transport.lastExpectHdr)

	transport.statusCode = http.StatusCreated
	req2, err := newPutRequest(100)
	require.NoError(t, err)
	resp, err := pl.Do(req2)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, "100-continue", transport.lastExpectHdr)
}

func TestAutoMode_ActivatesOn503(t *testing.T) {
	transport := &mockTransport{statusCode: http.StatusServiceUnavailable}
	ecPolicy := NewExpectContinuePolicy(nil)
	pl := newTestPipeline(ecPolicy, transport)

	req1, err := newPutRequest(100)
	require.NoError(t, err)
	_, err = pl.Do(req1)
	require.NoError(t, err)
	require.Empty(t, transport.lastExpectHdr)

	transport.statusCode = http.StatusCreated
	req2, err := newPutRequest(100)
	require.NoError(t, err)
	resp, err := pl.Do(req2)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, "100-continue", transport.lastExpectHdr)
}

func TestAutoMode_NotAppliedToGET(t *testing.T) {
	transport := &mockTransport{statusCode: http.StatusServiceUnavailable}
	ecPolicy := NewExpectContinuePolicy(nil)
	pl := newTestPipeline(ecPolicy, transport)

	// Activate via a PUT 503
	req1, err := newPutRequest(100)
	require.NoError(t, err)
	_, err = pl.Do(req1)
	require.NoError(t, err)

	transport.statusCode = http.StatusOK
	req2, err := runtime.NewRequest(context.Background(), http.MethodGet, "https://example.blob.core.windows.net/container/blob")
	require.NoError(t, err)

	resp, err := pl.Do(req2)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Empty(t, transport.lastExpectHdr)
}

func TestAutoMode_NotAppliedZeroContentLength(t *testing.T) {
	transport := &mockTransport{statusCode: http.StatusServiceUnavailable}
	ecPolicy := NewExpectContinuePolicy(nil)
	pl := newTestPipeline(ecPolicy, transport)

	req1, err := newPutRequest(100)
	require.NoError(t, err)
	_, err = pl.Do(req1)
	require.NoError(t, err)

	transport.statusCode = http.StatusCreated
	req2, err := runtime.NewRequest(context.Background(), http.MethodPut, "https://example.blob.core.windows.net/container/blob")
	require.NoError(t, err)
	req2.Raw().ContentLength = 0

	resp, err := pl.Do(req2)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Empty(t, transport.lastExpectHdr)
}

func TestAutoMode_AppliedToSmallBody(t *testing.T) {
	// With threshold 0, even a 1-byte PUT body qualifies
	transport := &mockTransport{statusCode: http.StatusServiceUnavailable}
	ecPolicy := NewExpectContinuePolicy(nil)
	pl := newTestPipeline(ecPolicy, transport)

	req1, err := newPutRequest(1)
	require.NoError(t, err)
	_, err = pl.Do(req1)
	require.NoError(t, err)

	transport.statusCode = http.StatusCreated
	req2, err := newPutRequest(1)
	require.NoError(t, err)
	resp, err := pl.Do(req2)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, "100-continue", transport.lastExpectHdr)
}

func TestAutoMode_ExpiresAfterDuration(t *testing.T) {
	currentTime := time.Now()
	ecPolicy := NewExpectContinuePolicy(nil)
	ecPolicy.now = func() time.Time { return currentTime }

	transport := &mockTransport{statusCode: http.StatusServiceUnavailable}
	pl := newTestPipeline(ecPolicy, transport)

	req1, err := newPutRequest(100)
	require.NoError(t, err)
	_, err = pl.Do(req1)
	require.NoError(t, err)

	currentTime = currentTime.Add(expectContinueDuration + time.Second)

	transport.statusCode = http.StatusCreated
	req2, err := newPutRequest(100)
	require.NoError(t, err)
	resp, err := pl.Do(req2)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Empty(t, transport.lastExpectHdr)
}

func TestAutoMode_ActiveWithinDuration(t *testing.T) {
	currentTime := time.Now()
	ecPolicy := NewExpectContinuePolicy(nil)
	ecPolicy.now = func() time.Time { return currentTime }

	transport := &mockTransport{statusCode: http.StatusServiceUnavailable}
	pl := newTestPipeline(ecPolicy, transport)

	req1, err := newPutRequest(100)
	require.NoError(t, err)
	_, err = pl.Do(req1)
	require.NoError(t, err)

	currentTime = currentTime.Add(expectContinueDuration - time.Second)

	transport.statusCode = http.StatusCreated
	req2, err := newPutRequest(100)
	require.NoError(t, err)
	resp, err := pl.Do(req2)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, "100-continue", transport.lastExpectHdr)
}

func TestAutoMode_ReactivatesOnSubsequentThrottle(t *testing.T) {
	currentTime := time.Now()
	ecPolicy := NewExpectContinuePolicy(nil)
	ecPolicy.now = func() time.Time { return currentTime }

	transport := &mockTransport{statusCode: http.StatusServiceUnavailable}
	pl := newTestPipeline(ecPolicy, transport)

	req1, err := newPutRequest(100)
	require.NoError(t, err)
	_, err = pl.Do(req1)
	require.NoError(t, err)

	currentTime = currentTime.Add(expectContinueDuration + time.Second)

	// Expired, but 503 reactivates
	req2, err := newPutRequest(100)
	require.NoError(t, err)
	_, err = pl.Do(req2)
	require.NoError(t, err)
	require.Empty(t, transport.lastExpectHdr)

	// Reactivated
	transport.statusCode = http.StatusCreated
	req3, err := newPutRequest(100)
	require.NoError(t, err)
	resp, err := pl.Do(req3)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, "100-continue", transport.lastExpectHdr)
}

// ---------------------------------------------------------------------------
// Auto mode with custom threshold
// ---------------------------------------------------------------------------

func TestAutoMode_CustomThreshold_NotAppliedBelowThreshold(t *testing.T) {
	customThreshold := int64(1024) // 1 KiB
	transport := &mockTransport{statusCode: http.StatusServiceUnavailable}
	ecPolicy := NewExpectContinuePolicy(&ExpectContinueOptions{ContentLengthThresholdBytes: customThreshold})
	pl := newTestPipeline(ecPolicy, transport)

	// Activate via a PUT that meets the threshold
	req1, err := newPutRequest(customThreshold)
	require.NoError(t, err)
	_, err = pl.Do(req1)
	require.NoError(t, err)

	// Below threshold — no header
	transport.statusCode = http.StatusCreated
	req2, err := newPutRequest(customThreshold - 1)
	require.NoError(t, err)
	resp, err := pl.Do(req2)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Empty(t, transport.lastExpectHdr)
}

func TestAutoMode_CustomThreshold_AppliedAtThreshold(t *testing.T) {
	customThreshold := int64(1024) // 1 KiB
	transport := &mockTransport{statusCode: http.StatusServiceUnavailable}
	ecPolicy := NewExpectContinuePolicy(&ExpectContinueOptions{ContentLengthThresholdBytes: customThreshold})
	pl := newTestPipeline(ecPolicy, transport)

	// Activate
	req1, err := newPutRequest(customThreshold)
	require.NoError(t, err)
	_, err = pl.Do(req1)
	require.NoError(t, err)

	// At threshold — header applied
	transport.statusCode = http.StatusCreated
	req2, err := newPutRequest(customThreshold)
	require.NoError(t, err)
	resp, err := pl.Do(req2)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, "100-continue", transport.lastExpectHdr)
}

// ---------------------------------------------------------------------------
// On mode tests
// ---------------------------------------------------------------------------

func TestOnMode_AlwaysAppliesHeader(t *testing.T) {
	transport := &mockTransport{statusCode: http.StatusCreated}
	ecPolicy := NewExpectContinuePolicy(&ExpectContinueOptions{Mode: ExpectContinueModeOn})
	pl := newTestPipeline(ecPolicy, transport)

	req, err := newPutRequest(100)
	require.NoError(t, err)
	resp, err := pl.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, "100-continue", transport.lastExpectHdr)
}

func TestOnMode_AppliedToSmallBody(t *testing.T) {
	// Threshold 0: even a 1-byte body gets the header in On mode
	transport := &mockTransport{statusCode: http.StatusCreated}
	ecPolicy := NewExpectContinuePolicy(&ExpectContinueOptions{Mode: ExpectContinueModeOn})
	pl := newTestPipeline(ecPolicy, transport)

	req, err := newPutRequest(1)
	require.NoError(t, err)
	_, err = pl.Do(req)
	require.NoError(t, err)
	require.Equal(t, "100-continue", transport.lastExpectHdr)
}

func TestOnMode_CustomThreshold_RespectsThreshold(t *testing.T) {
	customThreshold := int64(1024)
	transport := &mockTransport{statusCode: http.StatusCreated}
	ecPolicy := NewExpectContinuePolicy(&ExpectContinueOptions{Mode: ExpectContinueModeOn, ContentLengthThresholdBytes: customThreshold})
	pl := newTestPipeline(ecPolicy, transport)

	// Below threshold — no header
	req, err := newPutRequest(customThreshold - 1)
	require.NoError(t, err)
	_, err = pl.Do(req)
	require.NoError(t, err)
	require.Empty(t, transport.lastExpectHdr)

	// At threshold — header applied
	req2, err := newPutRequest(customThreshold)
	require.NoError(t, err)
	_, err = pl.Do(req2)
	require.NoError(t, err)
	require.Equal(t, "100-continue", transport.lastExpectHdr)
}

func TestOnMode_NotAppliedToGET(t *testing.T) {
	transport := &mockTransport{statusCode: http.StatusOK}
	ecPolicy := NewExpectContinuePolicy(&ExpectContinueOptions{Mode: ExpectContinueModeOn})
	pl := newTestPipeline(ecPolicy, transport)

	req, err := runtime.NewRequest(context.Background(), http.MethodGet, "https://example.blob.core.windows.net/container/blob")
	require.NoError(t, err)
	_, err = pl.Do(req)
	require.NoError(t, err)
	require.Empty(t, transport.lastExpectHdr)
}

// ---------------------------------------------------------------------------
// Off mode tests
// ---------------------------------------------------------------------------

func TestOffMode_NeverAppliesHeader(t *testing.T) {
	transport := &mockTransport{statusCode: http.StatusCreated}
	ecPolicy := NewExpectContinuePolicy(&ExpectContinueOptions{Mode: ExpectContinueModeOff})
	pl := newTestPipeline(ecPolicy, transport)

	req, err := newPutRequest(100)
	require.NoError(t, err)
	resp, err := pl.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Empty(t, transport.lastExpectHdr)
}

func TestOffMode_IgnoresThrottleResponses(t *testing.T) {
	transport := &mockTransport{statusCode: http.StatusServiceUnavailable}
	ecPolicy := NewExpectContinuePolicy(&ExpectContinueOptions{Mode: ExpectContinueModeOff})
	pl := newTestPipeline(ecPolicy, transport)

	// 503 should NOT activate anything
	req1, err := newPutRequest(100)
	require.NoError(t, err)
	_, err = pl.Do(req1)
	require.NoError(t, err)

	transport.statusCode = http.StatusCreated
	req2, err := newPutRequest(100)
	require.NoError(t, err)
	resp, err := pl.Do(req2)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Empty(t, transport.lastExpectHdr)
}
