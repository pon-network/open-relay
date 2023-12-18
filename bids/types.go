package bids

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/pon-network/open-relay/bulletinboard"
	"github.com/pon-network/open-relay/redisPackage"
)

var (
	builderKeyBid          = "builder-bid"
	builderTimeKeyBid      = "builder-bid-time"
	builderValueKeyBid     = "builder-bid-value"
	builderHighestKeyBid   = "builder-highest-bid"
	bidKeyBuilderUtils     = "builder-bid-utils"
	bidKeyOpenBuilderUtils = "builder-open-bid-utils"
)

var (
	slotProposerDeliveredKey = "slot-proposer-payload-delivered"
	slotBountyBidWinnerKey   = "slot-bounty-bid-winner"
)

type BidBoard struct {
	redisInterface redisPackage.RedisInterface
	log            *logrus.Entry
	bulletinBoard  bulletinboard.RelayMQTT
	bidTimeout     time.Duration
	bidMutex       sync.Mutex
}
