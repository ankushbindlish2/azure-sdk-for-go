// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package armslis_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/internal/recording"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/internal/v3/testutil"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armslis"
	"github.com/stretchr/testify/suite"
)

type SliTestSuite struct {
	suite.Suite

	ctx              context.Context
	cred             azcore.TokenCredential
	options          *arm.ClientOptions
	sliName          string
	serviceGroupName string
}

func (testsuite *SliTestSuite) SetupSuite() {
	testutil.StartRecording(testsuite.T(), pathToPackage)
	testsuite.ctx = context.Background()
	testsuite.cred, testsuite.options = testutil.GetCredAndClientOptions(testsuite.T())
	testsuite.sliName, _ = recording.GenerateAlphaNumericID(testsuite.T(), "sli", 10, false)
	testsuite.serviceGroupName = recording.GetEnvVariable("SERVICE_GROUP_NAME", "testSG")
}

func (testsuite *SliTestSuite) TearDownSuite() {
	testutil.StopRecording(testsuite.T())
}

func TestSliTestSuite(t *testing.T) {
	suite.Run(t, new(SliTestSuite))
}

func (testsuite *SliTestSuite) TestSliCRUD() {
	var err error

	// Create client
	client, err := armslis.NewClient(testsuite.cred, testsuite.options)
	testsuite.Require().NoError(err)

	// CreateOrUpdate
	fmt.Println("Call operation: Client.CreateOrUpdate")
	sliResource := armslis.Sli{
		Properties: &armslis.SliResource{
			Description:    to.Ptr("Test SLI for live test"),
			Category:       to.Ptr(armslis.CategoryLatency),
			EvaluationType: to.Ptr(armslis.EvaluationTypeWindowBased),
			EnableAlert:    to.Ptr(false),
			DestinationAmwAccounts: []*armslis.AmwAccount{
				{
					ResourceID: to.Ptr("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/microsoft.monitor/accounts/dest"),
					Identity:   to.Ptr("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/id"),
				},
			},
			BaselineProperties: &armslis.BaselineProperties{
				Baseline: &armslis.Baseline{
					Value:                     to.Ptr[float32](99),
					EvaluationPeriodDays:      to.Ptr[int32](30),
					EvaluationCalculationType: to.Ptr(armslis.EvaluationCalculationTypeCalendarDays),
				},
			},
			SliProperties: &armslis.SliProperties{
				WindowUptimeCriteria: &armslis.WindowUptimeCriteria{
					Target:     to.Ptr[float32](95),
					Comparator: to.Ptr(armslis.WindowUptimeCriteriaComparatorGreaterThanOrEqual),
				},
				Signals: &armslis.Signal{
					SignalSources: []*armslis.SignalSource{
						{
							SignalSourceID:                  to.Ptr("A"),
							SourceAmwAccountManagedIdentity: to.Ptr("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/id"),
							SourceAmwAccountResourceID:      to.Ptr("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/microsoft.monitor/accounts/src"),
							MetricNamespace:                 to.Ptr("TestMetrics"),
							MetricName:                      to.Ptr("Latency"),
							Filters:                         []*armslis.Condition{},
							SpatialAggregation: &armslis.SpatialAggregation{
								Type:       to.Ptr(armslis.SpatialAggregationTypeAverage),
								Dimensions: []*string{},
							},
							TemporalAggregation: &armslis.TemporalAggregation{
								Type:              to.Ptr(armslis.TemporalAggregationTypeAverage),
								WindowSizeMinutes: to.Ptr[int32](5),
							},
						},
					},
					SignalFormula: to.Ptr("A"),
				},
			},
		},
	}
	createResp, err := client.CreateOrUpdate(testsuite.ctx, testsuite.serviceGroupName, testsuite.sliName, sliResource, nil)
	testsuite.Require().NoError(err)
	testsuite.Require().NotNil(createResp.Properties)

	// Get
	fmt.Println("Call operation: Client.Get")
	getResp, err := client.Get(testsuite.ctx, testsuite.serviceGroupName, testsuite.sliName, nil)
	testsuite.Require().NoError(err)
	testsuite.Require().NotNil(getResp.Properties)

	// ListByParent
	fmt.Println("Call operation: Client.NewListByParentPager")
	pager := client.NewListByParentPager(testsuite.serviceGroupName, nil)
	testsuite.Require().True(pager.More())
	page, err := pager.NextPage(testsuite.ctx)
	testsuite.Require().NoError(err)
	testsuite.Require().NotEmpty(page.Value)

	// Delete
	fmt.Println("Call operation: Client.Delete")
	_, err = client.Delete(testsuite.ctx, testsuite.serviceGroupName, testsuite.sliName, nil)
	testsuite.Require().NoError(err)
}
