package domain

// Prix is the amount asked for an annonce, in centimes d'euro. Zero is a don;
// there is no ceiling.
type Prix struct {
	centimes int64
}

func NewPrix(centimes int64) (Prix, error) {
	if centimes < 0 {
		return Prix{}, ErrPrixNegatif
	}
	return Prix{centimes: centimes}, nil
}

func (p Prix) Centimes() int64 { return p.centimes }
func (p Prix) EstUnDon() bool  { return p.centimes == 0 }
