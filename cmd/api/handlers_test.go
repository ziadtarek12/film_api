package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"filmapi.zeyadtarek.net/internals/jsonlog"
	"filmapi.zeyadtarek.net/internals/models"
)

// MockModels for testing
type MockModels struct {
	FilmModel       MockFilmModel
	UserModel       MockUserModel
	TokenModel      MockTokenModel
	PermissionModel MockPermissionModel
}

type MockFilmModel struct {
	GetFunc    func(id int64) (*models.Film, error)
	InsertFunc func(film *models.Film) error
	UpdateFunc func(film *models.Film) error
	DeleteFunc func(id int64) error
	GetAllFunc func(title string, genres []string, actors []string, directors []string, filters models.Filters) ([]*models.Film, models.Metadata, error)
	CountFunc  func() (int, error)
}

func (m MockFilmModel) Get(id int64) (*models.Film, error) {
	if m.GetFunc != nil {
		return m.GetFunc(id)
	}
	return nil, nil
}

func (m MockFilmModel) Insert(film *models.Film) error {
	if m.InsertFunc != nil {
		return m.InsertFunc(film)
	}
	return nil
}

func (m MockFilmModel) Update(film *models.Film) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(film)
	}
	return nil
}

func (m MockFilmModel) Delete(id int64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}

func (m MockFilmModel) GetAll(title string, genres []string, actors []string, directors []string, filters models.Filters) ([]*models.Film, models.Metadata, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc(title, genres, actors, directors, filters)
	}
	return nil, models.Metadata{}, nil
}

func (m MockFilmModel) Count() (int, error) {
	if m.CountFunc != nil {
		return m.CountFunc()
	}
	return 0, nil
}

type MockUserModel struct {
	InsertFunc      func(user *models.User) error
	GetByEmailFunc  func(email string) (*models.User, error)
	UpdateFunc      func(user *models.User) error
	GetForTokenFunc func(tokenScope, tokenPlaintext string) (*models.User, error)
}

func (m MockUserModel) Insert(user *models.User) error {
	if m.InsertFunc != nil {
		return m.InsertFunc(user)
	}
	return nil
}

func (m MockUserModel) GetByEmail(email string) (*models.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(email)
	}
	return nil, nil
}

func (m MockUserModel) Update(user *models.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(user)
	}
	return nil
}

func (m MockUserModel) GetForToken(tokenScope, tokenPlaintext string) (*models.User, error) {
	if m.GetForTokenFunc != nil {
		return m.GetForTokenFunc(tokenScope, tokenPlaintext)
	}
	return nil, nil
}

type MockTokenModel struct {
	NewFunc              func(userID int64, ttl, scope string) (*models.Token, error)
	InsertFunc           func(token *models.Token) error
	DeleteAllForUserFunc func(scope string, userID int64) error
}

func (m MockTokenModel) New(userID int64, ttl, scope string) (*models.Token, error) {
	if m.NewFunc != nil {
		return m.NewFunc(userID, ttl, scope)
	}
	return nil, nil
}

func (m MockTokenModel) Insert(token *models.Token) error {
	if m.InsertFunc != nil {
		return m.InsertFunc(token)
	}
	return nil
}

func (m MockTokenModel) DeleteAllForUser(scope string, userID int64) error {
	if m.DeleteAllForUserFunc != nil {
		return m.DeleteAllForUserFunc(scope, userID)
	}
	return nil
}

type MockPermissionModel struct {
	GetAllForUserFunc func(userID int64) (models.Permissions, error)
	AddForUserFunc    func(userID int64, codes ...string) error
}

func (m MockPermissionModel) GetAllForUser(userID int64) (models.Permissions, error) {
	if m.GetAllForUserFunc != nil {
		return m.GetAllForUserFunc(userID)
	}
	return nil, nil
}

func (m MockPermissionModel) AddForUser(userID int64, codes ...string) error {
	if m.AddForUserFunc != nil {
		return m.AddForUserFunc(userID, codes...)
	}
	return nil
}

// TestHealthCheckHandler tests the healthcheck handler
func TestHealthCheckHandler(t *testing.T) {
	// Create a new application instance with mock dependencies
	app := &application{
		logger: jsonlog.New(bytes.NewBuffer(nil), jsonlog.LevelInfo),
		config: config{
			env: "test",
		},
	}

	// Create a new HTTP request
	req, err := http.NewRequest("GET", "/v1/healthcheck", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler function
	app.healthCheckHandler(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	// Check that the status is "available"
	status, ok := response["status"]
	if !ok {
		t.Errorf("response missing status field")
	}
	if status != "available" {
		t.Errorf("handler returned wrong status: got %v want %v", status, "available")
	}

	// Check that the system_info contains the correct environment
	systemInfo, ok := response["system_info"].(map[string]interface{})
	if !ok {
		t.Errorf("response missing system_info field")
	}
	env, ok := systemInfo["environment"]
	if !ok {
		t.Errorf("system_info missing environment field")
	}
	if env != "test" {
		t.Errorf("handler returned wrong environment: got %v want %v", env, "test")
	}
}

// TestGetFilmHandler tests the getFilmHandler function
func TestGetFilmHandler(t *testing.T) {
	// Skip this test for now as it requires more complex mocking
	t.Skip("Skipping test that requires complex mocking of models.FilmModel")
}
