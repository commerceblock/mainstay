package attestation

import (
	"mainstay/clients"
	"mainstay/test"
	"testing"
)

func TestgetNextHash(t *testing.T) {
	test := test.NewTest(false, false)
	testConfig := test.Config
	var sideClientFake *clients.SidechainClientFake
	sideClientFake = testConfig.OceanClient().(*clients.SidechainClientFake)
	var listener = NewAttestListener(nil, nil, sideClientFake)
	hash, _ := sideClientFake.GetBestBlockHash()
	if listener.getNextHash() != *hash {
		t.Errorf("getNextHash() bad return hash")
	}
}
