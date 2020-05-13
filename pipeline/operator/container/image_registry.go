package container

import (
	"encoding/base64"
	"encoding/json"
	"net/url"

	"github.com/docker/docker/api/types"
)

func ParseImageRegistry(uri string) (*ImageRegistry, error) {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return nil, err
	}

	endpoint := &ImageRegistry{}

	endpoint.Host = u.Host
	endpoint.Prefix = u.Path

	if u.User != nil {
		endpoint.Username = u.User.Username()
		endpoint.Password, _ = u.User.Password()
	}

	return endpoint, nil
}

type ImageRegistry struct {
	Host     string
	Username string
	Password string
	Prefix   string
}

func (s ImageRegistry) Fix(imageRef string) string {
	return s.Host + s.Prefix + imageRef
}

func (s ImageRegistry) RegistryAuth() string {
	authConfig := types.AuthConfig{Username: s.Username, Password: s.Password, ServerAddress: s.Host}
	b, _ := json.Marshal(authConfig)
	return base64.StdEncoding.EncodeToString(b)
}
