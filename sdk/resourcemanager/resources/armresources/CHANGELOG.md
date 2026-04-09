# Release History

## 4.0.0 (2026-04-09)
### Breaking Changes

- Enum `AliasPathAttributes` has been removed
- Enum `AliasPathTokenType` has been removed
- Enum `AliasPatternType` has been removed
- Enum `AliasType` has been removed
- Enum `ExportTemplateOutputFormat` has been removed
- Enum `ExtendedLocationType` has been removed
- Enum `ProviderAuthorizationConsentState` has been removed
- Enum `ResourceIdentityType` has been removed
- Enum `TagsPatchOperation` has been removed
- Function `NewClient` has been removed
- Function `*Client.CheckExistence` has been removed
- Function `*Client.CheckExistenceByID` has been removed
- Function `*Client.BeginCreateOrUpdate` has been removed
- Function `*Client.BeginCreateOrUpdateByID` has been removed
- Function `*Client.BeginDelete` has been removed
- Function `*Client.BeginDeleteByID` has been removed
- Function `*Client.Get` has been removed
- Function `*Client.GetByID` has been removed
- Function `*Client.NewListByResourceGroupPager` has been removed
- Function `*Client.NewListPager` has been removed
- Function `*Client.BeginMoveResources` has been removed
- Function `*Client.BeginUpdate` has been removed
- Function `*Client.BeginUpdateByID` has been removed
- Function `*Client.BeginValidateMoveResources` has been removed
- Function `NewClientFactory` has been removed
- Function `*ClientFactory.NewClient` has been removed
- Function `*ClientFactory.NewOperationsClient` has been removed
- Function `*ClientFactory.NewProviderResourceTypesClient` has been removed
- Function `*ClientFactory.NewProvidersClient` has been removed
- Function `*ClientFactory.NewResourceGroupsClient` has been removed
- Function `*ClientFactory.NewTagsClient` has been removed
- Function `NewOperationsClient` has been removed
- Function `*OperationsClient.NewListPager` has been removed
- Function `NewProviderResourceTypesClient` has been removed
- Function `*ProviderResourceTypesClient.List` has been removed
- Function `NewProvidersClient` has been removed
- Function `*ProvidersClient.Get` has been removed
- Function `*ProvidersClient.GetAtTenantScope` has been removed
- Function `*ProvidersClient.NewListAtTenantScopePager` has been removed
- Function `*ProvidersClient.NewListPager` has been removed
- Function `*ProvidersClient.ProviderPermissions` has been removed
- Function `*ProvidersClient.Register` has been removed
- Function `*ProvidersClient.RegisterAtManagementGroupScope` has been removed
- Function `*ProvidersClient.Unregister` has been removed
- Function `NewResourceGroupsClient` has been removed
- Function `*ResourceGroupsClient.CheckExistence` has been removed
- Function `*ResourceGroupsClient.CreateOrUpdate` has been removed
- Function `*ResourceGroupsClient.BeginDelete` has been removed
- Function `*ResourceGroupsClient.BeginExportTemplate` has been removed
- Function `*ResourceGroupsClient.Get` has been removed
- Function `*ResourceGroupsClient.NewListPager` has been removed
- Function `*ResourceGroupsClient.Update` has been removed
- Function `NewTagsClient` has been removed
- Function `*TagsClient.CreateOrUpdate` has been removed
- Function `*TagsClient.BeginCreateOrUpdateAtScope` has been removed
- Function `*TagsClient.CreateOrUpdateValue` has been removed
- Function `*TagsClient.Delete` has been removed
- Function `*TagsClient.BeginDeleteAtScope` has been removed
- Function `*TagsClient.DeleteValue` has been removed
- Function `*TagsClient.GetAtScope` has been removed
- Function `*TagsClient.NewListPager` has been removed
- Function `*TagsClient.BeginUpdateAtScope` has been removed
- Struct `APIProfile` has been removed
- Struct `Alias` has been removed
- Struct `AliasPath` has been removed
- Struct `AliasPathMetadata` has been removed
- Struct `AliasPattern` has been removed
- Struct `ClientFactory` has been removed
- Struct `ErrorAdditionalInfo` has been removed
- Struct `ErrorResponse` has been removed
- Struct `ExportTemplateRequest` has been removed
- Struct `ExtendedLocation` has been removed
- Struct `GenericResource` has been removed
- Struct `GenericResourceExpanded` has been removed
- Struct `GenericResourceFilter` has been removed
- Struct `Identity` has been removed
- Struct `IdentityUserAssignedIdentitiesValue` has been removed
- Struct `MoveInfo` has been removed
- Struct `Operation` has been removed
- Struct `OperationDisplay` has been removed
- Struct `OperationListResult` has been removed
- Struct `Permission` has been removed
- Struct `Plan` has been removed
- Struct `Provider` has been removed
- Struct `ProviderConsentDefinition` has been removed
- Struct `ProviderExtendedLocation` has been removed
- Struct `ProviderListResult` has been removed
- Struct `ProviderPermission` has been removed
- Struct `ProviderPermissionListResult` has been removed
- Struct `ProviderRegistrationRequest` has been removed
- Struct `ProviderResourceType` has been removed
- Struct `ProviderResourceTypeListResult` has been removed
- Struct `Resource` has been removed
- Struct `ResourceGroup` has been removed
- Struct `ResourceGroupExportResult` has been removed
- Struct `ResourceGroupFilter` has been removed
- Struct `ResourceGroupListResult` has been removed
- Struct `ResourceGroupPatchable` has been removed
- Struct `ResourceGroupProperties` has been removed
- Struct `ResourceListResult` has been removed
- Struct `ResourceProviderOperationDisplayProperties` has been removed
- Struct `RoleDefinition` has been removed
- Struct `SKU` has been removed
- Struct `SubResource` has been removed
- Struct `TagCount` has been removed
- Struct `TagDetails` has been removed
- Struct `TagValue` has been removed
- Struct `Tags` has been removed
- Struct `TagsListResult` has been removed
- Struct `TagsPatchResource` has been removed
- Struct `TagsResource` has been removed
- Struct `ZoneMapping` has been removed


## 3.0.1 (2025-12-17)

### Other Changes

- Consolidate package version used in test.
- All things related to `*DeploymentOperationsClient` and `*DeploymentsClient` have been moved to package `github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armdeployments@v0.2.0`

## 3.0.0 (2025-06-10)
### Breaking Changes

- Enum `ChangeType` has been removed
- Enum `DeploymentMode` has been removed
- Enum `ExpressionEvaluationOptionsScopeType` has been removed
- Enum `ExtensionConfigPropertyType` has been removed
- Enum `Level` has been removed
- Enum `OnErrorDeploymentType` has been removed
- Enum `PropertyChangeType` has been removed
- Enum `ProvisioningOperation` has been removed
- Enum `ProvisioningState` has been removed
- Enum `ValidationLevel` has been removed
- Enum `WhatIfResultFormat` has been removed
- Function `*ClientFactory.NewDeploymentOperationsClient` has been removed
- Function `*ClientFactory.NewDeploymentsClient` has been removed
- Function `NewDeploymentOperationsClient` has been removed
- Function `*DeploymentOperationsClient.Get` has been removed
- Function `*DeploymentOperationsClient.GetAtManagementGroupScope` has been removed
- Function `*DeploymentOperationsClient.GetAtScope` has been removed
- Function `*DeploymentOperationsClient.GetAtSubscriptionScope` has been removed
- Function `*DeploymentOperationsClient.GetAtTenantScope` has been removed
- Function `*DeploymentOperationsClient.NewListAtManagementGroupScopePager` has been removed
- Function `*DeploymentOperationsClient.NewListAtScopePager` has been removed
- Function `*DeploymentOperationsClient.NewListAtSubscriptionScopePager` has been removed
- Function `*DeploymentOperationsClient.NewListAtTenantScopePager` has been removed
- Function `*DeploymentOperationsClient.NewListPager` has been removed
- Function `NewDeploymentsClient` has been removed
- Function `*DeploymentsClient.CalculateTemplateHash` has been removed
- Function `*DeploymentsClient.Cancel` has been removed
- Function `*DeploymentsClient.CancelAtManagementGroupScope` has been removed
- Function `*DeploymentsClient.CancelAtScope` has been removed
- Function `*DeploymentsClient.CancelAtSubscriptionScope` has been removed
- Function `*DeploymentsClient.CancelAtTenantScope` has been removed
- Function `*DeploymentsClient.CheckExistence` has been removed
- Function `*DeploymentsClient.CheckExistenceAtManagementGroupScope` has been removed
- Function `*DeploymentsClient.CheckExistenceAtScope` has been removed
- Function `*DeploymentsClient.CheckExistenceAtSubscriptionScope` has been removed
- Function `*DeploymentsClient.CheckExistenceAtTenantScope` has been removed
- Function `*DeploymentsClient.BeginCreateOrUpdate` has been removed
- Function `*DeploymentsClient.BeginCreateOrUpdateAtManagementGroupScope` has been removed
- Function `*DeploymentsClient.BeginCreateOrUpdateAtScope` has been removed
- Function `*DeploymentsClient.BeginCreateOrUpdateAtSubscriptionScope` has been removed
- Function `*DeploymentsClient.BeginCreateOrUpdateAtTenantScope` has been removed
- Function `*DeploymentsClient.BeginDelete` has been removed
- Function `*DeploymentsClient.BeginDeleteAtManagementGroupScope` has been removed
- Function `*DeploymentsClient.BeginDeleteAtScope` has been removed
- Function `*DeploymentsClient.BeginDeleteAtSubscriptionScope` has been removed
- Function `*DeploymentsClient.BeginDeleteAtTenantScope` has been removed
- Function `*DeploymentsClient.ExportTemplate` has been removed
- Function `*DeploymentsClient.ExportTemplateAtManagementGroupScope` has been removed
- Function `*DeploymentsClient.ExportTemplateAtScope` has been removed
- Function `*DeploymentsClient.ExportTemplateAtSubscriptionScope` has been removed
- Function `*DeploymentsClient.ExportTemplateAtTenantScope` has been removed
- Function `*DeploymentsClient.Get` has been removed
- Function `*DeploymentsClient.GetAtManagementGroupScope` has been removed
- Function `*DeploymentsClient.GetAtScope` has been removed
- Function `*DeploymentsClient.GetAtSubscriptionScope` has been removed
- Function `*DeploymentsClient.GetAtTenantScope` has been removed
- Function `*DeploymentsClient.NewListAtManagementGroupScopePager` has been removed
- Function `*DeploymentsClient.NewListAtScopePager` has been removed
- Function `*DeploymentsClient.NewListAtSubscriptionScopePager` has been removed
- Function `*DeploymentsClient.NewListAtTenantScopePager` has been removed
- Function `*DeploymentsClient.NewListByResourceGroupPager` has been removed
- Function `*DeploymentsClient.BeginValidate` has been removed
- Function `*DeploymentsClient.BeginValidateAtManagementGroupScope` has been removed
- Function `*DeploymentsClient.BeginValidateAtScope` has been removed
- Function `*DeploymentsClient.BeginValidateAtSubscriptionScope` has been removed
- Function `*DeploymentsClient.BeginValidateAtTenantScope` has been removed
- Function `*DeploymentsClient.BeginWhatIf` has been removed
- Function `*DeploymentsClient.BeginWhatIfAtManagementGroupScope` has been removed
- Function `*DeploymentsClient.BeginWhatIfAtSubscriptionScope` has been removed
- Function `*DeploymentsClient.BeginWhatIfAtTenantScope` has been removed
- Struct `BasicDependency` has been removed
- Struct `DebugSetting` has been removed
- Struct `Dependency` has been removed
- Struct `Deployment` has been removed
- Struct `DeploymentDiagnosticsDefinition` has been removed
- Struct `DeploymentExportResult` has been removed
- Struct `DeploymentExtended` has been removed
- Struct `DeploymentExtendedFilter` has been removed
- Struct `DeploymentExtensionConfigItem` has been removed
- Struct `DeploymentExtensionDefinition` has been removed
- Struct `DeploymentListResult` has been removed
- Struct `DeploymentOperation` has been removed
- Struct `DeploymentOperationProperties` has been removed
- Struct `DeploymentOperationsListResult` has been removed
- Struct `DeploymentParameter` has been removed
- Struct `DeploymentProperties` has been removed
- Struct `DeploymentPropertiesExtended` has been removed
- Struct `DeploymentValidateResult` has been removed
- Struct `DeploymentWhatIf` has been removed
- Struct `DeploymentWhatIfProperties` has been removed
- Struct `DeploymentWhatIfSettings` has been removed
- Struct `ExpressionEvaluationOptions` has been removed
- Struct `HTTPMessage` has been removed
- Struct `KeyVaultParameterReference` has been removed
- Struct `KeyVaultReference` has been removed
- Struct `OnErrorDeployment` has been removed
- Struct `OnErrorDeploymentExtended` has been removed
- Struct `ParametersLink` has been removed
- Struct `ResourceReference` has been removed
- Struct `ScopedDeployment` has been removed
- Struct `ScopedDeploymentWhatIf` has been removed
- Struct `StatusMessage` has been removed
- Struct `TargetResource` has been removed
- Struct `TemplateHashResult` has been removed
- Struct `TemplateLink` has been removed
- Struct `WhatIfChange` has been removed
- Struct `WhatIfOperationProperties` has been removed
- Struct `WhatIfOperationResult` has been removed
- Struct `WhatIfPropertyChange` has been removed


## 2.1.0 (2025-05-06)
### Features Added

- New enum type `ExtensionConfigPropertyType` with values `ExtensionConfigPropertyTypeArray`, `ExtensionConfigPropertyTypeBool`, `ExtensionConfigPropertyTypeInt`, `ExtensionConfigPropertyTypeObject`, `ExtensionConfigPropertyTypeSecureObject`, `ExtensionConfigPropertyTypeSecureString`, `ExtensionConfigPropertyTypeString`
- New struct `DeploymentExtensionConfigItem`
- New struct `DeploymentExtensionDefinition`
- New field `ExtensionConfigs` in struct `DeploymentProperties`
- New field `Extensions` in struct `DeploymentPropertiesExtended`
- New field `ExtensionConfigs` in struct `DeploymentWhatIfProperties`
- New field `APIVersion`, `Extension`, `Identifiers`, `ResourceType` in struct `ResourceReference`
- New field `APIVersion`, `Extension`, `Identifiers`, `SymbolicName` in struct `TargetResource`


## 2.0.0 (2025-02-13)
### Breaking Changes

- Type of `DeploymentProperties.Parameters` has been changed from `any` to `map[string]*DeploymentParameter`
- Type of `DeploymentWhatIfProperties.Parameters` has been changed from `any` to `map[string]*DeploymentParameter`
- Operation `*TagsClient.CreateOrUpdateAtScope` has been changed to LRO, use `*TagsClient.BeginCreateOrUpdateAtScope` instead.
- Operation `*TagsClient.DeleteAtScope` has been changed to LRO, use `*TagsClient.BeginDeleteAtScope` instead.
- Operation `*TagsClient.UpdateAtScope` has been changed to LRO, use `*TagsClient.BeginUpdateAtScope` instead.

### Features Added

- New enum type `ExportTemplateOutputFormat` with values `ExportTemplateOutputFormatBicep`, `ExportTemplateOutputFormatJSON`
- New enum type `Level` with values `LevelError`, `LevelInfo`, `LevelWarning`
- New enum type `ValidationLevel` with values `ValidationLevelProvider`, `ValidationLevelProviderNoRbac`, `ValidationLevelTemplate`
- New struct `DeploymentDiagnosticsDefinition`
- New struct `DeploymentParameter`
- New struct `KeyVaultParameterReference`
- New struct `KeyVaultReference`
- New field `ValidationLevel` in struct `DeploymentProperties`
- New field `Diagnostics`, `ValidationLevel` in struct `DeploymentPropertiesExtended`
- New field `ID`, `Name`, `Type` in struct `DeploymentValidateResult`
- New field `ValidationLevel` in struct `DeploymentWhatIfProperties`
- New field `OutputFormat` in struct `ExportTemplateRequest`
- New field `Output` in struct `ResourceGroupExportResult`
- New field `DeploymentID`, `Identifiers`, `SymbolicName` in struct `WhatIfChange`
- New field `Diagnostics`, `PotentialChanges` in struct `WhatIfOperationProperties`


## 1.2.0 (2023-11-24)
### Features Added

- Support for test fakes and OpenTelemetry trace spans.


## 1.2.0-beta.3 (2023-10-09)

### Other Changes

- Updated to latest `azcore` beta.

## 1.2.0-beta.2 (2023-07-19)

### Bug Fixes

- Fixed a potential panic in faked paged and long-running operations.

## 1.2.0-beta.1 (2023-06-12)

### Features Added

- Support for test fakes and OpenTelemetry trace spans.

## 1.1.1 (2023-04-14)
### Bug Fixes

- Fix serialization bug of empty value of `any` type.


## 1.1.0 (2023-03-27)
### Features Added

- New struct `ClientFactory` which is a client factory used to create any client in this module


## 1.0.0 (2022-05-16)
## Other Changes

The package of `github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources` is using our [next generation design principles](https://azure.github.io/azure-sdk/general_introduction.html) since version 1.0.0, which contains breaking changes.

To migrate the existing applications to the latest version, please refer to [Migration Guide](https://aka.ms/azsdk/go/mgmt/migration).

To learn more, please refer to our documentation [Quick Start](https://aka.ms/azsdk/go/mgmt).