// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	"encoding/json"
	"net/http"

	"mainstay/config"
	"mainstay/log"
)

// Utility functions to get best bitcoin fees from a remote API
// Provide min/max values from config and increment fee based
// on schedule, timing and upper/lower limits

// default fee per byte values in satoshis
const (
	DefaultMinFee       = 10
	DefaultMaxFee       = 100
	DefaultFeeIncrement = 5
)

// warnings for arguments
const (
	WarningInvalidMinFeeArg       = "Invalid min fee config value"
	WarningInvalidMaxFeeArg       = "Invalid max fee config value"
	WarningInvalidFeeIncrementArg = "Invalid fee increment config value"
)

// fee api config
const (
	// response format:
	// { "fastestFee": 40, "halfHourFee": 20, "hourFee": 10 }
	FeeApiUrl = "https://bitcoinfees.earn.com/api/v1/fees/recommended"

	// default fee type to use from response
	// options: fastestFee, halfHourFee, hourFee
	DefaultBestFeeType = "hourFee"
)

// AttestFees struct
type AttestFees struct {
	// minimum fee allowed for attestation transactions
	minFee int

	// maximum fee allowed for attestation transactions
	maxFee int

	// constant fee increment on fee bumping case
	feeIncrement int

	// current fee used for attestation transactions
	currentFee int

	// previous fee used for attestation transactions
	prevFee int
}

// New AttestFees instance
// Limit values taken from configuration
// Current fee value reset from api
func NewAttestFees(feesConfig config.FeesConfig) AttestFees {

	// min fee with upper limit max_fee default
	minFee := DefaultMinFee
	if feesConfig.MinFee > 0 && feesConfig.MinFee < DefaultMaxFee {
		minFee = feesConfig.MinFee
	} else {
		log.Warnf("%s (%d)\n", WarningInvalidMinFeeArg, feesConfig.MinFee)
	}
	log.Infof("*Fees* Min fee set to: %d\n", minFee)

	// max fee with lower limit min_fee && 0 and max fee default
	maxFee := DefaultMaxFee
	if feesConfig.MaxFee > 0 && feesConfig.MaxFee > minFee && feesConfig.MaxFee < DefaultMaxFee {
		maxFee = feesConfig.MaxFee
	} else {
		log.Warnf("%s (%d)\n", WarningInvalidMaxFeeArg, feesConfig.MaxFee)
	}
	log.Infof("*Fees* Max fee set to: %d\n", maxFee)

	// fee increment with lower limit 0
	feeIncrement := DefaultFeeIncrement
	if feesConfig.FeeIncrement > 0 {
		feeIncrement = feesConfig.FeeIncrement
	} else {
		log.Warnf("%s (%d)\n", WarningInvalidFeeIncrementArg, feesConfig.FeeIncrement)
	}
	log.Infof("*Fees* Fee increment set to: %d\n", feeIncrement)

	attestFees := AttestFees{
		minFee:       minFee,
		maxFee:       maxFee,
		feeIncrement: feeIncrement,
		prevFee:      0}

	attestFees.ResetFee()
	return attestFees
}

// Get current fee
func (a AttestFees) GetFee() int {
	log.Infof("*Fees* Current fee value: %d\n", a.currentFee)
	return a.currentFee
}

// Get previous fee
func (a AttestFees) GetPrevFee() int {
	log.Infof("*Fees* Previous fee value: %d\n", a.prevFee)
	return a.prevFee
}

// Reset current fee, getting latest best value from API
// Minimum option value to set current fee to minFee
func (a *AttestFees) ResetFee(useMinimum ...bool) {
	var fee int
	if len(useMinimum) > 0 && useMinimum[0] {
		fee = a.minFee
	} else {
		fee = getBestFee()
		if fee < a.minFee {
			fee = a.minFee
		} else if fee > a.maxFee {
			fee = a.maxFee
		}
	}
	a.currentFee = fee
	a.prevFee = 0
	log.Infof("*Fees* Current fee set to value: %d\n", a.currentFee)
}

// Bump fee upon request using increment value and not allowing values higher than max configured fee
func (a *AttestFees) BumpFee() {
	a.prevFee = a.currentFee
	a.currentFee += a.feeIncrement
	log.Infof("*Fees* Bumping fee value to: %d\n", a.currentFee)
	if a.currentFee > a.maxFee {
		log.Infof("*Fees* Max allowed fee value reached: %d\n", a.currentFee)
		a.currentFee = a.maxFee
	}
}

// getBestFee returns the best fee for the type requested from the API
func getBestFee(customFeeType ...string) int {
	var feeType = DefaultBestFeeType
	if len(customFeeType) > 0 {
		feeType = customFeeType[0]
	}

	fee := getFeeFromAPI(feeType)
	return fee
}

// GetFeeFromAPI attempts to get the best bitcoinfee from the fee API specified
func getFeeFromAPI(feeType string) int {
	resp, getErr := http.Get(FeeApiUrl)
	if getErr != nil {
		log.Infoln("*Fees* API request failed")
		return -1
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	var respJson map[string]float64
	decErr := dec.Decode(&respJson)
	if decErr != nil {
		log.Infoln("*Fees* API response decoding failed")
		return -1
	}

	fee, ok := respJson[feeType]
	if !ok {
		log.Infoln("*Fees* API response incorrect format")
		return -1
	}

	return int(fee)
}
