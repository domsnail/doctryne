package npm

type Person struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type PersonAffiliation string

const (
	PersonAffiliation_Owner       PersonAffiliation = "owner"
	PersonAffiliation_Author      PersonAffiliation = "author"
	PersonAffiliation_Contributor PersonAffiliation = "contributor"
	PersonAffiliation_Sponsor     PersonAffiliation = "sponsor"
	PersonAffiliation_Maintainer  PersonAffiliation = "maintainer"
)
