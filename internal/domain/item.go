package domain

type Degree int

const (
	Low Degree = iota
	Medium
	High
)

type Item struct {
	ID           int64
	WishlistID   int64
	Name         string
	Description  string
	Link         string
	DesireDegree Degree
	Booked       bool
}
