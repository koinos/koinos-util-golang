package util

import (
	"fmt"

	"github.com/koinos/koinos-proto-golang/v2/koinos/chain"
	"google.golang.org/protobuf/proto"
)

// Void is an empty struct whose main use is making maps act like sets
type Void struct{}

// NonceBytesToUInt64 converts the given nonce bytes to a UInt64
func NonceBytesToUInt64(nonceBytes []byte) (uint64, error) {
	// Extract the uint64 nonce from the response
	var nonce chain.ValueType
	if err := proto.Unmarshal(nonceBytes, &nonce); err != nil {
		return 0, err
	}
	switch x := nonce.Kind.(type) {
	case *chain.ValueType_Uint64Value:
		return x.Uint64Value, nil
	default:
		return 0, fmt.Errorf("%w: expected uint64 value", ErrInvalidNonce)
	}
}

// UInt64ToNonceBytes converts the given nonce uint64 to nonce bytes
func UInt64ToNonceBytes(value uint64) ([]byte, error) {
	nonce := chain.ValueType{Kind: &chain.ValueType_Uint64Value{Uint64Value: value}}
	return proto.Marshal(&nonce)
}
