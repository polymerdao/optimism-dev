package scheduler

import "github.com/ethereum-optimism/optimism/op-service/txmgr"

type TxSender interface {
	SendAndWaitSimple(txPurpose string, txs ...txmgr.TxCandidate) error
}
