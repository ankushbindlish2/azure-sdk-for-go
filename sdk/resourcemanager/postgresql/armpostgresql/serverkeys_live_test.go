// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package armpostgresql_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/internal/recording"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/internal/v3/testutil"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armdeployments"
	"github.com/stretchr/testify/suite"
)

type ServerKeysTestSuite struct {
	suite.Suite

	ctx               context.Context
	cred              azcore.TokenCredential
	options           *arm.ClientOptions
	serverName        string
	principalId       string
	adminPassword     string
	location          string
	resourceGroupName string
	subscriptionId    string
}

func (testsuite *ServerKeysTestSuite) SetupSuite() {
	testutil.StartRecording(testsuite.T(), pathToPackage)

	testsuite.ctx = context.Background()
	testsuite.cred, testsuite.options = testutil.GetCredAndClientOptions(testsuite.T())
	testsuite.serverName, _ = recording.GenerateAlphaNumericID(testsuite.T(), "serverna", 14, true)
	testsuite.adminPassword = recording.GetEnvVariable("ADMIN_PASSWORD", "")
	testsuite.location = recording.GetEnvVariable("LOCATION", "westus")
	testsuite.resourceGroupName = recording.GetEnvVariable("RESOURCE_GROUP_NAME", "scenarioTestTempGroup")
	testsuite.subscriptionId = recording.GetEnvVariable("AZURE_SUBSCRIPTION_ID", "00000000-0000-0000-0000-000000000000")
	resourceGroup, _, err := testutil.CreateResourceGroup(testsuite.ctx, testsuite.subscriptionId, testsuite.cred, testsuite.options, testsuite.location)
	testsuite.Require().NoError(err)
	testsuite.resourceGroupName = *resourceGroup.Name
	testsuite.Prepare()
}

func (testsuite *ServerKeysTestSuite) TearDownSuite() {
	testsuite.Cleanup()
	_, err := testutil.DeleteResourceGroup(testsuite.ctx, testsuite.subscriptionId, testsuite.cred, testsuite.options, testsuite.resourceGroupName)
	testsuite.Require().NoError(err)
	testutil.StopRecording(testsuite.T())
}

func TestServerKeysTestSuite(t *testing.T) {
	suite.Run(t, new(ServerKeysTestSuite))
}

func (testsuite *ServerKeysTestSuite) Prepare() {
	var err error
	// From step Servers_Create
	fmt.Println("Call operation: Servers_Create")
	serversClient, err := armpostgresql.NewServersClient(testsuite.subscriptionId, testsuite.cred, testsuite.options)
	testsuite.Require().NoError(err)
	serversClientCreateResponsePoller, err := serversClient.BeginCreate(testsuite.ctx, testsuite.resourceGroupName, testsuite.serverName, armpostgresql.ServerForCreate{
		Location: to.Ptr(testsuite.location),
		Identity: &armpostgresql.ResourceIdentity{
			Type: to.Ptr(armpostgresql.IdentityTypeSystemAssigned),
		},
		Properties: &armpostgresql.ServerPropertiesForDefaultCreate{
			CreateMode:               to.Ptr(armpostgresql.CreateModeDefault),
			MinimalTLSVersion:        to.Ptr(armpostgresql.MinimalTLSVersionEnumTLS12),
			SSLEnforcement:           to.Ptr(armpostgresql.SSLEnforcementEnumEnabled),
			InfrastructureEncryption: to.Ptr(armpostgresql.InfrastructureEncryptionEnabled),
			StorageProfile: &armpostgresql.StorageProfile{
				BackupRetentionDays: to.Ptr[int32](7),
				GeoRedundantBackup:  to.Ptr(armpostgresql.GeoRedundantBackupDisabled),
				StorageMB:           to.Ptr[int32](128000),
			},
			AdministratorLogin:         to.Ptr("cloudsa"),
			AdministratorLoginPassword: to.Ptr(testsuite.adminPassword),
		},
		SKU: &armpostgresql.SKU{
			Name:   to.Ptr("GP_Gen5_8"),
			Family: to.Ptr("Gen5"),
			Tier:   to.Ptr(armpostgresql.SKUTierGeneralPurpose),
		},
		Tags: map[string]*string{
			"ElasticServer": to.Ptr("1"),
		},
	}, nil)
	testsuite.Require().NoError(err)
	serversClientCreateResponse, err := testutil.PollForTest(testsuite.ctx, serversClientCreateResponsePoller)
	testsuite.Require().NoError(err)
	testsuite.principalId = *serversClientCreateResponse.Identity.PrincipalID
}

// Microsoft.DBforPostgreSQL/servers/{serverName}/keys/{keyName}
func (testsuite *ServerKeysTestSuite) TestServerKeys() {
	var err error
	tenantId := recording.GetEnvVariable("AZURE_TENANT_ID", "00000000-0000-0000-0000-000000000000")
	keyVaultName, _ := recording.GenerateAlphaNumericID(testsuite.T(), "pgkv", 12, true)
	keyName := "pgkey"

	// From step KeyVault_Create
	fmt.Println("Deploy KeyVault and Key")
	template := map[string]any{
		"$schema":        "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
		"contentVersion": "1.0.0.0",
		"parameters": map[string]any{
			"keyVaultName": map[string]any{
				"type":         "string",
				"defaultValue": keyVaultName,
			},
			"location": map[string]any{
				"type":         "string",
				"defaultValue": testsuite.location,
			},
			"tenantId": map[string]any{
				"type":         "string",
				"defaultValue": tenantId,
			},
			"serverPrincipalId": map[string]any{
				"type":         "string",
				"defaultValue": testsuite.principalId,
			},
			"keyName": map[string]any{
				"type":         "string",
				"defaultValue": keyName,
			},
		},
		"outputs": map[string]any{
			"keyVaultKeyUri": map[string]any{
				"type":  "string",
				"value": "[reference(resourceId('Microsoft.KeyVault/vaults/keys', parameters('keyVaultName'), parameters('keyName')), '2021-10-01').keyUriWithVersion]",
			},
		},
		"resources": []any{
			map[string]any{
				"type":       "Microsoft.KeyVault/vaults",
				"apiVersion": "2021-10-01",
				"name":       "[parameters('keyVaultName')]",
				"location":   "[parameters('location')]",
				"properties": map[string]any{
					"sku": map[string]any{
						"family": "A",
						"name":   "standard",
					},
					"tenantId": "[parameters('tenantId')]",
					"accessPolicies": []any{
						map[string]any{
							"tenantId": "[parameters('tenantId')]",
							"objectId": "[parameters('serverPrincipalId')]",
							"permissions": map[string]any{
								"keys": []any{
									"get", "list", "create", "update", "delete", "wrapKey", "unwrapKey",
								},
							},
						},
					},
					"enableSoftDelete":    true,
					"enablePurgeProtection": false,
				},
			},
			map[string]any{
				"type":       "Microsoft.KeyVault/vaults/keys",
				"apiVersion": "2021-10-01",
				"name":       "[concat(parameters('keyVaultName'), '/', parameters('keyName'))]",
				"location":   "[parameters('location')]",
				"dependsOn": []any{
					"[resourceId('Microsoft.KeyVault/vaults', parameters('keyVaultName'))]",
				},
				"properties": map[string]any{
					"kty":     "RSA",
					"keySize": 2048,
					"keyOps":  []any{"encrypt", "decrypt", "sign", "verify", "wrapKey", "unwrapKey"},
				},
			},
		},
	}
	deployment := armdeployments.Deployment{
		Properties: &armdeployments.DeploymentProperties{
			Template: template,
			Mode:     to.Ptr(armdeployments.DeploymentModeIncremental),
		},
	}
	deploymentExtend, err := testutil.CreateDeployment(testsuite.ctx, testsuite.subscriptionId, testsuite.cred, testsuite.options, testsuite.resourceGroupName, "KeyVault_Create", &deployment)
	testsuite.Require().NoError(err)
	keyVaultKeyUri := deploymentExtend.Properties.Outputs.(map[string]interface{})["keyVaultKeyUri"].(map[string]interface{})["value"].(string)

	// Build the server key name from the key URI: {vaultName}_{keyName}_{keyVersion}
	// URI format: https://{vaultName}.vault.azure.net/keys/{keyName}/{version}
	uriParts := strings.Split(keyVaultKeyUri, "/")
	// uriParts: ["https:", "", "{vaultName}.vault.azure.net", "keys", "{keyName}", "{version}"]
	serverKeyName := strings.Split(uriParts[2], ".")[0] + "_" + uriParts[4] + "_" + uriParts[5]

	// From step ServerKeys_CreateOrUpdate
	fmt.Println("Call operation: ServerKeys_CreateOrUpdate")
	serverKeysClient, err := armpostgresql.NewServerKeysClient(testsuite.subscriptionId, testsuite.cred, testsuite.options)
	testsuite.Require().NoError(err)
	serverKeysClientCreateOrUpdateResponsePoller, err := serverKeysClient.BeginCreateOrUpdate(testsuite.ctx, testsuite.serverName, serverKeyName, testsuite.resourceGroupName, armpostgresql.ServerKey{
		Properties: &armpostgresql.ServerKeyProperties{
			ServerKeyType: to.Ptr(armpostgresql.ServerKeyTypeAzureKeyVault),
			URI:           to.Ptr(keyVaultKeyUri),
		},
	}, nil)
	testsuite.Require().NoError(err)
	_, err = testutil.PollForTest(testsuite.ctx, serverKeysClientCreateOrUpdateResponsePoller)
	testsuite.Require().NoError(err)

	// From step ServerKeys_List
	fmt.Println("Call operation: ServerKeys_List")
	serverKeysClientNewListPager := serverKeysClient.NewListPager(testsuite.resourceGroupName, testsuite.serverName, nil)
	for serverKeysClientNewListPager.More() {
		_, err := serverKeysClientNewListPager.NextPage(testsuite.ctx)
		testsuite.Require().NoError(err)
		break
	}

	// From step ServerKeys_Get
	fmt.Println("Call operation: ServerKeys_Get")
	_, err = serverKeysClient.Get(testsuite.ctx, testsuite.resourceGroupName, testsuite.serverName, serverKeyName, nil)
	testsuite.Require().NoError(err)

	// From step ServerKeys_Delete
	fmt.Println("Call operation: ServerKeys_Delete")
	serverKeysClientDeleteResponsePoller, err := serverKeysClient.BeginDelete(testsuite.ctx, testsuite.serverName, serverKeyName, testsuite.resourceGroupName, nil)
	testsuite.Require().NoError(err)
	_, err = testutil.PollForTest(testsuite.ctx, serverKeysClientDeleteResponsePoller)
	testsuite.Require().NoError(err)
}

func (testsuite *ServerKeysTestSuite) Cleanup() {
	var err error
	// From step Servers_Delete
	fmt.Println("Call operation: Servers_Delete")
	serversClient, err := armpostgresql.NewServersClient(testsuite.subscriptionId, testsuite.cred, testsuite.options)
	testsuite.Require().NoError(err)
	serversClientDeleteResponsePoller, err := serversClient.BeginDelete(testsuite.ctx, testsuite.resourceGroupName, testsuite.serverName, nil)
	testsuite.Require().NoError(err)
	_, err = testutil.PollForTest(testsuite.ctx, serversClientDeleteResponsePoller)
	testsuite.Require().NoError(err)
}
