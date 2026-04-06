// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

// This sample demonstrates new certificate features:
//   - IP addresses in Subject Alternative Names (SANs)
//   - URIs in Subject Alternative Names (SANs)
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azcertificates"
)

func main() {
	vaultURL := os.Getenv("AZURE_KEYVAULT_URL")
	if vaultURL == "" {
		log.Fatal("AZURE_KEYVAULT_URL environment variable must be set")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("failed to obtain a credential: %v", err)
	}

	client, err := azcertificates.NewClient(vaultURL, cred, nil)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	certName := fmt.Sprintf("go-san-cert-%d", time.Now().Unix())
	ctx := context.Background()

	// Create a self-signed certificate with IP addresses and URIs in SANs
	fmt.Println("Creating a self-signed certificate with IP and URI SANs...")
	createParams := azcertificates.CreateCertificateParameters{
		CertificatePolicy: &azcertificates.CertificatePolicy{
			IssuerParameters: &azcertificates.IssuerParameters{
				Name: to.Ptr("self"),
			},
			X509CertificateProperties: &azcertificates.X509CertificateProperties{
				Subject: to.Ptr("CN=GoSanTest"),
				SubjectAlternativeNames: &azcertificates.SubjectAlternativeNames{
					IPAddresses: []*string{
						to.Ptr("10.0.0.1"),
						to.Ptr("2001:db8::1"),
					},
					URIs: []*string{
						to.Ptr("https://mydomain.com"),
					},
				},
			},
			SecretProperties: &azcertificates.SecretProperties{
				ContentType: to.Ptr("application/x-pkcs12"),
			},
		},
	}

	_, err = client.CreateCertificate(ctx, certName, createParams, nil)
	if err != nil {
		log.Fatalf("failed to create certificate: %v", err)
	}

	// Poll until the certificate operation completes
	fmt.Println("Waiting for certificate creation to complete...")
	for {
		op, err := client.GetCertificateOperation(ctx, certName, nil)
		if err != nil {
			log.Fatalf("failed to get certificate operation: %v", err)
		}
		if op.Status != nil && strings.EqualFold(*op.Status, "completed") {
			break
		}
		if op.Error != nil && op.Error.Code != "" {
			log.Fatalf("certificate operation failed: %s", op.Error.Error())
		}
		time.Sleep(2 * time.Second)
	}

	// Retrieve the certificate and inspect SANs
	resp, err := client.GetCertificate(ctx, certName, "", nil)
	if err != nil {
		log.Fatalf("failed to get certificate: %v", err)
	}

	fmt.Printf("Certificate ID: %s\n", *resp.ID)

	if resp.Policy != nil && resp.Policy.X509CertificateProperties != nil && resp.Policy.X509CertificateProperties.SubjectAlternativeNames != nil {
		sans := resp.Policy.X509CertificateProperties.SubjectAlternativeNames

		fmt.Println("IP Address SANs:")
		for _, ip := range sans.IPAddresses {
			fmt.Printf("  %s\n", *ip)
		}

		fmt.Println("URI SANs:")
		for _, uri := range sans.URIs {
			fmt.Printf("  %s\n", *uri)
		}
	}

	// Clean up: delete the certificate
	fmt.Println("Deleting test certificate...")
	_, err = client.DeleteCertificate(ctx, certName, nil)
	if err != nil {
		log.Fatalf("failed to delete certificate: %v", err)
	}
	fmt.Println("Done.")
}
