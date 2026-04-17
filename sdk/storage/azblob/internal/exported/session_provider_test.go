// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package exported

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/internal/mock"
	"github.com/Azure/azure-sdk-for-go/sdk/internal/temporal"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/internal/generated"
	"github.com/stretchr/testify/require"
)

// newTestContainerClient creates a ContainerClient backed by a mock server for testing.
func newTestContainerClient(t *testing.T, srv *mock.Server) *generated.ContainerClient {
	azClient, err := azcore.NewClient("test", "v1.0.0", runtime.PipelineOptions{}, &policy.ClientOptions{Transport: srv})
	require.NoError(t, err)
	return generated.NewContainerClient(srv.URL()+"/testcontainer", azClient)
}

// createSessionResponseXML creates a session response XML body for testing.
func createSessionResponseXML(sessionKey, sessionToken string, expiration time.Time) []byte {
	return []byte(`<?xml version="1.0" encoding="utf-8"?>
<CreateSessionResponse>
	<AuthenticationType>HMAC</AuthenticationType>
	<Id>test-session-id</Id>
	<Credentials>
		<SessionKey>` + sessionKey + `</SessionKey>
		<SessionToken>` + sessionToken + `</SessionToken>
	</Credentials>
	<Expiration>` + expiration.Format(time.RFC1123) + `</Expiration>
</CreateSessionResponse>`)
}

// createErrorResponseXML creates an error response XML body for testing.
func createErrorResponseXML(code, message string) []byte {
	return []byte(`<?xml version="1.0" encoding="utf-8"?>
<Error>
	<Code>` + code + `</Code>
	<Message>` + message + `</Message>
</Error>`)
}

func TestAcquireSession_Success(t *testing.T) {
	srv, closeFn := mock.NewServer(mock.WithTransformAllRequestsToTestServerUrl())
	defer closeFn()

	expiration := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
	srv.AppendResponse(
		mock.WithStatusCode(http.StatusCreated),
		mock.WithBody(createSessionResponseXML("test-key", "test-token", expiration)),
	)

	client := newTestContainerClient(t, srv)
	state := sessionState{
		client: client,
		ctx:    context.Background(),
	}

	creds, exp, err := acquireSession(state)
	require.NoError(t, err)
	require.Equal(t, "test-key", creds.key)
	require.Equal(t, "test-token", creds.token)
	require.Equal(t, expiration.Format(time.RFC1123), exp.Format(time.RFC1123))
}

func TestAcquireSession_FallbackToBearer_Retryable(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		errorCode  string
	}{
		{
			name:       "ServerError500",
			statusCode: http.StatusInternalServerError,
			errorCode:  "InternalError",
		},
		{
			name:       "ServiceUnavailable503",
			statusCode: http.StatusServiceUnavailable,
			errorCode:  "ServiceUnavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, closeFn := mock.NewServer(mock.WithTransformAllRequestsToTestServerUrl())
			defer closeFn()

			srv.SetResponse(
				mock.WithStatusCode(tt.statusCode),
				mock.WithHeader("x-ms-error-code", tt.errorCode),
				mock.WithBody(createErrorResponseXML(tt.errorCode, "error message")),
			)

			client := newTestContainerClient(t, srv)
			state := sessionState{
				client: client,
				ctx:    context.Background(),
			}

			_, _, err := acquireSession(state)
			require.ErrorIs(t, err, errFallbackToBearer)
		})
	}
}

func TestAcquireSession_FallbackToBearer_NonRetryable(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		errorCode  string
	}{
		{
			name:       "FeatureNotEnabled",
			statusCode: http.StatusBadRequest,
			errorCode:  featureNotEnabled,
		},
		{
			name:       "Forbidden",
			statusCode: http.StatusForbidden,
			errorCode:  "AuthorizationFailure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, closeFn := mock.NewServer(mock.WithTransformAllRequestsToTestServerUrl())
			defer closeFn()

			srv.AppendResponse(
				mock.WithStatusCode(tt.statusCode),
				mock.WithHeader("x-ms-error-code", tt.errorCode),
				mock.WithBody(createErrorResponseXML(tt.errorCode, "error message")),
			)

			client := newTestContainerClient(t, srv)
			state := sessionState{
				client: client,
				ctx:    context.Background(),
			}

			_, _, err := acquireSession(state)
			require.ErrorIs(t, err, errFallbackToBearer)
		})
	}
}

func TestAcquireSession_Error(t *testing.T) {
	srv, closeFn := mock.NewServer(mock.WithTransformAllRequestsToTestServerUrl())
	defer closeFn()

	srv.AppendResponse(
		mock.WithStatusCode(http.StatusNotFound),
		mock.WithHeader("x-ms-error-code", "ContainerNotFound"),
		mock.WithBody(createErrorResponseXML("ContainerNotFound", "Container not found")),
	)

	client := newTestContainerClient(t, srv)
	state := sessionState{
		client: client,
		ctx:    context.Background(),
	}

	_, _, err := acquireSession(state)
	require.Error(t, err)
	// Should NOT be errFallbackToBearer - this is a real error
	require.False(t, errors.Is(err, errFallbackToBearer))

	var respErr *azcore.ResponseError
	require.True(t, errors.As(err, &respErr))
	require.Equal(t, http.StatusNotFound, respErr.StatusCode)
}

func TestSingleContainerProvider_GetSessionCredentials_MatchingContainer(t *testing.T) {
	srv, closeFn := mock.NewServer(mock.WithTransformAllRequestsToTestServerUrl())
	defer closeFn()

	expiration := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
	srv.AppendResponse(
		mock.WithStatusCode(http.StatusCreated),
		mock.WithBody(createSessionResponseXML("test-key", "test-token", expiration)),
	)

	client := newTestContainerClient(t, srv)
	provider := &singleContainerProvider{
		client:        client,
		containerName: "mycontainer",
		resource:      temporal.NewResource(acquireSession),
	}

	creds, err := provider.GetSessionCredentials(context.Background(), "mycontainer")
	require.NoError(t, err)
	require.Equal(t, "test-key", creds.key)
	require.Equal(t, "test-token", creds.token)
}

func TestSingleContainerProvider_GetSessionCredentials_MismatchedContainer(t *testing.T) {
	srv, closeFn := mock.NewServer(mock.WithTransformAllRequestsToTestServerUrl())
	defer closeFn()

	client := newTestContainerClient(t, srv)
	provider := &singleContainerProvider{
		client:        client,
		containerName: "mycontainer",
		resource:      temporal.NewResource(acquireSession),
	}

	_, err := provider.GetSessionCredentials(context.Background(), "othercontainer")
	require.ErrorIs(t, err, errFallbackToBearer)

	// Verify no requests were made to the server
	require.Equal(t, 0, srv.Requests())
}

func TestSingleContainerProvider_ExpireSessionCredentials_MatchingContainer(t *testing.T) {
	srv, closeFn := mock.NewServer(mock.WithTransformAllRequestsToTestServerUrl())
	defer closeFn()

	expiration := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
	// First response for initial session
	srv.AppendResponse(
		mock.WithStatusCode(http.StatusCreated),
		mock.WithBody(createSessionResponseXML("first-key", "first-token", expiration)),
	)
	// Second response after expiry
	srv.AppendResponse(
		mock.WithStatusCode(http.StatusCreated),
		mock.WithBody(createSessionResponseXML("second-key", "second-token", expiration)),
	)

	client := newTestContainerClient(t, srv)
	provider := &singleContainerProvider{
		client:        client,
		containerName: "mycontainer",
		resource:      temporal.NewResource(acquireSession),
	}

	// Get initial session
	creds1, err := provider.GetSessionCredentials(context.Background(), "mycontainer")
	require.NoError(t, err)
	require.Equal(t, "first-key", creds1.key)

	// Expire the session
	provider.ExpireSessionCredentials("mycontainer")

	// Get a new session after expiry
	creds2, err := provider.GetSessionCredentials(context.Background(), "mycontainer")
	require.NoError(t, err)
	require.Equal(t, "second-key", creds2.key)

	// Verify two requests were made
	require.Equal(t, 2, srv.Requests())
}

func TestSingleContainerProvider_ExpireSessionCredentials_MismatchedContainer(t *testing.T) {
	srv, closeFn := mock.NewServer(mock.WithTransformAllRequestsToTestServerUrl())
	defer closeFn()

	expiration := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
	// Only one response should be used
	srv.AppendResponse(
		mock.WithStatusCode(http.StatusCreated),
		mock.WithBody(createSessionResponseXML("test-key", "test-token", expiration)),
	)

	client := newTestContainerClient(t, srv)
	provider := &singleContainerProvider{
		client:        client,
		containerName: "mycontainer",
		resource:      temporal.NewResource(acquireSession),
	}

	// Get initial session
	creds1, err := provider.GetSessionCredentials(context.Background(), "mycontainer")
	require.NoError(t, err)
	require.Equal(t, "test-key", creds1.key)

	// Attempt to expire with a different container name - should have no effect
	provider.ExpireSessionCredentials("othercontainer")

	// Get session again - should return cached value, not make a new request
	creds2, err := provider.GetSessionCredentials(context.Background(), "mycontainer")
	require.NoError(t, err)
	require.Equal(t, "test-key", creds2.key)

	// Verify only one request was made (session was not expired)
	require.Equal(t, 1, srv.Requests())
}

func TestSingleContainerProvider_SessionCaching(t *testing.T) {
	srv, closeFn := mock.NewServer(mock.WithTransformAllRequestsToTestServerUrl())
	defer closeFn()

	expiration := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
	srv.AppendResponse(
		mock.WithStatusCode(http.StatusCreated),
		mock.WithBody(createSessionResponseXML("test-key", "test-token", expiration)),
	)

	client := newTestContainerClient(t, srv)
	provider := &singleContainerProvider{
		client:        client,
		containerName: "mycontainer",
		resource:      temporal.NewResource(acquireSession),
	}

	// Get session multiple times
	for i := 0; i < 5; i++ {
		creds, err := provider.GetSessionCredentials(context.Background(), "mycontainer")
		require.NoError(t, err)
		require.Equal(t, "test-key", creds.key)
	}

	// Only one request should have been made (session is cached)
	require.Equal(t, 1, srv.Requests())
}
