package s3util

import (
	"net/url"
	"strings"

	"github.com/minio/minio-go/v6"
)

// s3s?://<accessKey>:<secretKey>@<endpoint>?region=xx
func NewS3(uri string) (*minio.Client, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	accessKey := u.User.Username()
	secretKey, _ := u.User.Password()

	region := u.Query().Get("region")
	endpoint := u.Host
	ssl := strings.HasSuffix(u.Scheme, "s")

	if region != "" {
		return minio.NewWithRegion(endpoint, accessKey, secretKey, ssl, region)
	}

	return minio.New(endpoint, accessKey, secretKey, ssl)
}
