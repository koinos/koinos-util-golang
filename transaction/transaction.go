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

// PrepareTransaction prepares a transaction by filling the nonce, rcLimit, and payer, if they are not set
func PrepareTransaction(ctrx context.Context, trx *protocol.Transaction, client *rpc.KoinosRPCClient) error {
	if trx.Header.Payer == nil {
		return fmt.Errorf("no transaction payer set")
	}

	if trx.Header.Nonce == nil || len(trx.Header.Nonce) == 0 {
		if client == nil {
			return fmt.Errorf("rpc client is nil")
		}

		nonceValue, err := client.GetAccountNonce(ctrx, trx.Header.Payer)
		if err != nil {
			return err
		}

		nonce, err := util.UInt64ToNonceBytes(nonceValue + 1)
		if err != nil {
			return err
		}

		trx.Header.Nonce = nonce
	}

	if trx.Header.RcLimit == 0 {
		if client == nil {
			return fmt.Errorf("rpc client is nil")
		}

		rcLimit, err := client.GetAccountRc(ctrx, trx.Header.Payer)
		if err != nil {
			return err
		}

		trx.Header.RcLimit = rcLimit
	}

	if trx.Header.ChainId == nil || len(trx.Header.Nonce) == 0 {
		if client == nil {
			return fmt.Errorf("rpc client is nil")
		}

		chainId, err := client.GetChainID(ctrx)
		if err != nil {
			return err
		}

		trx.Header.ChainId = chainId
	}

	// Fill out the transaction ID, if not set
	if trx.Id == nil || len(trx.Id) == 0 {
		if trx.Header.OperationMerkleRoot == nil || len(trx.Header.OperationMerkleRoot) == 0 {
			// Get operation multihashes
			opHashes := make([][]byte, len(trx.Operations))
			var err error

			for i, op := range trx.Operations {
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

			trx.Header.OperationMerkleRoot = merkleRoot
		}

		headerBytes, err := canonical.Marshal(trx.Header)
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

		trx.Id = tid
	}

	return nil
}

// SignTransaction signs the transaction with the given key
func SignTransaction(trx *protocol.Transaction, key *util.KoinosKey) error {
	privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), key.PrivateBytes())

	// Decode the mutlihashed ID
	idBytes, err := multihash.Decode(trx.Id)
	if err != nil {
		return err
	}

	// Sign the transaction ID
	signatureBytes, err := btcec.SignCompact(btcec.S256(), privateKey, idBytes.Digest, true)
	if err != nil {
		return err
	}

	// Attach the signature data to the transaction
	if trx.Signatures == nil {
		trx.Signatures = [][]byte{signatureBytes}
	} else {
		trx.Signatures = append(trx.Signatures, signatureBytes)
	}

	return nil
}
