# Release History

## 1.0.0 (2026-04-09)
### Breaking Changes

- Function `NewPolicyStatesClient` parameter(s) have been changed from `(credential azcore.TokenCredential, options *arm.ClientOptions)` to `(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions)`
- Function `*PolicyStatesClient.BeginTriggerResourceGroupEvaluation` parameter(s) have been changed from `(ctx context.Context, subscriptionID string, resourceGroupName string, options *PolicyStatesClientBeginTriggerResourceGroupEvaluationOptions)` to `(ctx context.Context, resourceGroupName string, options *PolicyStatesClientBeginTriggerResourceGroupEvaluationOptions)`
- Function `*PolicyStatesClient.BeginTriggerSubscriptionEvaluation` parameter(s) have been changed from `(ctx context.Context, subscriptionID string, options *PolicyStatesClientBeginTriggerSubscriptionEvaluationOptions)` to `(ctx context.Context, options *PolicyStatesClientBeginTriggerSubscriptionEvaluationOptions)`
- Enum `ComponentPolicyStatesResource` has been removed
- Enum `PolicyEventsResourceType` has been removed
- Enum `PolicyStatesResource` has been removed
- Enum `PolicyStatesSummaryResourceType` has been removed
- Enum `PolicyTrackedResourcesResourceType` has been removed
- Function `*AttestationsClient.NewListForResourceGroupPager` has been removed
- Function `*AttestationsClient.NewListForResourcePager` has been removed
- Function `*AttestationsClient.NewListForSubscriptionPager` has been removed
- Function `*ClientFactory.NewComponentPolicyStatesClient` has been removed
- Function `*ClientFactory.NewPolicyEventsClient` has been removed
- Function `*ClientFactory.NewPolicyTrackedResourcesClient` has been removed
- Function `NewComponentPolicyStatesClient` has been removed
- Function `*ComponentPolicyStatesClient.ListQueryResultsForPolicyDefinition` has been removed
- Function `*ComponentPolicyStatesClient.ListQueryResultsForResource` has been removed
- Function `*ComponentPolicyStatesClient.ListQueryResultsForResourceGroup` has been removed
- Function `*ComponentPolicyStatesClient.ListQueryResultsForResourceGroupLevelPolicyAssignment` has been removed
- Function `*ComponentPolicyStatesClient.ListQueryResultsForSubscription` has been removed
- Function `*ComponentPolicyStatesClient.ListQueryResultsForSubscriptionLevelPolicyAssignment` has been removed
- Function `NewPolicyEventsClient` has been removed
- Function `*PolicyEventsClient.NewListQueryResultsForManagementGroupPager` has been removed
- Function `*PolicyEventsClient.NewListQueryResultsForPolicyDefinitionPager` has been removed
- Function `*PolicyEventsClient.NewListQueryResultsForPolicySetDefinitionPager` has been removed
- Function `*PolicyEventsClient.NewListQueryResultsForResourceGroupLevelPolicyAssignmentPager` has been removed
- Function `*PolicyEventsClient.NewListQueryResultsForResourceGroupPager` has been removed
- Function `*PolicyEventsClient.NewListQueryResultsForResourcePager` has been removed
- Function `*PolicyEventsClient.NewListQueryResultsForSubscriptionLevelPolicyAssignmentPager` has been removed
- Function `*PolicyEventsClient.NewListQueryResultsForSubscriptionPager` has been removed
- Function `*PolicyMetadataClient.NewListPager` has been removed
- Function `*PolicyStatesClient.NewListQueryResultsForManagementGroupPager` has been removed
- Function `*PolicyStatesClient.NewListQueryResultsForPolicyDefinitionPager` has been removed
- Function `*PolicyStatesClient.NewListQueryResultsForPolicySetDefinitionPager` has been removed
- Function `*PolicyStatesClient.NewListQueryResultsForResourceGroupLevelPolicyAssignmentPager` has been removed
- Function `*PolicyStatesClient.NewListQueryResultsForResourceGroupPager` has been removed
- Function `*PolicyStatesClient.NewListQueryResultsForResourcePager` has been removed
- Function `*PolicyStatesClient.NewListQueryResultsForSubscriptionLevelPolicyAssignmentPager` has been removed
- Function `*PolicyStatesClient.NewListQueryResultsForSubscriptionPager` has been removed
- Function `*PolicyStatesClient.SummarizeForManagementGroup` has been removed
- Function `*PolicyStatesClient.SummarizeForPolicyDefinition` has been removed
- Function `*PolicyStatesClient.SummarizeForPolicySetDefinition` has been removed
- Function `*PolicyStatesClient.SummarizeForResource` has been removed
- Function `*PolicyStatesClient.SummarizeForResourceGroup` has been removed
- Function `*PolicyStatesClient.SummarizeForResourceGroupLevelPolicyAssignment` has been removed
- Function `*PolicyStatesClient.SummarizeForSubscription` has been removed
- Function `*PolicyStatesClient.SummarizeForSubscriptionLevelPolicyAssignment` has been removed
- Function `NewPolicyTrackedResourcesClient` has been removed
- Function `*PolicyTrackedResourcesClient.NewListQueryResultsForManagementGroupPager` has been removed
- Function `*PolicyTrackedResourcesClient.NewListQueryResultsForResourceGroupPager` has been removed
- Function `*PolicyTrackedResourcesClient.NewListQueryResultsForResourcePager` has been removed
- Function `*PolicyTrackedResourcesClient.NewListQueryResultsForSubscriptionPager` has been removed
- Function `*RemediationsClient.NewListDeploymentsAtManagementGroupPager` has been removed
- Function `*RemediationsClient.NewListDeploymentsAtResourceGroupPager` has been removed
- Function `*RemediationsClient.NewListDeploymentsAtResourcePager` has been removed
- Function `*RemediationsClient.NewListDeploymentsAtSubscriptionPager` has been removed
- Function `*RemediationsClient.NewListForManagementGroupPager` has been removed
- Function `*RemediationsClient.NewListForResourceGroupPager` has been removed
- Function `*RemediationsClient.NewListForResourcePager` has been removed
- Function `*RemediationsClient.NewListForSubscriptionPager` has been removed
- Struct `AttestationListResult` has been removed
- Struct `ComplianceDetail` has been removed
- Struct `ComponentEventDetails` has been removed
- Struct `ComponentExpressionEvaluationDetails` has been removed
- Struct `ComponentPolicyEvaluationDetails` has been removed
- Struct `ComponentPolicyState` has been removed
- Struct `ComponentPolicyStatesQueryResults` has been removed
- Struct `ComponentStateDetails` has been removed
- Struct `ErrorDefinition` has been removed
- Struct `ErrorDefinitionAutoGenerated` has been removed
- Struct `ErrorDefinitionAutoGenerated2` has been removed
- Struct `ErrorResponse` has been removed
- Struct `ErrorResponseAutoGenerated` has been removed
- Struct `ErrorResponseAutoGenerated2` has been removed
- Struct `PolicyAssignmentSummary` has been removed
- Struct `PolicyDefinitionSummary` has been removed
- Struct `PolicyDetails` has been removed
- Struct `PolicyEvaluationDetails` has been removed
- Struct `PolicyEvent` has been removed
- Struct `PolicyEventsQueryResults` has been removed
- Struct `PolicyGroupSummary` has been removed
- Struct `PolicyMetadataCollection` has been removed
- Struct `PolicyMetadataSlimProperties` has been removed
- Struct `PolicyState` has been removed
- Struct `PolicyStatesQueryResults` has been removed
- Struct `PolicyTrackedResource` has been removed
- Struct `PolicyTrackedResourcesQueryResults` has been removed
- Struct `QueryFailure` has been removed
- Struct `QueryFailureError` has been removed
- Struct `QueryOptions` has been removed
- Struct `RemediationDeployment` has been removed
- Struct `RemediationDeploymentsListResult` has been removed
- Struct `RemediationListResult` has been removed
- Struct `Resource` has been removed
- Struct `SlimPolicyMetadata` has been removed
- Struct `SummarizeResults` has been removed
- Struct `Summary` has been removed
- Struct `SummaryResults` has been removed
- Struct `TrackedResourceModificationDetails` has been removed
- Struct `TypedErrorInfo` has been removed

### Features Added

- New field `SystemData` in struct `PolicyMetadata`


## 0.9.0 (2025-07-25)
### Breaking Changes

- Type of `PolicyEvaluationResult.EvaluationDetails` has been changed from `*PolicyEvaluationDetails` to `*CheckRestrictionEvaluationDetails`

### Features Added

- New value `FieldRestrictionResultAudit` added to enum type `FieldRestrictionResult`
- New enum type `ComponentPolicyStatesResource` with values `ComponentPolicyStatesResourceLatest`
- New function `*ClientFactory.NewComponentPolicyStatesClient() *ComponentPolicyStatesClient`
- New function `NewComponentPolicyStatesClient(azcore.TokenCredential, *arm.ClientOptions) (*ComponentPolicyStatesClient, error)`
- New function `*ComponentPolicyStatesClient.ListQueryResultsForPolicyDefinition(context.Context, string, string, ComponentPolicyStatesResource, *ComponentPolicyStatesClientListQueryResultsForPolicyDefinitionOptions) (ComponentPolicyStatesClientListQueryResultsForPolicyDefinitionResponse, error)`
- New function `*ComponentPolicyStatesClient.ListQueryResultsForResource(context.Context, string, ComponentPolicyStatesResource, *ComponentPolicyStatesClientListQueryResultsForResourceOptions) (ComponentPolicyStatesClientListQueryResultsForResourceResponse, error)`
- New function `*ComponentPolicyStatesClient.ListQueryResultsForResourceGroup(context.Context, string, string, ComponentPolicyStatesResource, *ComponentPolicyStatesClientListQueryResultsForResourceGroupOptions) (ComponentPolicyStatesClientListQueryResultsForResourceGroupResponse, error)`
- New function `*ComponentPolicyStatesClient.ListQueryResultsForResourceGroupLevelPolicyAssignment(context.Context, string, string, string, ComponentPolicyStatesResource, *ComponentPolicyStatesClientListQueryResultsForResourceGroupLevelPolicyAssignmentOptions) (ComponentPolicyStatesClientListQueryResultsForResourceGroupLevelPolicyAssignmentResponse, error)`
- New function `*ComponentPolicyStatesClient.ListQueryResultsForSubscription(context.Context, string, ComponentPolicyStatesResource, *ComponentPolicyStatesClientListQueryResultsForSubscriptionOptions) (ComponentPolicyStatesClientListQueryResultsForSubscriptionResponse, error)`
- New function `*ComponentPolicyStatesClient.ListQueryResultsForSubscriptionLevelPolicyAssignment(context.Context, string, string, ComponentPolicyStatesResource, *ComponentPolicyStatesClientListQueryResultsForSubscriptionLevelPolicyAssignmentOptions) (ComponentPolicyStatesClientListQueryResultsForSubscriptionLevelPolicyAssignmentResponse, error)`
- New struct `CheckRestrictionEvaluationDetails`
- New struct `ComponentExpressionEvaluationDetails`
- New struct `ComponentPolicyEvaluationDetails`
- New struct `ComponentPolicyState`
- New struct `ComponentPolicyStatesQueryResults`
- New struct `PolicyEffectDetails`
- New field `IncludeAuditEffect` in struct `CheckRestrictionsRequest`
- New field `PolicyEffect`, `Reason` in struct `FieldRestriction`
- New field `IsDataAction` in struct `Operation`
- New field `EffectDetails` in struct `PolicyEvaluationResult`
- New field `ResourceIDs` in struct `RemediationFilters`


## 0.8.0 (2023-11-24)
### Features Added

- Support for test fakes and OpenTelemetry trace spans.


## 0.7.1 (2023-04-14)
### Bug Fixes

- Fix serialization bug of empty value of `any` type.


## 0.7.0 (2023-03-31)
### Features Added

- New struct `ClientFactory` which is a client factory used to create any client in this module


## 0.6.0 (2022-10-07)
### Features Added

- New field `Metadata` in struct `AttestationProperties`
- New field `AssessmentDate` in struct `AttestationProperties`


## 0.5.0 (2022-05-17)

The package of `github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/policyinsights/armpolicyinsights` is using our [next generation design principles](https://azure.github.io/azure-sdk/general_introduction.html) since version 0.5.0, which contains breaking changes.

To migrate the existing applications to the latest version, please refer to [Migration Guide](https://aka.ms/azsdk/go/mgmt/migration).

To learn more, please refer to our documentation [Quick Start](https://aka.ms/azsdk/go/mgmt).