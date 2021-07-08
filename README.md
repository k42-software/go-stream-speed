# GO Stream Speed

This is a quick and dirty tool for testing stream performance in GO.
Specifically I'm looking to compare available network transports.

## Usage

You will need a file which will be sent over the stream. I tested with a 2 GB
mp4 video file, but any file will do.

If you don't have a big file handy, you can create one using

```shell
# create bigfile.data
go run ./cmd/bigfile
```

### Testing over the network

This uses the ports 7951 (KCP), 7961 (TCP), 7971 (QUIC), 7991 (TLS)
on 127.0.0.1. These are all hard coded.

```shell
# Increase network receive buffer sizes for best performance.
sudo sysctl -w net.core.rmem_max=21299200
sudo sysctl -w net.core.rmem_default=2500000

# execute in first terminal
go run ./cmd/receive

# simultaneously in second terminal
go run ./cmd/send -mode tcp -file <PathToYourBigFile>
```

### Testing over in memory ring buffer

This doesn't use the network - which means that any performance issues are
related to the protocol implementation and not the network interface.

```shell
go run ./cmd/send -mode mockquic -file <PathToYourBigFile>
```

### Available Sending modes

Available sending modes are

Name       | Description
-----------|-------------------------------------------------
tcp        | TCP
tls        | TLS over TCP
quic       | QUIC over UDP
kcp        | KCP over UDP
mocktls    | TLS over ring buffer (no network)
mockquic   | QUIC over ring buffer (no network)
mockkcp    | KCP over ring buffer (no network)
mockstream | raw stream over ring buffer (no network)

### Libraries

 - QUIC https://github.com/lucas-clemente/quic-go
 - KCP https://github.com/xtaci/kcp-go
