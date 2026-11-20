package domain

// RaisonSuppression is why a vendeur withdraws an annonce, one code of a
// closed list. It travels on the event only: the annonce does not keep it.
type RaisonSuppression struct {
	code string
}

// Codes of the raisons de suppression; they are stable (JSON, events).
const (
	RaisonVendueSurLeSite   = "vendue-sur-le-site"
	RaisonVendueAilleurs    = "vendue-ailleurs"
	RaisonPlusEnvieDeVendre = "plus-envie-de-vendre"
	RaisonAutre             = "autre"
)

// NewRaisonSuppression checks that code belongs to the closed list.
func NewRaisonSuppression(code string) (RaisonSuppression, error) {
	switch code {
	case RaisonVendueSurLeSite, RaisonVendueAilleurs, RaisonPlusEnvieDeVendre, RaisonAutre:
		return RaisonSuppression{code: code}, nil
	}
	return RaisonSuppression{}, ErrRaisonSuppressionInvalide
}

func (r RaisonSuppression) Code() string { return r.code }
