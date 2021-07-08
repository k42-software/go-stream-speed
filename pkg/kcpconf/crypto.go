package kcpconf

import (
	"github.com/xtaci/kcp-go/v5"
	"log"
)

const CryptoKey = "12345678901234567890123456789012"

func NewBlockCrypt() kcp.BlockCrypt {
	kcpCryptoBlock, err := kcp.NewSalsa20BlockCrypt(
		[]byte(CryptoKey),
	)
	if err != nil {
		log.Fatalf("[ERROR] %s", err)
	}
	return kcpCryptoBlock
}
