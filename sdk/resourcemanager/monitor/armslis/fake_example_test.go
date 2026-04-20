// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package armslis_test

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	azfake "github.com/Azure/azure-sdk-for-go/sdk/azcore/fake"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armslis"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armslis/fake"
)

func ExampleServer() {
	// first, create an instance of the fake server for the client you wish to test.
	fakeServer := fake.Server{

		// provide implementations for the APIs you wish to fake.
		// this fake corresponds to the Client.Get() API.
		Get: func(ctx context.Context, serviceGroupName string, sliName string, options *armslis.ClientGetOptions) (resp azfake.Responder[armslis.ClientGetResponse], errResp azfake.ErrorResponder) {
			// construct the response type, populating fields as required
			getResp := armslis.ClientGetResponse{}
			getResp.Properties = &armslis.SliResource{
				Description:    to.Ptr("Test SLI"),
				Category:       to.Ptr(armslis.CategoryLatency),
				EvaluationType: to.Ptr(armslis.EvaluationTypeWindowBased),
				EnableAlert:    to.Ptr(false),
			}

			// use resp to set the desired response
			resp.SetResponse(http.StatusOK, getResp, nil)

			return
		},
	}

	// now create the corresponding client, connecting the fake server via the client options
	client, err := armslis.NewClient(&azfake.TokenCredential{}, &arm.ClientOptions{
		ClientOptions: azcore.ClientOptions{
			Transport: fake.NewServerTransport(&fakeServer),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// call the API. the provided values will be passed to the fake's implementation.
	resp, err := client.Get(context.TODO(), "testSG", "testSli", nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(*resp.Properties.Description)
	// Output: Test SLI
}
