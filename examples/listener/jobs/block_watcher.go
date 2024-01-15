package jobs

import (
	"context"
	"encoding/csv"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/berachain/offchain-sdk/job"
	sdk "github.com/berachain/offchain-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

const DexTracExtraData = `6f6f676120626f6f6761`

// Compile time check to ensure that Listener implements job.Basic.
var _ job.Basic = &BlockWatcher{}

type TxInfo struct {
	Hash        common.Hash
	Status      string
	FirstSeenAt time.Time
}

type PoolStat struct {
	txPoolToBlock   uint64
	timePoolToBlock time.Duration
	txDropped       uint64
	timeDropped     time.Duration
	P2BTimes        []float64
	DTimes          []float64
}

// Listener is a simple job that logs the current block when it is run.
type BlockWatcher struct {
	txs  map[common.Hash]TxInfo
	stat PoolStat
}

func median(numbers []float64) float64 {
	sort.Float64s(numbers)
	if len(numbers) == 0 {
		return 0
	}
	middle := len(numbers) / 2
	if len(numbers)%2 == 0 {
		return (numbers[middle-1] + numbers[middle]) / 2
	} else {
		return numbers[middle]
	}
}

func (BlockWatcher) RegistryKey() string {
	return "BlockWatcher"
}

// Execute implements job.Basic.
func (w *BlockWatcher) Execute(ctx context.Context, args any) (any, error) {
	var file *os.File
	var err error
	if w.txs == nil {
		w.txs = make(map[common.Hash]TxInfo)
		file, err = os.Create("txpool.csv")
		if err != nil {
			panic(err)
		}

	} else {
		file, err = os.OpenFile("txpool.csv", os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}

	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	sCtx := sdk.UnwrapContext(ctx)

	newHead := args.(*coretypes.Header)
	if newHead == nil {
		sCtx.Logger().Error("HEADER IS NIL")
		return nil, nil
	}

	// sCtx.Logger().Info("New block", "extra", common.Bytes2Hex(newHead.Extra))

	block, err := sCtx.Chain().BlockByNumber(ctx, newHead.Number)
	if block == nil {
		sCtx.Logger().Error("BLOCK IS NIL. err:", err.Error())
		return nil, nil
	}

	sCtx.Logger().Debug("block_watcher", "height", newHead.Number.String())

	txPoolContent, err := sCtx.Chain().TxPoolContent(ctx)
	if err != nil {
		sCtx.Logger().Error("Cannot get TxPoolContent", "err", err.Error())
	}

	current := time.Unix(int64(block.Time()), 0)
	for txID := range txPoolContent["pending"] {
		for _, tx := range txPoolContent["pending"][txID] {
			if ti, ok := w.txs[tx.Hash()]; !ok {
				w.txs[tx.Hash()] = TxInfo{
					Hash:        tx.Hash(),
					Status:      "pending (" + newHead.Number.String() + ")",
					FirstSeenAt: current,
				}
			} else {
				ti.Status = ti.Status + " -> pending (" + newHead.Number.String() + ")"
			}
			sCtx.Logger().Debug("txpool",
				"tx", tx.Hash().Hex(),
				"status", w.txs[tx.Hash()].Status,
				"first_seen_at", w.txs[tx.Hash()].FirstSeenAt,
			)
		}
	}
	for txID := range txPoolContent["queued"] {
		for _, tx := range txPoolContent["queued"][txID] {
			if ti, ok := w.txs[tx.Hash()]; !ok {
				w.txs[tx.Hash()] = TxInfo{
					Hash:        tx.Hash(),
					Status:      "queued (" + newHead.Number.String() + ")",
					FirstSeenAt: current,
				}
			} else {
				ti.Status = ti.Status + " -> queued (" + newHead.Number.String() + ")"
			}
			sCtx.Logger().Debug("txpool",
				"tx", tx.Hash().Hex(),
				"status", w.txs[tx.Hash()].Status,
				"first_seen_at", w.txs[tx.Hash()].FirstSeenAt,
			)
		}
	}
	for _, tx := range block.Transactions() {
		if ti, ok := w.txs[tx.Hash()]; ok {
			sCtx.Logger().Debug("txpool",
				"tx", tx.Hash().Hex(),
				"status", w.txs[tx.Hash()].Status+" -> included ("+newHead.Number.String()+")",
				"time_in_pool", strconv.FormatFloat(current.Sub(ti.FirstSeenAt).Seconds(), 'f', -1, 64),
			)
			w.stat.txPoolToBlock++
			w.stat.timePoolToBlock += current.Sub(ti.FirstSeenAt)
			w.stat.P2BTimes = append(w.stat.P2BTimes, float64(current.Sub(ti.FirstSeenAt)))
			delete(w.txs, tx.Hash())
		} else {
			sCtx.Logger().Debug("txpool",
				"tx", tx.Hash().Hex(),
				"status", "included (not in pool)",
			)
		}
	}

	pooledTxs := make(map[common.Hash][]byte)
	for txID := range txPoolContent["pending"] {
		for _, tx := range txPoolContent["pending"][txID] {
			pooledTxs[tx.Hash()] = []byte{}
		}
	}
	for txID := range txPoolContent["queued"] {
		for _, tx := range txPoolContent["queued"][txID] {
			pooledTxs[tx.Hash()] = []byte{}
		}
	}
	for _, ti := range w.txs {
		if _, ok := pooledTxs[ti.Hash]; !ok {
			sCtx.Logger().Debug("txpool",
				"tx", ti.Hash.Hex(),
				"status", ti.Status+" -> dropped ("+newHead.Number.String()+")",
				"time_in_pool", strconv.FormatFloat(current.Sub(ti.FirstSeenAt).Seconds(), 'f', -1, 64),
			)
			w.stat.txDropped++
			w.stat.timeDropped += current.Sub(ti.FirstSeenAt)
			w.stat.DTimes = append(w.stat.DTimes, float64(current.Sub(ti.FirstSeenAt)))
			delete(w.txs, ti.Hash)
		}
	}

	writer.Write([]string{
		strconv.FormatFloat(float64(w.stat.timePoolToBlock)/float64(time.Second)/float64(w.stat.txPoolToBlock), 'f', -1, 64),
		strconv.FormatFloat(median(w.stat.P2BTimes)/float64(time.Second), 'f', -1, 64),
		strconv.FormatFloat(float64(w.stat.timeDropped)/float64(time.Second)/float64(w.stat.txDropped), 'f', -1, 64),
		strconv.FormatFloat(median(w.stat.DTimes)/float64(time.Second), 'f', -1, 64),
	})

	// pendingTxNum := 0
	// iterator := sCtx.DB().NewIterator(nil, nil)

	// for iterator.Next() {
	// 	pendingTxNum++
	// }
	// sCtx.Logger().Debug("block_watcher", "height", newHead.Number.String(), "pending", pendingTxNum, "txInBlock", len(block.Transactions()), "gasUsed", block.GasUsed(), "gasLimit", block.GasLimit(), "gasUsed %", float64(block.GasUsed())/float64(block.GasLimit())*100)

	// // filter on block proposed by DexTrac
	// if common.Bytes2Hex(newHead.Extra) == DexTracExtraData {
	// 	time.Sleep(5 * time.Second)
	// 	sCtx.Logger().Debug("DexTrac_block", "height", newHead.Number.String(), "hash", newHead.Hash().Hex())

	// 	txNum := 0
	// 	for _, tx := range block.Transactions() {
	// 		if ok, _ := sCtx.DB().Has(tx.Hash().Bytes()); ok {
	// 			txNum++
	// 		}
	// 	}
	// 	sCtx.Logger().Debug("DexTrac_block", "finalized", txNum, "pending", pendingTxNum, "%", float64(txNum)/float64(pendingTxNum)*100)
	// }

	// iterator = sCtx.DB().NewIterator(nil, nil)

	// for iterator.Next() {
	// 	sCtx.DB().Delete(iterator.Key())
	// }

	return nil, nil
}
