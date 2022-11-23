package transaction

import (
	"bytes"
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

type TransactionBuilder struct {
	ops             []*protocol.Operation
	rpcClient       *rpc.KoinosRPCClient
	nonceBytes      []byte
	rcLimit         uint64
	rcLimitAbsolute bool
	chainID         []byte
	payer           []byte
	key             *util.KoinosKey
}

// AddOperations adds a list of operations to the transaction
func (tb *TransactionBuilder) AddOperations(ops ...*protocol.Operation) {
	tb.ops = append(tb.ops, ops...)
}

// SetRPCClient sets the RPC client to use for the transaction builder
func (tb *TransactionBuilder) SetRPCClient(rpcClient *rpc.KoinosRPCClient) {
	tb.rpcClient = rpcClient
}

// SetNonce sets the nonce of the transaction
func (tb *TransactionBuilder) SetNonce(nonce uint64) error {
	bytes, err := util.UInt64ToNonceBytes(nonce)
	if err != nil {
		return err
	}

	tb.nonceBytes = bytes
	return nil
}

// SetRPCLimit sets the RC limit of the transaction
func (tb *TransactionBuilder) SetRCLimit(rcLimit uint64, absolute bool) {
	tb.rcLimit = rcLimit
	tb.rcLimitAbsolute = absolute
}

// SetChainID sets the chain ID of the transaction
func (tb *TransactionBuilder) SetChainID(chainID []byte) {
	tb.chainID = chainID
}

// SetPayer sets the payer of the transaction
func (tb *TransactionBuilder) SetPayer(payer []byte) {
	tb.payer = payer
}

// SetKey sets the key of the transaction
func (tb *TransactionBuilder) SetKey(key *util.KoinosKey) {
	tb.key = key
}

// CreateTransaction creates a transaction from the operations in the builder
func (tb *TransactionBuilder) Build(ctx context.Context, signed bool) (*protocol.Transaction, error) {
	if len(tb.ops) == 0 {
		return nil, fmt.Errorf("%w: no operations to build transaction", ErrInvalidTransactionBuilderRequest)
	}

	if tb.key == nil {
		return nil, fmt.Errorf("%w: no key given", ErrInvalidTransactionBuilderRequest)
	}

	address := tb.key.AddressBytes()

	// Fetch the nonce if it is not set
	nonce := tb.nonceBytes
	if tb.nonceBytes == nil {
		if tb.rpcClient == nil {
			return nil, fmt.Errorf("%w: no nonce given and no RPC client set", ErrInvalidTransactionBuilderRequest)
		}

		nonceVal, err := tb.rpcClient.GetAccountNonce(ctx, address)
		if err != nil {
			return nil, err
		}

		nonce, err = util.UInt64ToNonceBytes(nonceVal + 1)
		if err != nil {
			return nil, err
		}
	}

	// Get the RC limit
	rcLimit := tb.rcLimit
	if !tb.rcLimitAbsolute {
		if tb.rpcClient == nil {
			return nil, fmt.Errorf("%w: no RC limit given and no RPC client set", ErrInvalidTransactionBuilderRequest)
		}

		rcLimitVal, err := tb.rpcClient.GetAccountRc(ctx, address)
		if err != nil {
			return nil, err
		}

		rcDec, err := util.SatoshiToDecimal(rcLimitVal, 8)
		if err != nil {
			return nil, err
		}

		fracDec, err := util.SatoshiToDecimal(rcLimit, 8)
		if err != nil {
			return nil, err
		}

		// Multiply to get the absolute RC limit decimal
		rcLimitDec := rcDec.Mul(*fracDec)

		rcLimit, err = util.DecimalToSatoshi(&rcLimitDec, 8)
		if err != nil {
			return nil, err
		}
	}

	// Get the chain ID
	chainID := tb.chainID
	if chainID == nil {
		if tb.rpcClient == nil {
			return nil, fmt.Errorf("%w: no chain ID given and no RPC client set", ErrInvalidTransactionBuilderRequest)
		}

		var err error
		chainID, err = tb.rpcClient.GetChainID(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Get the payer
	payer := tb.payer
	if payer == nil {
		payer = address
	}

	// Create the transaction and return it if not requesting signed
	transaction, err := CreateTransaction(ctx, tb.ops, address, nonce, rcLimit, chainID, payer)
	if !signed {
		return transaction, err
	}

	// Sign the transaction
	err = SignTransaction(tb.key.PrivateBytes(), transaction)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

// CreateTransaction creates a transaction from a list of operations with a specified payer
func CreateTransaction(ctx context.Context, ops []*protocol.Operation, address []byte, nonce []byte, rcLimit uint64, chainID []byte, payer []byte) (*protocol.Transaction, error) {
	var err error

	// Get operation multihashes
	opHashes := make([][]byte, len(ops))
	for i, op := range ops {
		opHashes[i], err = util.HashMessage(op)
		if err != nil {
			return nil, err
		}
	}

	// Find merkle root
	merkleRoot, err := util.CalculateMerkleRoot(opHashes)
	if err != nil {
		return nil, err
	}

	// Create the header
	var header protocol.TransactionHeader
	if bytes.Equal(payer, address) {
		header = protocol.TransactionHeader{ChainId: chainID, RcLimit: rcLimit, Nonce: nonce, OperationMerkleRoot: merkleRoot, Payer: payer}
	} else {
		header = protocol.TransactionHeader{ChainId: chainID, RcLimit: rcLimit, Nonce: nonce, OperationMerkleRoot: merkleRoot, Payer: payer, Payee: address}
	}

	headerBytes, err := canonical.Marshal(&header)
	if err != nil {
		return nil, err
	}

	// Calculate the transaction ID
	sha256Hasher := sha256.New()
	sha256Hasher.Write(headerBytes)
	tid, err := multihash.Encode(sha256Hasher.Sum(nil), multihash.SHA2_256)
	if err != nil {
		return nil, err
	}

	// Create the transaction
	transaction := protocol.Transaction{Header: &header, Operations: ops, Id: tid}

	return &transaction, nil
}

// SignTransaction signs the transaction with the given key
func SignTransaction(private []byte, tx *protocol.Transaction) error {
	privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), private)

	// Decode the mutlihashed ID
	idBytes, err := multihash.Decode(tx.Id)
	if err != nil {
		return err
	}

	// Sign the transaction ID
	signatureBytes, err := btcec.SignCompact(btcec.S256(), privateKey, idBytes.Digest, true)
	if err != nil {
		return err
	}

	// Attach the signature data to the transaction
	if tx.Signatures == nil {
		tx.Signatures = [][]byte{signatureBytes}
	} else {
		tx.Signatures = append(tx.Signatures, signatureBytes)
	}

	return nil
}
