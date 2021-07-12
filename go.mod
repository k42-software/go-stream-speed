module github.com/k42-software/go-stream-speed

go 1.15

//replace github.com/astrolox/suft/ => /home/brian/astrolox/go/suft/
//replace github.com/astrolox/suft/protocol/ => /home/brian/astrolox/go/suft/protocol/

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/dolmen-go/contextio v0.0.0-20200217195037-68fc5150bcd5
	github.com/dustin/go-humanize v1.0.0
	github.com/jakecoffman/rely v1.0.1
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99
	github.com/k42-software/go-multierror/v2 v2.0.0
	github.com/lucas-clemente/quic-go v0.21.1
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/pkg/profile v1.6.0
	github.com/xtaci/kcp-go/v5 v5.6.1
	gvisor.dev/gvisor v0.0.0-20210625224711-17e7bcb5b604
)
