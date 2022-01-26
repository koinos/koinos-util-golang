package util

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// SatoshiToDecimal converts the given UInt64 value to a decimals with the given precision
func SatoshiToDecimal(balance int64, precision int) (*decimal.Decimal, error) {
	denominator, err := decimal.NewFromString(fmt.Sprintf("1e%d", precision))
	if err != nil {
		return nil, err
	}

	v := decimal.NewFromInt(balance).Div(denominator)
	return &v, nil
}

// DecimalToSatoshi converts the given decimal to a satoshi value
func DecimalToSatoshi(d *decimal.Decimal, precision int) (int64, error) {
	multiplier, err := decimal.NewFromString(fmt.Sprintf("1e%d", precision))
	if err != nil {
		return 0, err
	}

	return d.Mul(multiplier).BigInt().Int64(), nil
}
