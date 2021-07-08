package quicconf

import (
	"crypto/tls"
	"github.com/k42-software/go-stream-speed/pkg/tlsconf"
	"github.com/lucas-clemente/quic-go"
)

func ServerTlsConfig() (tlsConf *tls.Config) {
	tlsConf = tlsconf.ServerConfig()
	tlsConf.NextProtos = []string{"quic"}
	return tlsConf
}

func ServerConfig(withTracer bool) (quicConf *quic.Config) {
	return ClientConfig(withTracer)
}

func DefaultServerConfig(tlsConf *tls.Config, config *quic.Config) (*tls.Config, *quic.Config) {
	if tlsConf == nil {
		tlsConf = ServerTlsConfig()
	}
	if config == nil {
		config = ServerConfig(false)
	}
	return tlsConf, config
}
