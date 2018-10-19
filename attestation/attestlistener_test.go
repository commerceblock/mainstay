package attestation

import (
	"mainstay/clients"
	"mainstay/test"
	"testing"
)

func TestGetNextHash(t *testing.T) {
	test := test.NewTest(false, false)
	testConfig := test.Config
	var sideClientFake *clients.SidechainClientFake
	sideClientFake = testConfig.OceanClient().(*clients.SidechainClientFake)
	var listener = NewListener(sideClientFake)
	hash, _ := sideClientFake.GetBestBlockHash()
	if listener.GetNextHash() != *hash {
		t.Errorf("GetNextHash() bad return hash")
	}
}
