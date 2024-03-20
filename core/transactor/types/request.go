package types

import (
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/berachain/offchain-sdk/types/queue/types"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

// Request is a transaction request, using the go-ethereum call msg.
type Request struct {
	// CallMsg is used to provide the basic tx data. The From field is ignored for txs, only used
	// for eth calls.
	*ethereum.CallMsg

	// MsgID is the (optional) user-provided string id for this tx request.
	MsgID string

	// initialTime is the time at which this tx was initially requested; filled in automatically.
	initialTime time.Time
}

// NewRequest returns a new transaction request with the given input data. The ID is optional,
// but at most 1 is allowed per tx request.
func NewRequest(
	to common.Address, gasLimit uint64, gasFeeCap, gasTipCap, value *big.Int, data []byte,
	msgID ...string,
) *Request {
	if len(msgID) > 1 {
		panic("must only pass in 1 id for a new tx request")
	}
	return &Request{
		CallMsg: &ethereum.CallMsg{
			To:        &to,
			Gas:       gasLimit,
			GasFeeCap: gasFeeCap,
			GasTipCap: gasTipCap,
			Value:     value,
			Data:      data,
		},
		MsgID:       strings.Join(msgID, ""),
		initialTime: time.Now(),
	}
}

// Validate ensures that the initialTime is set on the tx request.
func (r *Request) Validate() error {
	if r.initialTime.Equal(time.Time{}) || (r.initialTime == time.Time{}) {
		return errors.New("timeFired must be set")
	}

	return nil
}

// Time returns the time this tx was initially requested.
func (r *Request) Time() time.Time {
	return r.initialTime
}

// String() implements fmt.Stringer.
func (r *Request) String() string {
	return r.MsgID
}

// New returns a new empty Request.
func (Request) New() types.Marshallable {
	return &Request{}
}

// NewTxRequest returns a new Request with the given type and error.
func (r Request) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// Unmarshal unmarshals a Request from the given data.
func (r *Request) Unmarshal(data []byte) error {
	return json.Unmarshal(data, r)
}

// Requests is a list of requests.
type Requests []*Request

func (rs Requests) Messages() []*ethereum.CallMsg {
	msgs := make([]*ethereum.CallMsg, len(rs))
	for i, req := range rs {
		msgs[i] = req.CallMsg
	}
	return msgs
}

func (rs Requests) MsgIDs() []string {
	ids := make([]string, len(rs))
	for i, r := range rs {
		ids[i] = r.MsgID
	}
	return ids
}

func (rs Requests) Times() []time.Time {
	times := make([]time.Time, len(rs))
	for i, r := range rs {
		times[i] = r.Time()
	}
	return times
}
