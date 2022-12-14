package svs

import (
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
)

const (
	TypeVector     enc.TLNum = 0xc9
	TypeEntry      enc.TLNum = 0xca
	TypeEntrySeqno enc.TLNum = 0xcc
)

type NamingScheme int

const (
	SourceOrientedNaming NamingScheme = 0
	GroupOrientedNaming  NamingScheme = 1
)

type HandlingOption int

const (
	NoHandling            HandlingOption = 0
	SourceCentricHandling HandlingOption = 1
	EqualTrafficHandling  HandlingOption = 2
)

type CoreState int32

const (
	Steady      CoreState = 0
	Suppression CoreState = 1
)

type Status int

const (
	Unseen  Status = 0
	Expired Status = 1
	Renewed Status = 2
)
