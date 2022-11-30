package transaction

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/koinos/koinos-proto-golang/koinos/canonical"
	"github.com/koinos/koinos-proto-golang/koinos/protocol"
	util "github.com/koinos/koinos-util-golang"
	"github.com/koinos/koinos-util-golang/rpc"
	"github.com/multiformats/go-multihash"
)

// Transaction adds builder functions to protocol.Transaction
type Transaction protocol.Transaction

// Prepare prepares a transaction by filling the nonce, rcLimit, and payer, if they are not set
func (t *Transaction) Prepare(ctx context.Context, client *rpc.KoinosRPCClient) error {
	if t.Header.Payer == nil {
		return fmt.Errorf("no transaction payer set")
	}

	if t.Header.Nonce == nil || len(t.Header.Nonce) == 0 {
		if client == nil {
			return fmt.Errorf("rpc client is nil")
		}

		nonceValue, err := client.GetAccountNonce(ctx, t.Header.Payer)
		if err != nil {
			return err
		}

		nonce, err := util.UInt64ToNonceBytes(nonceValue + 1)
		if err != nil {
			return err
		}

		t.Header.Nonce = nonce
	}

	if t.Header.RcLimit == 0 {
		if client == nil {
			return fmt.Errorf("rpc client is nil")
		}

		rcLimit, err := client.GetAccountRc(ctx, t.Header.Payer)
		if err != nil {
			return err
		}

		t.Header.RcLimit = rcLimit
	}

	if t.Header.ChainId == nil || len(t.Header.Nonce) == 0 {
		if client == nil {
			return fmt.Errorf("rpc client is nil")
		}

		chainId, err := client.GetChainID(ctx)
		if err != nil {
			return err
		}

		t.Header.ChainId = chainId
	}

	// Fill out the transaction ID, if not set
	if t.Id == nil || len(t.Id) == 0 {
		if t.Header.OperationMerkleRoot == nil || len(t.Header.OperationMerkleRoot) == 0 {
			// Get operation multihashes
			opHashes := make([][]byte, len(t.Operations))
			var err error

			for i, op := range t.Operations {
				opHashes[i], err = util.HashMessage(op)
				if err != nil {
					return err
				}
			}

			// Find merkle root
			merkleRoot, err := util.CalculateMerkleRoot(opHashes)
			if err != nil {
				return err
			}

			t.Header.OperationMerkleRoot = merkleRoot
		}

		headerBytes, err := canonical.Marshal(t.Header)
		if err != nil {
			return err
		}

		// Calculate the transaction ID
		sha256Hasher := sha256.New()
		sha256Hasher.Write(headerBytes)
		tid, err := multihash.Encode(sha256Hasher.Sum(nil), multihash.SHA2_256)
		if err != nil {
			return err
		}

		t.Id = tid
	}

	return nil
}

// Sign signs the transaction with the given key
func (t *Transaction) Sign(key *util.KoinosKey) error {
	privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), key.PrivateBytes())

	// Decode the mutlihashed ID
	idBytes, err := multihash.Decode(t.Id)
	if err != nil {
		return err
	}

	// Sign the transaction ID
	signatureBytes, err := btcec.SignCompact(btcec.S256(), privateKey, idBytes.Digest, true)
	if err != nil {
		return err
	}

	// Attach the signature data to the transaction
	if t.Signatures == nil {
		t.Signatures = [][]byte{signatureBytes}
	} else {
		t.Signatures = append(t.Signatures, signatureBytes)
	}

	return nil
}
