// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package exported

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/internal/mock"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/internal/generated"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/internal/shared"
	"github.com/stretchr/testify/require"
)

// mockSessionProvider is a mock implementation of the sessionProvider interface for testing.
type mockSessionProvider struct {
	getCredsFn   func(ctx context.Context, containerName string) (sessionCredentials, error)
	expireFn     func(containerName string)
	getCalls     int
	expireCalls  int
	lastGetCtx   context.Context
	lastGetCName string
}

func (m *mockSessionProvider) GetSessionCredentials(ctx context.Context, containerName string) (sessionCredentials, error) {
	m.getCalls++
	m.lastGetCtx = ctx
	m.lastGetCName = containerName
	if m.getCredsFn != nil {
		return m.getCredsFn(ctx, containerName)
	}
	return sessionCredentials{}, nil
}

func (m *mockSessionProvider) ExpireSessionCredentials(containerName string) {
	m.expireCalls++
	if m.expireFn != nil {
		m.expireFn(containerName)
	}
}

// mockBearerPolicy is a mock bearer token policy for testing.
type mockBearerPolicy struct {
	doFn    func(req *policy.Request) (*http.Response, error)
	doCalls int
}

func (m *mockBearerPolicy) Do(req *policy.Request) (*http.Response, error) {
	m.doCalls++
	if m.doFn != nil {
		return m.doFn(req)
	}
	return &http.Response{StatusCode: http.StatusOK}, nil
}

// newTestServiceClient creates a ServiceClient backed by a mock server for testing.
func newTestServiceClient(t *testing.T, srv *mock.Server) *generated.ServiceClient {
	azClient, err := azcore.NewClient("test", "v1.0.0", runtime.PipelineOptions{}, &policy.ClientOptions{Transport: srv})
	require.NoError(t, err)
	return generated.NewServiceClient(srv.URL(), azClient)
}

// TestNewSessionPolicy_Success tests successful creation of a session policy.
func TestNewSessionPolicy_Success(t *testing.T) {
	srv, closeFn := mock.NewServer(mock.WithTransformAllRequestsToTestServerUrl())
	defer closeFn()

	serviceClient := newTestServiceClient(t, srv)
	bearerPolicy := &mockBearerPolicy{}

	opts := SessionOptions{
		Mode:          SessionModeSingleContainer,
		AccountName:   "testaccount",
		ContainerName: "testcontainer",
	}

	pol, err := NewSessionPolicy(opts, bearerPolicy, serviceClient)
	require.NoError(t, err)
	require.NotNil(t, pol)
}

// TestNewSessionPolicy_Errors tests error cases when creating a session policy.
func TestNewSessionPolicy_Errors(t *testing.T) {
	tests := []struct {
		name          string
		opts          SessionOptions
		expectedError string
	}{
		{
			name: "MissingAccountName",
			opts: SessionOptions{
				Mode:          SessionModeSingleContainer,
				AccountName:   "",
				ContainerName: "testcontainer",
			},
			expectedError: "account name is required",
		},
		{
			name: "MissingContainerName",
			opts: SessionOptions{
				Mode:          SessionModeSingleContainer,
				AccountName:   "testaccount",
				ContainerName: "",
			},
			expectedError: "container name is required",
		},
		{
			name: "UnsupportedMode",
			opts: SessionOptions{
				Mode:          SessionMode("unsupported"),
				AccountName:   "testaccount",
				ContainerName: "testcontainer",
			},
			expectedError: "unsupported session mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, closeFn := mock.NewServer(mock.WithTransformAllRequestsToTestServerUrl())
			defer closeFn()

			serviceClient := newTestServiceClient(t, srv)
			bearerPolicy := &mockBearerPolicy{}

			pol, err := NewSessionPolicy(tt.opts, bearerPolicy, serviceClient)
			require.Error(t, err)
			require.Nil(t, pol)
			require.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// TestSessionPolicy_Do_FallbackToBearer tests scenarios where the session policy falls back to bearer token authentication.
func TestSessionPolicy_Do_FallbackToBearer(t *testing.T) {
	tests := []struct {
		name                  string
		method                string
		url                   string
		providerReturnsError  bool
		expectedProviderCalls int
	}{
		{
			name:                  "NonGetMethod",
			method:                http.MethodPost,
			url:                   "https://testaccount.blob.core.windows.net/container/blob",
			providerReturnsError:  false,
			expectedProviderCalls: 0,
		},
		{
			name:                  "CompParam",
			method:                http.MethodGet,
			url:                   "https://testaccount.blob.core.windows.net/container/blob?comp=metadata",
			providerReturnsError:  false,
			expectedProviderCalls: 0,
		},
		{
			name:                  "ContainerOnly",
			method:                http.MethodGet,
			url:                   "https://testaccount.blob.core.windows.net/container",
			providerReturnsError:  false,
			expectedProviderCalls: 0,
		},
		{
			name:                  "ProviderError",
			method:                http.MethodGet,
			url:                   "https://testaccount.blob.core.windows.net/container/blob",
			providerReturnsError:  true,
			expectedProviderCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bearerPolicy := &mockBearerPolicy{
				doFn: func(req *policy.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusOK}, nil
				},
			}

			mockProvider := &mockSessionProvider{}
			if tt.providerReturnsError {
				mockProvider.getCredsFn = func(ctx context.Context, containerName string) (sessionCredentials, error) {
					return sessionCredentials{}, errFallbackToBearer
				}
			}

			pol := &sessionPolicy{
				bearerTokenPolicy: bearerPolicy,
				opts:              SessionOptions{AccountName: "testaccount"},
				provider:          mockProvider,
			}

			req := createTestPolicyRequest(t, tt.method, tt.url)

			resp, err := pol.Do(req)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			require.Equal(t, 1, bearerPolicy.doCalls)
			require.Equal(t, tt.expectedProviderCalls, mockProvider.getCalls)
		})
	}
}

// TestCanUseSession tests the canUseSession helper function.
func TestCanUseSession(t *testing.T) {
	tests := []struct {
		name              string
		method            string
		urlStr            string
		expectedContainer string
		expectedOK        bool
	}{
		{
			name:              "ValidGETBlobRequest",
			method:            http.MethodGet,
			urlStr:            "https://account.blob.core.windows.net/container/blob",
			expectedContainer: "container",
			expectedOK:        true,
		},
		{
			name:              "ValidGETBlobRequestWithPath",
			method:            http.MethodGet,
			urlStr:            "https://account.blob.core.windows.net/container/path/to/blob",
			expectedContainer: "container",
			expectedOK:        true,
		},
		{
			name:              "NonGETMethod_POST",
			method:            http.MethodPost,
			urlStr:            "https://account.blob.core.windows.net/container/blob",
			expectedContainer: "",
			expectedOK:        false,
		},
		{
			name:              "NonGETMethod_PUT",
			method:            http.MethodPut,
			urlStr:            "https://account.blob.core.windows.net/container/blob",
			expectedContainer: "",
			expectedOK:        false,
		},
		{
			name:              "NonGETMethod_DELETE",
			method:            http.MethodDelete,
			urlStr:            "https://account.blob.core.windows.net/container/blob",
			expectedContainer: "",
			expectedOK:        false,
		},
		{
			name:              "RequestWithCompParam",
			method:            http.MethodGet,
			urlStr:            "https://account.blob.core.windows.net/container/blob?comp=metadata",
			expectedContainer: "",
			expectedOK:        false,
		},
		{
			name:              "EmptyPath",
			method:            http.MethodGet,
			urlStr:            "https://account.blob.core.windows.net/",
			expectedContainer: "",
			expectedOK:        false,
		},
		{
			name:              "ContainerOnly_NoBlob",
			method:            http.MethodGet,
			urlStr:            "https://account.blob.core.windows.net/container",
			expectedContainer: "",
			expectedOK:        false,
		},
		{
			name:              "ContainerOnly_TrailingSlash",
			method:            http.MethodGet,
			urlStr:            "https://account.blob.core.windows.net/container/",
			expectedContainer: "",
			expectedOK:        false,
		},
		{
			name:              "RootPath",
			method:            http.MethodGet,
			urlStr:            "https://account.blob.core.windows.net",
			expectedContainer: "",
			expectedOK:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.urlStr, nil)
			require.NoError(t, err)

			containerName, ok := canUseSession(req)
			require.Equal(t, tt.expectedOK, ok)
			require.Equal(t, tt.expectedContainer, containerName)
		})
	}
}

// TestHandleSessionRefresh_TriggersRefresh tests that session refresh is triggered when session_expiring or session_revoking header is present.
func TestHandleSessionRefresh_TriggersRefresh(t *testing.T) {
	tests := []struct {
		name        string
		headerValue string
	}{
		{"SessionExpiring", "session_expiring"},
		{"SessionRevoking", "session_revoking"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup
			wg.Add(1)

			mockProvider := &mockSessionProvider{
				getCredsFn: func(ctx context.Context, containerName string) (sessionCredentials, error) {
					wg.Done()
					return sessionCredentials{}, nil
				},
			}

			pol := &sessionPolicy{
				opts:     SessionOptions{AccountName: "testaccount"},
				provider: mockProvider,
			}

			resp := &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
			}
			// Use direct map assignment to set header with exact key (getHeader uses direct map access)
			resp.Header[shared.HeaderXmsAuthInfo] = []string{tt.headerValue}

			pol.handleSessionRefresh(resp, "testcontainer")

			// Wait for refresh goroutine to complete
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				// Success - refresh was triggered
			case <-time.After(5 * time.Second):
				t.Fatal("expected refresh to be triggered")
			}
		})
	}
}

// TestHandleSessionRefresh_NoRefresh tests that no refresh is triggered for normal responses.
func TestHandleSessionRefresh_NoRefresh(t *testing.T) {
	mockProvider := &mockSessionProvider{}

	pol := &sessionPolicy{
		opts:     SessionOptions{AccountName: "testaccount"},
		provider: mockProvider,
	}

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
	}

	pol.handleSessionRefresh(resp, "testcontainer")

	// Give some time for any potential goroutine to start
	time.Sleep(100 * time.Millisecond)

	require.Equal(t, 0, mockProvider.getCalls)
}

// TestHandleSessionError_NonResponseError tests that non-ResponseError errors are passed through.
func TestHandleSessionError_NonResponseError(t *testing.T) {
	pol := &sessionPolicy{
		opts: SessionOptions{AccountName: "testaccount"},
	}

	originalErr := errors.New("some random error")
	resp := &http.Response{StatusCode: http.StatusOK}

	retResp, retErr := pol.handleSessionError(nil, resp, originalErr, "testcontainer")
	require.Equal(t, resp, retResp)
	require.Equal(t, originalErr, retErr)
}

// TestHandleSessionError_ServiceUnavailable tests fallback to bearer on 503 with SessionOperationsTemporarilyUnavailable.
func TestHandleSessionError_ServiceUnavailable(t *testing.T) {
	pol := &sessionPolicy{
		opts: SessionOptions{AccountName: "testaccount"},
	}

	originalErr := &azcore.ResponseError{
		StatusCode: http.StatusServiceUnavailable,
		ErrorCode:  sessionUnavailable,
	}
	resp := &http.Response{StatusCode: http.StatusServiceUnavailable}

	retResp, retErr := pol.handleSessionError(nil, resp, originalErr, "testcontainer")
	require.Nil(t, retResp)
	require.ErrorIs(t, retErr, errFallbackToBearer)
}

// TestHandleSessionError_OtherError tests that other errors are passed through.
func TestHandleSessionError_OtherError(t *testing.T) {
	pol := &sessionPolicy{
		opts: SessionOptions{AccountName: "testaccount"},
	}

	originalErr := &azcore.ResponseError{
		StatusCode: http.StatusNotFound,
		ErrorCode:  "BlobNotFound",
	}
	resp := &http.Response{StatusCode: http.StatusNotFound}

	retResp, retErr := pol.handleSessionError(nil, resp, originalErr, "testcontainer")
	require.Equal(t, resp, retResp)
	require.Equal(t, originalErr, retErr)
}

// TestRetryWithNewSession_FallbackError tests retry returns errFallbackToBearer when provider does.
func TestRetryWithNewSession_FallbackError(t *testing.T) {
	expireCalled := false
	mockProvider := &mockSessionProvider{
		getCredsFn: func(ctx context.Context, containerName string) (sessionCredentials, error) {
			return sessionCredentials{}, errFallbackToBearer
		},
		expireFn: func(containerName string) {
			expireCalled = true
		},
	}

	pol := &sessionPolicy{
		opts:     SessionOptions{AccountName: "testaccount"},
		provider: mockProvider,
	}

	req := createTestPolicyRequest(t, http.MethodGet, "https://testaccount.blob.core.windows.net/container/blob")

	resp, err := pol.retryWithNewSession(req, "container")
	require.Nil(t, resp)
	require.ErrorIs(t, err, errFallbackToBearer)
	require.True(t, expireCalled)
}

// TestRetryWithNewSession_OtherError tests retry returns other provider errors.
func TestRetryWithNewSession_OtherError(t *testing.T) {
	expectedErr := errors.New("provider error")
	mockProvider := &mockSessionProvider{
		getCredsFn: func(ctx context.Context, containerName string) (sessionCredentials, error) {
			return sessionCredentials{}, expectedErr
		},
	}

	pol := &sessionPolicy{
		opts:     SessionOptions{AccountName: "testaccount"},
		provider: mockProvider,
	}

	req := createTestPolicyRequest(t, http.MethodGet, "https://testaccount.blob.core.windows.net/container/blob")

	resp, err := pol.retryWithNewSession(req, "container")
	require.Nil(t, resp)
	require.Equal(t, expectedErr, err)
}

// createTestPolicyRequest creates a policy.Request for testing with Next() support.
func createTestPolicyRequest(t *testing.T, method, urlStr string) *policy.Request {
	httpReq, err := http.NewRequestWithContext(context.Background(), method, urlStr, nil)
	require.NoError(t, err)

	// Create a minimal pipeline for testing
	_ = runtime.NewPipeline("test", "v1.0.0", runtime.PipelineOptions{}, &policy.ClientOptions{
		Transport: &mockTransport{},
	})

	req, err := runtime.NewRequest(context.Background(), method, urlStr)
	require.NoError(t, err)
	req.Raw().Header = httpReq.Header

	return req
}

// mockTransport is a mock HTTP transport for testing.
type mockTransport struct {
	response *http.Response
	err      error
}

func (m *mockTransport) Do(_ *http.Request) (*http.Response, error) {
	if m.response != nil {
		return m.response, m.err
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
	}, nil
}

// TestApplySessionReq_SetsAuthorizationHeader tests that applySessionReq sets the authorization header correctly.
func TestApplySessionReq_SetsAuthorizationHeader(t *testing.T) {
	sessionKey := "dGVzdC1rZXk=" // base64 encoded "test-key"
	sessionToken := "test-token"

	transport := &recordingTransport{}

	pol := &sessionPolicy{
		opts: SessionOptions{
			AccountName: "testaccount",
		},
	}

	creds := sessionCredentials{
		SessionKey:   &sessionKey,
		SessionToken: &sessionToken,
	}

	// Create a pipeline with our policy that will call applySessionReq
	testPolicy := &testApplyPolicy{
		pol:   pol,
		creds: creds,
	}

	pl := runtime.NewPipeline("test", "v1.0.0", runtime.PipelineOptions{
		PerCall: []policy.Policy{testPolicy},
	}, &policy.ClientOptions{
		Transport: transport,
	})

	req, err := runtime.NewRequest(context.Background(), http.MethodGet, "https://testaccount.blob.core.windows.net/container/blob")
	require.NoError(t, err)

	resp, err := pl.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify the Authorization header was set
	authHeader := transport.lastRequest.Header.Get(shared.HeaderAuthorization)
	require.True(t, strings.HasPrefix(authHeader, "Session "))
	require.Contains(t, authHeader, sessionToken)
}

// testApplyPolicy is a helper policy that calls applySessionReq for testing.
type testApplyPolicy struct {
	pol   *sessionPolicy
	creds sessionCredentials
}

func (p *testApplyPolicy) Do(req *policy.Request) (*http.Response, error) {
	return p.pol.applySessionReq(req, p.creds)
}

// recordingTransport records the last request for verification.
type recordingTransport struct {
	lastRequest *http.Request
}

func (r *recordingTransport) Do(req *http.Request) (*http.Response, error) {
	r.lastRequest = req
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
	}, nil
}

// TestHandleSessionError_Unauthorized_TriggersRetry tests that 401 triggers retry with new session.
func TestHandleSessionError_Unauthorized_TriggersRetry(t *testing.T) {
	sessionKey := "dGVzdC1rZXk=" // base64 encoded "test-key"
	sessionToken := "new-token"
	callCount := 0
	expireCalled := false

	mockProvider := &mockSessionProvider{
		getCredsFn: func(ctx context.Context, containerName string) (sessionCredentials, error) {
			callCount++
			return sessionCredentials{
				SessionKey:   &sessionKey,
				SessionToken: &sessionToken,
			}, nil
		},
		expireFn: func(containerName string) {
			expireCalled = true
		},
	}

	transport := &mockTransport{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
		},
	}

	pol := &sessionPolicy{
		opts: SessionOptions{
			AccountName: "testaccount",
		},
		provider: mockProvider,
	}

	// Create a helper policy to pass the request through
	testPolicy := &testRetryPolicy{
		pol:           pol,
		containerName: "container",
	}

	pl := runtime.NewPipeline("test", "v1.0.0", runtime.PipelineOptions{
		PerCall: []policy.Policy{testPolicy},
	}, &policy.ClientOptions{
		Transport: transport,
	})

	req, err := runtime.NewRequest(context.Background(), http.MethodGet, "https://testaccount.blob.core.windows.net/container/blob")
	require.NoError(t, err)

	originalErr := &azcore.ResponseError{
		StatusCode: http.StatusUnauthorized,
		ErrorCode:  "AuthenticationFailed",
	}
	testPolicy.originalErr = originalErr
	testPolicy.originalResp = &http.Response{StatusCode: http.StatusUnauthorized}

	resp, err := pl.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.True(t, expireCalled)
	require.Equal(t, 1, callCount)
}

// testRetryPolicy is a helper policy for testing handleSessionError with 401.
type testRetryPolicy struct {
	pol           *sessionPolicy
	containerName string
	originalErr   error
	originalResp  *http.Response
}

func (p *testRetryPolicy) Do(req *policy.Request) (*http.Response, error) {
	return p.pol.handleSessionError(req, p.originalResp, p.originalErr, p.containerName)
}

// TestIntegration_SessionPolicy_SuccessfulRequest tests the full flow of a successful session request.
func TestIntegration_SessionPolicy_SuccessfulRequest(t *testing.T) {
	sessionKey := "dGVzdC1rZXk=" // base64 encoded "test-key"
	sessionToken := "test-session-token"

	mockProvider := &mockSessionProvider{
		getCredsFn: func(ctx context.Context, containerName string) (sessionCredentials, error) {
			require.Equal(t, "testcontainer", containerName)
			return sessionCredentials{
				SessionKey:   &sessionKey,
				SessionToken: &sessionToken,
			}, nil
		},
	}

	transport := &recordingTransport{}

	bearerPolicy := &mockBearerPolicy{}

	pol := &sessionPolicy{
		bearerTokenPolicy: bearerPolicy,
		opts: SessionOptions{
			AccountName:   "testaccount",
			ContainerName: "testcontainer",
		},
		provider: mockProvider,
	}

	// Create request through runtime to get proper Next() support
	pl := runtime.NewPipeline("test", "v1.0.0", runtime.PipelineOptions{
		PerCall: []policy.Policy{pol},
	}, &policy.ClientOptions{
		Transport: transport,
	})

	req, err := runtime.NewRequest(context.Background(), http.MethodGet, "https://testaccount.blob.core.windows.net/testcontainer/blob")
	require.NoError(t, err)

	resp, err := pl.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify session credentials were used
	require.Equal(t, 1, mockProvider.getCalls)
	require.Equal(t, 0, bearerPolicy.doCalls)

	// Verify Authorization header was set with Session prefix
	authHeader := transport.lastRequest.Header.Get(shared.HeaderAuthorization)
	require.True(t, strings.HasPrefix(authHeader, "Session "))
}

// TestIntegration_SessionPolicy_FallbackToBearer tests that non-session requests fallback to bearer.
func TestIntegration_SessionPolicy_FallbackToBearer(t *testing.T) {
	mockProvider := &mockSessionProvider{}

	bearerPolicy := &mockBearerPolicy{
		doFn: func(req *policy.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
			}, nil
		},
	}

	pol := &sessionPolicy{
		bearerTokenPolicy: bearerPolicy,
		opts: SessionOptions{
			AccountName:   "testaccount",
			ContainerName: "testcontainer",
		},
		provider: mockProvider,
	}

	// POST request should not use session
	req := createTestPolicyRequest(t, http.MethodPost, "https://testaccount.blob.core.windows.net/testcontainer/blob")

	resp, err := pol.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify bearer was used, not session
	require.Equal(t, 0, mockProvider.getCalls)
	require.Equal(t, 1, bearerPolicy.doCalls)
}

// TestHandleSessionRefresh_ConcurrentRefresh tests that only one goroutine refreshes at a time.
func TestHandleSessionRefresh_ConcurrentRefresh(t *testing.T) {
	var mu sync.Mutex
	refreshCount := 0
	refreshStarted := make(chan struct{})
	refreshComplete := make(chan struct{})

	mockProvider := &mockSessionProvider{
		getCredsFn: func(ctx context.Context, containerName string) (sessionCredentials, error) {
			mu.Lock()
			refreshCount++
			mu.Unlock()
			// Signal that refresh has started, but only once
			select {
			case refreshStarted <- struct{}{}:
			default:
			}
			<-refreshComplete
			return sessionCredentials{}, nil
		},
	}

	pol := &sessionPolicy{
		opts:     SessionOptions{AccountName: "testaccount"},
		provider: mockProvider,
	}

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
	}
	// Use direct map assignment to set header with exact key (getHeader uses direct map access)
	resp.Header[shared.HeaderXmsAuthInfo] = []string{"session_expiring"}

	// Trigger first refresh
	pol.handleSessionRefresh(resp, "testcontainer")

	// Wait for first refresh to start with timeout
	select {
	case <-refreshStarted:
		// First refresh has started
	case <-time.After(5 * time.Second):
		t.Fatal("first refresh did not start")
	}

	// Try to trigger another refresh while first is in progress (should be skipped due to TryLock)
	pol.handleSessionRefresh(resp, "testcontainer")
	pol.handleSessionRefresh(resp, "testcontainer")

	// Allow first refresh to complete
	close(refreshComplete)

	// Give time for mutex to be released and any potential extra goroutines to execute
	time.Sleep(100 * time.Millisecond)

	// Only one refresh should have occurred because TryLock prevents concurrent refreshes
	mu.Lock()
	count := refreshCount
	mu.Unlock()
	require.Equal(t, 1, count)
}

// TestDoWithSession_ProviderError tests doWithSession when provider returns an error.
func TestDoWithSession_ProviderError(t *testing.T) {
	expectedErr := errors.New("provider error")
	mockProvider := &mockSessionProvider{
		getCredsFn: func(ctx context.Context, containerName string) (sessionCredentials, error) {
			return sessionCredentials{}, expectedErr
		},
	}

	pol := &sessionPolicy{
		opts:     SessionOptions{AccountName: "testaccount"},
		provider: mockProvider,
	}

	req := createTestPolicyRequest(t, http.MethodGet, "https://testaccount.blob.core.windows.net/container/blob")

	resp, err := pol.doWithSession(req, "container")
	require.Nil(t, resp)
	require.Equal(t, expectedErr, err)
}

// TestDoWithSession_Success tests successful doWithSession flow.
func TestDoWithSession_Success(t *testing.T) {
	sessionKey := "dGVzdC1rZXk=" // base64 encoded "test-key"
	sessionToken := "test-token"

	mockProvider := &mockSessionProvider{
		getCredsFn: func(ctx context.Context, containerName string) (sessionCredentials, error) {
			return sessionCredentials{
				SessionKey:   &sessionKey,
				SessionToken: &sessionToken,
			}, nil
		},
	}

	transport := &mockTransport{}

	pol := &sessionPolicy{
		opts: SessionOptions{
			AccountName: "testaccount",
		},
		provider: mockProvider,
	}

	// Create a helper policy to call doWithSession
	testPolicy := &testDoWithSessionPolicy{
		pol:           pol,
		containerName: "container",
	}

	pl := runtime.NewPipeline("test", "v1.0.0", runtime.PipelineOptions{
		PerCall: []policy.Policy{testPolicy},
	}, &policy.ClientOptions{
		Transport: transport,
	})

	req, err := runtime.NewRequest(context.Background(), http.MethodGet, "https://testaccount.blob.core.windows.net/container/blob")
	require.NoError(t, err)

	resp, err := pl.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// testDoWithSessionPolicy is a helper policy that calls doWithSession for testing.
type testDoWithSessionPolicy struct {
	pol           *sessionPolicy
	containerName string
}

func (p *testDoWithSessionPolicy) Do(req *policy.Request) (*http.Response, error) {
	return p.pol.doWithSession(req, p.containerName)
}

// TestApplySessionReq_NilSessionKey tests applySessionReq with nil session key.
func TestApplySessionReq_NilSessionKey(t *testing.T) {
	sessionToken := "test-token"

	pol := &sessionPolicy{
		opts: SessionOptions{
			AccountName: "testaccount",
		},
	}

	req := createTestPolicyRequest(t, http.MethodGet, "https://testaccount.blob.core.windows.net/container/blob")

	creds := sessionCredentials{
		SessionKey:   nil,
		SessionToken: &sessionToken,
	}

	// Should fail because session key is empty (invalid base64)
	_, err := pol.applySessionReq(req, creds)
	require.Error(t, err)
}

// TestApplySessionReq_NilSessionToken tests applySessionReq with nil session token.
func TestApplySessionReq_NilSessionToken(t *testing.T) {
	sessionKey := "dGVzdC1rZXk=" // base64 encoded "test-key"

	transport := &mockTransport{}

	pol := &sessionPolicy{
		opts: SessionOptions{
			AccountName: "testaccount",
		},
	}

	creds := sessionCredentials{
		SessionKey:   &sessionKey,
		SessionToken: nil,
	}

	// Create a pipeline with our policy that will call applySessionReq
	testPolicy := &testApplyPolicy{
		pol:   pol,
		creds: creds,
	}

	pl := runtime.NewPipeline("test", "v1.0.0", runtime.PipelineOptions{
		PerCall: []policy.Policy{testPolicy},
	}, &policy.ClientOptions{
		Transport: transport,
	})

	req, err := runtime.NewRequest(context.Background(), http.MethodGet, "https://testaccount.blob.core.windows.net/container/blob")
	require.NoError(t, err)

	resp, err := pl.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
}
