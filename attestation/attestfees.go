package attestation

import (
	"encoding/json"
	"log"
	"net/http"
)

// Utility functions to get best bitcoin fees from a remote API

// default fee in satoshis
const FEE_PER_BYTE = 20

// response format:
// { "fastestFee": 40, "halfHourFee": 20, "hourFee": 10 }
const FEE_API_URL = "https://bitcoinfees.earn.com/api/v1/fees/recommended"

// default fee type to use from response
// options: fastestFee, halfHourFee, hourFee
const BEST_FEE_TYPE = "hourFee"

// GetFee returns the best fee based on the parameters provided
func GetFee(defaultFee bool, customFeeType ...string) int {
	if defaultFee {
		log.Printf("*Fees* Using default fee value: %d\n", FEE_PER_BYTE)
		return FEE_PER_BYTE
	}

	var feeType = BEST_FEE_TYPE
	if len(customFeeType) > 0 {
		feeType = customFeeType[0]
	}

	fee := GetFeeFromAPI(feeType)
	if fee < FEE_PER_BYTE {
		log.Printf("*Fees* Using default fee value: %d\n", FEE_PER_BYTE)
		return FEE_PER_BYTE
	}
	log.Printf("*Fees* Using fee of type %s and value %d from API\n", feeType, int(fee))
	return fee
}

// GetFeeFromAPI attempts to get the best bitcoinfee from the fee API specified
func GetFeeFromAPI(feeType string) int {
	resp, getErr := http.Get(FEE_API_URL)
	if getErr != nil {
		log.Printf("*Fees* API request failed - Using default fee value: %d\n", FEE_PER_BYTE)
		return FEE_PER_BYTE
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	var respJson map[string]float64
	decErr := dec.Decode(&respJson)
	if decErr != nil {
		log.Printf("*Fees* API response decoding failed - Using default fee value: %d\n", FEE_PER_BYTE)
		return FEE_PER_BYTE
	}

	fee, ok := respJson[feeType]
	if !ok {
		log.Printf("*Fees* API response incorrect format - Using default fee value: %d\n", FEE_PER_BYTE)
		return FEE_PER_BYTE
	}

	return int(fee)
}
