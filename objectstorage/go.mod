module github.com/nrdcg/oci-go-sdk/objectstorage/v1065

go 1.25.0

replace github.com/nrdcg/oci-go-sdk/common/v1065 => ../common

replace github.com/nrdcg/oci-go-sdk/helpers/v1065 => ../helpers

require (
	github.com/nrdcg/oci-go-sdk/common/v1065 v1065.124.1
	github.com/nrdcg/oci-go-sdk/helpers/v1065 v1065.124.1
	github.com/stretchr/testify v1.12.1
)

require (
	github.com/gofrs/flock v0.13.1 // indirect
	github.com/sony/gobreaker/v2 v2.4.0 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.52.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
)
