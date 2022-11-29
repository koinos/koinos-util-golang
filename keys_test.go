package util

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"github.com/btcsuite/btcutil/base58"
	"github.com/stretchr/testify/assert"
)

func TestPrivateWIF(t *testing.T) {
	secret := "foobar"
	wif := "5KJTiKfLEzvFuowRMJqDZnSExxxwspVni1G4RcggoPtDqP5XgM1"

	hasher := sha256.New()

	hasher.Write([]byte(secret))
	key1, err := NewKoinosKeyFromBytes(hasher.Sum(nil))
	assert.NoError(t, err)

	key2Bytes, err := DecodeWIF(wif)
	assert.NoError(t, err)
	key2, err := NewKoinosKeyFromBytes(key2Bytes)
	assert.NoError(t, err)

	assert.True(t, bytes.Equal(key1.PrivateBytes(), key2.PrivateBytes()))

	// Wrong checksum, change last octal (4->3)
	wif = "5KJTiKfLEzvFuowRMJqDZnSExxxwspVni1G4RcggoPtDqP5XgLz"
	_, err = DecodeWIF(wif)
	assert.Error(t, err, assert.AnError)

	// Wrong seed, change first octal of secret (C->D)
	wif = "5KRWQqW5riLTcB39nLw6K7iv2HWBMYvbP7Ch4kUgRd8kEvLH5jH"
	_, err = DecodeWIF(wif)
	assert.Error(t, err, assert.AnError)

	// Wrong prefix, change first octal of prefix (8->7)
	wif = "4nCYtcUpcC6dkge8r2uEJeqrK97TUZ1n7n8LXDgLtun1wRyxU2P"
	_, err = DecodeWIF(wif)
	assert.Error(t, err, assert.AnError)
}

func TestPublicAddress(t *testing.T) {
	wif := "5J1F7GHadZG3sCCKHCwg8Jvys9xUbFsjLnGec4H125Ny1V9nR6V"
	keyBytes, err := DecodeWIF(wif)
	assert.NoError(t, err)
	key, err := NewKoinosKeyFromBytes(keyBytes)
	assert.NoError(t, err)
	addrBytes := key.AddressBytes()

	expectedBytes := []byte{0x00, 0xf5, 0x4a, 0x58, 0x51, 0xe9, 0x37, 0x2b, 0x87, 0x81, 0x0a, 0x8e, 0x60, 0xcd, 0xd2, 0xe7, 0xcf, 0xd8, 0x0b, 0x6e, 0x31, 0xc7, 0xf1, 0x8f, 0xe8}

	assert.True(t, bytes.Equal(addrBytes, expectedBytes))
}

func TestCompressedKey(t *testing.T) {
	uncompressedWIF := "5JtU2c2MHKb8xSeNvsZJpxZRXeRg6iq6uwc6EUtDA9zsWM6B4c5"
	keyBytes, err := DecodeWIF(uncompressedWIF)
	assert.NoError(t, err)
	key, err := NewKoinosKeyFromBytes(keyBytes)
	assert.NoError(t, err)

	expectedWIF := "L1xAJ5axX33g7iBynn9bggE7GGBuaFdK6g1t6W52fQiRvQi73evQ"
	compresseedWIF := key.Private()
	assert.Equal(t, expectedWIF, compresseedWIF)

	expectedAddress := "13Sqw4TrwdZ8RZ9UVfqqA2i3mrbeumcWba"
	assert.Equal(t, expectedAddress, base58.Encode(key.AddressBytes()))

	keyBytes, err = DecodeWIF(expectedWIF)
	assert.NoError(t, err)
	key, err = NewKoinosKeyFromBytes(keyBytes)
	assert.NoError(t, err)
	assert.Equal(t, expectedAddress, base58.Encode(key.AddressBytes()))
}
