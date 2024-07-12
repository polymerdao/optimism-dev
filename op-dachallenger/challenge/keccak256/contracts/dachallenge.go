package contracts

import (
	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/contracts/metrics"
	"github.com/ethereum-optimism/optimism/op-service/sources/batching"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)


type DAChallengeContract struct {
	metrics     metrics.ContractMetricer
	multiCaller *batching.MultiCaller
	contract    *batching.BoundContract
	abi         *abi.ABI
}

func NewDAChallengeContract(m metrics.ContractMetricer, addr common.Address, caller *batching.MultiCaller) *DAChallengeContract {
	panic("implement me")
}
