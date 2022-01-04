module github.com/koinos/koinos-util-golang

go 1.15

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/ethereum/go-ethereum v1.10.8
	github.com/koinos/koinos-proto-golang v0.2.1-0.20211223221426-6213a3ada193
	github.com/multiformats/go-multihash v0.1.0
	github.com/stretchr/testify v1.7.0
	github.com/ybbus/jsonrpc/v2 v2.1.6
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
)

replace google.golang.org/protobuf => github.com/koinos/protobuf-go v1.27.2-0.20211026185306-2456c83214fe
