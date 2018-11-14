package attestation

import (
	"encoding/json"
	"log"
	"net/http"
)

// Utility functions to get best bitcoin fees from a remote API
// Provide min/max values from config and increment fee based
// on schedule, timing and upper/lower limits

// default fee per byte values in satoshis
const (
	DEFAULT_MIN_FEE       = 10
	DEFAULT_MAX_FEE       = 100
	DEFAULT_FEE_INCREMENT = 5
)

// fee api config
const (
	// response format:
	// { "fastestFee": 40, "halfHourFee": 20, "hourFee": 10 }
	FEE_API_URL = "https://bitcoinfees.earn.com/api/v1/fees/recommended"

	// default fee type to use from response
	// options: fastestFee, halfHourFee, hourFee
	DEFAULT_BEST_FEE_TYPE = "hourFee"
)

// AttestFees struct
type AttestFees struct {
	minFee       int
	maxFee       int
	feeIncrement int
	currentFee   int
}

// New AttestFees instance
// Limit values taken from configuration
// Current fee value reset from api
func NewAttestFees() AttestFees {

	attestFees := AttestFees{
		minFee:       DEFAULT_MIN_FEE,
		maxFee:       DEFAULT_MAX_FEE,
		feeIncrement: DEFAULT_FEE_INCREMENT}

	attestFees.ResetFee()
	return attestFees
}

// Get current fee
func (a AttestFees) GetFee() int {
	log.Printf("*Fees* Current fee value: %d\n", a.currentFee)
	return a.currentFee
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
	log.Printf("*Fees* Current fee set to value: %d\n", a.currentFee)
}

// Bump fee upon request using increment value and not allowing values higher than max configured fee
func (a *AttestFees) BumpFee() {
	a.currentFee += a.feeIncrement
	log.Printf("*Fees* Bumping fee value to: %d\n", a.currentFee)
	if a.currentFee > a.maxFee {
		log.Printf("*Fees* Max allowed fee value reached: %d\n", a.currentFee)
		a.currentFee = a.maxFee
	}
}

// getBestFee returns the best fee for the type requested from the API
func getBestFee(customFeeType ...string) int {
	var feeType = DEFAULT_BEST_FEE_TYPE
	if len(customFeeType) > 0 {
		feeType = customFeeType[0]
	}

	fee := getFeeFromAPI(feeType)
	return fee
}

// GetFeeFromAPI attempts to get the best bitcoinfee from the fee API specified
func getFeeFromAPI(feeType string) int {
	resp, getErr := http.Get(FEE_API_URL)
	if getErr != nil {
		log.Println("*Fees* API request failed")
		return -1
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	var respJson map[string]float64
	decErr := dec.Decode(&respJson)
	if decErr != nil {
		log.Println("*Fees* API response decoding failed")
		return -1
	}

	fee, ok := respJson[feeType]
	if !ok {
		log.Println("*Fees* API response incorrect format")
		return -1
	}

	return int(fee)
}
