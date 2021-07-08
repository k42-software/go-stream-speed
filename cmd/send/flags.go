package main

import (
	"context"
	"flag"
	"github.com/pkg/profile"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

type testType string

const (
	testTCP         testType = "tcp"
	testMockTCP              = "mocktcp" // Not implemented here
	testKCP                  = "kcp"
	testMockKCP              = "mockkcp"
	testTLS                  = "tls"
	testMockTLS              = "mocktls"
	testQUIC                 = "quic"
	testMockQUIC             = "mockquic"
	testMockStream           = "mockstream"
	testMockPackets          = "mockpackets"
)

const defaultTestFile = "bigfile.data" // "data/BrianTestVideo3.mp4"
const defaultTestMode = testMockQUIC

// Cli flags
var runProfiler = flag.Bool(
	"profile",
	false,
	"run profiler",
)
var modeStr = flag.String(
	"mode",
	"",
	"which test to perform",
)
var testFilePath = flag.String(
	"file",
	defaultTestFile,
	"path to file to upload",
)

func processCliFlags() (dialer testDialer, shutdown context.CancelFunc) {

	flag.Parse()

	if len(*modeStr) == 0 {
		log.Printf("[ERROR] Testing mode not specified")
		log.Printf("[DEBUG] Defaulting to mode: %s", defaultTestMode)
		*modeStr = defaultTestMode
	}

	switch testType(strings.ToLower(*modeStr)) {
	case testTCP:
		dialer = dialTCP
	case testMockTCP:
		log.Fatalf("[ERROR] Testing Mode not implemented")

	case testTLS:
		dialer = dialTLS
	case testMockTLS:
		dialer = dialMockTLS

	case testQUIC:
		dialer = dialQUIC
	case testMockQUIC:
		dialer = dialMockQUIC

	case testKCP:
		dialer = dialKCP
	case testMockKCP:
		dialer = dialMockKCP

	case testMockStream:
		dialer = dialMockStream

	default:
		log.Printf("[ERROR] Unknown mode: %q", *modeStr)
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *runProfiler {
		var profiler interface{ Stop() }
		timer := time.AfterFunc(
			1*time.Second,
			func() {
				profiler = profile.Start(
					profile.CPUProfile,
					//profile.ProfilePath(fmt.Sprintf("./%d/", time.Now().Unix())),
					profile.ProfilePath("."),
				)
			},
		)
		shutdown = func() {
			timer.Stop()
			if profiler != nil {
				profiler.Stop()
			}
			runtime.Gosched()
		}
	}

	if len(*testFilePath) == 0 {
		log.Fatal("[ERROR] You must specify a test file")
	} else {
		if _, err := os.Stat(*testFilePath); err != nil {
			log.Fatalf("[ERROR] test file: %s: %s", *testFilePath, err)
		}
	}

	return
}
