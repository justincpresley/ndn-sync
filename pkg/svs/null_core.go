package svs

import (
	"time"

	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
	ndn "github.com/zjkmxy/go-ndn/pkg/ndn"
)

type nullCore struct{}

func newNullCore() *nullCore                              { return &nullCore{} }
func (c *nullCore) Listen()                               {}
func (c *nullCore) Activate(immediateStart bool)          {}
func (c *nullCore) Shutdown()                             {}
func (c *nullCore) Update(dataset enc.Name, seqno uint64) {}
func (c *nullCore) StateVector() StateVector              { return NewStateVector() }
func (c *nullCore) FeedInterest(interest ndn.Interest, rawInterest enc.Wire, sigCovered enc.Wire, reply ndn.ReplyFunc, deadline time.Time) {
}
func (c *nullCore) Subscribe() chan SyncUpdate { return nil }
