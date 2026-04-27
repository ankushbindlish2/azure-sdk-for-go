// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package shared

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

const (
	// expectContinueDuration is how long 100-continue remains active after
	// receiving a throttling response from the service.
	expectContinueDuration = 1 * time.Minute
)

// ExpectContinueMode controls how the Expect: 100-continue header is applied to
// PUT requests in the storage SDK pipeline.
type ExpectContinueMode int

const (
	// ExpectContinueModeAuto is the default mode. The SDK will NOT send the
	// "Expect: 100-continue" header under normal conditions. However, if the
	// service returns a throttling response (429 Too Many Requests, 500 Internal
	// Server Error, or 503 Service Unavailable), the SDK will automatically begin
	// adding the header to qualifying PUT requests for a one-minute window after
	// the last throttling response. This allows the SDK to back off a storage
	// tenant that is under load—by waiting for a "100 Continue" before sending the
	// request body—while not adding overhead to every PUT request during normal
	// operation. When the one-minute window expires without another throttling
	// response, the header is no longer sent and the SDK returns to normal behavior.
	//
	// Qualifying requests are PUT requests with Content-Length > 0 (and >=
	// the configured threshold, if one is set).
	// The header is never sent on requests with zero Content-Length (per HTTP spec).
	ExpectContinueModeAuto ExpectContinueMode = iota

	// ExpectContinueModeOn unconditionally adds the "Expect: 100-continue" header
	// to every qualifying PUT request, regardless of whether the service has
	// returned any throttling responses. This is useful in scenarios where you know
	// the target storage account is under heavy load and want to proactively avoid
	// sending request bodies that will be rejected. Only PUT requests with
	// Content-Length > 0 (and >= the configured threshold, if one is set) will
	// receive the header.
	ExpectContinueModeOn

	// ExpectContinueModeOff completely disables the Expect: 100-continue behavior.
	// The header is never added to any request, even if the service returns
	// throttling responses. Use this mode if you have determined that 100-continue
	// is not beneficial for your workload or if it interferes with a custom
	// transport or proxy configuration.
	ExpectContinueModeOff
)

// ExpectContinueOptions configures the "Expect: 100-continue" pipeline policy
// for PUT requests. Combine this with [ExpectContinueMode] and an optional
// content-length threshold in a single options struct.
type ExpectContinueOptions struct {
	// Mode controls when the "Expect: 100-continue" header is added to PUT requests.
	// The default zero value is [ExpectContinueModeAuto], which only enables the
	// header after receiving a throttling response (429, 500, or 503) for a
	// one-minute window.
	Mode ExpectContinueMode

	// ContentLengthThresholdBytes is the minimum Content-Length (in bytes) for a PUT request to
	// receive the "Expect: 100-continue" header. The default zero value means all
	// PUT requests with a body (Content-Length > 0) qualify. Set this to a positive
	// value (e.g. 8 * 1024 * 1024 for 8 MiB) to limit the header to larger
	// uploads only.
	ContentLengthThresholdBytes int64
}

// ExpectContinuePolicy is a per-retry pipeline policy that conditionally adds
// the "Expect: 100-continue" header to PUT requests.
//
// The policy supports three modes (see [ExpectContinueMode]):
//   - Auto (default): adds the header only after receiving 429, 500, or 503
//     responses, for a one-minute window after the last such response.
//   - On: always adds the header to qualifying requests.
//   - Off: never adds the header.
//
// By default, the header is sent on all PUT requests with Content-Length > 0.
// A custom threshold can be set so that only requests with Content-Length >=
// that threshold receive the header.
//
// The activation timestamp is stored as an int64 (UnixNano) and accessed via
// atomic operations for lock-free, thread-safe reads and writes.
type ExpectContinuePolicy struct {
	// lastThrottleNano stores the UnixNano timestamp of the last throttling
	// response. Accessed atomically — zero means never throttled.
	lastThrottleNano atomic.Int64

	mode ExpectContinueMode

	// threshold is the minimum Content-Length (in bytes) for a request to qualify
	// for the Expect header. A value of 0 means all PUT requests with
	// Content-Length > 0 qualify (the default, aligning with the .NET SDK).
	threshold int64

	// now is injectable for testing; if nil, time.Now is used.
	now func() time.Time
}

// NewExpectContinuePolicy creates a new ExpectContinuePolicy from the given options.
// If opts is nil, the policy uses default settings (Auto mode, no threshold).
func NewExpectContinuePolicy(opts *ExpectContinueOptions) *ExpectContinuePolicy {
	p := &ExpectContinuePolicy{}
	if opts != nil {
		p.mode = opts.Mode
		p.threshold = opts.ContentLengthThresholdBytes
	}
	return p
}

func (p *ExpectContinuePolicy) timeNow() time.Time {
	if p.now != nil {
		return p.now()
	}
	return time.Now()
}

func (p *ExpectContinuePolicy) isThrottleActive() bool {
	lastNano := p.lastThrottleNano.Load()
	if lastNano == 0 {
		return false
	}
	return p.timeNow().UnixNano()-lastNano < int64(expectContinueDuration)
}

func (p *ExpectContinuePolicy) recordThrottle() {
	p.lastThrottleNano.Store(p.timeNow().UnixNano())
}

// shouldApplyHeader returns true if the Expect header should be added to this request.
func (p *ExpectContinuePolicy) shouldApplyHeader(raw *http.Request) bool {
	if raw.Method != http.MethodPut || raw.ContentLength <= 0 {
		return false
	}
	if p.threshold > 0 && raw.ContentLength < p.threshold {
		return false
	}
	switch p.mode {
	case ExpectContinueModeOn:
		return true
	case ExpectContinueModeOff:
		return false
	default: // ExpectContinueModeAuto
		return p.isThrottleActive()
	}
}

// Do implements the policy.Policy interface.
func (p *ExpectContinuePolicy) Do(req *policy.Request) (*http.Response, error) {
	if p.shouldApplyHeader(req.Raw()) {
		req.Raw().Header.Set("Expect", "100-continue")
	}

	resp, err := req.Next()

	// In Auto mode, record throttling responses to activate the header for subsequent requests.
	if p.mode == ExpectContinueModeAuto && err == nil && resp != nil {
		switch resp.StatusCode {
		case http.StatusTooManyRequests, // 429
			http.StatusInternalServerError, // 500
			http.StatusServiceUnavailable:  // 503
			p.recordThrottle()
		}
	}

	return resp, err
}
