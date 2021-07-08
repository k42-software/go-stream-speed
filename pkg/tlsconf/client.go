package tlsconf

import (
	"crypto/tls"
)

func ClientConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
	}
}
