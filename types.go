package util

import (
	"fmt"

	"github.com/koinos/koinos-proto-golang/koinos/protocol"
	"google.golang.org/protobuf/proto"
)

// Void is an empty struct whose main use is making maps act like sets
type Void struct{}

// NonceBytesToUInt64 converts the given nonce bytes to a UInt64
func NonceBytesToUInt64(nonceBytes []byte) (uint64, error) {
	// Extract the uint64 nonce from the response
	var nonce protocol.ValueType
	proto.Unmarshal(nonceBytes, &nonce)
	switch x := nonce.Kind.(type) {
	case *protocol.ValueType_Uint64Value:
		return x.Uint64Value, nil
	default:
		return 0, fmt.Errorf("%w: expected uint64 value", ErrInvalidNonce)
	}
}
