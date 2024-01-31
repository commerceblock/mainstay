// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"strings"

	"mainstay/crypto"
	"mainstay/log"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
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
		log.Infoln("REGTEST")
		chainCfg = chaincfg.RegressionNetParams
		doRegtest()
	} else {
		if chain == "testnet" {
			log.Infoln("TESTNET")
			chainCfg = chaincfg.TestNet3Params
		} else {
			log.Infoln("MAINNET")
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
			log.Errorf("failed decoding pub %s %v", pub, pubBytesErr)
		}
		pubP2pkh, pubP2pkhErr := btcutil.NewAddressPubKeyHash(btcutil.Hash160(pubBytes), &chainCfg)
		if pubP2pkhErr != nil {
			log.Errorf("failed generating addr from pub %s %v", pub, pubP2pkhErr)
		}
		log.Infof("pub P2PKH:\t%s\n", pubP2pkh)

		pubmultistr += "21" + pub
	}

	pubmultistr += fmt.Sprintf("5%d", nKeys)
	pubmultistr += "ae"
	log.Infof("%d-of-%d MULTISIG script: %s\n", nSigs, nKeys, pubmultistr)

	// generate P2SH address
	pubmultibytes, _ := hex.DecodeString(pubmultistr)
	addr, err := btcutil.NewAddressScriptHash(pubmultibytes, &chainCfg)
	if err != nil {
		log.Infoln(err)
	}
	log.Infof("%d-of-%d P2SH address: %s\n", nSigs, nKeys, addr.String())
}

// Generate multisig script and p2sh address for mainstay
// from a list of pubkeys or pubX/pubY coordinates
func doMain() {
	if nKeys <= 0 || nKeys > 15 || nSigs <= 0 || nSigs > 15 || nSigs > nKeys {
		log.Errorf("invalid nSigs(%d) or nKeys(%d)", nSigs, nKeys)
	}
	if keys == "" && keysX == "" && keysY == "" {
		log.Error("Keys missing. Either provide -keys or -keysX and -keysY.")
	}

	if keys == "" {
		keysXSplit := strings.Split(keysX, ",")
		keysYSplit := strings.Split(keysY, ",")

		if len(keysXSplit) != nKeys && len(keysYSplit) != nKeys {
			log.Errorf("nKeys(%d) but %d keysX and %d keysY provided",
				nKeys, len(keysXSplit), len(keysYSplit))
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
			log.Errorf("nKeys(%d) but %d keys provided", nKeys, len(keysSplit))
		}
		infoFromPubs(keysSplit, nKeys, nSigs)
	}
}

// Get btcec PublicKey from x/y coordinates
func pubFromCoordinates(xStr string, yStr string) *btcec.PublicKey {
	return btcec.NewPublicKey(crypto.HexToFieldVal(xStr), crypto.HexToFieldVal(yStr))
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
	log.Infof("pub:\t\t%s\n", pubEnc)
	p2pkh, _ := crypto.GetAddressFromPubKey(pub, &chainCfg)
	log.Infof("pub P2PKH:\t%s\n\n", p2pkh)

	// main pubkey
	log.Infof("pubMain:\t%s\n", mainPub)
	pubmainbytes, _ := hex.DecodeString(mainPub)
	pubMainp2pkh, _ := btcutil.NewAddressPubKeyHash(btcutil.Hash160(pubmainbytes), &chainCfg)
	log.Infof("pubMain P2PKH:\t%s\n\n", pubMainp2pkh)

	infoFromPubs([]string{mainPub, pubEnc}, 2, 1)
}
