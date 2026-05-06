package server

import (
	"GoApp/internal/database"
	"GoApp/internal/paging"
	"GoApp/internal/views"
	"log"
	"net/http"
	"net/url"
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

	searchQuery := c.DefaultQuery("q", "")

	var products []database.GetProductSummariesRow
	var total int64

	if searchQuery != "" {
		results, err := s.db.SearchProductSummaries(ctx, database.SearchProductSummariesParams{
			Query:       searchQuery,
			Source:      selectedSource,
			Category:    selectedCategory,
			Subcategory: selectedSubcategory,
			MinPrice:    minPriceParam,
			MaxPrice:    maxPriceParam,
			PageLimit:   int32(pageable.Size),
			PageOffset:  int32(pageable.Offset()),
		})
		if err != nil {
			log.Printf("error searching products: %v", err)
		}
		for _, r := range results {
			products = append(products, database.GetProductSummariesRow{
				ID:              r.ID,
				Source:          r.Source,
				Name:            r.Name,
				Brand:           r.Brand,
				Category:        r.Category,
				Price:           r.Price,
				OriginalPrice:   r.OriginalPrice,
				DiscountPercent: r.DiscountPercent,
				Currency:        r.Currency,
				InStock:         r.InStock,
				Quantity:        r.Quantity,
				Rating:          r.Rating,
				ReviewCount:     r.ReviewCount,
				Images:          r.Images,
			})
		}
		total, err = s.db.CountSearchProducts(ctx, database.CountSearchProductsParams{
			Query:       searchQuery,
			Source:      selectedSource,
			Category:    selectedCategory,
			Subcategory: selectedSubcategory,
			MinPrice:    minPriceParam,
			MaxPrice:    maxPriceParam,
		})
		if err != nil {
			log.Printf("error counting search products: %v", err)
		}
	} else {
		total, err = s.db.CountProductsBySourceAndCategory(ctx, database.CountProductsBySourceAndCategoryParams{
			Source:      selectedSource,
			Category:    selectedCategory,
			Subcategory: selectedSubcategory,
			MinPrice:    minPriceParam,
			MaxPrice:    maxPriceParam,
		})
		if err != nil {
			log.Printf("error counting products: %v", err)
		}
		products, err = s.db.GetProductSummaries(ctx, database.GetProductSummariesParams{
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
		SearchQuery:         searchQuery,
		Page:                page.Pageable.Page,
		TotalPages:          page.TotalPages,
		TotalCount:          page.TotalElements,
	}

	userName, _ := c.Get("userName")
	userNameStr, _ := userName.(string)
	lang := getLangStr(c)

	if c.GetHeader("HX-Request") == "true" {
		c.Header("Content-Type", "text/html")
		trigger := c.Query("trigger")

		// Always needed
		if err := views.ActiveChipsOOB(data, lang).Render(ctx, c.Writer); err != nil {
			log.Printf("error rendering ActiveChipsOOB: %v", err)
		}
		if err := views.FilterHeaderOOB(data, lang).Render(ctx, c.Writer); err != nil {
			log.Printf("error rendering FilterHeaderOOB: %v", err)
		}
		if err := views.PriceFilterOOB(data, lang).Render(ctx, c.Writer); err != nil {
			log.Printf("error rendering PriceFilterOOB: %v", err)
		}
		if err := views.ProductGrid(data, lang).Render(ctx, c.Writer); err != nil {
			log.Printf("error rendering ProductGrid: %v", err)
		}

		if trigger == "source" {
			if err := views.SourceFilterOOB(sources, selectedSource, lang).Render(ctx, c.Writer); err != nil {
				log.Printf("error rendering SourceFilterOOB: %v", err)
			}
			if err := views.CategoryFilterOOB(categories, selectedSource, selectedCategory, lang).Render(ctx, c.Writer); err != nil {
				log.Printf("error rendering CategoryFilterOOB: %v", err)
			}
			if err := views.SubcategoryFilterOOB(nil, selectedSource, "", "", lang).Render(ctx, c.Writer); err != nil {
				log.Printf("error rendering SubcategoryFilterOOB: %v", err)
			}
		}

		if trigger == "category" {
			if err := views.CategoryFilterOOB(categories, selectedSource, selectedCategory, lang).Render(ctx, c.Writer); err != nil {
				log.Printf("error rendering CategoryFilterOOB: %v", err)
			}
			if err := views.SubcategoryFilterOOB(subcategories, selectedSource, selectedCategory, selectedSubcategory, lang).Render(ctx, c.Writer); err != nil {
				log.Printf("error rendering SubcategoryFilterOOB: %v", err)
			}
		}

		if trigger == "subcategory" {
			if err := views.SubcategoryFilterOOB(subcategories, selectedSource, selectedCategory, selectedSubcategory, lang).Render(ctx, c.Writer); err != nil {
				log.Printf("error rendering SubcategoryFilterOOB: %v", err)
			}
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
	fragmentData := views.ProductsPageData{
		SelectedSource:      c.Query("source"),
		SelectedCategory:    c.Query("category"),
		SelectedSubcategory: c.Query("subcategory"),
	}
	for _, p := range products {
		if err := views.ProductCard(p, fragmentData).Render(ctx, c.Writer); err != nil {
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

	// build back URL
	backURL := "/products"
	params := url.Values{}
	if src := c.Query("source"); src != "" {
		params.Set("source", src)
	}
	if cat := c.Query("category"); cat != "" {
		params.Set("category", cat)
	}
	if sub := c.Query("subcategory"); sub != "" {
		params.Set("subcategory", sub)
	}
	if len(params) > 0 {
		backURL = "/products?" + params.Encode()
	}

	userName, _ := c.Get("userName")
	userNameStr, _ := userName.(string)
	lang := getLangStr(c)

	c.Header("Content-Type", "text/html")
	if err := views.ProductDetailPage(userNameStr, product, backURL, lang).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering ProductDetailPage: %v", err)
	}
}

func (s *Server) productPriceHistoryHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	history, err := s.db.GetProductPriceHistory(c.Request.Context(), id)
	if err != nil {
		log.Printf("error getting price history for product %s: %v", id, err)
		c.Status(http.StatusInternalServerError)
		return
	}

	lang := getLangStr(c)

	c.Header("Content-Type", "text/html")
	if err := views.PriceHistoryPanel(history, lang).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering PriceHistoryPanel: %v", err)
	}
}

func (s *Server) productSuggestHandler(c *gin.Context) {
	q := c.DefaultQuery("q", "")
	if len([]rune(q)) < 2 {
		c.Header("Content-Type", "text/html")
		if err := views.SearchSuggestions(nil, getLangStr(c)).Render(c.Request.Context(), c.Writer); err != nil {
			log.Printf("error rendering SearchSuggestions: %v", err)
		}
		return
	}

	suggestions, err := s.db.SuggestProductNames(c.Request.Context(), database.SuggestProductNamesParams{
		Query:     q,
		PageLimit: 8,
	})
	if err != nil {
		log.Printf("error getting suggestions: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Header("Content-Type", "text/html")
	if err := views.SearchSuggestions(suggestions, getLangStr(c)).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering SearchSuggestions: %v", err)
	}
}
