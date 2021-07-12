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
udptcp     | TCP (user mode) over UDP*
mocktls    | TLS over ring buffer (no network)
mockquic   | QUIC over ring buffer (no network)
mockkcp    | KCP over ring buffer (no network)
mockstream | raw stream over ring buffer (no network)

* *Note that the TCP over UDP mode has not been optimised - Actually its running
  a full user mode network stack (not just TCP), so its also passing IP headers
  and doing a lot of extra work that isn't strictly necessary for running just
  one TCP session over a UDP socket. However, that was the easiest way to
  implement it quickly. There's plenty of potential for improving the
  performance of this test.*

### Libraries

- QUIC https://github.com/lucas-clemente/quic-go
- KCP https://github.com/xtaci/kcp-go

### Results (on my laptop)

Results vary with each run, but stay approximately the same. Your mileage may
vary.

Test              | ~ Time            | ~ Speed
------------------|-------------------|---------------
TCP               | ~ 2.6522s         | ~ 772 MiB/s
TLS+TCP           | ~ 3.3561s         | ~ 610 MiB/s
QUIC              | ~ 7.9078s         | ~ 259 MiB/s
KCP               | ~ timeout         | ~ 4 MiB/s
UDPTCP (1500 MTU) | ~ 9.4912s         | ~ 216 MiB/s
UDPTCP (65k MTU)  | ~ 4.9213s         | ~ 416 MiB/s
in memory         | ~ 847ms           | ~ 2.4 GiB/s
TLS in memory     | ~ 1.4193s         | ~ 1.4 GiB/s
QUIC in memory    | ~ 6.1452s         | ~ 333 MiB/s
KCP in memory     | ~ timeout         | ~ 4 MiB/s
