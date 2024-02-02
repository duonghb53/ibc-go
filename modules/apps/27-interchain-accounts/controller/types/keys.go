package types

import "cosmossdk.io/collections"

const (
	// SubModuleName defines the interchain accounts controller module name
	SubModuleName = "icacontroller"

	// StoreKey is the store key string for the interchain accounts controller module
	StoreKey = SubModuleName

	// ParamsKey is the store key for the interchain accounts controller parameters
	ParamsKey = "params"
)

var DeregisteredAccountsKey = collections.NewPrefix(0)
