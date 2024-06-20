package poolsidev2

import (
	"context"
	"encoding/json"
	"math/big"
	"strings"
	"time"

	"github.com/KyberNetwork/ethrpc"
	"github.com/KyberNetwork/kyberswap-dex-lib/pkg/entity"
	"github.com/KyberNetwork/kyberswap-dex-lib/pkg/util"
	"github.com/KyberNetwork/logger"
	"github.com/ethereum/go-ethereum/common"
)

type PoolsListUpdater struct {
	config       *Config
	ethrpcClient *ethrpc.Client
}

func NewPoolsListUpdater(
	cfg *Config,
	ethrpcClient *ethrpc.Client,
) *PoolsListUpdater {
	return &PoolsListUpdater{
		config:       cfg,
		ethrpcClient: ethrpcClient,
	}
}

func (u *PoolsListUpdater) GetNewPools(ctx context.Context, metadataBytes []byte) ([]entity.Pool, []byte, error) {
	var (
		dexID     = u.config.DexID
		startTime = time.Now()
	)

	logger.WithFields(logger.Fields{"dex_id": dexID}).Info("Started getting new pools")

	ctx = util.NewContextWithTimestamp(ctx)

	allPairsLength, err := u.getAllPairsLength(ctx)
	if err != nil {
		logger.
			WithFields(logger.Fields{"dex_id": dexID}).
			Error("getAllPairsLength failed")

		return nil, metadataBytes, err
	}

	offset, err := u.getOffset(metadataBytes)
	if err != nil {
		logger.
			WithFields(logger.Fields{"dex_id": dexID, "err": err}).
			Warn("getOffset failed")
	}

	batchSize := u.config.NewPoolLimit

	if offset == allPairsLength {
		batchSize = 0
	}

	if offset+batchSize > allPairsLength {
		batchSize = allPairsLength - offset
	}

	pairAddresses, err := u.listPairAddresses(ctx, offset, batchSize)
	if err != nil {
		logger.
			WithFields(logger.Fields{"dex_id": dexID, "err": err}).
			Error("listPairAddresses failed")

		return nil, metadataBytes, err
	}

	pools, err := u.initPools(ctx, pairAddresses)
	if err != nil {
		logger.
			WithFields(logger.Fields{"dex_id": dexID, "err": err}).
			Error("initPools failed")

		return nil, metadataBytes, err
	}

	newMetadataBytes, err := json.Marshal(PoolsListUpdaterMetadata{
		Offset: offset + len(pairAddresses),
	})
	if err != nil {
		logger.
			WithFields(logger.Fields{"dex_id": dexID, "err": err}).
			Error("newMetadataBytes failed")

		return nil, metadataBytes, err
	}

	logger.
		WithFields(
			logger.Fields{
				"dex_id":      dexID,
				"pools_len":   len(pools),
				"offset":      offset,
				"duration_ms": time.Since(startTime).Milliseconds(),
			},
		).
		Info("Finished getting new pools")

	return pools, newMetadataBytes, nil
}

func (u *PoolsListUpdater) getAllPairsLength(ctx context.Context) (int, error) {
	var allPairsLength *big.Int

	getAllPairsLengthRequest := u.ethrpcClient.NewRequest().SetContext(ctx)

	getAllPairsLengthRequest.AddCall(&ethrpc.Call{
		ABI:    poolsideV2FactoryABI,
		Target: u.config.FactoryAddress,
		Method: factoryMethodAllPairsLength,
		Params: nil,
	}, []interface{}{&allPairsLength})

	if _, err := getAllPairsLengthRequest.Call(); err != nil {
		return 0, err
	}

	return int(allPairsLength.Int64()), nil
}

func (u *PoolsListUpdater) getOffset(metadataBytes []byte) (int, error) {
	if len(metadataBytes) == 0 {
		return 0, nil
	}

	var metadata PoolsListUpdaterMetadata
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return 0, err
	}

	return metadata.Offset, nil
}

func (u *PoolsListUpdater) listPairAddresses(ctx context.Context, offset int, batchSize int) ([]common.Address, error) {
	listPairAddressesResult := make([]common.Address, batchSize)
	listPairAddressesRequest := u.ethrpcClient.NewRequest().SetContext(ctx)

	for i := 0; i < batchSize; i++ {
		index := big.NewInt(int64(offset + i))

		listPairAddressesRequest.AddCall(&ethrpc.Call{
			ABI:    poolsideV2FactoryABI,
			Target: u.config.FactoryAddress,
			Method: factoryMethodGetPair,
			Params: []interface{}{index},
		}, []interface{}{&listPairAddressesResult[i]})
	}

	resp, err := listPairAddressesRequest.TryAggregate()
	if err != nil {
		return nil, err
	}

	var pairAddresses []common.Address
	for i, isSuccess := range resp.Result {
		if !isSuccess {
			continue
		}

		pairAddresses = append(pairAddresses, listPairAddressesResult[i])
	}

	return pairAddresses, nil
}

func (u *PoolsListUpdater) listPairDatas(ctx context.Context, pairAddresses []common.Address) ([]PairData, error) {
	pairDatasResult := make([]PairData, len(pairAddresses))

	listPairDatasRequest := u.ethrpcClient.NewRequest().SetContext(ctx)

	for i, pairAddress := range pairAddresses {
		listPairDatasRequest.AddCall(&ethrpc.Call{
			ABI:    poolsideV2PairABI,
			Target: pairAddress.Hex(),
			Method: pairMethodToken0,
			Params: nil,
		}, []interface{}{&pairDatasResult[i].Token0})

		listPairDatasRequest.AddCall(&ethrpc.Call{
			ABI:    poolsideV2PairABI,
			Target: pairAddress.Hex(),
			Method: pairMethodToken1,
			Params: nil,
		}, []interface{}{&pairDatasResult[i].Token1})

		listPairDatasRequest.AddCall(&ethrpc.Call{
			ABI:    poolsideV2PairABI,
			Target: pairAddress.Hex(),
			Method: pairMethodPlBps,
			Params: nil,
		}, []interface{}{&pairDatasResult[i].PlBps})

		listPairDatasRequest.AddCall(&ethrpc.Call{
			ABI:    poolsideV2PairABI,
			Target: pairAddress.Hex(),
			Method: pairMethodFeeBps,
			Params: nil,
		}, []interface{}{&pairDatasResult[i].FeeBps})
	}

	if _, err := listPairDatasRequest.Aggregate(); err != nil {
		return nil, err
	}

	return pairDatasResult, nil
}

func (u *PoolsListUpdater) getDecimals(ctx context.Context, tokenAddress string) (uint8, error) {
	var decimals uint8

	getDecimalsRequest := u.ethrpcClient.NewRequest().SetContext(ctx)

	getDecimalsRequest.AddCall(&ethrpc.Call{
		ABI:    erc20ABI,
		Target: tokenAddress,
		Method: erc20TokenDecimals,
		Params: nil,
	}, []interface{}{&decimals})

	if _, err := getDecimalsRequest.Call(); err != nil {
		return 0, err
	}

	return decimals, nil
}

func (u *PoolsListUpdater) getRebaseTokenInfo(ctx context.Context, tokenAddress string) RebaseTokenInfo {
	var underlyingToken common.Address

	decimals, err := u.getDecimals(ctx, tokenAddress)

	if err != nil {
		decimals = defaultDecimals
	}

	getUnderlyingTokenRequest := u.ethrpcClient.NewRequest().SetContext(ctx)

	getUnderlyingTokenRequest.AddCall(&ethrpc.Call{
		ABI:    buttonTokenABI,
		Target: tokenAddress,
		Method: buttonTokenMethodGetUnderlyingToken,
		Params: nil,
	}, []interface{}{&underlyingToken})

	if _, err := getUnderlyingTokenRequest.Call(); err != nil {
		return RebaseTokenInfo{
			UnderlyingToken: "",
			WrapRatio:       nil,
			UnwrapRatio:     nil,
			Decimals:        decimals,
		}
	}

	return RebaseTokenInfo{
		UnderlyingToken: underlyingToken.Hex(),
		WrapRatio:       nil,
		UnwrapRatio:     nil,
		Decimals:        decimals,
	}
}

func (u *PoolsListUpdater) initPools(ctx context.Context, pairAddresses []common.Address) ([]entity.Pool, error) {
	pairDatas, err := u.listPairDatas(ctx, pairAddresses)

	if err != nil {
		return nil, err
	}

	pools := make([]entity.Pool, 0, len(pairAddresses))

	for i, pairData := range pairDatas {
		token0Address := strings.ToLower(pairData.Token0.Hex())
		token1Address := strings.ToLower(pairData.Token1.Hex())

		token0 := &entity.PoolToken{
			Address:   token0Address,
			Swappable: true,
		}

		token1 := &entity.PoolToken{
			Address:   token1Address,
			Swappable: true,
		}

		rebaseTokenInfoMap := make(map[string]RebaseTokenInfo)

		rebaseTokenInfoMap[token0Address] = u.getRebaseTokenInfo(ctx, token0Address)
		rebaseTokenInfoMap[token1Address] = u.getRebaseTokenInfo(ctx, token1Address)

		extra, err := json.Marshal(Extra{
			PlBps:              pairData.PlBps,
			FeeBps:             pairData.FeeBps,
			RebaseTokenInfoMap: rebaseTokenInfoMap,
		})

		if err != nil {
			return nil, err
		}

		var newPool = entity.Pool{
			Address:   strings.ToLower(pairAddresses[i].Hex()),
			Exchange:  u.config.DexID,
			Type:      DexType,
			Timestamp: time.Now().Unix(),
			Reserves:  []string{"0", "0"},
			Tokens:    []*entity.PoolToken{token0, token1},
			Extra:     string(extra),
		}

		pools = append(pools, newPool)
	}

	return pools, nil
}
