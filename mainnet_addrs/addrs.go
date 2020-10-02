package mainnet_addrs

import "github.com/ethereum/go-ethereum/common"

const (
	router_mainnet_addr  = "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D"
	factory_mainnet_addr = "0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f"
)

var (
	UNISWAP_ROUTER       = common.HexToAddress(router_mainnet_addr)
	UNISWAP_FACTORY_ADDR = common.HexToAddress(factory_mainnet_addr)
)
