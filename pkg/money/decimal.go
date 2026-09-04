// Package money provides exact decimal arithmetic for monetary aggregates.
package money

import (
	"fmt"
	"math/big"
	"regexp"
)

var decimalPattern = regexp.MustCompile(`^-?[0-9]+(\.[0-9]{1,2})?$`)

// Parse accepts a finite decimal amount with at most two fractional digits.
func Parse(value string) (*big.Rat, error) {
	if !decimalPattern.MatchString(value) {
		return nil, fmt.Errorf("invalid monetary amount %q", value)
	}
	amount, ok := new(big.Rat).SetString(value)
	if !ok {
		return nil, fmt.Errorf("invalid monetary amount %q", value)
	}
	return amount, nil
}
