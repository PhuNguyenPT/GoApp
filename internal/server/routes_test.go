package server

import (
	"GoApp/internal/database"
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
	"golang.org/x/crypto/bcrypt"
)

type mockDB struct {
	deleteExpiredSessionsCalled int
}

func (m *mockDB) Health() map[string]string {
	return map[string]string{"status": "up", "message": "It's healthy"}
}

func (m *mockDB) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return database.User{}, err
	}

	return database.User{ID: id, Name: arg.Name, Email: arg.Email, PasswordHash: arg.PasswordHash, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (m *mockDB) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return database.User{}, err
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	return database.User{
		ID:           id,
		Name:         "Test User",
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

func (m *mockDB) GetUserByID(ctx context.Context, id uuid.UUID) (database.User, error) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	return database.User{
		ID:           id,
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

func (m *mockDB) UpdateUserName(ctx context.Context, arg database.UpdateUserNameParams) (database.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return database.User{}, err
	}
	return database.User{
		ID:           arg.ID,
		Name:         arg.Name,
		Email:        "test@example.com",
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

func (m *mockDB) UpdateUserPassword(ctx context.Context, arg database.UpdateUserPasswordParams) error {
	return nil
}

func (m *mockDB) DeleteUser(ctx context.Context, id uuid.UUID) error { return nil }

func (m *mockDB) CreateSession(ctx context.Context, arg database.CreateSessionParams) (database.Session, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return database.Session{}, err
	}
	return database.Session{ID: id, UserID: arg.UserID, Token: arg.Token, CreatedAt: time.Now(), ExpiresAt: time.Now().Add(1 * time.Hour.Abs())}, nil
}

func (m *mockDB) GetSessionByToken(ctx context.Context, token string) (database.Session, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return database.Session{}, err
	}
	userID, err := uuid.NewV7()
	if err != nil {
		return database.Session{}, err
	}
	return database.Session{ID: id, UserID: userID, Token: token, CreatedAt: time.Now(), ExpiresAt: time.Now().Add(1 * time.Hour.Abs())}, nil
}

func (m *mockDB) DeleteSession(ctx context.Context, token string) error {
	return nil
}

func (m *mockDB) DeleteExpiredSessions(ctx context.Context) error {
	m.deleteExpiredSessionsCalled++
	return nil
}

func (m *mockDB) GetActiveSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]database.Session, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return []database.Session{}, err
	}
	token, err := uuid.NewV7()
	if err != nil {
		return []database.Session{}, err
	}
	return []database.Session{
		{ID: id, UserID: userID, Token: token.String(), UserAgent: "Mozilla/5.0 Test Browser", CreatedAt: time.Now(), ExpiresAt: time.Now().Add(1 * time.Hour.Abs()), IpAddress: "127.0.0.1"},
	}, nil
}

func (m *mockDB) DeleteSessionByID(ctx context.Context, arg database.DeleteSessionByIDParams) error {
	return nil
}

func (m *mockDB) CreateContact(ctx context.Context, arg database.CreateContactParams) (database.Contact, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return database.Contact{}, err
	}
	return database.Contact{
		ID:        id,
		Name:      arg.Name,
		Email:     arg.Email,
		Subject:   arg.Subject,
		Message:   arg.Message,
		IpAddress: "127.0.0.1",
		CreatedAt: time.Now(),
	}, nil
}

func (m *mockDB) CountContactsByIPToday(ctx context.Context, ipAddress string) (int64, error) {
	return 0, nil
}

func (m *mockDB) CountContactsByEmailToday(ctx context.Context, email string) (int64, error) {
	return 0, nil
}

func (m *mockDB) GetDistinctSources(ctx context.Context) ([]sql.NullString, error) {
	allProducts, err := m.GetProductsBySourceAndCategory(ctx, database.GetProductsBySourceAndCategoryParams{})
	if err != nil {
		return []sql.NullString{}, err
	}
	unique := make(map[string]sql.NullString)
	for _, p := range allProducts {
		if p.Source.Valid {
			unique[p.Source.String] = p.Source
		}
	}

	var result []sql.NullString
	for _, s := range unique {
		result = append(result, s)
	}
	return result, nil
}

func (m *mockDB) GetCategoriesBySource(ctx context.Context, source string) ([]sql.NullString, error) {
	products, err := m.GetProductsBySourceAndCategory(ctx, database.GetProductsBySourceAndCategoryParams{
		Source: source,
	})
	if err != nil {
		return []sql.NullString{}, err
	}
	unique := make(map[string]sql.NullString)
	for _, p := range products {
		if p.Category.Valid {
			unique[p.Category.String] = p.Category
		}
	}
	var result []sql.NullString
	for _, c := range unique {
		result = append(result, c)
	}
	return result, nil
}

func (m *mockDB) GetSubcategoriesBySourceAndCategory(ctx context.Context, arg database.GetSubcategoriesBySourceAndCategoryParams) ([]sql.NullString, error) {
	products, err := m.GetProductsBySourceAndCategory(ctx, database.GetProductsBySourceAndCategoryParams{
		Source:   arg.Source,
		Category: arg.Category,
	})
	if err != nil {
		return []sql.NullString{}, err
	}
	unique := make(map[string]sql.NullString)
	for _, p := range products {
		if p.Subcategory.Valid {
			unique[p.Subcategory.String] = p.Subcategory
		}
	}
	var result []sql.NullString
	for _, s := range unique {
		result = append(result, s)
	}
	return result, nil
}

func (m *mockDB) GetProductsBySourceAndCategory(ctx context.Context, arg database.GetProductsBySourceAndCategoryParams) ([]database.Product, error) {
	ids := make(uuid.UUIDs, 4)
	for i := range ids {
		id, err := uuid.NewV7()
		if err != nil {
			return nil, err
		}
		ids[i] = id
	}

	fptPhone, err := json.Marshal(map[string]string{
		"Kích thước màn hình": "6.3 inch",
		"Thời lượng pin":      "31 Giờ",
		"Kết nối NFC":         "NFC",
	})
	if err != nil {
		return []database.Product{}, err
	}
	tgddPhone, err := json.Marshal(map[string]string{
		"Hệ điều hành":       "iOS 26",
		"Chip xử lý (CPU)":   "Apple A19 Pro 6 nhân",
		"RAM":                "12 GB",
		"Dung lượng lưu trữ": "256 GB",
		"Công nghệ màn hình": "OLED",
		"Màn hình rộng":      "6.9\"",
		"Dung lượng pin":     "37 giờ",
		"Mạng di động":       "Hỗ trợ 5G",
		"Bluetooth":          "v6.0",
		"Cổng kết nối/sạc":   "Type-C",
		"Kết nối khác":       "NFC",
		"Kháng nước, bụi":    "IP68",
	})
	if err != nil {
		return []database.Product{}, err
	}
	fptLaptop, err := json.Marshal(map[string]string{
		"CPU":                 "Core Ultra 9",
		"Card đồ hoạ":         "NVIDIA GeForce RTX 5090 24GB GDDR7 (1824 TOPS)",
		"Kích thước màn hình": "18 inch",
		"Tấm nền":             "IPS",
	})
	if err != nil {
		return []database.Product{}, err
	}
	tgddLaptop, err := json.Marshal(map[string]string{
		"Công nghệ CPU":       "AMD Ryzen 5 - 6600H",
		"RAM":                 "16 GB",
		"Loại RAM":            "DDR5",
		"Ổ cứng":              "512 GB SSD NVMe PCIe",
		"Kích thước màn hình": "15.6\"",
		"Độ phân giải":        "Full HD (1920 x 1080)",
		"Tấm nền":             "IPS",
		"Tần số quét":         "165 Hz",
		"Card màn hình":       "NVIDIA GeForce RTX 2050, 4 GB",
		"Hệ điều hành":        "Windows 11 Home SL",
		"Thông tin Pin":       "4-cell, 57Wh",
	})
	if err != nil {
		return []database.Product{}, err
	}
	crawledAt := sql.NullTime{Time: time.Now(), Valid: true}

	all := []database.Product{
		{
			ID:              ids[0],
			Url:             "https://fptshop.com.vn/dien-thoai/iphone-17-pro",
			Source:          sql.NullString{String: "fptshop", Valid: true},
			Name:            sql.NullString{String: "iPhone 17 Pro", Valid: true},
			Brand:           sql.NullString{String: "Apple", Valid: true},
			Category:        sql.NullString{String: "Điện thoại", Valid: true},
			Subcategory:     sql.NullString{String: "Apple", Valid: true},
			Price:           sql.NullString{String: "33590000.0", Valid: true},
			OriginalPrice:   sql.NullString{String: "34990000.0", Valid: true},
			DiscountPercent: sql.NullInt32{Int32: 4, Valid: true},
			Currency:        sql.NullString{String: "VND", Valid: true},
			InStock:         sql.NullBool{Bool: true, Valid: true},
			Quantity:        sql.NullInt32{Valid: false},
			Rating:          sql.NullString{String: "4.9", Valid: true},
			ReviewCount:     sql.NullInt32{Int32: 25, Valid: true},
			Images:          pqtype.NullRawMessage{RawMessage: []byte(`["https://cdn2.fptshop.com.vn/unsafe/iphone_17_pro_cosmic_orange_1_12e8ea1358.png"]`), Valid: true},
			Specs:           pqtype.NullRawMessage{RawMessage: fptPhone, Valid: true},
			CrawledAt:       crawledAt,
			CreatedAt:       crawledAt,
			UpdatedAt:       crawledAt,
		},
		{
			ID:              ids[1],
			Url:             "https://www.thegioididong.com/dtdd/iphone-17-pro-max",
			Source:          sql.NullString{String: "thegioididong", Valid: true},
			Name:            sql.NullString{String: "Điện thoại iPhone 17 Pro Max 256GB", Valid: true},
			Brand:           sql.NullString{String: "Apple", Valid: true},
			Category:        sql.NullString{String: "Điện thoại", Valid: true},
			Subcategory:     sql.NullString{String: "Điện thoại iPhone (Apple)", Valid: true},
			Price:           sql.NullString{String: "37590000.0", Valid: true},
			OriginalPrice:   sql.NullString{Valid: false},
			DiscountPercent: sql.NullInt32{Valid: false},
			Currency:        sql.NullString{String: "VND", Valid: true},
			InStock:         sql.NullBool{Bool: true, Valid: true},
			Quantity:        sql.NullInt32{Valid: false},
			Rating:          sql.NullString{String: "4.9", Valid: true},
			ReviewCount:     sql.NullInt32{Valid: false},
			Images:          pqtype.NullRawMessage{RawMessage: []byte(`["https://cdn.tgdd.vn/Products/Images/42/342679/Slider/vi-vn-iphone-17-pro-max-1.jpg"]`), Valid: true},
			Specs:           pqtype.NullRawMessage{RawMessage: tgddPhone, Valid: true},
			CrawledAt:       crawledAt,
			CreatedAt:       crawledAt,
			UpdatedAt:       crawledAt,
		},
		{
			ID:              ids[2],
			Url:             "https://fptshop.com.vn/may-tinh-xach-tay/acer-predator-helios-18-gaming-ai-ph18-73-98aq-u9-275hx",
			Source:          sql.NullString{String: "fptshop", Valid: true},
			Name:            sql.NullString{String: "Acer Predator Helios 18 Gaming AI PH18-73-98AQ U9 275HX", Valid: true},
			Brand:           sql.NullString{String: "Acer", Valid: true},
			Category:        sql.NullString{String: "Laptop", Valid: true},
			Subcategory:     sql.NullString{String: "Acer", Valid: true},
			Price:           sql.NullString{String: "168990000.0", Valid: true},
			OriginalPrice:   sql.NullString{String: "171690000.0", Valid: true},
			DiscountPercent: sql.NullInt32{Int32: 2, Valid: true},
			Currency:        sql.NullString{String: "VND", Valid: true},
			InStock:         sql.NullBool{Bool: true, Valid: true},
			Quantity:        sql.NullInt32{Valid: false},
			Rating:          sql.NullString{String: "4.9", Valid: true},
			ReviewCount:     sql.NullInt32{Int32: 89, Valid: true},
			Images:          pqtype.NullRawMessage{RawMessage: []byte(`["https://cdn2.fptshop.com.vn/unsafe/acer_predator_helios_18_gaming_ai_ph18_73_den_1_e6a6d14282.jpg"]`), Valid: true},
			Specs:           pqtype.NullRawMessage{RawMessage: fptLaptop, Valid: true},
			CrawledAt:       crawledAt,
			CreatedAt:       crawledAt,
			UpdatedAt:       crawledAt,
		},
		{
			ID:              ids[3],
			Url:             "https://www.thegioididong.com/laptop/acer-nitro-v-15-anv15-41-r2up-r5-nhqpgsv004",
			Source:          sql.NullString{String: "thegioididong", Valid: true},
			Name:            sql.NullString{String: "Laptop Acer Gaming Nitro V 15 ANV15 41 R2UP - NH.QPGSV.004", Valid: true},
			Brand:           sql.NullString{String: "Acer", Valid: true},
			Category:        sql.NullString{String: "Laptop", Valid: true},
			Subcategory:     sql.NullString{String: "Laptop Acer", Valid: true},
			Price:           sql.NullString{String: "18690000.0", Valid: true},
			OriginalPrice:   sql.NullString{String: "20190000.0", Valid: true},
			DiscountPercent: sql.NullInt32{Int32: 7, Valid: true},
			Currency:        sql.NullString{String: "VND", Valid: true},
			InStock:         sql.NullBool{Bool: true, Valid: true},
			Quantity:        sql.NullInt32{Valid: false},
			Rating:          sql.NullString{String: "4.9", Valid: true},
			ReviewCount:     sql.NullInt32{Valid: false},
			Images: pqtype.NullRawMessage{
				RawMessage: []byte(`[
					"https://cdn.tgdd.vn/Products/Images/44/333430/Slider/bao-hanh-1020x570.png",
					"https://cdn.tgdd.vn/Products/Images/2102/214612/balo-predator-3-600x600.jpg"
				]`),
				Valid: true,
			},
			Specs:     pqtype.NullRawMessage{RawMessage: tgddLaptop, Valid: true},
			CrawledAt: crawledAt,
			CreatedAt: crawledAt,
			UpdatedAt: crawledAt,
		},
	}

	// filter by source and category
	src := arg.Source
	cat := arg.Category
	sub := arg.Subcategory

	var result []database.Product
	for _, p := range all {
		sourceMatch := src == "" || strings.EqualFold(p.Source.String, src)
		categoryMatch := cat == "" || strings.EqualFold(p.Category.String, cat)
		subcategoryMatch := sub == "" || strings.EqualFold(p.Subcategory.String, sub)
		if sourceMatch && categoryMatch && subcategoryMatch {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockDB) CountProductsBySourceAndCategory(ctx context.Context, arg database.CountProductsBySourceAndCategoryParams) (int64, error) {
	products, err := m.GetProductsBySourceAndCategory(ctx, database.GetProductsBySourceAndCategoryParams{
		Source:      arg.Source,
		Category:    arg.Category,
		Subcategory: arg.Subcategory,
	})
	if err != nil {
		return 0, err
	}
	return int64(len(products)), nil
}

func (m *mockDB) GetProductByID(ctx context.Context, id uuid.UUID) (database.Product, error) {
	fptPhone, err := json.Marshal(map[string]string{
		"Kích thước màn hình": "6.3 inch",
		"Thời lượng pin":      "31 Giờ",
		"Kết nối NFC":         "NFC",
	})
	if err != nil {
		return database.Product{}, err
	}
	crawledAt := sql.NullTime{Time: time.Now(), Valid: true}

	return database.Product{
		ID:              id,
		Url:             "https://fptshop.com.vn/dien-thoai/iphone-17-pro",
		Source:          sql.NullString{String: "fptshop", Valid: true},
		Name:            sql.NullString{String: "iPhone 17 Pro", Valid: true},
		Brand:           sql.NullString{String: "Apple", Valid: true},
		Category:        sql.NullString{String: "Điện thoại", Valid: true},
		Subcategory:     sql.NullString{String: "Apple", Valid: true},
		Price:           sql.NullString{String: "33590000.0", Valid: true},
		OriginalPrice:   sql.NullString{String: "34990000.0", Valid: true},
		DiscountPercent: sql.NullInt32{Int32: 4, Valid: true},
		Currency:        sql.NullString{String: "VND", Valid: true},
		InStock:         sql.NullBool{Bool: true, Valid: true},
		Quantity:        sql.NullInt32{Valid: false},
		Rating:          sql.NullString{String: "4.9", Valid: true},
		ReviewCount:     sql.NullInt32{Int32: 25, Valid: true},
		Images:          pqtype.NullRawMessage{RawMessage: []byte(`["https://cdn2.fptshop.com.vn/unsafe/iphone_17_pro_cosmic_orange_1_12e8ea1358.png"]`), Valid: true},
		Specs:           pqtype.NullRawMessage{RawMessage: fptPhone, Valid: true},
		CrawledAt:       crawledAt,
		UpdatedAt:       crawledAt,
		CreatedAt:       crawledAt,
	}, nil
}

func (m *mockDB) GetProductSummaries(ctx context.Context, arg database.GetProductSummariesParams) ([]database.GetProductSummariesRow, error) {
	products, err := m.GetProductsBySourceAndCategory(ctx, database.GetProductsBySourceAndCategoryParams(arg))
	if err != nil {
		return nil, err
	}

	rows := make([]database.GetProductSummariesRow, len(products))
	for i, p := range products {
		rows[i] = database.GetProductSummariesRow{
			ID:              p.ID,
			Source:          p.Source,
			Name:            p.Name,
			Brand:           p.Brand,
			Category:        p.Category,
			Price:           p.Price,
			OriginalPrice:   p.OriginalPrice,
			DiscountPercent: p.DiscountPercent,
			Currency:        p.Currency,
			InStock:         p.InStock,
			Quantity:        p.Quantity,
			Rating:          p.Rating,
			ReviewCount:     p.ReviewCount,
			Images:          p.Images,
		}
	}
	return rows, nil
}

var testHandler http.Handler

func TestMain(m *testing.M) {
	if err := os.Chdir("../../"); err != nil {
		log.Fatalf("failed to change directory: %v", err)
	}
	s := &Server{
		db:  &mockDB{},
		cfg: &Config{AppEnv: EnvTest, GinMode: gin.TestMode},
	}
	testHandler = s.RegisterRoutes(s.cfg)
	os.Exit(m.Run())
}

func TestHomePageHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}

	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("expected HTML content type, got %v", ct)
	}
}

func TestUnknownRoute(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/does-not-exist", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %v", rr.Code)
	}
}
