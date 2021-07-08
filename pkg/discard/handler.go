package discard

import (
	"context"
	"github.com/dolmen-go/contextio"
	"github.com/dustin/go-humanize"
	"github.com/k42-software/go-stream-speed/pkg/counter"
	"io"
	"log"
	"net"
	"time"
)

func StreamConnection(
	parentContext context.Context,
	networkConnection net.Conn,
) {
	started := time.Now()

	log.Printf("[DEBUG] Receive: Handling stream with discard")

	ctx, ctxCancel := context.WithTimeout(parentContext, 15*time.Minute)
	writeCounter := counter.NewCounter("Receive: Discard")
	writer := contextio.NewWriter(ctx, writeCounter)

	n, _ := io.Copy(writer, networkConnection)

	duration := time.Since(started)
	log.Printf(
		"[DEBUG] Receive: Discarded %s in %s (%s/s)",
		humanize.IBytes(uint64(n)),
		duration,
		humanize.IBytes(uint64(float64(n)/duration.Seconds())),
	)

	ctxCancel()
	_ = writeCounter.Close()

	if err := networkConnection.Close(); err != nil {
		log.Printf("[ERROR] Receive: Connection close error: %s", err)
	}

	log.Printf("[DEBUG] Receive: Connection closed")
}
