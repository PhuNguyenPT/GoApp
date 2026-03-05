package server

import (
	"GoApp/internal/database"
	"GoApp/internal/views"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Server) productsPageHandler(c *gin.Context) {
	ctx := c.Request.Context()

	selectedSource := c.Query("source")
	selectedCategory := c.Query("category")
	selectedSubcategory := c.Query("subcategory")

	if c.GetHeader("HX-Request") == "true" && c.Query("trigger") == "source" {
		selectedCategory = ""
		selectedSubcategory = ""
	}
	if c.GetHeader("HX-Request") == "true" && c.Query("trigger") == "category" {
		selectedSubcategory = ""
	}

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

	subcategories, err := s.db.GetSubcategoriesBySourceAndCategory(ctx, database.GetSubcategoriesBySourceAndCategoryParams{
		Source:   selectedSource,
		Category: selectedCategory,
	})
	if err != nil {
		log.Printf("error getting subcategories: %v", err)
	}

	total, err := s.db.CountProductsBySourceAndCategory(ctx, database.CountProductsBySourceAndCategoryParams{
		Source:      selectedSource,
		Category:    selectedCategory,
		Subcategory: selectedSubcategory,
	})
	if err != nil {
		log.Printf("error counting products: %v", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(s.cfg.PageSize)))
	offset := (pageNumber - 1) * s.cfg.PageSize

	products, err := s.db.GetProductSummaries(ctx, database.GetProductSummariesParams{
		Source:      selectedSource,
		Category:    selectedCategory,
		Subcategory: selectedSubcategory,
		PageLimit:   int32(s.cfg.PageSize),
		PageOffset:  int32(offset),
	})
	if err != nil {
		log.Printf("error getting products: %v", err)
	}

	data := views.ProductsPageData{
		Sources:             sources,
		Categories:          categories,
		Subcategories:       subcategories,
		Products:            products,
		SelectedSource:      selectedSource,
		SelectedCategory:    selectedCategory,
		SelectedSubcategory: selectedSubcategory,
		Page:                pageNumber,
		TotalPages:          totalPages,
		TotalCount:          total,
	}

	userName, _ := c.Get("userName")
	userNameStr, _ := userName.(string)
	lang := getLangStr(c)

	if c.GetHeader("HX-Request") == "true" {
		c.Header("Content-Type", "text/html")

		if err := views.SourceFilterOOB(sources, selectedSource, lang).Render(ctx, c.Writer); err != nil {
			log.Printf("error rendering SourceFilterOOB: %v", err)
		}
		if err := views.ActiveChipsOOB(data, lang).Render(ctx, c.Writer); err != nil {
			log.Printf("error rendering ActiveChipsOOB: %v", err)
		}
		if c.Query("trigger") == "source" {
			if err := views.CategoryFilterOOB(categories, selectedSource, selectedCategory, lang).Render(ctx, c.Writer); err != nil {
				log.Printf("error rendering CategoryFilterOOB: %v", err)
			}
			if err := views.SubcategoryFilterOOB(nil, selectedSource, "", "", lang).Render(ctx, c.Writer); err != nil {
				log.Printf("error rendering SubcategoryFilterOOB: %v", err)
			}
		}
		if c.Query("trigger") == "category" {
			if err := views.SubcategoryFilterOOB(subcategories, selectedSource, selectedCategory, selectedSubcategory, lang).Render(ctx, c.Writer); err != nil {
				log.Printf("error rendering SubcategoryFilterOOB: %v", err)
			}
		}
		if err := views.ProductGrid(data, lang).Render(ctx, c.Writer); err != nil {
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

	products, err := s.db.GetProductSummaries(ctx, database.GetProductSummariesParams{
		Source:      c.Query("source"),
		Category:    c.Query("category"),
		Subcategory: c.Query("subcategory"),
		PageLimit:   int32(limit),
		PageOffset:  0,
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

func (s *Server) productDetailPageHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.Redirect(http.StatusFound, "/products")
		return
	}

	product, err := s.db.GetProductByID(c.Request.Context(), id)
	if err != nil {
		c.Redirect(http.StatusFound, "/products")
		return
	}

	userName, _ := c.Get("userName")
	userNameStr, _ := userName.(string)
	lang := getLangStr(c)

	c.Header("Content-Type", "text/html")
	if err := views.ProductDetailPage(userNameStr, product, lang).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering ProductDetailPage: %v", err)
	}
}
