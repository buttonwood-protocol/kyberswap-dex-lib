package poolsidev2

import (
	"context"
	"encoding/json"
	"math/big"
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

	newMetadataBytes, err := json.Marshal(PoolsListUpdaterMetadata{
		Offset: offset + len(pairAddresses),
	})
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
