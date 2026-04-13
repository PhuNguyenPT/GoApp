package paging

type Pageable[T ~string] struct {
	Page int
	Size int
	Sort Order[T]
}

func NewPageable[T ~string](page, size int, sort Order[T]) Pageable[T] {
	return Pageable[T]{Page: page, Size: size, Sort: sort}
}

func (p Pageable[T]) Offset() int {
	return (p.Page - 1) * p.Size
}

func (p Pageable[T]) Next() Pageable[T] {
	return Pageable[T]{Page: p.Page + 1, Size: p.Size, Sort: p.Sort}
}

func (p Pageable[T]) PreviousOrFirst() Pageable[T] {
	if p.Page <= 1 {
		return p.First()
	}
	return Pageable[T]{Page: p.Page - 1, Size: p.Size, Sort: p.Sort}
}

func (p Pageable[T]) First() Pageable[T] {
	return Pageable[T]{Page: 1, Size: p.Size, Sort: p.Sort}
}
