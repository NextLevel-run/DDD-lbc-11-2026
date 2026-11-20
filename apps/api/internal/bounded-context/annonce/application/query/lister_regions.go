package query

import "ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"

// RegionAvecDepartementsView is a region and its departements, as a lieu
// selector shows them.
type RegionAvecDepartementsView struct {
	Code         string
	Nom          string
	Departements []DepartementView
}

// ListerRegionsQuery lists every region, metropolitan first, each with its
// departements in code order: the lieux a recherche accepts.
type ListerRegionsQuery func() []RegionAvecDepartementsView

func BuildListerRegionsQuery() ListerRegionsQuery {
	return func() []RegionAvecDepartementsView {
		regions := domain.Regions()
		views := make([]RegionAvecDepartementsView, len(regions))
		for i, region := range regions {
			departements := domain.DepartementsDe(region)
			view := RegionAvecDepartementsView{
				Code:         region.Code(),
				Nom:          region.Nom(),
				Departements: make([]DepartementView, len(departements)),
			}
			for j, departement := range departements {
				view.Departements[j] = DepartementView{Code: departement.Code(), Nom: departement.Nom()}
			}
			views[i] = view
		}
		return views
	}
}
