// Copyright (c) 2016, 2018, 2020, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.

package transfer

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nrdcg/oci-go-sdk/common/v1065"
	"github.com/nrdcg/oci-go-sdk/helpers/v1065"
	"github.com/nrdcg/oci-go-sdk/objectstorage/v1065"
	"github.com/stretchr/testify/assert"
)

type fakeFileUpload struct{}

// split file into multiple parts and uploads them to blob storage, then merge
func (fake fakeFileUpload) UploadFileMultiparts(ctx context.Context, request UploadFileRequest) (response UploadResponse, err error) {
	response = UploadResponse{
		Type: MultipartUpload,
	}

	return
}

// uploads a file to blob storage via PutObject API
func (fake fakeFileUpload) UploadFilePutObject(ctx context.Context, request UploadFileRequest) (response UploadResponse, err error) {
	response = UploadResponse{
		Type: SinglepartUpload,
	}

	return
}

// resume a file upload, use it when UploadFile failed
func (fake fakeFileUpload) ResumeUploadFile(ctx context.Context, uploadID string) (response UploadResponse, err error) {
	return
}

func TestUploadManager_UploadFile(t *testing.T) {
	type testData struct {
		FileSize            int
		PartSize            *int64
		FileUploader        FileUploader
		ExpectedResponsType UploadResponseType
		ExpectedError       error
	}

	testDataSet := []testData{
		{
			FileSize:            100,
			PartSize:            common.Int64(1000),
			FileUploader:        fakeFileUpload{},
			ExpectedResponsType: SinglepartUpload,
		}, {
			FileSize:            60,
			PartSize:            common.Int64(50),
			FileUploader:        fakeFileUpload{},
			ExpectedResponsType: MultipartUpload,
		}, {
			FileSize:      60,
			PartSize:      common.Int64(50),
			FileUploader:  nil,
			ExpectedError: errorInvalidFileUploader,
		},
	}

	for _, testData := range testDataSet {
		uploadManager := UploadManager{FileUploader: testData.FileUploader}
		// small file fits into one part
		filePath, _ := helpers.WriteTempFileOfSize(int64(testData.FileSize))
		req := UploadFileRequest{
			UploadRequest: UploadRequest{
				NamespaceName: common.String("namespace"),
				BucketName:    common.String("bname"),
				ObjectName:    common.String("objectName"),
				PartSize:      testData.PartSize,
			},
			FilePath: filePath,
		}

		resp, err := uploadManager.UploadFile(context.Background(), req)
		assert.Equal(t, err, testData.ExpectedError)
		assert.Equal(t, testData.ExpectedResponsType, resp.Type)
	}

}

func TestUploadManager_ResumeUploadFile(t *testing.T) {
	fileUploader := fakeFileUpload{}
	uploadManager := UploadManager{FileUploader: fileUploader}
	_, err := uploadManager.ResumeUploadFile(context.Background(), "")
	assert.Error(t, err)
}

type fakeReader struct{}

func (fr fakeReader) Read(p []byte) (n int, err error) {
	return
}

type fakeSigner struct{}

func (fakeSigner) Sign(_ *http.Request) error {
	return nil
}

func TestUploadManager_UploadStream(t *testing.T) {

	req := UploadStreamRequest{
		UploadRequest: UploadRequest{
			NamespaceName: common.String("namespace"),
			BucketName:    common.String("bname"),
			ObjectName:    common.String("objectName"),
		},
		StreamReader: fakeReader{},
	}
	uploadManager := UploadManager{StreamUploader: nil}
	_, err := uploadManager.UploadStream(context.Background(), req)
	assert.Equal(t, errorInvalidStreamUploader, err)
}

// TestUploadManager_UploadStreamEmptyReader tests that empty readers upload one empty object for known reader types like `*bytes.Reader` and generic readers like `io.Reader`
func TestUploadManager_UploadStreamEmptyReader(t *testing.T) {
	tests := []struct {
		name   string
		reader io.Reader
	}{
		{
			name:   "known empty reader",
			reader: bytes.NewReader(nil),
		},
		{
			name:   "generic empty reader",
			reader: io.LimitReader(bytes.NewReader(nil), 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var putObjectRequests int32
			var createMultipartRequests int32
			var commitMultipartRequests int32

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodPut && r.URL.Path == "/n/namespace/b/bname/o/objectName":
					atomic.AddInt32(&putObjectRequests, 1)

					body, err := io.ReadAll(r.Body)
					assert.NoError(t, err)
					// Empty streams should be sent as an empty PutObject body
					assert.Empty(t, body)
					assert.Equal(t, int64(0), r.ContentLength)

					w.Header().Set("etag", "empty-etag")
					w.WriteHeader(http.StatusOK)
				case r.Method == http.MethodPost && r.URL.Path == "/n/namespace/b/bname/u" && r.URL.RawQuery == "":
					atomic.AddInt32(&createMultipartRequests, 1)

					w.Header().Set("content-type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"namespace":"namespace","bucket":"bname","object":"objectName","uploadId":"upload-id","timeCreated":"2020-01-01T00:00:00Z"}`))
				case r.Method == http.MethodPost && r.URL.Path == "/n/namespace/b/bname/u/objectName" && r.URL.Query().Get("uploadId") == "upload-id":
					atomic.AddInt32(&commitMultipartRequests, 1)

					w.Header().Set("content-type", "application/json")
					w.Header().Set("opc-request-id", "request-id")
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte(`{"code":"InvalidUploadPart","message":"There are no parts to commit"}`))
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			client := &objectstorage.ObjectStorageClient{
				BaseClient: common.BaseClient{
					HTTPClient: server.Client(),
					Signer:     fakeSigner{},
					UserAgent:  "upload-manager-test",
					Host:       server.URL,
				},
			}

			noRetryPolicy := common.NoRetryPolicy()
			req := UploadStreamRequest{
				UploadRequest: UploadRequest{
					NamespaceName:       common.String("namespace"),
					BucketName:          common.String("bname"),
					ObjectName:          common.String("objectName"),
					PartSize:            common.Int64(10),
					ObjectStorageClient: client,
					RequestMetadata: common.RequestMetadata{
						RetryPolicy: &noRetryPolicy,
					},
				},
				StreamReader: tt.reader,
			}

			resp, err := NewUploadManager().UploadStream(context.Background(), req)

			// UploadStream should succeed for both known and generic empty readers
			assert.NoError(t, err)
			// Empty streams should use PutObject instead of multipart upload
			assert.Equal(t, SinglepartUpload, resp.Type)
			// PutObject should be called exactly once for the empty object
			assert.Equal(t, int32(1), atomic.LoadInt32(&putObjectRequests))
			// Multipart upload should not start for an empty stream
			assert.Equal(t, int32(0), atomic.LoadInt32(&createMultipartRequests))
			// No multipart commit should be attempted with zero parts
			assert.Equal(t, int32(0), atomic.LoadInt32(&commitMultipartRequests))
		})
	}
}

// TestUploadManager_UploadStreamEmptyStreamPreservesSSEHeaders tests that when an empty stream is uploaded, the SSE headers are preserved in the PutObject request and not lost
func TestUploadManager_UploadStreamEmptyStreamPreservesSSEHeaders(t *testing.T) {
	type capturedRequest struct {
		header  http.Header
		path    string
		body    []byte
		readErr error
	}

	capturedRequestCh := make(chan capturedRequest, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		capturedRequestCh <- capturedRequest{
			header:  r.Header.Clone(),
			path:    r.URL.EscapedPath(),
			body:    body,
			readErr: err,
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	objectStorageClient := &objectstorage.ObjectStorageClient{
		BaseClient: common.BaseClient{
			HTTPClient: server.Client(),
			Signer:     fakeSigner{},
			Host:       server.URL,
			UserAgent:  "test-user-agent",
		},
	}

	req := UploadStreamRequest{
		UploadRequest: UploadRequest{
			NamespaceName:           common.String("namespace"),
			BucketName:              common.String("bucket"),
			ObjectName:              common.String("empty-object"),
			ObjectStorageClient:     objectStorageClient,
			OpcSseCustomerAlgorithm: common.String("AES256"),
			OpcSseCustomerKey:       common.String("customer-key"),
			OpcSseCustomerKeySha256: common.String("customer-key-sha256"),
			OpcSseKmsKeyId:          common.String("kms-key-id"),
		},
		StreamReader: strings.NewReader(""),
	}
	uploadManager := NewUploadManager()

	resp, err := uploadManager.UploadStream(context.Background(), req)

	assert.NoError(t, err)
	var captured capturedRequest
	select {
	case captured = <-capturedRequestCh:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for PutObject request")
	}

	assert.NoError(t, captured.readErr)
	assert.Equal(t, SinglepartUpload, resp.Type)
	assert.Equal(t, "/n/namespace/b/bucket/o/empty-object", captured.path)
	assert.Empty(t, captured.body)
	assert.Equal(t, "AES256", captured.header.Get("opc-sse-customer-algorithm"))
	assert.Equal(t, "customer-key", captured.header.Get("opc-sse-customer-key"))
	assert.Equal(t, "customer-key-sha256", captured.header.Get("opc-sse-customer-key-sha256"))
	assert.Equal(t, "kms-key-id", captured.header.Get("opc-sse-kms-key-id"))
}
