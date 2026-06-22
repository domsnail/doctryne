package npm

import "encoding/json"

type Bugs struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}

type otherBugs struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}

func (b *Bugs) UnmarshalJSON(data []byte) error {
	var b_ otherBugs
	err := json.Unmarshal(data, &b_)
	if err == nil {
		b.Url = b_.Url
		if b_.Name != "" {
			b.Url = b_.Name
		}

		return nil
	}

	var s string
	err = json.Unmarshal(data, &s)
	if err == nil {
		b.Url = s

		return nil
	}

	return nil
}
