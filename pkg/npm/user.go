package npm

import "encoding/json"

type Person struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (p *Person) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		p.Name = s
		return nil
	}

	if err := json.Unmarshal(data, &p); err != nil {
		return err
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
