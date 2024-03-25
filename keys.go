package util

import (
	"bytes"
	"crypto/ecdsa"
	"errors"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/koinos/koinos-proto-golang/v2/koinos/protocol"
	"github.com/multiformats/go-multihash"
)

const compressMagic byte = 0x01

// KoinosKey represents a set of keys
type KoinosKey struct {
	PrivateKey *ecdsa.PrivateKey
}

// GenerateKoinosKey generates a new set of keys
func GenerateKoinosKey() (*KoinosKey, error) {
	k, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	keys := &KoinosKey{PrivateKey: k}
	return keys, nil
}

// NewKoinosKeyFromBytes creates a new key set from a private key byte slice
func NewKoinosKeyFromBytes(private []byte) (*KoinosKey, error) {
	pk, err := crypto.ToECDSA(private)
	if err != nil {
		return nil, err
	}

	return &KoinosKey{PrivateKey: pk}, nil
}

// AddressBytes fetches the byte address associated with this key set
func (keys *KoinosKey) AddressBytes() []byte {
	_, pubkey := btcec.PrivKeyFromBytes(btcec.S256(), keys.PrivateBytes())
	mainNetAddr, _ := btcutil.NewAddressPubKey(pubkey.SerializeCompressed(), &chaincfg.MainNetParams)
	return base58.Decode(mainNetAddr.EncodeAddress())
}

// Private gets the private key in WIF format
func (keys *KoinosKey) Private() string {
	return EncodeWIF(crypto.FromECDSA(keys.PrivateKey), true, 128)
}

// Public gets the compressed public key in base58
func (keys *KoinosKey) Public() string {
	return base58.Encode(crypto.CompressPubkey(&keys.PrivateKey.PublicKey))
}

// PublicBytes get the public key bytes
func (keys *KoinosKey) PublicBytes() []byte {
	return crypto.CompressPubkey(&keys.PrivateKey.PublicKey)
}

// PrivateBytes gets the private key bytes
func (keys *KoinosKey) PrivateBytes() []byte {
	return crypto.FromECDSA(keys.PrivateKey)
}

// EncodeWIF encodes a private key into a WIF format string
func EncodeWIF(privKey []byte, compress bool, netID byte) string {
	// Precalculate size.  Maximum number of bytes before base58 encoding
	// is one byte for the network, 32 bytes of private key, possibly one
	// extra byte if the pubkey is to be compressed, and finally four
	// bytes of checksum.
	encodeLen := 1 + 32 + 4
	if compress {
		encodeLen++
	}

	a := make([]byte, 0, encodeLen)
	a = append(a, netID)
	// Pad and append bytes manually, instead of using Serialize, to
	// avoid another call to make.
	a = paddedAppend(btcec.PrivKeyBytesLen, a, privKey)
	if compress {
		a = append(a, compressMagic)
	}
	cksum := chainhash.DoubleHashB(a)[:4]
	a = append(a, cksum...)
	return base58.Encode(a)
}

// paddedAppend appends the src byte slice to dst, returning the new slice.
// If the length of the source is smaller than the passed size, leading zero
// bytes are appended to the dst slice before appending src.
func paddedAppend(size uint, dst, src []byte) []byte {
	for i := 0; i < int(size)-len(src); i++ {
		dst = append(dst, 0)
	}
	return append(dst, src...)
}

// DecodeWIF decodes a WIF format string into bytes
func DecodeWIF(wif string) ([]byte, error) {
	decoded := base58.Decode(wif)
	if len(wif) > 0 && len(decoded) == 0 {
		return nil, errors.New("unable to decode base 58 string")
	}

	decodedLen := len(decoded)
	var compress bool

	// Length of base58 decoded WIF must be 32 bytes + an optional 1 byte
	// (0x01) if compressed, plus 1 byte for netID + 4 bytes of checksum.
	switch decodedLen {
	case 1 + btcec.PrivKeyBytesLen + 1 + 4:
		if decoded[33] != compressMagic {
			return nil, btcutil.ErrMalformedPrivateKey
		}
		compress = true
	case 1 + btcec.PrivKeyBytesLen + 4:
		compress = false
	default:
		return nil, btcutil.ErrMalformedPrivateKey
	}

	// Checksum is first four bytes of double SHA256 of the identifier byte
	// and privKey.  Verify this matches the final 4 bytes of the decoded
	// private key.
	var tosum []byte
	if compress {
		tosum = decoded[:1+btcec.PrivKeyBytesLen+1]
	} else {
		tosum = decoded[:1+btcec.PrivKeyBytesLen]
	}
	cksum := chainhash.DoubleHashB(tosum)[:4]
	if !bytes.Equal(cksum, decoded[decodedLen-4:]) {
		return nil, btcutil.ErrChecksumMismatch
	}

	//netID := decoded[0]
	privKeyBytes := decoded[1 : 1+btcec.PrivKeyBytesLen]

	return privKeyBytes, nil
}

// SignTransaction signs the transaction with the given key
func SignTransaction(key []byte, tx *protocol.Transaction) error {
	privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), key)

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
