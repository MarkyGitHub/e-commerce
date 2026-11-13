package auth

import (
	"context"
	"testing"
	"testing/quick"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/MarkyGitHub/e-commerce/internal/shared"
	"github.com/MarkyGitHub/e-commerce/internal/testutil"
)

func setupAuthTestDB(t *testing.T) *gorm.DB {
	testDB := testutil.SetupTestDB(t)
	if err := testDB.AutoMigrate(&User{}, &Session{}); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}
	return testDB
}

func TestUserRepository_Create(t *testing.T) {
	testDB := setupAuthTestDB(t)
	repo := NewUserRepository(testDB)
	ctx := context.Background()

	tests := []struct {
		name    string
		user    *User
		wantErr error
	}{
		{
			name: "create valid user",
			user: &User{
				ID:           uuid.New(),
				Email:        "test@example.com",
				PasswordHash: "hashed_password",
				Username:     "Test User",
				Role:         RoleCustomer,
			},
			wantErr: nil,
		},
		{
			name: "create user with admin role",
			user: &User{
				ID:           uuid.New(),
				Email:        "admin@example.com",
				PasswordHash: "hashed_password",
				Username:     "Admin User",
				Role:         RoleAdmin,
			},
			wantErr: nil,
		},
		{
			name: "duplicate email",
			user: &User{
				ID:           uuid.New(),
				Email:        "duplicate@example.com",
				PasswordHash: "hashed_password",
				Username:     "First User",
				Role:         RoleCustomer,
			},
			wantErr: nil, // First create should succeed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.user)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("Create() error = %v, want nil", err)
				}
				// Verify user was created
				found, err := repo.FindByID(ctx, tt.user.ID)
				if err != nil {
					t.Errorf("FindByID() error = %v", err)
				}
				if found.Email != tt.user.Email {
					t.Errorf("FindByID() email = %s, want %s", found.Email, tt.user.Email)
				}
			} else {
				if err != tt.wantErr {
					t.Errorf("Create() error = %v, want %v", err, tt.wantErr)
				}
			}
		})
	}

	// Test duplicate email error
	t.Run("duplicate email error", func(t *testing.T) {
		duplicateUser := &User{
			ID:           uuid.New(),
			Email:        "duplicate@example.com", // Same as above
			PasswordHash: "hashed_password",
			Username:     "Second User",
			Role:         RoleCustomer,
		}
		err := repo.Create(ctx, duplicateUser)
		if err != shared.ErrConflict {
			t.Errorf("Create() duplicate email error = %v, want %v", err, shared.ErrConflict)
		}
	})
}

func TestUserRepository_FindByEmail(t *testing.T) {
	testDB := setupAuthTestDB(t)
	repo := NewUserRepository(testDB)
	ctx := context.Background()

	// Create test user
	user := &User{
		ID:           uuid.New(),
		Email:        "findbyemail@example.com",
		PasswordHash: "hashed_password",
		Username:     "Find User",
		Role:         RoleCustomer,
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name    string
		email   string
		wantErr error
		wantID  uuid.UUID
	}{
		{
			name:    "find existing user",
			email:   "findbyemail@example.com",
			wantErr: nil,
			wantID:  user.ID,
		},
		{
			name:    "user not found",
			email:   "nonexistent@example.com",
			wantErr: shared.ErrNotFound,
			wantID:  uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindByEmail(ctx, tt.email)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("FindByEmail() error = %v, want %v", err, tt.wantErr)
				}
				if found != nil {
					t.Errorf("FindByEmail() found = %v, want nil", found)
				}
			} else {
				if err != nil {
					t.Errorf("FindByEmail() error = %v, want nil", err)
				}
				if found.ID != tt.wantID {
					t.Errorf("FindByEmail() id = %v, want %v", found.ID, tt.wantID)
				}
			}
		})
	}
}

func TestUserRepository_FindByID(t *testing.T) {
	testDB := setupAuthTestDB(t)
	repo := NewUserRepository(testDB)
	ctx := context.Background()

	// Create test user
	user := &User{
		ID:           uuid.New(),
		Email:        "findbyid@example.com",
		PasswordHash: "hashed_password",
		Username:     "Find ID User",
		Role:         RoleCustomer,
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr error
		wantID  uuid.UUID
	}{
		{
			name:    "find existing user",
			id:      user.ID,
			wantErr: nil,
			wantID:  user.ID,
		},
		{
			name:    "user not found",
			id:      uuid.New(),
			wantErr: shared.ErrNotFound,
			wantID:  uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindByID(ctx, tt.id)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("FindByID() error = %v, want %v", err, tt.wantErr)
				}
				if found != nil {
					t.Errorf("FindByID() found = %v, want nil", found)
				}
			} else {
				if err != nil {
					t.Errorf("FindByID() error = %v, want nil", err)
				}
				if found.ID != tt.wantID {
					t.Errorf("FindByID() id = %v, want %v", found.ID, tt.wantID)
				}
			}
		})
	}
}

func TestSessionRepository_Create(t *testing.T) {
	testDB := setupAuthTestDB(t)
	userRepo := NewUserRepository(testDB)
	sessionRepo := NewSessionRepository(testDB)
	ctx := context.Background()

	// Create test user
	user := &User{
		ID:           uuid.New(),
		Email:        "session@example.com",
		PasswordHash: "hashed_password",
		Username:     "Session User",
		Role:         RoleCustomer,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name    string
		session *Session
		wantErr bool
	}{
		{
			name: "create valid session",
			session: &Session{
				ID:        uuid.New(),
				UserID:    user.ID,
				Token:     "test_token_123",
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
		},
		{
			name: "create session with different token",
			session: &Session{
				ID:        uuid.New(),
				UserID:    user.ID,
				Token:     "different_token_456",
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sessionRepo.Create(ctx, tt.session)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				// Verify session was created
				found, err := sessionRepo.FindByToken(ctx, tt.session.Token)
				if err != nil {
					t.Errorf("FindByToken() error = %v", err)
				}
				if found.Token != tt.session.Token {
					t.Errorf("FindByToken() token = %s, want %s", found.Token, tt.session.Token)
				}
			}
		})
	}
}

func TestSessionRepository_FindByToken(t *testing.T) {
	testDB := setupAuthTestDB(t)
	userRepo := NewUserRepository(testDB)
	sessionRepo := NewSessionRepository(testDB)
	ctx := context.Background()

	// Create test user
	user := &User{
		ID:           uuid.New(),
		Email:        "findtoken@example.com",
		PasswordHash: "hashed_password",
		Username:     "Find Token User",
		Role:         RoleCustomer,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test session
	session := &Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     "find_token_123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := sessionRepo.Create(ctx, session); err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantErr error
		wantID  uuid.UUID
	}{
		{
			name:    "find existing session",
			token:   "find_token_123",
			wantErr: nil,
			wantID:  session.ID,
		},
		{
			name:    "session not found",
			token:   "nonexistent_token",
			wantErr: shared.ErrNotFound,
			wantID:  uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := sessionRepo.FindByToken(ctx, tt.token)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("FindByToken() error = %v, want %v", err, tt.wantErr)
				}
				if found != nil {
					t.Errorf("FindByToken() found = %v, want nil", found)
				}
			} else {
				if err != nil {
					t.Errorf("FindByToken() error = %v, want nil", err)
				}
				if found.ID != tt.wantID {
					t.Errorf("FindByToken() id = %v, want %v", found.ID, tt.wantID)
				}
			}
		})
	}
}

func TestSessionRepository_Delete(t *testing.T) {
	testDB := setupAuthTestDB(t)
	userRepo := NewUserRepository(testDB)
	sessionRepo := NewSessionRepository(testDB)
	ctx := context.Background()

	// Create test user
	user := &User{
		ID:           uuid.New(),
		Email:        "delete@example.com",
		PasswordHash: "hashed_password",
		Username:     "Delete User",
		Role:         RoleCustomer,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test session
	session := &Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     "delete_token_123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := sessionRepo.Create(ctx, session); err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	// Verify session exists
	_, err := sessionRepo.FindByToken(ctx, "delete_token_123")
	if err != nil {
		t.Fatalf("Session should exist before delete: %v", err)
	}

	// Delete session
	if err := sessionRepo.Delete(ctx, "delete_token_123"); err != nil {
		t.Errorf("Delete() error = %v, want nil", err)
	}

	// Verify session is deleted
	_, err = sessionRepo.FindByToken(ctx, "delete_token_123")
	if err != shared.ErrNotFound {
		t.Errorf("Delete() session should not exist after delete, got error = %v", err)
	}
}

// Property 3: Session Authentication Round-Trip
func TestProperty_SessionAuthenticationRoundTrip(t *testing.T) {
	testDB := setupAuthTestDB(t)
	userRepo := NewUserRepository(testDB)
	sessionRepo := NewSessionRepository(testDB)
	ctx := context.Background()

	f := func(email string, password string) bool {
		// Skip invalid inputs
		if len(email) < 5 || len(password) < 8 {
			return true
		}

		// Create user
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			// Some inputs (e.g., very long passwords) are rejected by bcrypt.
			// Treat these as invalid inputs for this property.
			return true
		}

		user := &User{
			ID:           uuid.New(),
			Email:        email,
			PasswordHash: string(hashedPassword),
			Username:     "Test User",
			Role:         RoleCustomer,
		}
		if err := userRepo.Create(ctx, user); err != nil {
			// Duplicate email or other DB constraints can happen with random inputs.
			// Treat these as invalid inputs for this property.
			return true
		}

		// Create session
		token := uuid.New().String()
		session := &Session{
			ID:        uuid.New(),
			UserID:    user.ID,
			Token:     token,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		if err := sessionRepo.Create(ctx, session); err != nil {
			return false
		}

		// Property: Session can be found by token
		foundSession, err := sessionRepo.FindByToken(ctx, token)
		if err != nil || foundSession == nil {
			return false
		}

		// Property: Session token can be used to get user
		foundUser, err := userRepo.FindByID(ctx, foundSession.UserID)
		if err != nil || foundUser == nil {
			return false
		}

		// Property: Logout invalidates session
		if err := sessionRepo.Delete(ctx, token); err != nil {
			return false
		}

		// Property: After logout, token cannot be used
		_, err = sessionRepo.FindByToken(ctx, token)
		return err == shared.ErrNotFound
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}
