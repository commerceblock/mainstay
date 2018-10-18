// Package crypto contains utilities for generating and validation attestation addresses.
package crypto

import (
    "crypto/ecdsa"
    "math/big"
    "log"
    "strconv"
    "encoding/hex"
    "fmt"

    "github.com/btcsuite/btcutil"
    "github.com/btcsuite/btcd/btcec"
    "github.com/btcsuite/btcd/chaincfg"
)

// Get private key wallet readable format from a string encoded private key
func GetWalletPrivKey(privKey string) *btcutil.WIF {
    key, err := btcutil.DecodeWIF(privKey)
    if err!=nil {
        log.Fatal(err)
    }
    return key
}

// Tweak a private key by adding the tweak to it's integer representation
func TweakPrivKey(walletPrivKey *btcutil.WIF, tweak []byte, chainCfg *chaincfg.Params) *btcutil.WIF {
    // Convert private key and tweak to big Int
    keyVal := new(big.Int).SetBytes(walletPrivKey.PrivKey.Serialize())
    twkVal := new(big.Int).SetBytes(tweak)

    // Add the two Ints (tweaking by scalar addition of private keys)
    resVal := new(big.Int)
    resVal.Add(keyVal, twkVal)

    // In case of addition overflow, apply modulo of the max allowed Int by the curve params, to the result
    n := new(big.Int).Set(btcec.S256().Params().N)
    resVal.Mod(resVal, n)

    // Conver the result back to bytes - new private key
    resPrivKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), resVal.Bytes())

    // Return priv key in wallet readable format
    resWalletPrivKey, err := btcutil.NewWIF(resPrivKey, chainCfg, walletPrivKey.CompressPubKey)
    if err!=nil {
        log.Fatal(err)
    }
    return resWalletPrivKey
}

// Get pay to pub key hash address from a private key
func GetAddressFromPrivKey(walletPrivKey *btcutil.WIF, chainCfg *chaincfg.Params) btcutil.Address {
    return GetAddressFromPubKey(walletPrivKey.PrivKey.PubKey(), chainCfg)
}

// Tweak a pub key by adding the elliptic curve representation of the tweak to the pub key
func TweakPubKey(pubKey *btcec.PublicKey, tweak []byte) *btcec.PublicKey {
    // Get elliptic curve point for the tweak
    _, twkPubKey := btcec.PrivKeyFromBytes(btcec.S256(), tweak)

    // Add the two pub keys using addition on the elliptic curve
    resX, resY := btcec.S256().Add(pubKey.ToECDSA().X, pubKey.ToECDSA().Y, twkPubKey.ToECDSA().X, twkPubKey.ToECDSA().Y)

    return (*btcec.PublicKey)(&ecdsa.PublicKey{btcec.S256(), resX, resY})
}

// Get pay to pub key hash address from a pub key
func GetAddressFromPubKey(pubkey *btcec.PublicKey, chainCfg *chaincfg.Params) btcutil.Address {
    pubKeyHash := btcutil.Hash160(pubkey.SerializeCompressed()) // get pub key hash
    addr, err := btcutil.NewAddressPubKeyHash(pubKeyHash, chainCfg) // encode to address
    if err!=nil {
        log.Fatal(err)
    }
    return addr
}

// Check whether an address has been tweaked by a specific hash
func IsAddrTweakedFromHash(address string, hash []byte, walletPrivKey *btcutil.WIF, chainCfg *chaincfg.Params) bool {
    tweakedPriv := TweakPrivKey(walletPrivKey, hash, chainCfg)
    tweakedAddr := GetAddressFromPrivKey(tweakedPriv, chainCfg)

    return address == tweakedAddr.String()
}

// Raw method to parse a multisig script and get pubkeys and num of sigs
func ParseRedeemScript(script string) ([]*btcec.PublicKey, int) {

    // check op codes
    lscript := len(script)
    op := script[0]
    op1 := script[lscript-4]
    if ! (string(op) == string(op1)) && (string(op1) == "5") {
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
    for i:=0; i<numOfKeys; i++ {
        keysize, _ := strconv.ParseInt(string(script[startIndex:startIndex+2]), 16, 16)
        if ! (keysize == 65 || keysize == 33) {
            log.Fatal("Incorrect pubkey size")
        }
        keystr := script[startIndex+2:startIndex+2+2*keysize]
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
    script += fmt.Sprintf("5%d",nSigs)

    for _, pub := range(pubkeys) {
        script += "21"
        script += hex.EncodeToString(pub.SerializeCompressed())
    }

    script += fmt.Sprintf("5%d",len(pubkeys))
    script += "ae"

    scriptBytes, _ := hex.DecodeString(script)
    multisigAddr, err := btcutil.NewAddressScriptHash(scriptBytes, chainCfg)
    if err != nil {
        log.Fatal(err)
    }

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
        script := scriptSig[it+1:it+1+int(scriptSize)]
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
