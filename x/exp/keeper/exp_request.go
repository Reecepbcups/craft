package keeper

import (
	"encoding/binary"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/notional-labs/craft/x/exp/types"
)

func (k ExpKeeper) addAddressToBurnRequestList(ctx sdk.Context, memberAccount string, tokenLeft *sdk.Coin) error {
	burnReques := types.BurnRequest{
		Account:       memberAccount,
		BurnTokenLeft: tokenLeft,
		RequestTime:   ctx.BlockTime(),
		Status:        types.StatusOnGoingRequest,
	}
	k.SetBurnRequest(ctx, burnReques)

	return nil
}

func (k ExpKeeper) completeBurnRequest(ctx sdk.Context, burnRequest types.BurnRequest) {
	k.removeBurnRequest(ctx, burnRequest)
	// unexcepted status
	if burnRequest.Status == types.StatusOnGoingRequest {
		burnRequest.Status = -1
	}
	k.setEndedBurnRequest(ctx, burnRequest)
}

func (k ExpKeeper) addAddressToMintRequestList(ctx sdk.Context, memberAccount sdk.AccAddress, tokenLeft sdk.Dec) {
	mintRequest := types.MintRequest{
		Account:        memberAccount.String(),
		DaoTokenLeft:   tokenLeft,
		DaoTokenMinted: sdk.NewDec(0),
		Status:         types.StatusOnGoingRequest,
		RequestTime:    ctx.BlockHeader().Time,
	}

	k.SetMintRequest(ctx, mintRequest)
}

func (k ExpKeeper) completeMintRequest(ctx sdk.Context, mintRequest types.MintRequest) {
	k.removeMintRequest(ctx, mintRequest)
	// unexcepted status
	if mintRequest.Status == types.StatusOnGoingRequest {
		mintRequest.Status = -1
	}
	k.setEndedMintRequest(ctx, mintRequest)
}

// need modify for better performance .
func (k ExpKeeper) GetDaoTokenPrice(ctx sdk.Context) sdk.Dec {
	asset, _ := k.GetDaoAssetInfo(ctx)

	return asset.DaoTokenPrice
}

// calculate exp value by ibc asset .
func (k ExpKeeper) calculateDaoTokenValue(ctx sdk.Context, amount math.Int) sdk.Dec {
	daoTokenPrice := k.GetDaoTokenPrice(ctx)

	decCoin := sdk.NewDecFromInt(amount)
	return decCoin.Quo(daoTokenPrice)
}

func (k ExpKeeper) SetBurnRequest(ctx sdk.Context, burnRequest types.BurnRequest) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&burnRequest)

	accAddress, err := sdk.AccAddressFromBech32(burnRequest.Account)
	if err != nil {
		panic(err)
	}
	store.Set(types.GetBurnRequestAddressBytes(accAddress), bz)
}

func (k ExpKeeper) GetBurnRequestByKey(ctx sdk.Context, key []byte) (types.BurnRequest, error) {
	var burnRequest types.BurnRequest

	store := ctx.KVStore(k.storeKey)
	if !store.Has(key) {
		return types.BurnRequest{}, sdkerrors.Wrapf(types.ErrInvalidKey, "burnRequest")
	}

	bz := store.Get(key)
	err := k.cdc.Unmarshal(bz, &burnRequest)
	if err != nil {
		return types.BurnRequest{}, err
	}

	return burnRequest, nil
}

func (k ExpKeeper) setEndedBurnRequest(ctx sdk.Context, burnRequest types.BurnRequest) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&burnRequest)

	accAddress, err := sdk.AccAddressFromBech32(burnRequest.Account)
	if err != nil {
		panic(err)
	}
	store.Set(types.GetEndedBurnRequestKey(accAddress), bz)
}

// GetAllBurnRequest returns all the burn request from store .
func (k ExpKeeper) GetAllBurnRequests(ctx sdk.Context) (burnRequests types.BurnRequests) {
	k.IterateBurnRequests(ctx, func(burnRequest types.BurnRequest) bool {
		burnRequests = append(burnRequests, burnRequest)
		return false
	})
	return
}

// IterateBurnRequest iterates over the all the BurnRequest and performs a callback function .
func (k ExpKeeper) IterateBurnRequests(ctx sdk.Context, cb func(burnRequest types.BurnRequest) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.KeyBurnRequestList)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var burnRequest types.BurnRequest
		err := k.cdc.Unmarshal(iterator.Value(), &burnRequest)
		if err != nil {
			panic(err)
		}

		if cb(burnRequest) {
			break
		}
	}
}

func (k ExpKeeper) removeBurnRequest(ctx sdk.Context, burnRequest types.BurnRequest) {
	store := ctx.KVStore(k.storeKey)
	accAddress, _ := sdk.AccAddressFromBech32(burnRequest.Account)
	if store.Has(types.GetBurnRequestAddressBytes(accAddress)) {
		store.Delete(types.GetBurnRequestAddressBytes(accAddress))
	}
}

func (k ExpKeeper) GetBurnRequestsByStatus(ctx sdk.Context, status int) (burnRequests types.BurnRequests) {
	k.IterateStatusBurnRequests(ctx, status, func(burnRequest types.BurnRequest) bool {
		burnRequests = append(burnRequests, burnRequest)
		return false
	})
	return
}

func (k ExpKeeper) GetBurnRequest(ctx sdk.Context, accAddress sdk.AccAddress) (types.BurnRequest, error) {
	return k.GetBurnRequestByKey(ctx, types.GetBurnRequestAddressBytes(accAddress))
}

// IterateBurnRequest iterates over the all the BurnRequest and performs a callback function .
func (k ExpKeeper) IterateStatusBurnRequests(ctx sdk.Context, status int, cb func(burnRequest types.BurnRequest) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.KeyBurnRequestList)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var burnRequest types.BurnRequest
		err := k.cdc.Unmarshal(iterator.Value(), &burnRequest)
		if err != nil {
			panic(err)
		}

		if cb(burnRequest) {
			break
		}
	}
}

func (k ExpKeeper) removeMintRequest(ctx sdk.Context, mintRequest types.MintRequest) {
	store := ctx.KVStore(k.storeKey)

	accAddress, _ := sdk.AccAddressFromBech32(mintRequest.Account)
	if store.Has(types.GetMintRequestAddressBytes(accAddress)) {
		store.Delete(types.GetMintRequestAddressBytes(accAddress))
	}
}

func (k ExpKeeper) SetMintRequest(ctx sdk.Context, mintRequest types.MintRequest) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&mintRequest)

	accAddress, err := sdk.AccAddressFromBech32(mintRequest.Account)
	if err != nil {
		panic(err)
	}
	store.Set(types.GetMintRequestAddressBytes(accAddress), bz)
}

func (k ExpKeeper) GetMintRequestsByStatus(ctx sdk.Context, status int) (mintRequests types.MintRequests) {
	k.IterateStatusMintRequests(ctx, status, func(mintRequest types.MintRequest) bool {
		mintRequests = append(mintRequests, mintRequest)
		return false
	})
	return
}

// GetAllMintRequest returns all the MintRequest from store .
func (k ExpKeeper) GetAllMintRequest(ctx sdk.Context) (mintRequests types.MintRequests) {
	k.IterateMintRequest(ctx, func(mintRequest types.MintRequest) bool {
		mintRequests = append(mintRequests, mintRequest)
		return false
	})
	return
}

func (k ExpKeeper) setEndedMintRequest(ctx sdk.Context, mintRequest types.MintRequest) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&mintRequest)

	accAddress, err := sdk.AccAddressFromBech32(mintRequest.Account)
	if err != nil {
		panic(err)
	}
	store.Set(types.GetEndedMintRequestKey(accAddress), bz)
}

func (k ExpKeeper) GetMintRequestByKey(ctx sdk.Context, key []byte) (mintRequest types.MintRequest, found bool) {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(key) {
		return types.MintRequest{}, false
	}

	bz := store.Get(key)
	err := k.cdc.Unmarshal(bz, &mintRequest)
	if err != nil {
		return types.MintRequest{}, false
	}

	return mintRequest, true
}

// IterateMintRequest iterates over the all the MintRequest and performs a callback function .
func (k ExpKeeper) IterateMintRequest(ctx sdk.Context, cb func(mintRequest types.MintRequest) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.KeyMintRequestList)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var mintRequest types.MintRequest
		err := k.cdc.Unmarshal(iterator.Value(), &mintRequest)
		if err != nil {
			panic(err)
		}

		if cb(mintRequest) {
			break
		}
	}
}

// IterateStatusMintRequests iterates over the all the BurnRequest and performs a callback function .
func (k ExpKeeper) IterateStatusMintRequests(ctx sdk.Context, status int, cb func(mintRequest types.MintRequest) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.KeyMintRequestList)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var mintRequest types.MintRequest
		err := k.cdc.Unmarshal(iterator.Value(), &mintRequest)
		if err != nil {
			panic(err)
		}

		if cb(mintRequest) {
			break
		}
	}
}

// IncreaseOracleID increase oracle ID by 1.
func (k ExpKeeper) IncreaseOracleID(ctx sdk.Context) {
	k.SetOracleID(ctx, k.GetNextOracleID(ctx))
}

func (k ExpKeeper) SetOracleID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.KeyOracleID, GetOracleIDBytes(id))
}

// GetOracleIDBytes returns the byte representation of the OracleID.
func GetOracleIDBytes(id uint64) (idbz []byte) {
	idbz = make([]byte, 8)
	binary.BigEndian.PutUint64(idbz, id)
	return
}

// GetNextOracleID return next oracle ID.
func (k ExpKeeper) GetNextOracleID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyOracleID)
	return binary.BigEndian.Uint64(bz) + 1
}

// GetCurreentOracleID return current oracle ID.
func (k ExpKeeper) GetCurreentOracleID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyOracleID)
	return binary.BigEndian.Uint64(bz)
}

// SetNextOracleRequest set oracle request and increase oracle ID by 1.
func (k ExpKeeper) SetNextOracleRequest(ctx sdk.Context, oracleRequest types.OracleRequest) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&oracleRequest)
	nextID := k.GetNextOracleID(ctx)
	// NOTE: append only to an already created slice.
	// You could use types.KeyOracleRequest = append(types.KeyOracleRequest, GetOracleIDBytes(nextID)...).
	// But wanted to keep readability simple.
	key := types.KeyOracleRequest
	key = append(key, GetOracleIDBytes(nextID)...)
	k.IncreaseOracleID(ctx)

	store.Set(key, bz)
}

// GetOracleRequest get oracle request by oracleID.
func (k ExpKeeper) GetOracleRequest(ctx sdk.Context, oracleID uint64) (oracleRequest types.OracleRequest) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyOracleRequest
	key = append(key, GetOracleIDBytes(oracleID)...)
	bz := store.Get(key)

	if bz == nil {
		return types.OracleRequest{}
	}

	_ = k.cdc.Unmarshal(bz, &oracleRequest)
	// if err != nil { // if we did this, it breaks all of app.go.
	// 	panic(err)
	// }

	return oracleRequest
}

// // IterateOracleRequest iterates over the all the Oracle Request .
// func (k ExpKeeper) IterateOracleRequest(ctx sdk.Context, cb func(oracleRequest types.OracleRequest) (stop bool)) {
// 	store := ctx.KVStore(k.storeKey)

// 	iterator := sdk.KVStorePrefixIterator(store, types.KeyOracleRequest)
// 	defer iterator.Close()
// 	for ; iterator.Valid(); iterator.Next() {
// 		var oracleRequest types.OracleRequest
// 		err := k.cdc.Unmarshal(iterator.Value(), &oracleRequest)
// 		if err != nil {
// 			panic(err)
// 		}

// 		if cb(oracleRequest) {
// 			break
// 		}
// 	}
// }

// func (k ExpKeeper) GetAllOracleRequest(ctx sdk.Context) (oracleRequests []types.OracleRequest) {
// 	k.IterateOracleRequest(ctx, func(oracleRequest types.OracleRequest) bool {
// 		oracleRequests = append(oracleRequests, oracleRequest)
// 		return false
// 	})
// 	return
// }

func (k ExpKeeper) SetBurnRequestOracle(ctx sdk.Context, oracle types.OracleRequest) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&oracle)
	// https://github.com/go-critic/go-critic/issues/865
	key := types.KeyOracleRequest
	key = append(key, []byte(oracle.AddressRequest)...)

	store.Set(key, bz)
}

func (k ExpKeeper) GetBurnRequestOracle(ctx sdk.Context, addressBurn string) (oracle types.OracleRequest, found bool) {
	store := ctx.KVStore(k.storeKey)
	// https://github.com/go-critic/go-critic/issues/865
	key := types.KeyOracleRequest
	key = append(key, []byte(addressBurn)...)

	bz := store.Get(key)
	if bz == nil {
		return types.OracleRequest{}, false
	}

	err := k.cdc.Unmarshal(bz, &oracle)
	if err != nil {
		panic(err)
	}
	return oracle, true
}

func (k ExpKeeper) RemoveBurnRequestOracle(ctx sdk.Context, addressBurn string) {
	store := ctx.KVStore(k.storeKey)
	// https://github.com/go-critic/go-critic/issues/865
	key := types.KeyOracleRequest
	key = append(key, []byte(addressBurn)...)

	if store.Has(key) {
		store.Delete(key)
	}
}
