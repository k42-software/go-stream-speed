package main

import (
	"os"
	"testing"
)

func TestSendMain(t *testing.T) {
	_ = os.Chdir("../../")
	main()
}
