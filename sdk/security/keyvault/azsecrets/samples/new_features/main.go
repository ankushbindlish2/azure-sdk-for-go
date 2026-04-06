// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

// This sample demonstrates new secret features:
//   - OutContentType: request PEM format for a certificate-backed secret originally stored as PFX
//   - PreviousVersion: inspect the previous version identifier of a secret
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
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
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

	secretClient, err := azsecrets.NewClient(vaultURL, cred, nil)
	if err != nil {
		log.Fatalf("failed to create secrets client: %v", err)
	}

	certClient, err := azcertificates.NewClient(vaultURL, cred, nil)
	if err != nil {
		log.Fatalf("failed to create certificates client: %v", err)
	}

	ctx := context.Background()

	// ── Part 1: PreviousVersion ──────────────────────────────────────────
	secretName := fmt.Sprintf("go-prev-ver-%d", time.Now().Unix())
	fmt.Println("=== PreviousVersion Demo ===")

	// Create a secret with two versions
	fmt.Println("Creating secret version 1...")
	_, err = secretClient.SetSecret(ctx, secretName, azsecrets.SetSecretParameters{
		Value: to.Ptr("version-one"),
	}, nil)
	if err != nil {
		log.Fatalf("failed to set secret v1: %v", err)
	}

	fmt.Println("Creating secret version 2...")
	_, err = secretClient.SetSecret(ctx, secretName, azsecrets.SetSecretParameters{
		Value: to.Ptr("version-two"),
	}, nil)
	if err != nil {
		log.Fatalf("failed to set secret v2: %v", err)
	}

	// Retrieve the latest secret and check PreviousVersion
	resp, err := secretClient.GetSecret(ctx, secretName, "", nil)
	if err != nil {
		log.Fatalf("failed to get secret: %v", err)
	}
	fmt.Printf("Secret Name: %s\n", resp.ID.Name())
	fmt.Printf("Secret Value: %s\n", *resp.Value)
	if resp.PreviousVersion != nil {
		fmt.Printf("Previous Version: %s\n", *resp.PreviousVersion)
	} else {
		fmt.Println("Previous Version: (not populated by service yet)")
	}

	// Clean up the secret
	fmt.Println("Deleting test secret...")
	_, err = secretClient.DeleteSecret(ctx, secretName, nil)
	if err != nil {
		log.Fatalf("failed to delete secret: %v", err)
	}

	// ── Part 2: OutContentType (PFX → PEM conversion) ───────────────────
	certName := fmt.Sprintf("go-out-ct-%d", time.Now().Unix())
	fmt.Println("\n=== OutContentType Demo ===")
	fmt.Println("Creating a PFX certificate to demonstrate out_content_type conversion...")

	createParams := azcertificates.CreateCertificateParameters{
		CertificatePolicy: &azcertificates.CertificatePolicy{
			IssuerParameters: &azcertificates.IssuerParameters{
				Name: to.Ptr("self"),
			},
			X509CertificateProperties: &azcertificates.X509CertificateProperties{
				Subject: to.Ptr("CN=GoOutContentTypeTest"),
			},
			SecretProperties: &azcertificates.SecretProperties{
				ContentType: to.Ptr("application/x-pkcs12"),
			},
		},
	}

	_, err = certClient.CreateCertificate(ctx, certName, createParams, nil)
	if err != nil {
		log.Fatalf("failed to create certificate: %v", err)
	}

	// Poll until the certificate is ready
	fmt.Println("Waiting for certificate creation to complete...")
	for {
		op, err := certClient.GetCertificateOperation(ctx, certName, nil)
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

	// Retrieve the certificate-backed secret with OutContentType=PEM
	fmt.Println("Retrieving certificate-backed secret as PEM via OutContentType...")
	secretResp, err := secretClient.GetSecret(ctx, certName, "", &azsecrets.GetSecretOptions{
		OutContentType: to.Ptr(azsecrets.ContentTypePEM),
	})
	if err != nil {
		log.Fatalf("failed to get secret with OutContentType: %v", err)
	}

	if secretResp.Value != nil && strings.Contains(*secretResp.Value, "-----BEGIN") {
		// Print just the first 80 characters to confirm PEM format
		preview := *secretResp.Value
		if len(preview) > 80 {
			preview = preview[:80] + "..."
		}
		fmt.Printf("Successfully converted to PEM format: %s\n", preview)
	} else {
		fmt.Println("Warning: response does not appear to be PEM-encoded")
	}

	// Check PreviousVersion on the cert-backed secret too
	if secretResp.PreviousVersion != nil {
		fmt.Printf("Certificate secret's PreviousVersion: %s\n", *secretResp.PreviousVersion)
	} else {
		fmt.Println("Certificate secret's PreviousVersion: (not populated by service yet)")
	}

	// Clean up: delete the certificate (which also removes the cert-backed secret)
	fmt.Println("Deleting test certificate...")
	_, err = certClient.DeleteCertificate(ctx, certName, nil)
	if err != nil {
		log.Fatalf("failed to delete certificate: %v", err)
	}

	fmt.Println("Done.")
}
