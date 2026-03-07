package server

import (
	"GoApp/internal/paging"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func resolvePageable[T ~string](c *gin.Context, defaultSize, maxSize int) paging.Pageable[T] {
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(c.Query("size"))
	if size < 1 || size > maxSize {
		size = defaultSize
	}
	field, dir, _ := strings.Cut(c.Query("sort"), ",")
	d := paging.Direction(dir)
	if d != paging.Asc && d != paging.Desc {
		d = paging.Asc
	}
	return paging.NewPageable[T](page, size, paging.Order[T]{Field: T(field), Direction: d})
}
