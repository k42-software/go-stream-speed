package quicconf

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/k42-software/go-stream-speed/pkg/buffered"
	"github.com/k42-software/go-stream-speed/pkg/tlsconf"
	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/logging"
	"github.com/lucas-clemente/quic-go/qlog"
	"io"
	"log"
	"os"
)

func ClientTlsConfig() (tlsConf *tls.Config) {
	tlsConf = tlsconf.ClientConfig()
	tlsConf.NextProtos = []string{"quic"}
	return tlsConf
}

func ClientConfig(withTracer bool) (quicConf *quic.Config) {

	const KB = 1 << 10
	const MB = 1 << 20

	// This is the value that Chromium is using
	const QuicConnectionFlowControlMultiplier = 1.5

	const QuicInitialStreamReceiveWindow = 512 * KB // default: 512 kb
	const QuicMaxStreamReceiveWindow = 32 * MB      // default: 6 MB

	const QuicInitialConnectionReceiveWindow = QuicInitialStreamReceiveWindow * QuicConnectionFlowControlMultiplier // default: 512 kb * 1.5
	const QuicMaxConnectionReceiveWindow = QuicMaxStreamReceiveWindow * QuicConnectionFlowControlMultiplier         // default: 15 MB

	quicConf = &quic.Config{
		InitialStreamReceiveWindow:     QuicInitialStreamReceiveWindow,
		MaxStreamReceiveWindow:         QuicMaxStreamReceiveWindow,
		InitialConnectionReceiveWindow: QuicInitialConnectionReceiveWindow,
		MaxConnectionReceiveWindow:     QuicMaxConnectionReceiveWindow,
		MaxIncomingStreams:             10, // default: 100
		MaxIncomingUniStreams:          10, // default: 100
		StatelessResetKey:              nil,
		KeepAlive:                      false, // default: false
		DisablePathMTUDiscovery:        false, // default: false
		EnableDatagrams:                false, // default: false
	}

	if withTracer {
		quicConf.Tracer = qlog.NewTracer(func(_ logging.Perspective, connID []byte) io.WriteCloser {
			filename := fmt.Sprintf("client_%x.qlog", connID)
			f, err := os.Create(filename)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Creating qlog file %s.\n", filename)
			return buffered.NewWriteCloser(bufio.NewWriter(f), f)
		})
	}

	return quicConf
}
