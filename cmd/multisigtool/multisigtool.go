// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"flag"
	"fmt"
	"math/big"

	"mainstay/crypto"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
)

var (
	isRegtest bool
)

func init() {
	flag.BoolVar(&isRegtest, "regtest", true, "Do regtest work")
	flag.Parse()
}

func main() {
	if isRegtest {
		doRegtest()
	}
}

func pubFromCoordinates(xStr string, yStr string) *btcec.PublicKey {
	x := new(big.Int)
	y := new(big.Int)
	_, errX := fmt.Sscan(xStr, x)
	if errX != nil {
		fmt.Println("fail-x")
	}
	_, errY := fmt.Sscan(yStr, y)
	if errY != nil {
		fmt.Println("fail-y")
	}

	return (*btcec.PublicKey)(&ecdsa.PublicKey{btcec.S256(), x, y})
}

// Generate multisig script and p2sh address for mainstay
// from a bitcoin public key (pubMain) and an hsm key (pubX/pubY)
func doRegtest() {
	// with cc
	hsmPubX := "52188485757047374213594781872026731452065763678534687787431199968762674330343"
	hsmPubY := "69859013696334556830093077820796214361531655269017624516106175968498561616239"

	mainPub := "03e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b33"

	fmt.Println("REGTEST")

	// hsm pubkey
	pub := pubFromCoordinates(hsmPubX, hsmPubY)
	pubEnc := hex.EncodeToString(pub.SerializeCompressed())
	fmt.Printf("pub:\t\t%s\n", pubEnc)
	p2pkh, _ := crypto.GetAddressFromPubKey(pub, &chaincfg.RegressionNetParams)
	fmt.Printf("pub P2PKH:\t%s\n\n", p2pkh)

	// main pubkey
	fmt.Printf("pubMain:\t%s\n", mainPub)
	pubmainbytes, _ := hex.DecodeString(mainPub)
	pubMainp2pkh, _ := btcutil.NewAddressPubKeyHash(btcutil.Hash160(pubmainbytes), &chaincfg.RegressionNetParams)
	fmt.Printf("pubMain P2PKH:\t%s\n\n", pubMainp2pkh)

	// generate 1 of 2 multisig
	pubmultistr := fmt.Sprintf("5121%s21%s52ae", mainPub, pubEnc)
	fmt.Printf("1-of-2 MULTISIG script: %s\n", pubmultistr)

	// generate P2SH address
	pubmultibytes, _ := hex.DecodeString(pubmultistr)
	addr, err := btcutil.NewAddressScriptHash(pubmultibytes, &chaincfg.TestNet3Params)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("1-of-2 P2SH address: %s\n", addr.String())
}
