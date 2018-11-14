package attestation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Attest Fees test
func TestAttestFees(t *testing.T) {
	attestFees := NewAttestFees()

	// test reset to minimum
	attestFees.ResetFee(true)
	assert.Equal(t, attestFees.minFee, attestFees.GetFee())

	// test reset never sets a value smaller than min and larger than max
	attestFees.ResetFee()
	assert.Equal(t, true, attestFees.GetFee() >= attestFees.minFee)
	assert.Equal(t, true, attestFees.GetFee() <= attestFees.maxFee)

	// test fee bumping maintains current fee within limits
	attestFees.feeIncrement = 20
	attestFees.minFee = 10
	attestFees.maxFee = 100
	attestFees.ResetFee(true)
	fee := attestFees.GetFee()
	for _, i := range []int{1, 2, 3, 4} {
		attestFees.BumpFee()
		assert.Equal(t, fee+i*attestFees.feeIncrement, attestFees.GetFee())
	}

	attestFees.BumpFee()
	assert.Equal(t, attestFees.maxFee, attestFees.GetFee())
}
