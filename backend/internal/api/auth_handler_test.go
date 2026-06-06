package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jj/novelist/internal/auth"
	"github.com/jj/novelist/internal/model"
	"github.com/jj/novelist/internal/store"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDBWithUUID(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// Register callback to auto-generate UUIDs for primary keys
	_ = db.Callback().Create().Before("gorm:create").Register("assign_uuid", func(tx *gorm.DB) {
		if tx.Statement.Schema == nil {
			return
		}
		pkField := tx.Statement.Schema.PrioritizedPrimaryField
		if pkField == nil {
			return
		}
		field := tx.Statement.Schema.LookUpField(pkField.Name)
		if field == nil {
			return
		}
		val, _ := field.ValueOf(tx.Statement.Context, tx.Statement.ReflectValue)
		if id, ok := val.(uuid.UUID); ok && id == uuid.Nil {
			field.Set(tx.Statement.Context, tx.Statement.ReflectValue, uuid.New())
		}
	})

	if err := db.AutoMigrate(
		&model.User{},
		&model.Project{},
		&model.Chapter{},
		&model.Character{},
		&model.WorldSetting{},
		&model.Outline{},
		&model.Discussion{},
		&model.Conversation{},
		&model.Setting{},
	); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	store.DB = db
}

func init() {
	auth.SetSecret("test-secret-key-for-unit-tests")
}

func setupGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestRegister_Success(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupGin()
	r.POST("/register", Register)

	body := `{"username":"testuser","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Register() status = %d, want %d; body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if _, ok := resp["token"]; !ok {
		t.Error("Register() response missing token")
	}
	if user, ok := resp["user"].(map[string]interface{}); ok {
		if user["username"] != "testuser" {
			t.Errorf("Register() username = %v, want testuser", user["username"])
		}
	} else {
		t.Error("Register() response missing user")
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupGin()
	r.POST("/register", Register)

	body := `{"username":"testuser","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("first Register() status = %d, want %d", w.Code, http.StatusCreated)
	}

	// Try duplicate
	req2 := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("duplicate Register() status = %d, want %d; body: %s", w2.Code, http.StatusConflict, w2.Body.String())
	}
}

func TestRegister_InvalidInput(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupGin()
	r.POST("/register", Register)

	tests := []struct {
		name string
		body string
		want int
	}{
		{"empty username", `{"username":"","password":"password123"}`, http.StatusBadRequest},
		{"short username", `{"username":"ab","password":"password123"}`, http.StatusBadRequest},
		{"short password", `{"username":"testuser","password":"12345"}`, http.StatusBadRequest},
		{"empty body", `{}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.want {
				t.Errorf("Register(%s) status = %d, want %d; body: %s", tt.name, w.Code, tt.want, w.Body.String())
			}
		})
	}
}

func TestLogin_Success(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupGin()
	r.POST("/login", Login)

	// Create user directly
	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := model.User{Username: "loginuser", PasswordHash: string(hashed)}
	store.GetDB().Create(&user)

	body := `{"username":"loginuser","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Login() status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if _, ok := resp["token"]; !ok {
		t.Error("Login() response missing token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupGin()
	r.POST("/login", Login)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := model.User{Username: "loginuser", PasswordHash: string(hashed)}
	store.GetDB().Create(&user)

	body := `{"username":"loginuser","password":"wrongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Login() wrong password status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupGin()
	r.POST("/login", Login)

	body := `{"username":"nonexistent","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Login() nonexistent user status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestGetMe(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupGin()

	userID := uuid.New()
	r.GET("/me", func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("username", "testuser")
		GetMe(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetMe() status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["username"] != "testuser" {
		t.Errorf("GetMe() username = %v, want testuser", resp["username"])
	}
}
