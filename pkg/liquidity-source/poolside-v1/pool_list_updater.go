package poolsidev1

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

	//batchSize := getBatchSize(allPairsLength, u.config.NewPoolLimit, offset)
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
		Offset: offset + batchSize,
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
		ABI:    poolsideV1FactoryABI,
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

// get pair addresses from Factory by limit
func (u *PoolsListUpdater) listPairAddresses(ctx context.Context, offset int, batchSize int) ([]common.Address, error) {
	listPairAddressesResult := make([]common.Address, batchSize)
	listPairAddressesRequest := u.ethrpcClient.NewRequest().SetContext(ctx)

	for i := 0; i < batchSize; i++ {
		index := big.NewInt(int64(offset + i))

		listPairAddressesRequest.AddCall(&ethrpc.Call{
			ABI:    poolsideV1FactoryABI,
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

// get token pairs from pair list
func (u *PoolsListUpdater) listPairTokens(ctx context.Context, pairAddresses []common.Address) ([]common.Address, []common.Address, error) {
	var (
		listToken0Result = make([]common.Address, len(pairAddresses))
		listToken1Result = make([]common.Address, len(pairAddresses))
	)

	listTokensRequest := u.ethrpcClient.NewRequest().SetContext(ctx)

	for i, pairAddress := range pairAddresses {
		listTokensRequest.AddCall(&ethrpc.Call{
			ABI:    poolsideV1PairABI,
			Target: pairAddress.Hex(),
			Method: pairMethodToken0,
			Params: nil,
		}, []interface{}{&listToken0Result[i]})

		listTokensRequest.AddCall(&ethrpc.Call{
			ABI:    poolsideV1PairABI,
			Target: pairAddress.Hex(),
			Method: pairMethodToken1,
			Params: nil,
		}, []interface{}{&listToken1Result[i]})
	}

	if _, err := listTokensRequest.Aggregate(); err != nil {
		return nil, nil, err
	}

	return listToken0Result, listToken1Result, nil
}

func (u *PoolsListUpdater) initPools(ctx context.Context, pairAddresses []common.Address) ([]entity.Pool, error) {
	token0List, token1List, err := u.listPairTokens(ctx, pairAddresses)

	if err != nil {
		return nil, err
	}

	extra, err := json.Marshal(Extra{
		Fee:          u.config.Fee,
		FeePrecision: u.config.FeePrecision,
	})
	if err != nil {
		return nil, err
	}

	pools := make([]entity.Pool, 0, len(pairAddresses))

	for i, pairAddress := range pairAddresses {
		token0 := &entity.PoolToken{
			Address:   strings.ToLower(token0List[i].Hex()),
			Swappable: true,
		}

		token1 := &entity.PoolToken{
			Address:   strings.ToLower(token1List[i].Hex()),
			Swappable: true,
		}

		var newPool = entity.Pool{
			Address:   strings.ToLower(pairAddress.Hex()),
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