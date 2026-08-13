package entity

type QueryFilter struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type Sorting struct {
	Field string `json:"field"`
	Order string `json:"order"`
}
