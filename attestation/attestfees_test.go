// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	"testing"

	"mainstay/config"

	"github.com/stretchr/testify/assert"
)

// Attest Fees test
func TestAttestFees(t *testing.T) {

	attestFees := NewAttestFees(config.FeesConfig{-1, -1, -1})

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

// Attest Fees test with custom feesConfig
func TestAttestFeesWithConfig(t *testing.T) {

	// test attest fees with new config
	attestFees := NewAttestFees(config.FeesConfig{0, 10, 20})
	assert.Equal(t, DefaultMinFee, attestFees.minFee)
	assert.Equal(t, DefaultMaxFee, attestFees.maxFee)
	assert.Equal(t, 20, attestFees.feeIncrement)

	attestFees.ResetFee(true)
	assert.Equal(t, DefaultMinFee, attestFees.GetFee())

	// test attest fees with new config
	attestFees = NewAttestFees(config.FeesConfig{10, 5, 20})
	assert.Equal(t, 10, attestFees.minFee)
	assert.Equal(t, DefaultMaxFee, attestFees.maxFee)
	assert.Equal(t, 20, attestFees.feeIncrement)

	attestFees.ResetFee(true)
	assert.Equal(t, 10, attestFees.GetFee())

	// test attest fees with new config
	attestFees = NewAttestFees(config.FeesConfig{10, 30, 0})
	assert.Equal(t, 10, attestFees.minFee)
	assert.Equal(t, 30, attestFees.maxFee)
	assert.Equal(t, DefaultFeeIncrement, attestFees.feeIncrement)

	attestFees.ResetFee(true)
	assert.Equal(t, 10, attestFees.GetFee())

	// test attest fees with new config
	attestFees = NewAttestFees(config.FeesConfig{10, 0, 40})
	assert.Equal(t, 10, attestFees.minFee)
	assert.Equal(t, DefaultMaxFee, attestFees.maxFee)
	assert.Equal(t, 40, attestFees.feeIncrement)

	attestFees.ResetFee(true)
	assert.Equal(t, 10, attestFees.GetFee())

	// test attest fees with new config
	attestFees = NewAttestFees(config.FeesConfig{110, 110, -30})
	assert.Equal(t, DefaultMinFee, attestFees.minFee)
	assert.Equal(t, DefaultMaxFee, attestFees.maxFee)
	assert.Equal(t, DefaultFeeIncrement, attestFees.feeIncrement)

	attestFees.ResetFee(true)
	assert.Equal(t, DefaultMinFee, attestFees.GetFee())
}
