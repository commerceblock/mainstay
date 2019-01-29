// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"math/big"
	"strings"

	"mainstay/crypto"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
)

// Multisig and pay to address generation for Mainstay

var (
	chain    string
	chainCfg chaincfg.Params

	nKeys int
	nSigs int

	keysX string
	keysY string
	keys  string
)

// init - flag parse
func init() {
	flag.StringVar(&chain, "chain", "", "Bitcoin chain configuration (regtest, testnet or mainnet)")

	flag.IntVar(&nKeys, "nKeys", 0, "Number of keys")
	flag.IntVar(&nSigs, "nSigs", 0, "Number of signatures")

	flag.StringVar(&keysX, "keysX", "", "List of pubkey X coordinates")
	flag.StringVar(&keysY, "keysY", "", "List of pubkey Y coordinates")
	flag.StringVar(&keys, "keys", "", "List of pubkeys")

	flag.Parse()
}

// main
func main() {
	if chain == "regtest" {
		fmt.Println("REGTEST")
		chainCfg = chaincfg.RegressionNetParams
		doRegtest()
	} else {
		if chain == "testnet" {
			fmt.Println("TESTNET")
			chainCfg = chaincfg.TestNet3Params
		} else {
			fmt.Println("MAINNET")
			chainCfg = chaincfg.MainNetParams
		}
		doMain()
	}
}

// Generate multisig and P2Sh info required
// from a list of pubkeys and nKeys/nSigs params
func infoFromPubs(pubs []string, nKeys int, nSigs int) {
	// multisig script
	pubmultistr := fmt.Sprintf("5%d", nSigs)

	// iterate through pubs
	for _, pub := range pubs {
		pubBytes, pubBytesErr := hex.DecodeString(pub)
		if pubBytesErr != nil {
			log.Fatal(fmt.Sprintf("failed decoding pub %s %v", pub, pubBytesErr))
		}
		pubP2pkh, pubP2pkhErr := btcutil.NewAddressPubKeyHash(btcutil.Hash160(pubBytes), &chainCfg)
		if pubP2pkhErr != nil {
			log.Fatal(fmt.Sprintf("failed generating addr from pub %s %v", pub, pubP2pkhErr))
		}
		fmt.Printf("pub P2PKH:\t%s\n", pubP2pkh)

		pubmultistr += "21" + pub
	}

	pubmultistr += fmt.Sprintf("5%d", nKeys)
	pubmultistr += "ae"
	fmt.Printf("%d-of-%d MULTISIG script: %s\n", nSigs, nKeys, pubmultistr)

	// generate P2SH address
	pubmultibytes, _ := hex.DecodeString(pubmultistr)
	addr, err := btcutil.NewAddressScriptHash(pubmultibytes, &chainCfg)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%d-of-%d P2SH address: %s\n", nSigs, nKeys, addr.String())
}

// Generate multisig script and p2sh address for mainstay
// from a list of pubkeys or pubX/pubY coordinates
func doMain() {
	if nKeys <= 0 || nKeys > 15 || nSigs <= 0 || nSigs > 15 || nSigs > nKeys {
		log.Fatal(fmt.Sprintf("invalid nSigs(%d) or nKeys(%d)", nSigs, nKeys))
	}
	if keys == "" && keysX == "" && keysY == "" {
		log.Fatal("Keys missing. Either provide -keys or -keysX and -keysY.")
	}

	if keys == "" {
		keysXSplit := strings.Split(keysX, ",")
		keysYSplit := strings.Split(keysY, ",")

		if len(keysXSplit) != nKeys && len(keysYSplit) != nKeys {
			log.Fatal(fmt.Sprintf("nKeys(%d) but %d keysX and %d keysY provided",
				nKeys, len(keysXSplit), len(keysYSplit)))
		}

		pubs := make([]string, nKeys)
		for i := 0; i < nKeys; i++ {
			pub := pubFromCoordinates(keysXSplit[i], keysYSplit[i])
			pubs[i] = hex.EncodeToString(pub.SerializeCompressed())
		}
		infoFromPubs(pubs, nKeys, nSigs)
	} else {
		keysSplit := strings.Split(keys, ",")
		if len(keysSplit) != nKeys {
			log.Fatal(fmt.Sprintf("nKeys(%d) but %d keys provided", nKeys, len(keysSplit)))
		}
		infoFromPubs(keysSplit, nKeys, nSigs)
	}
}

// Get btcec PublicKey from x/y coordinates
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
	hsmPubX := "17073944010873801765385810419928396464299027769026919728232198509972577863206"
	hsmPubY := "475813022329769762590164284448176075334749443379722569322944728779216384721"

	mainPub := "03e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b33"

	// hsm pubkey
	pub := pubFromCoordinates(hsmPubX, hsmPubY)
	pubEnc := hex.EncodeToString(pub.SerializeCompressed())
	fmt.Printf("pub:\t\t%s\n", pubEnc)
	p2pkh, _ := crypto.GetAddressFromPubKey(pub, &chainCfg)
	fmt.Printf("pub P2PKH:\t%s\n\n", p2pkh)

	// main pubkey
	fmt.Printf("pubMain:\t%s\n", mainPub)
	pubmainbytes, _ := hex.DecodeString(mainPub)
	pubMainp2pkh, _ := btcutil.NewAddressPubKeyHash(btcutil.Hash160(pubmainbytes), &chainCfg)
	fmt.Printf("pubMain P2PKH:\t%s\n\n", pubMainp2pkh)

	infoFromPubs([]string{mainPub, pubEnc}, 2, 1)
}
