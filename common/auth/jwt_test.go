// Copyright (c) 2016, 2018, 2026, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.

package auth

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

const syntheticJwtSignature = "synthetic-signature"

func TestJwtToken_ParseJwt(t *testing.T) {
	expectedHeader := testJwtHeader()
	expectedPayload := testJwtPayload()
	token, err := parseJwt(testJwtString(t, encodeJwtPart(t, expectedHeader), encodeJwtPart(t, expectedPayload)))

	assert.NoError(t, err)
	assert.Equal(t, expectedHeader, token.header)
	assert.Equal(t, expectedPayload, token.payload)
	assert.Equal(t, true, token.expired())
}

func testJwtHeader() map[string]interface{} {
	return map[string]interface{}{
		"alg": "RS256",
		"kid": "asw",
		"typ": "JWT",
	}
}

func testJwtPayload() map[string]interface{} {
	return map[string]interface{}{
		"aud":             "opc.oracle.com",
		"exp":             float64(1511838793),
		"iat":             float64(1511817193),
		"iss":             "authService.oracle.com",
		"opc-certtype":    "instance",
		"opc-compartment": "ocid1.compartment.oc1..bluhbluhbluh",
		"opc-instance":    "ocid1.instance.oc1.phx.bluhbluhbluh",
		"opc-tenant":      "ocidv1:tenancy:oc1:phx:1234567890:bluhbluhbluh",
		"ptype":           "instance",
		"sub":             "ocid1.instance.oc1.phx.bluhbluhbluh",
		"tenant":          "ocidv1:tenancy:oc1:phx:1234567890:bluhbluhbluh",
		"ttype":           "x509",
	}
}

func encodeJwtPart(t *testing.T, value map[string]interface{}) string {
	t.Helper()
	bytes, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("failed to marshal JWT test value: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

func testJwtString(t *testing.T, headerPart, payloadPart string) string {
	t.Helper()
	return headerPart + "." + payloadPart + "." + syntheticJwtSignature
}

func TestJwtToken_ParseJwtInvalidNumberOfParts(t *testing.T) {
	token, err := parseJwt(encodeJwtPart(t, testJwtPayload()) + "." + syntheticJwtSignature)

	assert.Nil(t, token)
	assert.EqualError(t, err, "the given token string contains an invalid number of parts")
}

func TestJwtToken_ParseJwtInvalidHeader(t *testing.T) {
	token, err := parseJwt(testJwtString(t, "INVALIDHEADER", encodeJwtPart(t, testJwtPayload())))

	assert.Nil(t, token)
	assert.Error(t, err)
}

func TestJwtToken_ParseJwtInvalidPayload(t *testing.T) {
	token, err := parseJwt(testJwtString(t, encodeJwtPart(t, testJwtHeader()), "INVALIDPAYLOAD"))

	assert.Nil(t, token)
	assert.Error(t, err)
}
