package container

import (
	"context"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/internal/recording"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/internal/generated"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/internal/testcommon"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func Test(t *testing.T) {
	recordMode := recording.GetRecordMode()
	t.Logf("Running container Tests in %s mode\n", recordMode)
	switch recordMode {
	case recording.LiveMode:
		suite.Run(t, &ContainerInternalRecordedTestsSuite{})
	case recording.PlaybackMode:
		suite.Run(t, &ContainerInternalRecordedTestsSuite{})
	case recording.RecordingMode:
		suite.Run(t, &ContainerInternalRecordedTestsSuite{})
	}
}

func (s *ContainerInternalRecordedTestsSuite) SetupSuite() {
	s.proxy = testcommon.SetupSuite(&s.Suite)
}

func (s *ContainerInternalRecordedTestsSuite) TearDownSuite() {
	testcommon.TearDownSuite(&s.Suite, s.proxy)
}

func (s *ContainerInternalRecordedTestsSuite) BeforeTest(suite string, test string) {
	testcommon.BeforeTest(s.T(), suite, test)
}

type ContainerInternalRecordedTestsSuite struct {
	suite.Suite
	proxy *recording.TestProxyInstance
}

func (s *ContainerInternalRecordedTestsSuite) TestContainerCreateSession() {
	_require := require.New(s.T())
	testName := s.T().Name()

	accountName, _ := testcommon.GetGenericAccountInfo(testcommon.TestAccountDefault)
	_require.Greater(len(accountName), 0)

	cred, err := testcommon.GetGenericTokenCredential()
	_require.NoError(err)

	svcClient, err := service.NewClient("https://"+accountName+".blob.core.windows.net/", cred, nil)
	_require.NoError(err)

	containerName := testcommon.GenerateContainerName(testName)
	containerClient := testcommon.GetContainerClient(containerName, svcClient)

	_, err = containerClient.Create(context.Background(), nil)
	_require.NoError(err)
	defer testcommon.DeleteContainer(context.Background(), _require, containerClient)

	resp, err := containerClient.generated().CreateSession(context.Background(), CreateSessionConfiguration{
		AuthenticationType: to.Ptr(generated.AuthenticationTypeHMAC),
	}, nil)
	_require.NoError(err)
	_require.NotNil(resp.ID)
	_require.NotEmpty(*resp.ID)
	_require.NotNil(resp.Expiration)
	_require.True(resp.Expiration.After(time.Now().Add(-time.Minute)), "Expiration should be in the future or very recent")
	_require.NotNil(resp.AuthenticationType)
	_require.NotNil(resp.Credentials)
	_require.NotNil(resp.Credentials.SessionKey)
	_require.NotEmpty(*resp.Credentials.SessionKey)
	_require.NotNil(resp.Credentials.SessionToken)
	_require.NotEmpty(*resp.Credentials.SessionToken)
}

func (s *ContainerInternalRecordedTestsSuite) TestContainerCreateSessionNonExistentContainer() {
	_require := require.New(s.T())
	testName := s.T().Name()

	accountName, _ := testcommon.GetGenericAccountInfo(testcommon.TestAccountDefault)
	_require.Greater(len(accountName), 0)

	cred, err := testcommon.GetGenericTokenCredential()
	_require.NoError(err)

	svcClient, err := service.NewClient("https://"+accountName+".blob.core.windows.net/", cred, nil)
	_require.NoError(err)

	containerName := testcommon.GenerateContainerName(testName)
	containerClient := testcommon.GetContainerClient(containerName, svcClient)

	// Attempting to create a session on a non-existent container should fail
	_, err = containerClient.generated().CreateSession(context.Background(), CreateSessionConfiguration{
		AuthenticationType: to.Ptr(generated.AuthenticationTypeHMAC),
	}, nil)
	_require.Error(err)
}

func (s *ContainerInternalRecordedTestsSuite) TestContainerCreateSessionMultipleTimes() {
	_require := require.New(s.T())
	testName := s.T().Name()

	accountName, _ := testcommon.GetGenericAccountInfo(testcommon.TestAccountDefault)
	_require.Greater(len(accountName), 0)

	cred, err := testcommon.GetGenericTokenCredential()
	_require.NoError(err)

	svcClient, err := service.NewClient("https://"+accountName+".blob.core.windows.net/", cred, nil)
	_require.NoError(err)

	containerName := testcommon.GenerateContainerName(testName)
	containerClient := testcommon.GetContainerClient(containerName, svcClient)

	_, err = containerClient.Create(context.Background(), nil)
	_require.NoError(err)
	defer testcommon.DeleteContainer(context.Background(), _require, containerClient)

	// Create multiple sessions and verify they have different IDs
	resp1, err := containerClient.generated().CreateSession(context.Background(), CreateSessionConfiguration{
		AuthenticationType: to.Ptr(generated.AuthenticationTypeHMAC),
	}, nil)
	_require.NoError(err)
	_require.NotNil(resp1.ID)

	resp2, err := containerClient.generated().CreateSession(context.Background(), CreateSessionConfiguration{
		AuthenticationType: to.Ptr(generated.AuthenticationTypeHMAC),
	}, nil)
	_require.NoError(err)
	_require.NotNil(resp2.ID)

	// Each session should have a unique ID
	_require.NotEqual(*resp1.ID, *resp2.ID)
}

func (s *ContainerInternalRecordedTestsSuite) TestContainerCreateSessionWithDifferentContainers() {
	_require := require.New(s.T())
	testName := s.T().Name()

	accountName, _ := testcommon.GetGenericAccountInfo(testcommon.TestAccountDefault)
	_require.Greater(len(accountName), 0)

	cred, err := testcommon.GetGenericTokenCredential()
	_require.NoError(err)

	svcClient, err := service.NewClient("https://"+accountName+".blob.core.windows.net/", cred, nil)
	_require.NoError(err)

	// Create first container
	containerName1 := testcommon.GenerateContainerName(testName + "1")
	containerClient1 := testcommon.GetContainerClient(containerName1, svcClient)
	_, err = containerClient1.Create(context.Background(), nil)
	_require.NoError(err)
	defer testcommon.DeleteContainer(context.Background(), _require, containerClient1)

	// Create second container
	containerName2 := testcommon.GenerateContainerName(testName + "2")
	containerClient2 := testcommon.GetContainerClient(containerName2, svcClient)
	_, err = containerClient2.Create(context.Background(), nil)
	_require.NoError(err)
	defer testcommon.DeleteContainer(context.Background(), _require, containerClient2)

	// Create sessions for each container
	resp1, err := containerClient1.generated().CreateSession(context.Background(), CreateSessionConfiguration{
		AuthenticationType: to.Ptr(generated.AuthenticationTypeHMAC),
	}, nil)
	_require.NoError(err)
	_require.NotNil(resp1.ID)

	resp2, err := containerClient2.generated().CreateSession(context.Background(), CreateSessionConfiguration{
		AuthenticationType: to.Ptr(generated.AuthenticationTypeHMAC),
	}, nil)
	_require.NoError(err)
	_require.NotNil(resp2.ID)

	// Sessions for different containers should be independent
	_require.NotEqual(*resp1.ID, *resp2.ID)
}

func (s *ContainerInternalRecordedTestsSuite) TestContainerCreateSessionWithSharedKeyFails() {
	_require := require.New(s.T())
	testName := s.T().Name()

	svcClient, err := testcommon.GetServiceClient(s.T(), testcommon.TestAccountDefault, nil)
	_require.NoError(err)

	containerName := testcommon.GenerateContainerName(testName)
	containerClient := testcommon.CreateNewContainer(context.Background(), _require, containerName, svcClient)
	defer testcommon.DeleteContainer(context.Background(), _require, containerClient)

	_, err = containerClient.generated().CreateSession(context.Background(), CreateSessionConfiguration{
		AuthenticationType: to.Ptr(generated.AuthenticationTypeHMAC),
	}, nil)
	_require.Error(err)
}
