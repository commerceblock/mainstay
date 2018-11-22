// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package crypto

import (
	"encoding/hex"
	"fmt"
	"log"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
)

// Various utility functions concerning multisig and scripts

// Raw method to parse a multisig script and get pubkeys and num of sigs
// Allow fatals here as this is only used in AttestClient initialisation
// NOTE: Handle errors if this is used somewhere else in the future
func ParseRedeemScript(script string) ([]*btcec.PublicKey, int) {

	// check op codes
	lscript := len(script)
	op := script[0]
	op1 := script[lscript-4]
	if !(string(op) == string(op1)) && (string(op1) == "5") {
		log.Fatal("Incorrect opcode in redeem script")
	}

	// check multisig
	if script[lscript-2:] != "ae" {
		log.Fatal("Checkmultisig missing from redeem script")
	}

	numOfSigs, _ := strconv.Atoi(string(script[1]))
	numOfKeys, _ := strconv.Atoi(string(script[lscript-3]))

	var startIndex int64 = 2
	var keys []*btcec.PublicKey
	for i := 0; i < numOfKeys; i++ {
		keysize, _ := strconv.ParseInt(string(script[startIndex:startIndex+2]), 16, 16)
		if !(keysize == 65 || keysize == 33) {
			log.Fatal("Incorrect pubkey size")
		}
		keystr := script[startIndex+2 : startIndex+2+2*keysize]
		keybytes, _ := hex.DecodeString(keystr)
		pubkey, err := btcec.ParsePubKey(keybytes, btcec.S256())
		if err != nil {
			log.Fatal(err)
		}
		startIndex += 2 + 2*keysize
		keys = append(keys, pubkey)
	}
	return keys, numOfSigs
}

// Raw method to create a multisig from pubkeys and return P2SH address and redeemScript
func CreateMultisig(pubkeys []*btcec.PublicKey, nSigs int, chainCfg *chaincfg.Params) (btcutil.Address, string) {

	var script string
	script += fmt.Sprintf("5%d", nSigs)

	for _, pub := range pubkeys {
		script += "21"
		script += hex.EncodeToString(pub.SerializeCompressed())
	}

	script += fmt.Sprintf("5%d", len(pubkeys))
	script += "ae"

	scriptBytes, _ := hex.DecodeString(script)
	multisigAddr, _ := btcutil.NewAddressScriptHash(scriptBytes, chainCfg)

	return multisigAddr, script
}

// Parse scriptSig and return sigs and redeemScript
func ParseScriptSig(scriptSig []byte) ([][]byte, []byte) {

	if len(scriptSig) == 0 {
		return [][]byte{}, []byte{}
	}

	var scripts [][]byte
	it := 1
	for {
		scriptSize := scriptSig[it]
		script := scriptSig[it+1 : it+1+int(scriptSize)]
		scripts = append(scripts, script)

		it += 1 + int(scriptSize)

		if len(scriptSig) <= it {
			break
		}
	}

	return scripts[:len(scripts)-1], scripts[len(scripts)-1]
}

// Create scriptSig from sigs and redeemScript
func CreateScriptSig(sigs [][]byte, script []byte) []byte {

	var scriptSig []byte
	scriptSig = append(scriptSig, byte(0))

	for _, sig := range sigs {
		scriptSig = append(scriptSig, byte(len(sig)))
		scriptSig = append(scriptSig, sig...)
	}

	scriptSig = append(scriptSig, byte(len(script)))
	scriptSig = append(scriptSig, script...)

	return scriptSig
}
