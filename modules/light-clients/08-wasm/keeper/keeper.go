package keeper

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strings"

	wasmvm "github.com/CosmWasm/wasmvm"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"
)

// Keeper defines the 08-wasm keeper
type Keeper struct {
	// implements gRPC QueryServer interface
	types.QueryServer

	storeKey  storetypes.StoreKey
	cdc       codec.BinaryCodec
	wasmVM    types.WasmEngine
	authority string
}

// NewKeeperWithVM creates a new Keeper instance with the provided Wasm VM.
// This constructor function is meant to be used when the chain uses x/wasm
// and the same Wasm VM instance should be shared with it.
func NewKeeperWithVM(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	authority string,
	vm types.WasmEngine,
) Keeper {
	if vm == nil {
		panic(errors.New("wasm VM must be not nil"))
	}

	if types.WasmVM != nil && !reflect.DeepEqual(types.WasmVM, vm) {
		panic(errors.New("global Wasm VM instance should not be set to a different instance"))
	}

	if strings.TrimSpace(authority) == "" {
		panic(errors.New("authority must be non-empty"))
	}

	types.WasmVM = vm
	types.WasmStoreKey = key

	return Keeper{
		cdc:       cdc,
		storeKey:  key,
		wasmVM:    vm,
		authority: authority,
	}
}

// NewKeeperWithConfig creates a new Keeper instance with the provided Wasm configuration.
// This constructor function is meant to be used when the chain does not use x/wasm
// and a Wasm VM needs to be instantiated using the provided parameters.
func NewKeeperWithConfig(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	authority string,
	wasmConfig types.WasmConfig,
) Keeper {
	vm, err := wasmvm.NewVM(wasmConfig.DataDir, wasmConfig.SupportedFeatures, types.ContractMemoryLimit, wasmConfig.ContractDebugMode, wasmConfig.MemoryCacheSize)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate new Wasm VM instance: %v", err))
	}

	return NewKeeperWithVM(cdc, key, authority, vm)
}

// GetAuthority returns the 08-wasm module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

func generateWasmCodeHash(code []byte) []byte {
	hash := sha256.Sum256(code)
	return hash[:]
}

func (k Keeper) storeWasmCode(ctx sdk.Context, code []byte) ([]byte, error) {
	store := ctx.KVStore(k.storeKey)

	var err error
	if types.IsGzip(code) {
		ctx.GasMeter().ConsumeGas(types.VMGasRegister.UncompressCosts(len(code)), "Uncompress gzip bytecode")
		code, err = types.Uncompress(code, types.MaxWasmByteSize())
		if err != nil {
			return nil, errorsmod.Wrap(err, "failed to store contract")
		}
	}

	// Check to see if the store has a code with the same code it
	expectedHash := generateWasmCodeHash(code)
	codeHashKey := types.CodeHashKey(expectedHash)
	if store.Has(codeHashKey) {
		return nil, types.ErrWasmCodeExists
	}

	// run the code through the wasm light client validation process
	if err := types.ValidateWasmCode(code); err != nil {
		return nil, errorsmod.Wrapf(err, "wasm bytecode validation failed")
	}

	// create the code in the vm
	ctx.GasMeter().ConsumeGas(types.VMGasRegister.CompileCosts(len(code)), "Compiling wasm bytecode")
	codeHash, err := k.wasmVM.StoreCode(code)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to store contract")
	}

	// pin the code to the vm in-memory cache
	if err := k.wasmVM.Pin(codeHash); err != nil {
		return nil, errorsmod.Wrapf(err, "failed to pin contract with code hash (%s) to vm cache", codeHash)
	}

	// safety check to assert that code hash returned by WasmVM equals to code hash
	if !bytes.Equal(codeHash, expectedHash) {
		return nil, errorsmod.Wrapf(types.ErrInvalidCodeHash, "expected %s, got %s", hex.EncodeToString(expectedHash), hex.EncodeToString(codeHash))
	}

	store.Set(codeHashKey, code)
	return codeHash, nil
}