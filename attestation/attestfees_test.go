package attestation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Attest Fees test
func TestGetFee(t *testing.T) {
	feePerByte := GetFee(true)
	assert.Equal(t, FEE_PER_BYTE, feePerByte)
}
