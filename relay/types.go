package relay

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/attestantio/go-builder-client/api"
	"github.com/attestantio/go-builder-client/api/capella"
	apiv1 "github.com/attestantio/go-builder-client/api/v1"
	consensusspec "github.com/attestantio/go-eth2-client/spec"
	consensusbellatrix "github.com/attestantio/go-eth2-client/spec/bellatrix"
	capellaSpec "github.com/attestantio/go-eth2-client/spec/capella"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	bulletinBoardTypes "github.com/bsn-eng/pon-golang-types/bulletinBoard"
	commonTypes "github.com/bsn-eng/pon-golang-types/common"
	databaseTypes "github.com/bsn-eng/pon-golang-types/database"
	relayTypes "github.com/bsn-eng/pon-golang-types/relay"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/flashbots/go-boost-utils/types"
	boostTypes "github.com/flashbots/go-boost-utils/types"
	"github.com/holiman/uint256"
	"github.com/pingcap/errors"
	"github.com/sirupsen/logrus"

	beaconclient "github.com/pon-network/open-relay/beaconinterface"
	bidBoard "github.com/pon-network/open-relay/bids"
	"github.com/pon-network/open-relay/bls"
	"github.com/pon-network/open-relay/bulletinboard"
	"github.com/pon-network/open-relay/database"
	ponpool "github.com/pon-network/open-relay/ponPool"
	"github.com/pon-network/open-relay/reporter"
	"github.com/pon-network/open-relay/signing"
	"github.com/pon-network/open-relay/utils"
)

type Signature phase0.BLSSignature
type EcdsaAddress [20]byte
type EcdsaSignature [65]byte
type Hash [32]byte
type PublicKey [48]byte
type Transaction []byte

func (h Hash) String() string {
	return hexutil.Bytes(h[:]).String()
}

func (e EcdsaAddress) String() string {
	return hexutil.Bytes(e[:]).String()
}

func (e EcdsaSignature) String() string {
	return hexutil.Bytes(e[:]).String()
}

func (t Transaction) String() string {
	transaction, _ := json.Marshal(t)
	return string(transaction)
}

type Relay struct {
	db             *database.DatabaseInterface
	ponPool        *ponpool.PonRegistrySubgraph
	bulletinBoard  *bulletinboard.RelayMQTT
	beaconClient   *beaconclient.MultiBeaconClient
	bidBoard       *bidBoard.BidBoard
	URL            string
	blsSk          *bls.SecretKey
	log            *logrus.Entry
	reporterServer *reporter.ReporterServer
	network        EthNetwork
	publicKey      phase0.BLSPubKey
	client         *http.Client
	server         *http.Server
	relayutils     *utils.RelayUtils
	version        string
	openRelay      bool
}

type RelayParams struct {
	DbURL          string
	DatabaseParams databaseTypes.DatabaseOpts
	DbDriver       databaseTypes.DatabaseDriver
	DeleteTables   bool

	URL string

	PonPoolURL    string
	PonPoolAPIKey string

	BulletinBoardParams bulletinBoardTypes.RelayMQTTOpts

	BeaconClientUrls []string

	ReporterURL string

	Network string

	RedisURI string

	BidTimeOut time.Duration

	Sk *bls.SecretKey

	Version string

	DiscordWebhook string
	OpenRelay      bool
}

type EthNetwork struct {
	Network             uint64
	GenesisTime         uint64
	DomainBuilder       signing.Domain
	DomainBeaconCapella signing.Domain
}

type RelayServerParams struct {
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

type ProposerReqParams struct {
	Slot              uint64
	ProposerPubKeyHex string
	ParentHashHex     string
}

type BuilderWinningBid struct {
	BidID             string  `json:"bid_id"`
	HighestBidValue   big.Int `json:"highest_bid_value"`
	HighestBidBuilder string  `json:"highest_bid_builder"`
}

type RelayConfig struct {
	MQTTBroker string `json:"mqtt_broker"`
	MQTTPort   uint16 `json:"mqtt_port"`
	PublicKey  string `json:"public_key"`
	Chain      uint64 `json:"chain"`
	Slot       uint64 `json:"current_slot"`
}

type Address [20]byte
type U256Str Hash
type (
	Root = Hash
)
type Bloom [256]byte

var (
	ErrLength = fmt.Errorf("incorrect byte length")
	ErrSign   = fmt.Errorf("negative value casted as unsigned int")
)

func reverse(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	for i := len(dst)/2 - 1; i >= 0; i-- {
		opp := len(dst) - 1 - i
		dst[i], dst[opp] = dst[opp], dst[i]
	}
	return dst
}

func (n U256Str) MarshalText() ([]byte, error) {
	return []byte(new(big.Int).SetBytes(reverse(n[:])).String()), nil
}

func (n *U256Str) UnmarshalJSON(input []byte) error {
	if len(input) < 2 {
		return ErrLength
	}
	x := new(big.Int)
	err := x.UnmarshalJSON(input[1 : len(input)-1])
	if err != nil {
		return err
	}
	return n.FromBig(x)
}

func (n *U256Str) UnmarshalText(input []byte) error {
	x := new(big.Int)
	err := x.UnmarshalText(input)
	if err != nil {
		return err
	}
	return n.FromBig(x)
}

func (n *U256Str) String() string {
	return new(big.Int).SetBytes(reverse(n[:])).String()
}

func (n *U256Str) FromSlice(x []byte) error {
	if len(x) > 32 {
		return ErrLength
	}
	copy(n[:], x)
	return nil
}

func (n *U256Str) FromBig(x *big.Int) error {
	if x.BitLen() > 256 {
		return ErrLength
	}
	if x.Sign() == -1 {
		return ErrSign
	}
	copy(n[:], reverse(x.FillBytes(n[:])))
	return nil
}

type BuilderOpenBid struct {
	Capella *capella.SubmitBlockRequest
}

type BidTrace struct {
	Slot                 uint64    `json:"slot,string"`
	ParentHash           Hash      `json:"parent_hash" ssz-size:"32"`
	BlockHash            Hash      `json:"block_hash" ssz-size:"32"`
	BuilderPubkey        PublicKey `json:"builder_pubkey" ssz-size:"48"`
	ProposerPubkey       PublicKey `json:"proposer_pubkey" ssz-size:"48"`
	ProposerFeeRecipient Address   `json:"proposer_fee_recipient" ssz-size:"20"`
	GasLimit             uint64    `json:"gas_limit,string"`
	GasUsed              uint64    `json:"gas_used,string"`
	Value                U256Str   `json:"value" ssz-size:"32"`
}

type ExecutionPayload struct {
	ParentHash    Hash            `json:"parent_hash" ssz-size:"32"`
	FeeRecipient  Address         `json:"fee_recipient" ssz-size:"20"`
	StateRoot     Root            `json:"state_root" ssz-size:"32"`
	ReceiptsRoot  Root            `json:"receipts_root" ssz-size:"32"`
	LogsBloom     Bloom           `json:"logs_bloom" ssz-size:"256"`
	Random        Hash            `json:"prev_randao" ssz-size:"32"`
	BlockNumber   uint64          `json:"block_number,string"`
	GasLimit      uint64          `json:"gas_limit,string"`
	GasUsed       uint64          `json:"gas_used,string"`
	Timestamp     uint64          `json:"timestamp,string"`
	ExtraData     []byte          `json:"extra_data" ssz-max:"32"`
	BaseFeePerGas U256Str         `json:"base_fee_per_gas" ssz-max:"32"`
	BlockHash     Hash            `json:"block_hash" ssz-size:"32"`
	Transactions  []hexutil.Bytes `json:"transactions" ssz-max:"1048576,1073741824" ssz-size:"?,?"`
}

func (b *BuilderOpenBid) Message() *apiv1.BidTrace {
	if b.Capella != nil {
		return b.Capella.Message
	}
	return nil
}
func (b *BuilderOpenBid) HasExecutionPayload() bool {
	if b.Capella != nil {
		return b.Capella.ExecutionPayload != nil
	}
	return false
}

func (b *BuilderOpenBid) Signature() phase0.BLSSignature {
	if b.Capella != nil {
		return b.Capella.Signature
	}
	return phase0.BLSSignature{}
}

func (b *BuilderOpenBid) BuilderPubkey() phase0.BLSPubKey {
	if b.Capella != nil {
		return b.Capella.Message.BuilderPubkey
	}
	return phase0.BLSPubKey{}
}

var ZeroU256 = boostTypes.IntToU256(0)

func (b *BuilderOpenBid) Value() *big.Int {
	if b.Capella != nil {
		return b.Capella.Message.Value.ToBig()
	}
	return nil
}

func (b *BuilderOpenBid) NumTx() int {
	if b.Capella != nil {
		return len(b.Capella.ExecutionPayload.Transactions)
	}
	return 0
}

func (b *BuilderOpenBid) ExecutionPayloadResponse() (*utils.GetPayloadResponse, error) {

	if b.Capella != nil {
		return &utils.GetPayloadResponse{
			Capella: &api.VersionedExecutionPayload{
				Version:   consensusspec.DataVersionCapella,
				Capella:   b.Capella.ExecutionPayload,
				Bellatrix: nil,
			},
		}, nil
	}

	return nil, errors.New("no execution payload")
}

func (b *BuilderOpenBid) GetHeaderResponse(signedBid *relayTypes.SignedBuilderBlockBid) (*utils.GetHeaderResponse, error) {
	if b.Capella != nil {
		return &utils.GetHeaderResponse{
			Version: "capella",
			Data:    signedBid,
		}, nil
	}

	return nil, errors.New("no execution payload")
}

func (b *BuilderOpenBid) ExecutionPayloadHeader() (*commonTypes.VersionedExecutionPayloadHeader, error) {
	if b.Capella != nil {
		//convert b.Capella.ExecutionPayload to commonTypes.VersionedExecutionPayloadHeader
		versionedExecutionPayloadHeader := commonTypes.VersionedExecutionPayloadHeader{
			Bellatrix: nil,
			Capella: &capellaSpec.ExecutionPayloadHeader{
				ParentHash:    phase0.Hash32(b.Capella.ExecutionPayload.ParentHash),
				FeeRecipient:  consensusbellatrix.ExecutionAddress(b.Capella.ExecutionPayload.FeeRecipient),
				StateRoot:     phase0.Hash32(b.Capella.ExecutionPayload.StateRoot),
				ReceiptsRoot:  phase0.Hash32(b.Capella.ExecutionPayload.ReceiptsRoot),
				BlockNumber:   b.Capella.ExecutionPayload.BlockNumber,
				GasLimit:      b.Capella.ExecutionPayload.GasLimit,
				GasUsed:       b.Capella.ExecutionPayload.GasUsed,
				Timestamp:     b.Capella.ExecutionPayload.Timestamp,
				ExtraData:     b.Capella.ExecutionPayload.ExtraData,
				BaseFeePerGas: b.Capella.ExecutionPayload.BaseFeePerGas,
				BlockHash:     phase0.Hash32(b.Capella.ExecutionPayload.BlockHash),
			},
			Deneb: nil,
		}
		return &versionedExecutionPayloadHeader, nil
	}
	return nil, errors.New("no capella execution payload provided")
}

func (b *BuilderOpenBid) Slot() uint64 {
	if b.Capella != nil {
		return b.Capella.Message.Slot
	}
	return 0
}

func BoostBidToBidTrace(bidTrace *boostTypes.BidTrace) *apiv1.BidTrace {
	if bidTrace == nil {
		return nil
	}
	return &apiv1.BidTrace{
		BuilderPubkey:        phase0.BLSPubKey(bidTrace.BuilderPubkey),
		Slot:                 bidTrace.Slot,
		ProposerPubkey:       phase0.BLSPubKey(bidTrace.ProposerPubkey),
		ProposerFeeRecipient: consensusbellatrix.ExecutionAddress(bidTrace.ProposerFeeRecipient),
		BlockHash:            phase0.Hash32(bidTrace.BlockHash),
		Value:                U256StrToUint256(bidTrace.Value),
		ParentHash:           phase0.Hash32(bidTrace.ParentHash),
		GasLimit:             bidTrace.GasLimit,
		GasUsed:              bidTrace.GasUsed,
	}
}

func U256StrToUint256(s types.U256Str) *uint256.Int {
	i := new(uint256.Int)
	i.SetBytes(reverse(s[:]))
	return i
}

type BuilderGetValidatorsResponseEntry struct {
	Slot           uint64                                  `json:"slot,string"`
	ValidatorIndex uint64                                  `json:"validator_index,string"`
	Entry          *boostTypes.SignedValidatorRegistration `json:"entry"`
}
