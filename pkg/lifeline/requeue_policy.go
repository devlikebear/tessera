package lifeline

type RequeuePolicy struct {
	MaxRecoveries int
}

func (p RequeuePolicy) Allows(recoveries int) bool {
	if p.MaxRecoveries <= 0 {
		return true
	}
	return recoveries < p.MaxRecoveries
}
