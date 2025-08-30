package gorest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)


type Dummy struct {
	ID    int    `json:"id" gorm:"primaryKey"`
	Name  string `json:"name" binding:"required"`
	Value int    `json:"value"`
	Skip  string `json:"-"`
}

func setupTestRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Set up GORM middleware
	r.Use(GormHandlerFunc(db))

	// Setup routes
	CRUD[Dummy](r, "/dummies", &Config{
		AddQueryParams: true,
	})

	return r
}

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&Dummy{})

	return db
}

func setupTestDummiesData(n int, db *gorm.DB) {
	var dummies []Dummy
	for i := 1; i <= n; i++ {
		name := "Dummy " + strconv.Itoa(i)
		dummies = append(dummies, Dummy{ID: i, Name: name, Value: i})
	}
	db.Create(&dummies)
}


func requestTest(r *gin.Engine, method string, url string, body any) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		reader = bytes.NewBuffer(jsonBytes)
	}

	req, _ := http.NewRequest(method, url, reader)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

func TestGetDummy(t *testing.T) {
	db := setupTestDB()
	r := setupTestRouter(db)
	count := 4

	// Insert a dummy record
	setupTestDummiesData(count, db)

	// Perform GET request
	w := requestTest(r, "GET", "/dummies", nil)

	// Test response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 but got %d", w.Code)
	}

	var dummies []Dummy
	if err := json.Unmarshal(w.Body.Bytes(), &dummies); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if len(dummies) != count || dummies[0].Name != "Dummy 1" {
		t.Fatalf("Unexpected response data: %v", dummies)
	}
}

func TestGetDummyByQueryParam(t *testing.T) {
	db := setupTestDB()
	r := setupTestRouter(db)
	count := 5

	// Insert dummy records
	setupTestDummiesData(count, db)

	// Perform GET request with query param
	w := requestTest(r, "GET", "/dummies?name=Dummy 4&value=4&id=4", nil)

	// Test response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 but got %d", w.Code)
	}
	var dummies []Dummy
	if err := json.Unmarshal(w.Body.Bytes(), &dummies); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if len(dummies) != 1 || dummies[0].Name != "Dummy 4" {
		t.Fatalf("Unexpected response data: %v", dummies)
	}
}

func TestGetDummyByQueryParamNotFound(t *testing.T) {
	db := setupTestDB()
	r := setupTestRouter(db)

	// Insert dummy records
	setupTestDummiesData(3, db)

	// Perform GET request with non-matching query param
	w := requestTest(r, "GET", "/dummies?name=Dummy 1&value=2", nil)

	// Test response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 but got %d", w.Code)
	}
	var dummies []Dummy
	if err := json.Unmarshal(w.Body.Bytes(), &dummies); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if len(dummies) != 0 {
		t.Fatalf("Expected empty result but got: %v", dummies)
	}
}

func TestGetDummyByID(t *testing.T) {
	db := setupTestDB()
	r := setupTestRouter(db)

	// Insert dummy records
	setupTestDummiesData(5, db)

	// Perform GET request
	w := requestTest(r, "GET", "/dummies/3", nil)

	// Test response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 but got %d", w.Code)
	}
	var dummy Dummy
	if err := json.Unmarshal(w.Body.Bytes(), &dummy); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if dummy.Name != "Dummy 3" {
		t.Fatalf("Unexpected response data: %v", dummy)
	}
}

func TestGetDummyByIDNotFound(t *testing.T) {
	db := setupTestDB()
	r := setupTestRouter(db)

	// Insert dummy records
	setupTestDummiesData(1, db)

	// Perform GET request for non-existent ID
	w := requestTest(r, "GET", "/dummies/999", nil)

	// Test response
	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected status 404 but got %d", w.Code)
	}
}

func TestCreateDummy(t *testing.T) {
	db := setupTestDB()
	r := setupTestRouter(db)
	count := 5

	// Insert dummy records
	setupTestDummiesData(count, db)

	// Prepare request body
	newDummy := map[string]string{"name": "New Dummy"}

	// Perform POST request
	w := requestTest(r, "POST", "/dummies", newDummy)

	// Test response
	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201 but got %d", w.Code)
	}
	var createdDummy Dummy
	if err := json.Unmarshal(w.Body.Bytes(), &createdDummy); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if createdDummy.ID != count + 1 || createdDummy.Name != "New Dummy" {
		t.Fatalf("Unexpected response data: %v", createdDummy)
	}

	// Verify record in database
	var obj Dummy
	if err := db.Last(&obj, createdDummy.ID).Error; err != nil {
		t.Fatalf("Failed to find created record in DB: %v", err)
	}
	if obj.Name != "New Dummy" {
		t.Fatalf("Database record mismatch: %v", obj)
	}
}

func TestCreateDummyInvalidData(t *testing.T) {
	db := setupTestDB()
	r := setupTestRouter(db)

	// Prepare invalid request body (missing 'name' field)
	invalidDummy := map[string]string{"invalid_field": "Invalid"}

	// Perform POST request
	w := requestTest(r, "POST", "/dummies", invalidDummy)

	// Test response
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 but got %d", w.Code)
	}
}

func TestUpdateDummy(t *testing.T) {
	db := setupTestDB()
	r := setupTestRouter(db)

	// Insert dummy records
	setupTestDummiesData(3, db)

	// Prepare update data
	updateData := map[string]string{"name": "Updated Dummy"}

	// Perform PUT request
	w := requestTest(r, "PUT", "/dummies/2", updateData)

	// Test response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 but got %d", w.Code)
	}
	var updatedDummy Dummy
	if err := json.Unmarshal(w.Body.Bytes(), &updatedDummy); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if updatedDummy.ID != 2 || updatedDummy.Name != "Updated Dummy" {
		t.Fatalf("Unexpected response data: %v", updatedDummy)
	}

	// Verify record in database
	var obj Dummy
	if err := db.First(&obj, 2).Error; err != nil {
		t.Fatalf("Failed to find updated record in DB: %v", err)
	}
	if obj.Name != "Updated Dummy" {
		t.Fatalf("Database record mismatch: %v", obj)
	}
}

func TestUpdateDummyNotFound(t *testing.T) {
	db := setupTestDB()
	r := setupTestRouter(db)

	// Prepare update data
	updateData := map[string]string{"name": "Updated Dummy"}

	// Perform PUT request for non-existent ID
	w := requestTest(r, "PUT", "/dummies/999", updateData)

	// Test response
	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected status 404 but got %d", w.Code)
	}
}

func TestDeleteDummy(t *testing.T) {
	db := setupTestDB()
	r := setupTestRouter(db)

	// Insert dummy records
	setupTestDummiesData(3, db)

	// Perform DELETE request
	w := requestTest(r, "DELETE", "/dummies/2", nil)

	// Test response
	if w.Code != http.StatusNoContent {
		t.Fatalf("Expected status 204 but got %d", w.Code)
	}

	// Verify record is deleted in database
	var obj Dummy
	if err := db.First(&obj, 2).Error; err == nil {
		t.Fatalf("Record was not deleted from DB: %v", obj)
	}
}

func TestDeleteDummyNotFound(t *testing.T) {
	db := setupTestDB()
	r := setupTestRouter(db)

	// Perform DELETE request for non-existent ID
	w := requestTest(r, "DELETE", "/dummies/999", nil)

	// Test response
	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected status 404 but got %d", w.Code)
	}
}
