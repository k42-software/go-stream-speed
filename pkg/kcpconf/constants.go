package kcpconf

//
// Honestly, I have no idea how to configure KCP. I keep tweaking these values
// and am not sure what's best, I guess that's why they are configurable. Seems
// to me like the decision on the correct values is very situational.
//
// The values below are the Turbo mode values recommended by;
// See https://github.com/skywind3000/kcp/blob/master/README.en.md#protocol-configuration
//
// Note that KCP seems to be prioritising low latency over high throughput.
// I might be wrong, but that's my impression. This tool measures throughput,
// not latency.
//

const NoDelay = 1
const Interval = 10
const Resend = 2
const NoFlowControl = 1 // aka nc, 0 = ON, 1 = OFF

const WriteDelay = true
const StreamMode = true

// These help on bad networks - but they're just overhead on a clean link.
const DataShards = 0   // suggested: 10
const ParityShards = 0 // suggested: 3
