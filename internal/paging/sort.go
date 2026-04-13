package paging

type Direction string

const (
	Asc  Direction = "asc"
	Desc Direction = "desc"
)

type Order[T ~string] struct {
	Field     T
	Direction Direction
}
