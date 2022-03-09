package util

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestSatoshi(t *testing.T) {
	// Sanity check a few values
	checkSatoshi(t, "1234.56789", 123456789000, 8)
	checkSatoshi(t, "0.1", 10000000, 8)
	checkSatoshi(t, "0.02", 2000000, 8)
	checkSatoshi(t, "0.00000001", 1, 8)
}

// Convert values to satoshi and back, make sure values match
func checkSatoshi(t *testing.T, value string, expected uint64, precision int) {
	d, err := decimal.NewFromString(value)
	assert.NoError(t, err)

	v, err := DecimalToSatoshi(&d, precision)
	assert.NoError(t, err)
	assert.Equal(t, expected, v)

	d2, err := SatoshiToDecimal(v, precision)
	assert.NoError(t, err)
	assert.True(t, d.Equals(*d2))
}
