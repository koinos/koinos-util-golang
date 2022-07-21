module github.com/koinos/koinos-util-golang

go 1.15

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/ethereum/go-ethereum v1.10.8
	github.com/koinos/koinos-proto-golang v0.3.1-0.20220708180354-16481ac5469c
	github.com/multiformats/go-multihash v0.1.0
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/ybbus/jsonrpc/v2 v2.1.6
	github.com/ybbus/jsonrpc/v3 v3.1.1 // indirect
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
)

replace google.golang.org/protobuf => github.com/koinos/protobuf-go v1.27.2-0.20211026185306-2456c83214fe
