package rpc

import (
	"context"
	"crypto/sha256"
	"encoding/json"

	kjson "github.com/koinos/koinos-proto-golang/encoding/json"
	"github.com/koinos/koinos-proto-golang/koinos/canonical"
	"github.com/koinos/koinos-proto-golang/koinos/contract_meta_store"
	"github.com/koinos/koinos-proto-golang/koinos/contracts/token"
	"github.com/koinos/koinos-proto-golang/koinos/protocol"
	"github.com/koinos/koinos-proto-golang/koinos/rpc/chain"
	contract_meta_store_rpc "github.com/koinos/koinos-proto-golang/koinos/rpc/contract_meta_store"
	util "github.com/koinos/koinos-util-golang"
	"github.com/multiformats/go-multihash"
	jsonrpc "github.com/ybbus/jsonrpc/v3"
	"google.golang.org/protobuf/proto"
)

// These are the rpc calls that the wallet uses
const (
	ReadContractCall      = "chain.read_contract"
	GetAccountNonceCall   = "chain.get_account_nonce"
	GetAccountRcCall      = "chain.get_account_rc"
	SubmitTransactionCall = "chain.submit_transaction"
	GetChainIDCall        = "chain.get_chain_id"
	GetContractMetaCall   = "contract_meta_store.get_contract_meta"
)

// SubmissionParams is the parameters for a transaction submission
type SubmissionParams struct {
	Nonce   uint64
	RCLimit uint64
}

// KoinosRPCError is a golang error that also contains log messages from a reverted transaction
type KoinosRPCError struct {
	Logs    []string
	message string
}

// Error returns the error message
func (e KoinosRPCError) Error() string {
	return e.message
}

// KoinosRPCClient is a wrapper around the jsonrpc client
type KoinosRPCClient struct {
	client jsonrpc.RPCClient
}

// NewKoinosRPCClient creates a new koinos rpc client
func NewKoinosRPCClient(url string) *KoinosRPCClient {
	client := jsonrpc.NewClient(url)
	return &KoinosRPCClient{client: client}
}

// Call wraps the rpc client call and handles some of the boilerplate
func (c *KoinosRPCClient) Call(ctx context.Context, method string, params proto.Message, returnType proto.Message) error {
	req, err := kjson.Marshal(params)
	if err != nil {
		return err
	}

	// Make the rpc call
	resp, err := c.client.Call(ctx, method, json.RawMessage(req))
	if err != nil {
		return err
	}
	if resp.Error != nil {
		err := KoinosRPCError{message: resp.Error.Message}

		if data, ok := resp.Error.Data.(string); ok {
			dataMap := make(map[string][]string)
			e := json.Unmarshal([]byte(data), &dataMap)
			if e == nil {
				if logs, ok := dataMap["logs"]; ok {
					err.Logs = logs
				}
			}
		}

		return err
	}

	// Fetch the contract response
	raw := json.RawMessage{}

	err = resp.GetObject(&raw)
	if err != nil {
		return err
	}

	err = kjson.Unmarshal([]byte(raw), returnType)
	if err != nil {
		return err
	}

	return nil
}

// GetAccountBalance gets the balance of a given account
func (c *KoinosRPCClient) GetAccountBalance(ctx context.Context, address []byte, contractID []byte, balanceOfEntry uint32) (uint64, error) {
	// Make the rpc call
	balanceOfArgs := &token.BalanceOfArguments{
		Owner: address,
	}
	argBytes, err := proto.Marshal(balanceOfArgs)
	if err != nil {
		return 0, err
	}

	cResp, err := c.ReadContract(ctx, argBytes, contractID, balanceOfEntry)
	if err != nil {
		return 0, err
	}

	balanceOfReturn := &token.BalanceOfResult{}
	err = proto.Unmarshal(cResp.Result, balanceOfReturn)
	if err != nil {
		return 0, err
	}

	return balanceOfReturn.Value, nil
}

// ReadContract reads from the given contract and returns the response
func (c *KoinosRPCClient) ReadContract(ctx context.Context, args []byte, contractID []byte, entryPoint uint32) (*chain.ReadContractResponse, error) {
	// Build the contract request
	params := chain.ReadContractRequest{ContractId: contractID, EntryPoint: entryPoint, Args: args}

	// Make the rpc call
	var cResp chain.ReadContractResponse
	err := c.Call(ctx, ReadContractCall, &params, &cResp)
	if err != nil {
		return nil, err
	}

	return &cResp, nil
}

// GetAccountRc gets the rc of a given account
func (c *KoinosRPCClient) GetAccountRc(ctx context.Context, address []byte) (uint64, error) {
	// Build the contract request
	params := chain.GetAccountRcRequest{
		Account: address,
	}

	// Make the rpc call
	var cResp chain.GetAccountRcResponse
	err := c.Call(ctx, GetAccountRcCall, &params, &cResp)
	if err != nil {
		return 0, err
	}

	return cResp.Rc, nil
}

// GetAccountNonce gets the nonce of a given account
func (c *KoinosRPCClient) GetAccountNonce(ctx context.Context, address []byte) (uint64, error) {
	// Build the contract request
	params := chain.GetAccountNonceRequest{
		Account: address,
	}

	// Make the rpc call
	var cResp chain.GetAccountNonceResponse
	err := c.Call(ctx, GetAccountNonceCall, &params, &cResp)
	if err != nil {
		return 0, err
	}

	nonce, err := util.NonceBytesToUInt64(cResp.Nonce)
	if err != nil {
		return 0, err
	}

	return nonce, nil
}

// GetContractMeta gets the metadata of a given contract
func (c *KoinosRPCClient) GetContractMeta(ctx context.Context, contractID []byte) (*contract_meta_store.ContractMetaItem, error) {
	// Build the contract request
	params := contract_meta_store_rpc.GetContractMetaRequest{
		ContractId: contractID,
	}

	// Make the rpc call
	var cResp contract_meta_store_rpc.GetContractMetaResponse
	err := c.Call(ctx, GetContractMetaCall, &params, &cResp)
	if err != nil {
		return nil, err
	}

	return cResp.Meta, nil
}

func (c *KoinosRPCClient) SubmitTransaction(ctx context.Context, ops []*protocol.Operation, key *util.KoinosKey, subParams *SubmissionParams, broadcast bool) (*protocol.TransactionReceipt, error) {
	return c.SubmitTransactionWithPayer(ctx, ops, key, subParams, key.AddressBytes(), broadcast, nil)
}

// SubmitTransaction creates and submits a transaction from a list of operations
func (c *KoinosRPCClient) SubmitTransactionWithPayer(ctx context.Context, ops []*protocol.Operation, key *util.KoinosKey, subParams *SubmissionParams, payer []byte, broadcast bool) (*protocol.TransactionReceipt, error) {
	// Cache the public address
	address := key.AddressBytes()

	var err error
	var nonce uint64 = 0
	var rcLimit uint64 = 0

	if subParams != nil {
		nonce = subParams.Nonce
		rcLimit = subParams.RCLimit
	}

	// If the nonce is not provided, get it from the chain
	if nonce == 0 {
		nonce, err = c.GetAccountNonce(ctx, address)
		if err != nil {
			return nil, err
		}
		nonce++
	}

	// Convert nonce to bytes
	nonceBytes, err := util.UInt64ToNonceBytes(nonce)
	if err != nil {
		return nil, err
	}

	// If the rc limit is not provided, get it from the chain
	if rcLimit == 0 {
		rcLimit, err = c.GetAccountRc(ctx, address)
		if err != nil {
			return nil, err
		}
	}

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

	chainID, err := c.GetChainID(ctx)
	if err != nil {
		return nil, err
	}

	// Create the header
	header := protocol.TransactionHeader{ChainId: chainID, RcLimit: rcLimit, Nonce: nonceBytes, OperationMerkleRoot: merkleRoot, Payer: payer}
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

	// Sign the transaction
	err = util.SignTransaction(key.PrivateBytes(), &transaction)

	if err != nil {
		return nil, err
	}

	// Submit the transaction
	params := chain.SubmitTransactionRequest{}
	params.Transaction = &transaction
	params.Broadcast = broadcast

	// Make the rpc call
	var cResp chain.SubmitTransactionResponse
	err = c.Call(ctx, SubmitTransactionCall, &params, &cResp)
	if err != nil {
		return nil, err
	}

	return cResp.Receipt, nil
}

// GetChainID gets the chain id
func (c *KoinosRPCClient) GetChainID(ctx context.Context) ([]byte, error) {
	// Build the contract request
	params := chain.GetChainIdRequest{}

	// Make the rpc call
	var cResp chain.GetChainIdResponse
	err := c.Call(ctx, GetChainIDCall, &params, &cResp)
	if err != nil {
		return nil, err
	}

	return cResp.ChainId, nil
}
