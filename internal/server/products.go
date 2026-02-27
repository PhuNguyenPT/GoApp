package server

import (
	"GoApp/internal/database"
	"GoApp/internal/views"
	"log"
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *Server) productsPageHandler(c *gin.Context) {
	ctx := c.Request.Context()

	selectedSource := c.Query("source")
	selectedCategory := c.Query("category")
	pageStr := c.DefaultQuery("page", "1")
	pageNumber, err := strconv.Atoi(pageStr)
	if err != nil {
		log.Printf("error parsing pageNumber: %v", err)
		pageNumber = 1
	}
	if pageNumber < 1 {
		pageNumber = 1
	}

	sources, err := s.db.GetDistinctSources(ctx)
	if err != nil {
		log.Printf("error getting distinct sources: %v", err)
	}

	categories, err := s.db.GetCategoriesBySource(ctx, selectedSource)
	if err != nil {
		log.Printf("error getting categories by source: %v", err)
	}

	total, err := s.db.CountProductsBySourceAndCategory(ctx, database.CountProductsBySourceAndCategoryParams{
		Column1: selectedSource,
		Column2: selectedCategory,
	})
	if err != nil {
		log.Printf("error counting products: %v", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(s.cfg.PageSize)))
	offset := (pageNumber - 1) * s.cfg.PageSize

	products, err := s.db.GetProductsBySourceAndCategory(ctx, database.GetProductsBySourceAndCategoryParams{
		Column1: selectedSource,
		Column2: selectedCategory,
		Limit:   int32(s.cfg.PageSize),
		Offset:  int32(offset),
	})
	if err != nil {
		log.Printf("error getting products: %v", err)
	}
	data := views.ProductsPageData{
		Sources:          sources,
		Categories:       categories,
		Products:         products,
		SelectedSource:   selectedSource,
		SelectedCategory: selectedCategory,
		Page:             pageNumber,
		TotalPages:       totalPages,
		TotalCount:       total,
	}

	userName, _ := c.Get("userName")
	userNameStr, _ := userName.(string)
	lang := getLangStr(c)

	if c.GetHeader("HX-Request") == "true" {
		c.Header("Content-Type", "text/html")
		if err := views.CategorySelectOOB(categories, selectedCategory, lang).Render(c.Request.Context(), c.Writer); err != nil {
			log.Printf("error rendering CategorySelectOOB: %v", err)
		}
		if err := views.ProductGrid(data, lang).Render(c.Request.Context(), c.Writer); err != nil {
			log.Printf("error rendering ProductGrid: %v", err)
		}
		return
	}

	c.Header("Content-Type", "text/html")
	if err := views.ProductsPage(userNameStr, data, lang).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering ProductsPage: %v", err)
	}
}

func (s *Server) productsFragmentHandler(c *gin.Context) {
	ctx := c.Request.Context()

	limitStr := c.DefaultQuery("limit", strconv.Itoa(s.cfg.PageSize))
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		log.Printf("error parsing limit: %v", err)
		limit = s.cfg.PageSize
	}
	if limit < 1 || limit > 100 {
		limit = s.cfg.PageSize
	}

	products, err := s.db.GetProductsBySourceAndCategory(ctx, database.GetProductsBySourceAndCategoryParams{
		Column1: c.Query("source"),
		Column2: c.Query("category"),
		Limit:   int32(limit),
		Offset:  0,
	})
	if err != nil {
		log.Printf("error getting products fragment: %v", err)
		return
	}

	c.Header("Content-Type", "text/html")
	for _, p := range products {
		if err := views.ProductCard(p).Render(ctx, c.Writer); err != nil {
			log.Printf("error rendering ProductCard: %v", err)
		}
	}
}
