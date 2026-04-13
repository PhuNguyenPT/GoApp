package paging

import "math"

type Page[C any, S ~string] struct {
	Content       []C
	TotalElements int64
	TotalPages    int
	Pageable      Pageable[S]
}

func NewPage[C any, S ~string](content []C, total int64, pageable Pageable[S]) Page[C, S] {
	totalPages := int(math.Ceil(float64(total) / float64(pageable.Size)))
	if totalPages < 1 {
		totalPages = 1
	}
	return Page[C, S]{
		Content:       content,
		TotalElements: total,
		TotalPages:    totalPages,
		Pageable:      pageable,
	}
}

func (p Page[C, S]) HasNext() bool     { return p.Pageable.Page < p.TotalPages }
func (p Page[C, S]) HasPrevious() bool { return p.Pageable.Page > 1 }
func (p Page[C, S]) IsFirst() bool     { return p.Pageable.Page == 1 }
func (p Page[C, S]) IsLast() bool      { return p.Pageable.Page >= p.TotalPages }
func (p Page[C, S]) IsEmpty() bool     { return len(p.Content) == 0 }
