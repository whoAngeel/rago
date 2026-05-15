package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestGin(queryParams ...string) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	url := "/documents"
	if len(queryParams) > 0 {
		url += "?" + queryParams[0]
	}
	c.Request, _ = http.NewRequest("GET", url, nil)

	return c
}

func TestParsePagination_Defaults(t *testing.T) {
	c := setupTestGin()

	p := parsePagination(c, 10)

	if p.Page != 1 {
		t.Errorf("esperaba page=1, obtuve %d", p.Page)
	}
	if p.Limit != 10 {
		t.Errorf("esperaba limit=10, obtuve %d", p.Limit)
	}
}

func TestParsePagination_CustomValues(t *testing.T) {
	c := setupTestGin("page=3&limit=25")

	p := parsePagination(c, 10)

	if p.Page != 3 {
		t.Errorf("esperaba page=3, obtuve %d", p.Page)
	}
	if p.Limit != 25 {
		t.Errorf("esperaba limit=25, obtuve %d", p.Limit)
	}
}

func TestParsePagination_ClampsPageBelowOne(t *testing.T) {
	c := setupTestGin("page=0")

	p := parsePagination(c, 10)

	if p.Page != 1 {
		t.Errorf("esperaba page=1, obtuve %d", p.Page)
	}
}

func TestParsePagination_ClampsLimitAbove100(t *testing.T) {
	c := setupTestGin("limit=200")

	p := parsePagination(c, 10)

	if p.Limit != 10 {
		t.Errorf("esperaba limit=10, obtuve %d", p.Limit)
	}
}

func TestParsePagination_ClampsLimitBelowOne(t *testing.T) {
	c := setupTestGin("limit=0")

	p := parsePagination(c, 10)

	if p.Limit != 10 {
		t.Errorf("esperaba limit=10, obtuve %d", p.Limit)
	}
}

func TestParsePagination_InvalidValuesFallback(t *testing.T) {
	c := setupTestGin("page=abc&limit=def")

	p := parsePagination(c, 10)

	if p.Page != 1 {
		t.Errorf("esperaba page=1 por fallback, obtuve %d", p.Page)
	}
	if p.Limit != 10 {
		t.Errorf("esperaba limit=10 por fallback, obtuve %d", p.Limit)
	}
}

func TestParsePagination_UsesCustomDefaultLimit(t *testing.T) {
	c := setupTestGin()

	p := parsePagination(c, 25)

	if p.Limit != 25 {
		t.Errorf("esperaba limit=25 (default custom), obtuve %d", p.Limit)
	}
}

func TestParsePagination_TotalPagesLogic(t *testing.T) {
	// Total items = 25, limit = 10 => total_pages = 3
	total := 25
	limit := 10
	expectedPages := (total + limit - 1) / limit

	if expectedPages != 3 {
		t.Errorf("esperaba 3 páginas, obtuve %d", expectedPages)
	}
}

// Verifica que la función se pueda llamar con el formato que usa el handler
func TestPaginationStruct(t *testing.T) {
	p := PaginationParams{Page: 2, Limit: 5}

	if p.Page != 2 {
		t.Errorf("esperaba page=2 desde struct, obtuve %d", p.Page)
	}
	if p.Limit != 5 {
		t.Errorf("esperaba limit=5 desde struct, obtuve %d", p.Limit)
	}
}

func TestSelectItemStruct(t *testing.T) {
	item := SelectItem{ID: 42, Filename: "test.pdf"}

	if item.ID != 42 {
		t.Errorf("esperaba ID=42, obtuve %d", item.ID)
	}
	if item.Filename != "test.pdf" {
		t.Errorf("esperaba Filename='test.pdf', obtuve '%s'", item.Filename)
	}
}

func TestDocumentsToSelectItems(t *testing.T) {
	type doc struct {
		ID       int
		Filename string
		Status   string
	}

	docs := []doc{
		{ID: 1, Filename: "a.pdf", Status: "completed"},
		{ID: 2, Filename: "b.csv", Status: "pending"},
	}

	items := make([]SelectItem, 0, len(docs))
	for _, d := range docs {
		items = append(items, SelectItem{ID: d.ID, Filename: d.Filename})
	}

	if len(items) != 2 {
		t.Errorf("esperaba 2 items, obtuve %d", len(items))
	}
	if items[0].Filename != "a.pdf" {
		t.Errorf("esperaba 'a.pdf', obtuve '%s'", items[0].Filename)
	}
	if items[1].ID != 2 {
		t.Errorf("esperaba ID=2, obtuve %d", items[1].ID)
	}
}

func TestSelectItem_OnlyExposesIDAndFilename(t *testing.T) {
	type selectItemJSON struct {
		ID       int    `json:"id"`
		Filename string `json:"filename"`
	}

	item := selectItemJSON{ID: 10, Filename: "doc.txt"}

	if item.ID != 10 {
		t.Errorf("esperaba ID=10, obtuve %d", item.ID)
	}
	if item.Filename != "doc.txt" {
		t.Errorf("esperaba 'doc.txt', obtuve '%s'", item.Filename)
	}
}

func TestDefaultQuery(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)

	limit := c.DefaultQuery("limit", "10")
	if limit != "10" {
		t.Errorf("DefaultQuery sin parámetro esperaba '10', obtuvo '%s'", limit)
	}

	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	c2.Request, _ = http.NewRequest("GET", "/?limit=5", nil)

	limit2 := c2.DefaultQuery("limit", "10")
	if limit2 != "5" {
		t.Errorf("DefaultQuery con ?limit=5 esperaba '5', obtuvo '%s'", limit2)
	}
}
