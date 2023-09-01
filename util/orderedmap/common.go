package orderedmap

type Ordering int

const (
	Canonical          Ordering = 0
	LatestEntriesFirst Ordering = 1
)

type MetaV struct {
	Old bool
}
