package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/koinos/koinos-proto-golang/koinos"
	"github.com/koinos/koinos-proto-golang/koinos/canonical"
	"github.com/koinos/koinos-proto-golang/koinos/protocol"
	"github.com/multiformats/go-multihash"

	"google.golang.org/protobuf/proto"
)

// MultihashString returns a hex string representation of the given multihash
func MultihashString(hash []byte) string {
	return "0x" + hex.EncodeToString(hash)
}

// BlockString returns a string containing the given block's height and ID
func BlockString(block *protocol.Block) string {
	id := MultihashString(block.Id)
	prevID := MultihashString(block.Header.Previous)
	return fmt.Sprintf("Height: %d ID: %s Prev: %s", block.Header.Height, id, prevID)
}

// TransactionString returns a string containing the given transaction's ID
func TransactionString(transaction *protocol.Transaction) string {
	id := MultihashString(transaction.Id)
	return fmt.Sprintf("ID: %s", string(id))
}

// BlockTopologyString returns a string representation of the BlockTopologyCmp
func BlockTopologyString(topo *koinos.BlockTopology) string {
	id := MultihashString(topo.Id)
	prevID := MultihashString(topo.Previous)
	return fmt.Sprintf("Height: %d ID: %s Prev: %s", topo.Height, id, prevID)
}

// HashMessage takes a protobuf message and returns the multihash of the message
func HashMessage(message proto.Message) ([]byte, error) {
	data, err := canonical.Marshal(message)
	if err != nil {
		panic(err)
	}

	hasher := sha256.New()
	hasher.Write(data)

	// Encode as multihash
	mh, err := multihash.Encode(hasher.Sum(nil), multihash.SHA2_256)
	if err != nil {
		return nil, err
	}

	return mh, nil
}

// DisplayAddress takes address bytes and returns a properly formatted human-readable string
func DisplayAddress(addressBytes []byte) string {
	return fmt.Sprintf("0x%s", hex.EncodeToString(addressBytes))
}

// HexStringToBytes decodes a hex string to a byte slice
func HexStringToBytes(s string) ([]byte, error) {
	return hex.DecodeString(s[2:])
}
