package svs

type StatusChange interface {
	Source() string
	OldStatus() Status
	NewStatus() Status
}

type statusChange struct {
	source  string
	oldStat Status
	newStat Status
}

func NewStatusChange(s string, o Status, n Status) StatusChange {
	return statusChange{source: s, oldStat: o, newStat: n}
}

func (sc statusChange) Source() string    { return sc.source }
func (sc statusChange) OldStatus() Status { return sc.oldStat }
func (sc statusChange) NewStatus() Status { return sc.newStat }
