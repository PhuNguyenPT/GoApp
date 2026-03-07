package server

import (
	"GoApp/internal/database"
	"GoApp/internal/paging"
	"GoApp/internal/views"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductSortField string

const (
	ProductSortDefault   ProductSortField = ""
	ProductSortPrice     ProductSortField = "price"
	ProductSortCrawledAt ProductSortField = "crawled_at"
	ProductSortRating    ProductSortField = "rating"
	ProductSortName      ProductSortField = "name"
)

var allowedProductSorts = map[ProductSortField]string{
	ProductSortPrice:     "price",
	ProductSortCrawledAt: "crawled_at",
	ProductSortRating:    "rating",
	ProductSortName:      "name",
}

func (s *Server) productsPageHandler(c *gin.Context) {
	ctx := c.Request.Context()

	selectedSource := c.Query("source")
	selectedCategory := c.Query("category")
	selectedSubcategory := c.Query("subcategory")
	log.Printf("selected: source=%q category=%q subcategory=%q", selectedSource, selectedCategory, selectedSubcategory)

	minPriceStr := c.DefaultQuery("min_price", "0")
	maxPriceStr := c.DefaultQuery("max_price", "0")
	minPrice, err := strconv.ParseFloat(minPriceStr, 64)
	if err != nil || minPrice < 0 {
		minPrice = 0
	}
	maxPrice, err := strconv.ParseFloat(maxPriceStr, 64)
	if err != nil || maxPrice < 0 {
		maxPrice = 0
	}
	minPriceParam := strconv.FormatFloat(minPrice, 'f', -1, 64)
	maxPriceParam := strconv.FormatFloat(maxPrice, 'f', -1, 64)

	if c.GetHeader("HX-Request") == "true" && c.Query("trigger") == "source" {
		selectedCategory = ""
		selectedSubcategory = ""
	}
	if c.GetHeader("HX-Request") == "true" && c.Query("trigger") == "category" {
		selectedSubcategory = ""
	}

	pageable := resolvePageable[ProductSortField](c, s.cfg.PageSize, s.cfg.MaxPageSize)
	sortExpr, ok := allowedProductSorts[pageable.Sort.Field]
	selectedSort := ""
	if ok {
		selectedSort = string(pageable.Sort.Field) + "," + string(pageable.Sort.Direction)
	} else {
		sortExpr = "crawled_at"
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
		MinPrice:    minPriceParam,
		MaxPrice:    maxPriceParam,
	})
	if err != nil {
		log.Printf("error counting products: %v", err)
	}

	products, err := s.db.GetProductSummaries(ctx, database.GetProductSummariesParams{
		Source:      selectedSource,
		Category:    selectedCategory,
		Subcategory: selectedSubcategory,
		MinPrice:    minPriceParam,
		MaxPrice:    maxPriceParam,
		SortField:   sortExpr,
		SortDir:     string(pageable.Sort.Direction),
		PageLimit:   int32(pageable.Size),
		PageOffset:  int32(pageable.Offset()),
	})
	if err != nil {
		log.Printf("error getting products: %v", err)
	}

	page := paging.NewPage(products, total, pageable)

	percentiles, err := s.db.GetPricePercentiles(ctx, database.GetPricePercentilesParams{
		Source:      selectedSource,
		Category:    selectedCategory,
		Subcategory: selectedSubcategory,
	})
	if err != nil {
		log.Printf("error getting price percentiles: %v", err)
	}

	var globalMinPrice, globalMaxPrice, p25, p50, p75 float64
	if err == nil {
		globalMinPrice, _ = strconv.ParseFloat(percentiles.MinPrice, 64)
		globalMaxPrice, _ = strconv.ParseFloat(percentiles.MaxPrice, 64)
		p25, _ = strconv.ParseFloat(percentiles.P25, 64)
		p50, _ = strconv.ParseFloat(percentiles.P50, 64)
		p75, _ = strconv.ParseFloat(percentiles.P75, 64)
	}
	currency := "VND"
	if err == nil && percentiles.Currency != "" {
		currency = percentiles.Currency
	}

	data := views.ProductsPageData{
		Sources:             sources,
		Categories:          categories,
		Subcategories:       subcategories,
		MinPrice:            minPrice,
		MaxPrice:            maxPrice,
		PricePercentiles:    [5]float64{globalMinPrice, p25, p50, p75, globalMaxPrice},
		Currency:            currency,
		Products:            page.Content,
		SelectedSource:      selectedSource,
		SelectedCategory:    selectedCategory,
		SelectedSubcategory: selectedSubcategory,
		SelectedSort:        selectedSort,
		Page:                page.Pageable.Page,
		TotalPages:          page.TotalPages,
		TotalCount:          page.TotalElements,
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
		if err := views.CategoryFilterOOB(categories, selectedSource, selectedCategory, lang).Render(ctx, c.Writer); err != nil {
			log.Printf("error rendering CategoryFilterOOB: %v", err)
		}
		if err := views.PriceFilterOOB(data, lang).Render(ctx, c.Writer); err != nil {
			log.Printf("error rendering PriceFilterOOB: %v", err)
		}
		if err := views.FilterHeaderOOB(data, lang).Render(ctx, c.Writer); err != nil {
			log.Printf("error rendering FilterHeaderOOB: %v", err)
		}
		if c.Query("trigger") == "source" {
			if err := views.SubcategoryFilterOOB(nil, selectedSource, "", "", lang).Render(ctx, c.Writer); err != nil {
				log.Printf("error rendering SubcategoryFilterOOB: %v", err)
			}
		} else {
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
		MinPrice:    "0",
		MaxPrice:    "0",
		SortField:   "crawled_at",
		SortDir:     string(paging.Desc),
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
