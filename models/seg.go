package models

type SegMent struct {
	Index    int
	Start    int
	End      int
	Url      string
	Count    int
	Complete bool
	Cache    []byte
}
