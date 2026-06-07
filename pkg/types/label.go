package types

type Label uint8

const (
	Label_Indirect Label = iota
	Label_Dev
	Label_Optional
	Label_Root
)
