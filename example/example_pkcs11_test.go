// Copyright (c) 2016, 2018, 2026, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.

// Example code for using a PKCS#11-backed user principal provider.

package example

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/nrdcg/oci-go-sdk/common/v1065"
	"github.com/nrdcg/oci-go-sdk/common/v1065/pkcs11auth"
	"github.com/nrdcg/oci-go-sdk/identity/v1065"
)

func callListAvailabilityDomains(provider common.ConfigurationProvider) error {
	client, err := identity.NewIdentityClientWithConfigurationProvider(provider)
	if err != nil {
		return fmt.Errorf("failed to create identity client: %w", err)
	}

	tenancyID, err := provider.TenancyOCID()
	if err != nil {
		return fmt.Errorf("failed to get tenancy OCID: %w", err)
	}

	request := identity.ListAvailabilityDomainsRequest{
		CompartmentId: &tenancyID,
	}

	response, err := client.ListAvailabilityDomains(context.Background(), request)
	if err != nil {
		return fmt.Errorf("failed to list availability domains: %w", err)
	}

	log.Printf("list of available domains: %v", response.Items)
	return nil
}

// Example_pkcs11ConfigurationProviderExplicitValues demonstrates constructing the provider directly.
func Example_pkcs11ConfigurationProviderExplicitValues() {
	tenancyID := os.Getenv("OCI_TENANCY_OCID")
	userID := os.Getenv("OCI_USER_OCID")
	region := os.Getenv("OCI_REGION")
	passPhrase := os.Getenv("OCI_PKCS11_PASSPHRASE")

	provider, err := pkcs11auth.NewPKCS11ConfigurationProvider(
		tenancyID,
		userID,
		region,
		passPhrase,
	)
	if err != nil {
		log.Printf("failed to create PKCS#11 provider: %v", err)
		return
	}
	defer provider.Close()

	if err := callListAvailabilityDomains(provider); err != nil {
		log.Print(err)
		return
	}
	fmt.Println("pkcs11 explicit values completed")

	// Output:
	// pkcs11 explicit values completed
}

// Example_pkcs11ConfigurationProviderDefaultConfig demonstrates loading tenancy, user, and region
// from the default OCI config file while still using the standalone PKCS#11 provider.
func Example_pkcs11ConfigurationProviderDefaultConfig() {
	provider, err := pkcs11auth.NewPKCS11ConfigurationProviderFromDefaultConfig(os.Getenv("OCI_PKCS11_PASSPHRASE"))
	if err != nil {
		log.Printf("failed to create PKCS#11 provider: %v", err)
		return
	}
	defer provider.Close()

	if err := callListAvailabilityDomains(provider); err != nil {
		log.Print(err)
		return
	}
	fmt.Println("pkcs11 default config completed")

	// Output:
	// pkcs11 default config completed
}

// Example_pkcs11ConfigurationProviderCustomProfile demonstrates loading tenancy, user, and region
// from a custom OCI config file profile supplied through common.CustomProfileConfigProvider.
func Example_pkcs11ConfigurationProviderCustomProfile() {
	configPath := os.Getenv("OCI_CONFIG_FILE")
	profileName := "pkcs11"
	passPhrase := os.Getenv("OCI_PKCS11_PASSPHRASE")
	configProvider := common.CustomProfileConfigProvider(configPath, profileName)

	provider, err := pkcs11auth.NewPKCS11ConfigurationProviderFromConfigurationProvider(
		configProvider,
		passPhrase,
	)
	if err != nil {
		log.Printf("failed to create PKCS#11 provider: %v", err)
		return
	}
	defer provider.Close()

	if err := callListAvailabilityDomains(provider); err != nil {
		log.Print(err)
		return
	}
	fmt.Println("pkcs11 custom profile completed")

	// Output:
	// pkcs11 custom profile completed
}

// Example_pkcs11ConfigurationProviderCustomProfileWithPKCS11Config demonstrates loading tenancy,
// user, and region from a custom OCI config file profile while supplying additional PKCS#11-specific settings.
func Example_pkcs11ConfigurationProviderCustomProfileWithPKCS11Config() {
	objectID := uint32(4) // CARD AUTH

	pkcs11Config := pkcs11auth.PKCS11Config{
		// All field values are optional. If not provided, defaults will be used.
		ModulePath:  os.Getenv("OCI_PKCS11_MODULE"),
		TokenLabel:  os.Getenv("OCI_PKCS11_TOKEN_LABEL"), // Either set TokenLabel or TokenSerial, not both. Defaults to first token seen
		TokenSerial: os.Getenv("OCI_PKCS11_TOKEN_SERIAL"),
		ObjectID:    &objectID, // Either set ObjectID or ObjectLabel, not both. 0 maps to PIV and 4 maps to CARD AUTH
	}

	configPath := os.Getenv("OCI_CONFIG_FILE")
	profileName := "pkcs11"
	passPhrase := os.Getenv("OCI_PKCS11_PASSPHRASE")
	configProvider := common.CustomProfileConfigProvider(configPath, profileName)

	provider, err := pkcs11auth.NewPKCS11ConfigurationProviderFromConfigurationProviderWithConfig(
		configProvider,
		passPhrase,
		&pkcs11Config,
	)
	if err != nil {
		log.Printf("failed to create PKCS#11 provider: %v", err)
		return
	}
	defer provider.Close()

	if err := callListAvailabilityDomains(provider); err != nil {
		log.Print(err)
		return
	}
	fmt.Println("pkcs11 custom profile with config completed")

	// Output:
	// pkcs11 custom profile with config completed
}
