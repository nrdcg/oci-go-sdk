// Copyright (c) 2016, 2018, 2026, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.

package pkcs11auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"io"
	"math/big"
	"testing"
	"time"

	crypto11 "github.com/ThalesGroup/crypto11"
	"github.com/nrdcg/oci-go-sdk/common/v1065"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePKCS11Context struct {
	findKeyPairID    []byte
	findKeyPairLabel []byte
	findCertID       []byte
	findCertLabel    []byte
	signer           crypto11.Signer
	certificate      *x509.Certificate
	closeCalls       int
}

type testSigner struct {
	crypto.Signer
	signCalls int
}

type fakeCrypto11Signer struct {
	crypto.Signer
}

type fakeSourceConfigurationProvider struct {
	tenancyID string
	userID    string
	region    string
}

func (f *fakePKCS11Context) FindCertificate(id []byte, label []byte, serial *big.Int) (*x509.Certificate, error) {
	f.findCertID = append([]byte(nil), id...)
	f.findCertLabel = append([]byte(nil), label...)
	return f.certificate, nil
}

func (f *fakePKCS11Context) FindKeyPair(id []byte, label []byte) (crypto11.Signer, error) {
	f.findKeyPairID = append([]byte(nil), id...)
	f.findKeyPairLabel = append([]byte(nil), label...)
	return f.signer, nil
}

func (f *fakePKCS11Context) Close() error {
	f.closeCalls++
	return nil
}

func (s *testSigner) Public() crypto.PublicKey {
	return s.Signer.Public()
}

func (s *testSigner) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	s.signCalls++
	return s.Signer.Sign(rand, digest, opts)
}

func (s *fakeCrypto11Signer) Delete() error {
	return nil
}

func (f fakeSourceConfigurationProvider) PrivateRSAKey() (*rsa.PrivateKey, error) {
	return nil, assert.AnError
}

func (f fakeSourceConfigurationProvider) KeyID() (string, error) {
	return "", assert.AnError
}

func (f fakeSourceConfigurationProvider) TenancyOCID() (string, error) {
	return f.tenancyID, nil
}

func (f fakeSourceConfigurationProvider) UserOCID() (string, error) {
	return f.userID, nil
}

func (f fakeSourceConfigurationProvider) KeyFingerprint() (string, error) {
	return "", assert.AnError
}

func (f fakeSourceConfigurationProvider) Region() (string, error) {
	return f.region, nil
}

func (f fakeSourceConfigurationProvider) AuthType() (common.AuthConfig, error) {
	return common.AuthConfig{AuthType: common.UnknownAuthenticationType}, nil
}

func TestResolveObjectDefinition_Default(t *testing.T) {
	object, err := resolveObjectDefinition("", nil)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x01}, object.objectID)
}

func TestResolveObjectDefinition_Label(t *testing.T) {
	object, err := resolveObjectDefinition(ObjectLabelCardAuth, nil)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x04}, object.objectID)
}

func TestResolveObjectDefinition_ArbitraryLabel(t *testing.T) {
	object, err := resolveObjectDefinition("SIGNING KEY", nil)
	require.NoError(t, err)
	assert.Nil(t, object.objectID)
	assert.Equal(t, "SIGNING KEY", object.objectLabel)
}

func TestResolveObjectDefinition_ID(t *testing.T) {
	objectID := uint32(4)
	object, err := resolveObjectDefinition("", &objectID)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x04}, object.objectID)
}

func TestResolveObjectDefinition_IDZeroMapsToPIV(t *testing.T) {
	objectID := uint32(0)
	object, err := resolveObjectDefinition("", &objectID)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x01}, object.objectID)
}

func TestResolveObjectDefinition_ArbitraryID(t *testing.T) {
	objectID := uint32(25)
	object, err := resolveObjectDefinition("", &objectID)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x19}, object.objectID)
}

func TestResolveTokenSelector_RejectsLabelAndSerial(t *testing.T) {
	_, err := resolveTokenSelector(PKCS11Config{
		TokenLabel:  "label",
		TokenSerial: "serial",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "can not be provided simultaneously")
}

func TestNewPKCS11ConfigurationProviderFromContext(t *testing.T) {
	privateKey, cert := generateTestKeyAndCertificate(t, "PKCS11 Test")
	context := &fakePKCS11Context{
		signer:      &fakeCrypto11Signer{Signer: privateKey},
		certificate: cert,
	}

	provider, err := newPKCS11ConfigurationProviderFromContext(
		context,
		"tenancy",
		"user",
		"us-phoenix-1",
		PKCS11Config{},
	)
	require.NoError(t, err)

	keyID, err := provider.KeyID()
	require.NoError(t, err)
	assert.Equal(t, "tenancy/user/"+provider.fingerprint, keyID)

	signer, err := provider.PrivateKeySigner()
	require.NoError(t, err)
	assert.Equal(t, privateKey.Public(), signer.Public())

	fingerprint, err := provider.KeyFingerprint()
	require.NoError(t, err)
	assert.Equal(t, provider.fingerprint, fingerprint)

	assert.Equal(t, []byte{0x01}, context.findKeyPairID)
	assert.Nil(t, context.findKeyPairLabel)
	assert.Equal(t, []byte{0x01}, context.findCertID)
	assert.Nil(t, context.findCertLabel)
}

func TestNewPKCS11ConfigurationProviderFromContext_UsesCardAuthObject(t *testing.T) {
	privateKey, cert := generateTestKeyAndCertificate(t, "PKCS11 Test Card Auth")
	context := &fakePKCS11Context{
		signer:      &fakeCrypto11Signer{Signer: privateKey},
		certificate: cert,
	}
	objectID := uint32(4)

	_, err := newPKCS11ConfigurationProviderFromContext(
		context,
		"tenancy",
		"user",
		"us-phoenix-1",
		PKCS11Config{ObjectID: &objectID},
	)
	require.NoError(t, err)

	assert.Equal(t, []byte{0x04}, context.findKeyPairID)
	assert.Nil(t, context.findKeyPairLabel)
	assert.Equal(t, []byte{0x04}, context.findCertID)
	assert.Nil(t, context.findCertLabel)
}

func TestNewPKCS11ConfigurationProviderFromContext_UsesArbitraryObjectID(t *testing.T) {
	privateKey, cert := generateTestKeyAndCertificate(t, "PKCS11 Generic Object")
	context := &fakePKCS11Context{
		signer:      &fakeCrypto11Signer{Signer: privateKey},
		certificate: cert,
	}
	objectID := uint32(25)

	_, err := newPKCS11ConfigurationProviderFromContext(
		context,
		"tenancy",
		"user",
		"us-phoenix-1",
		PKCS11Config{ObjectID: &objectID},
	)
	require.NoError(t, err)

	assert.Equal(t, []byte{0x19}, context.findKeyPairID)
	assert.Nil(t, context.findKeyPairLabel)
	assert.Equal(t, []byte{0x19}, context.findCertID)
	assert.Nil(t, context.findCertLabel)
}

func TestNewPKCS11ConfigurationProviderFromContext_UsesArbitraryLabel(t *testing.T) {
	privateKey, cert := generateTestKeyAndCertificate(t, "PKCS11 Label")
	context := &fakePKCS11Context{
		signer:      &fakeCrypto11Signer{Signer: privateKey},
		certificate: cert,
	}

	_, err := newPKCS11ConfigurationProviderFromContext(
		context,
		"tenancy",
		"user",
		"us-phoenix-1",
		PKCS11Config{ObjectLabel: "SIGNING KEY"},
	)
	require.NoError(t, err)

	assert.Nil(t, context.findKeyPairID)
	assert.Equal(t, []byte("SIGNING KEY"), context.findKeyPairLabel)
	assert.Nil(t, context.findCertID)
	assert.Equal(t, []byte("SIGNING KEY"), context.findCertLabel)
}

func TestNewPKCS11ConfigurationProviderFromContext_Close(t *testing.T) {
	privateKey, cert := generateTestKeyAndCertificate(t, "PKCS11 Close")
	context := &fakePKCS11Context{
		signer:      &fakeCrypto11Signer{Signer: privateKey},
		certificate: cert,
	}

	provider, err := newPKCS11ConfigurationProviderFromContext(
		context,
		"tenancy",
		"user",
		"us-phoenix-1",
		PKCS11Config{},
	)
	require.NoError(t, err)

	require.NoError(t, provider.Close())
	assert.Equal(t, 1, context.closeCalls)
}

func generateTestKeyAndCertificate(t *testing.T, commonName string) (*rsa.PrivateKey, *x509.Certificate) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore: time.Now().Add(-time.Hour),
		NotAfter:  time.Now().Add(time.Hour),
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	return privateKey, cert
}
