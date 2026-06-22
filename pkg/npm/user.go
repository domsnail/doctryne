package npm

import "encoding/json"

type Person struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type otherPerson struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (p *Person) UnmarshalJSON(data []byte) error {
	var p_ otherPerson
	err := json.Unmarshal(data, &p_)
	if err == nil {
		p.Name = p_.Name
		p.Email = p_.Email

		return nil
	}

	var s string
	err = json.Unmarshal(data, &s)
	if err == nil {
		p.Name = s

		return nil
	}

	return nil
}

type PersonAffiliation string

const (
	PersonAffiliation_Owner       PersonAffiliation = "owner"
	PersonAffiliation_Author      PersonAffiliation = "author"
	PersonAffiliation_Contributor PersonAffiliation = "contributor"
	PersonAffiliation_Sponsor     PersonAffiliation = "sponsor"
	PersonAffiliation_Maintainer  PersonAffiliation = "maintainer"
)
