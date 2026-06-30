// Copyright (c) 2016, 2018, 2026, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.

package pkcs11auth

import (
	"crypto"
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"runtime"
	"strings"

	crypto11 "github.com/ThalesGroup/crypto11"
	"github.com/nrdcg/oci-go-sdk/common/v1065"
)

const (
	// ObjectLabelPIV selects the PIV Authentication object.
	ObjectLabelPIV = "PIV"
	// ObjectLabelCardAuth selects the Card Authentication object.
	ObjectLabelCardAuth = "CARD AUTH"
)

// pkcs11Context represents an initialized PKCS#11 session/context that the provider can use to talk to an external authenticator.
type pkcs11Context interface {
	FindCertificate(id []byte, label []byte, serial *big.Int) (*x509.Certificate, error)
	FindKeyPair(id []byte, label []byte) (crypto11.Signer, error)
	Close() error
}

var (
	currentGOOS = runtime.GOOS
)

// objectDefinition is an internal mapping from a user-facing object choice to the concrete PKCS#11 object.
type objectDefinition struct {
	objectID    []byte
	objectLabel string
}

// PKCS11Config contains PKCS#11-specific configuration.
type PKCS11Config struct {
	// ObjectLabel selects a well-known PKCS#11 object such as "PIV" or "CARD AUTH", or an arbitrary PKCS#11 object label.
	ObjectLabel string
	// ObjectID selects a PKCS#11 object numerically.
	ObjectID *uint32
	// KeyID optionally overrides the default tenancy/user/fingerprint key ID.
	KeyID string
	// ModulePath optionally selects the PKCS#11 shared library to use.
	ModulePath string
	// TokenLabel optionally selects a token by PKCS#11 token label.
	TokenLabel string
	// TokenSerial optionally selects a token by PKCS#11 token serial.
	TokenSerial string
}

// PKCS11ConfigurationProvider provides user principal authentication backed by a PKCS#11 token.
type PKCS11ConfigurationProvider struct {
	context     pkcs11Context
	keySigner   crypto.Signer
	tenancyID   string
	userID      string
	region      string
	fingerprint string
	keyID       string
}

// promptingSigner is a wrapper around crypto11.Signer that adds a prompt to touch the external authenticator
type promptingSigner struct {
	signer      crypto11.Signer
	touchPrompt func()
}

func (s *promptingSigner) Public() crypto.PublicKey {
	return s.signer.Public()
}

func (s *promptingSigner) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	s.touchPrompt()
	return s.signer.Sign(rand, digest, opts)
}

func (s *promptingSigner) Delete() error {
	return s.signer.Delete()
}

// NewPKCS11ConfigurationProvider creates a PKCS#11-backed configuration provider using
// explicit tenancy, user, region, and passphrase values.
func NewPKCS11ConfigurationProvider(tenancyID, userID, region, passphrase string) (*PKCS11ConfigurationProvider, error) {
	return NewPKCS11ConfigurationProviderWithConfig(tenancyID, userID, region, passphrase, nil)
}

// NewPKCS11ConfigurationProviderWithConfig creates a PKCS#11-backed configuration provider using
// explicit tenancy, user, region, passphrase, and PKCS#11-specific settings. Uses default values for any PKCS#11 settings not provided.
func NewPKCS11ConfigurationProviderWithConfig(tenancyID, userID, region, passphrase string, pkcs11Config *PKCS11Config) (*PKCS11ConfigurationProvider, error) {
	config := normalizeConfig(pkcs11Config)

	modulePath, err := resolveModulePath(config.ModulePath)
	if err != nil {
		return nil, err
	}

	tokenSelector, err := resolveTokenSelector(config)
	if err != nil {
		return nil, err
	}

	passphrase, err = resolvePassphrase(passphrase)
	if err != nil {
		return nil, err
	}

	// Create an instance of crypto11.Context based on the provided configuration.
	context, err := crypto11.Configure(&crypto11.Config{
		Path:        modulePath,
		TokenLabel:  tokenSelector.tokenLabel,
		TokenSerial: tokenSelector.tokenSerial,
		SlotNumber:  tokenSelector.tokenNumber,
		Pin:         passphrase,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to configure PKCS#11 context: %w", err)
	}

	provider, err := newPKCS11ConfigurationProviderFromContext(context, tenancyID, userID, region, config)
	if err != nil {
		_ = context.Close()
		return nil, err
	}
	return provider, nil
}

// NewPKCS11ConfigurationProviderFromDefaultConfig creates a PKCS#11-backed configuration provider
// using the DEFAULT profile from the default OCI config file path.
func NewPKCS11ConfigurationProviderFromDefaultConfig(passphrase string) (*PKCS11ConfigurationProvider, error) {
	return NewPKCS11ConfigurationProviderFromConfigurationProvider(common.DefaultConfigProvider(), passphrase)
}

// NewPKCS11ConfigurationProviderFromConfigurationProvider creates a PKCS#11-backed configuration provider
// using tenancy, user, and region values sourced from another configuration provider.
func NewPKCS11ConfigurationProviderFromConfigurationProvider(configProvider common.ConfigurationProvider, passphrase string) (*PKCS11ConfigurationProvider, error) {
	return NewPKCS11ConfigurationProviderFromConfigurationProviderWithConfig(configProvider, passphrase, nil)
}

// NewPKCS11ConfigurationProviderFromConfigurationProviderWithConfig creates a PKCS#11-backed configuration provider
// using tenancy, user, and region values sourced from another configuration provider and additional
// PKCS#11-specific settings.
func NewPKCS11ConfigurationProviderFromConfigurationProviderWithConfig(configProvider common.ConfigurationProvider, passphrase string, pkcs11Config *PKCS11Config) (*PKCS11ConfigurationProvider, error) {
	tenancyID, err := configProvider.TenancyOCID()
	if err != nil {
		return nil, fmt.Errorf("failed to get tenancy OCID from configuration provider: %w", err)
	}
	userID, err := configProvider.UserOCID()
	if err != nil {
		return nil, fmt.Errorf("failed to get user OCID from configuration provider: %w", err)
	}
	region, err := configProvider.Region()
	if err != nil {
		return nil, fmt.Errorf("failed to get region from configuration provider: %w", err)
	}
	return NewPKCS11ConfigurationProviderWithConfig(tenancyID, userID, region, passphrase, pkcs11Config)
}

// NewPKCS11ConfigurationProviderFromFile creates a PKCS#11-backed configuration provider
// using tenancy, user, and region values from the DEFAULT profile of an OCI config file.
func NewPKCS11ConfigurationProviderFromFile(configFilePath, passphrase string) (*PKCS11ConfigurationProvider, error) {
	return NewPKCS11ConfigurationProviderFromFileWithProfile(configFilePath, "DEFAULT", passphrase)
}

// NewPKCS11ConfigurationProviderFromFileWithProfile creates a PKCS#11-backed configuration provider
// using tenancy, user, and region values from the selected profile of an OCI config file.
func NewPKCS11ConfigurationProviderFromFileWithProfile(configFilePath, profile, passphrase string) (*PKCS11ConfigurationProvider, error) {
	return NewPKCS11ConfigurationProviderFromFileWithProfileAndConfig(configFilePath, profile, passphrase, nil)
}

// NewPKCS11ConfigurationProviderFromFileWithProfileAndConfig creates a PKCS#11-backed configuration provider
// using tenancy, user, and region values from the selected profile of an OCI config file and additional
// PKCS#11-specific settings.
func NewPKCS11ConfigurationProviderFromFileWithProfileAndConfig(configFilePath, profile, passphrase string, pkcs11Config *PKCS11Config) (*PKCS11ConfigurationProvider, error) {
	configProvider, err := common.ConfigurationProviderFromFileWithProfile(configFilePath, profile, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create configuration provider from file: %w", err)
	}
	return NewPKCS11ConfigurationProviderFromConfigurationProviderWithConfig(configProvider, passphrase, pkcs11Config)
}

func newPKCS11ConfigurationProviderFromContext(context pkcs11Context, tenancyID, userID, region string, config PKCS11Config) (*PKCS11ConfigurationProvider, error) {
	if tenancyID == "" {
		return nil, fmt.Errorf("tenancy OCID can not be empty")
	}
	if userID == "" {
		return nil, fmt.Errorf("user OCID can not be empty")
	}
	if region == "" {
		return nil, fmt.Errorf("region cannot be empty")
	}

	object, err := resolveObjectDefinition(config.ObjectLabel, config.ObjectID)
	if err != nil {
		return nil, err
	}

	objectID := append([]byte(nil), object.objectID...)
	var objectLabel []byte
	if object.objectLabel != "" {
		objectLabel = []byte(object.objectLabel)
	}

	keySigner, err := context.FindKeyPair(objectID, objectLabel)
	if err != nil {
		return nil, fmt.Errorf("failed to find PKCS#11 private key: %w", err)
	}

	cert, err := context.FindCertificate(objectID, objectLabel, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to find PKCS#11 certificate: %w", err)
	}
	if cert == nil {
		return nil, fmt.Errorf("no certificate found on PKCS#11 token")
	}

	if _, ok := cert.PublicKey.(*rsa.PublicKey); !ok {
		return nil, fmt.Errorf("PKCS#11 token contains unsupported public key type %T", cert.PublicKey)
	}

	// Add a touch prompt for better user experience if the certificate's common name contains "touch"
	if certificateAppearsToRequireTouch(cert) {
		keySigner = &promptingSigner{
			signer:      keySigner,
			touchPrompt: logPKCS11TouchPrompt,
		}
	}

	fingerprint, err := publicKeyFingerprint(cert.PublicKey)
	if err != nil {
		return nil, err
	}

	keyID := config.KeyID
	if keyID == "" {
		keyID = fmt.Sprintf("%s/%s/%s", tenancyID, userID, fingerprint)
	}

	return &PKCS11ConfigurationProvider{
		context:     context,
		keySigner:   keySigner,
		tenancyID:   tenancyID,
		userID:      userID,
		region:      region,
		fingerprint: fingerprint,
		keyID:       keyID,
	}, nil
}

type tokenSelector struct {
	tokenLabel  string
	tokenSerial string
	tokenNumber *int
}

func resolveTokenSelector(config PKCS11Config) (tokenSelector, error) {
	label := strings.TrimSpace(config.TokenLabel)
	serial := strings.TrimSpace(config.TokenSerial)

	if label != "" && serial != "" {
		return tokenSelector{}, fmt.Errorf("token label and token serial can not be provided simultaneously")
	}

	if label != "" {
		return tokenSelector{tokenLabel: label}, nil
	}
	if serial != "" {
		return tokenSelector{tokenSerial: serial}, nil
	}

	defaultTokenNumber := 0
	return tokenSelector{tokenNumber: &defaultTokenNumber}, nil
}

func resolvePassphrase(passphrase string) (string, error) {
	passphrase = strings.TrimSpace(passphrase)
	if passphrase == "" {
		return "", fmt.Errorf("PKCS#11 passphrase can not be empty")
	}
	return passphrase, nil
}

func normalizeConfig(config *PKCS11Config) PKCS11Config {
	if config == nil {
		return PKCS11Config{}
	}
	return *config
}

func logPKCS11TouchPrompt() {
	log.Println("Touch the external authenticator to authorize signing")
}

func resolveModulePath(modulePath string) (string, error) {
	if modulePath != "" {
		info, err := os.Stat(modulePath)
		if err != nil || info.IsDir() {
			return "", fmt.Errorf("PKCS#11 module path %q does not exist", modulePath)
		}
		return modulePath, nil
	}

	for _, candidate := range defaultModulePaths() {
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("unable to locate a supported PKCS#11 module for %s", currentGOOS)
}

func defaultModulePaths() []string {
	switch currentGOOS {
	case "darwin":
		return []string{
			"/opt/homebrew/lib/libykcs11.dylib",
			"/usr/local/lib/libykcs11.dylib",
			"/Library/OpenSC/lib/opensc-pkcs11.so",
			"/Library/OpenSC/lib/pkcs11/opensc-pkcs11.so",
			"/usr/local/lib/opensc-pkcs11.so",
		}
	case "linux":
		return []string{
			"/usr/lib64/libykcs11.so",
			"/usr/lib/libykcs11.so",
			"/usr/lib64/pkcs11/opensc-pkcs11.so",
			"/usr/lib64/opensc-pkcs11.so",
			"/usr/lib/pkcs11/opensc-pkcs11.so",
			"/usr/lib/opensc-pkcs11.so",
		}
	case "windows":
		return []string{
			`C:\Program Files\Yubico\Yubico PIV Tool\bin\libykcs11.dll`,
			`C:\Windows\System32\opensc-pkcs11.dll`,
		}
	default:
		return nil
	}
}

// resolveObjectDefinition maps a user-friendly object label or object ID to a PKCS#11 object identifier.
func resolveObjectDefinition(objectLabel string, objectID *uint32) (objectDefinition, error) {
	if objectLabel != "" && objectID != nil {
		return objectDefinition{}, fmt.Errorf("object label and object id can not be provided simultaneously")
	}

	if objectID != nil {
		return objectDefinitionFromID(*objectID)
	}

	switch objectLabel {
	case "", ObjectLabelPIV:
		return objectDefinition{objectID: []byte{0x01}}, nil
	case ObjectLabelCardAuth:
		return objectDefinition{objectID: []byte{0x04}}, nil
	default:
		return objectDefinition{objectLabel: objectLabel}, nil
	}
}

func objectDefinitionFromID(objectID uint32) (objectDefinition, error) {
	if objectID > 0xff {
		return objectDefinition{}, fmt.Errorf("unsupported PKCS#11 object id %d: object IDs must fit in one byte", objectID)
	}

	switch objectID {
	case 0:
		return resolveObjectDefinition(ObjectLabelPIV, nil)
	case 1:
		return resolveObjectDefinition(ObjectLabelPIV, nil)
	case 4:
		return resolveObjectDefinition(ObjectLabelCardAuth, nil)
	default:
		return objectDefinition{objectID: []byte{byte(objectID)}}, nil
	}
}

func certificateAppearsToRequireTouch(cert *x509.Certificate) bool {
	if cert == nil {
		return false
	}
	return strings.Contains(strings.ToLower(cert.Subject.CommonName), "touch")
}

func publicKeyFingerprint(publicKey crypto.PublicKey) (string, error) {
	der, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal PKCS#11 public key: %w", err)
	}

	sum := md5.Sum(der)
	hexStr := hex.EncodeToString(sum[:])

	var sb strings.Builder
	for i := 0; i < len(hexStr); i += 2 {
		if i > 0 {
			sb.WriteByte(':')
		}
		sb.WriteString(hexStr[i : i+2])
	}
	return sb.String(), nil
}

func (p *PKCS11ConfigurationProvider) PrivateRSAKey() (*rsa.PrivateKey, error) {
	return nil, fmt.Errorf("PKCS#11-backed keys are non-exportable")
}

func (p *PKCS11ConfigurationProvider) PrivateKeySigner() (crypto.Signer, error) {
	return p.keySigner, nil
}

func (p *PKCS11ConfigurationProvider) KeyID() (string, error) {
	return p.keyID, nil
}

func (p *PKCS11ConfigurationProvider) TenancyOCID() (string, error) {
	return p.tenancyID, nil
}

func (p *PKCS11ConfigurationProvider) UserOCID() (string, error) {
	return p.userID, nil
}

func (p *PKCS11ConfigurationProvider) KeyFingerprint() (string, error) {
	return p.fingerprint, nil
}

func (p *PKCS11ConfigurationProvider) Region() (string, error) {
	return p.region, nil
}

func (p *PKCS11ConfigurationProvider) AuthType() (common.AuthConfig, error) {
	return common.AuthConfig{common.PKCS11Authentication, false, nil},
		nil
}

// Close releases the PKCS#11 context.
func (p *PKCS11ConfigurationProvider) Close() error {
	if p.context == nil {
		return nil
	}
	return p.context.Close()
}
