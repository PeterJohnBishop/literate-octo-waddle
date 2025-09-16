package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"fmt"
	"net/http"
	"rliterate-octo-waddle/server/middleware"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Email    string         `json:"email"`
	Password string         `json:"password"`
	Online   bool           `json:"online"`
	Files    pq.StringArray `json:"files" sql:"type:text[]"`
	Created  int64          `json:"created"`
	Updated  int64          `json:"updated"`
}

func CreateUsersTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT UNIQUE NOT NULL PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		online BOOL DEFAULT false,
		files TEXT[],
		created BIGINT DEFAULT (EXTRACT(EPOCH FROM now())),
    	updated BIGINT DEFAULT (EXTRACT(EPOCH FROM now()))
	);`

	_, err := db.Exec(query)
	return err
}

func GenerateUserID(email string) string {
	hash := sha256.Sum256([]byte(email))
	return fmt.Sprintf("user_%d", binary.BigEndian.Uint64(hash[:8]))
}

func HashedPassword(password string) (string, error) {
	hashedPassword, error := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(hashedPassword), error
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func RegisterUser(db *sql.DB, c *gin.Context) {
	var user User

	if err := c.ShouldBindJSON(&user); err != nil {
		fmt.Println("Failed to bind JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("Registering user with email:", user.Email)

	userId := GenerateUserID(user.Email)
	fmt.Println("Generated user ID:", userId)

	hashedPassword, err := HashedPassword(user.Password)
	if err != nil {
		fmt.Println("Error hashing password:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	fmt.Println("Password hashed successfully")

	query := `INSERT INTO users (id, name, email, password, online, files)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = db.ExecContext(c, query, userId, user.Name, user.Email, hashedPassword, user.Online, pq.Array(user.Files))
	if err != nil {
		fmt.Println("Database insert error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	access, refresh, err := middleware.GenerateTokens(userId)
	if err != nil {
		fmt.Println("Failed to generate tokens:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	middleware.StoreTokens(userId, access, refresh)

	fmt.Println("User registered and logged in successfully:", userId)
	c.JSON(http.StatusCreated, gin.H{
		"message":      "User created and logged in!",
		"token":        access,
		"refreshToken": refresh,
		"user": gin.H{
			"id":    userId,
			"name":  user.Name,
			"email": user.Email,
			// omit password hash
			"files":  user.Files,
			"online": user.Online,
		},
	})
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(db *sql.DB, c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println("Failed to bind JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	fmt.Println("Login attempt for email:", req.Email)

	var user User
	var files pq.StringArray
	query := `SELECT id, name, email, password, online, files, created, updated FROM users WHERE email = $1`
	err := db.QueryRowContext(c, query, req.Email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password,
		&user.Online, &files, &user.Created, &user.Updated,
	)
	user.Files = files
	if user.Files == nil {
		user.Files = []string{}
	}
	if err == sql.ErrNoRows {
		fmt.Println("No user found with email:", req.Email)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	} else if err != nil {
		fmt.Println("Database query error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("User found:", user.ID)

	if !CheckPasswordHash(req.Password, user.Password) {
		fmt.Println("Password verification failed for user:", user.ID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Password Verification Failed"})
		return
	}
	fmt.Println("Password verified for user:", user.ID)

	access, refresh, err := middleware.GenerateTokens(user.ID)
	if err != nil {
		fmt.Println("Failed to generate tokens:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	middleware.StoreTokens(user.ID, access, refresh)

	fmt.Println("Login successful for user:", user.ID)
	c.JSON(http.StatusOK, gin.H{
		"message":      "Login Success",
		"token":        access,
		"refreshToken": refresh,
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			// omit password hash
			"files":  user.Files,
			"online": user.Online,
		},
	})
}

func Refresh(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing refresh token"})
		return
	}

	claims, err := middleware.ValidateToken(body.RefreshToken, true)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	newAccess, _, err := middleware.GenerateTokens(claims.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": newAccess})
}

func GetUsers(db *sql.DB, c *gin.Context) {
	fmt.Println("Fetching all users")
	rows, err := db.QueryContext(c, "SELECT id, name, email, password, online, files, created, updated FROM users;")
	if err != nil {
		fmt.Println("Query failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Online, &user.Files, &user.Created, &user.Updated); err != nil {
			fmt.Println("Row scan failed:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if user.Files == nil {
			user.Files = []string{}
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		fmt.Println("Row iteration error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("Total users fetched:", len(users))
	c.JSON(http.StatusOK, users)
}

func GetUserByID(db *sql.DB, c *gin.Context) {
	id := c.Param("id")
	fmt.Println("Fetching user with ID:", id)

	var user User
	query := `SELECT id, name, email, password, online, files, created, updated FROM users WHERE id = $1`
	err := db.QueryRowContext(c, query, id).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Online, &user.Files, &user.Created, &user.Updated)
	if err == sql.ErrNoRows {
		fmt.Println("User not found:", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	} else if err != nil {
		fmt.Println("Query error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Println("User fetched:", user.ID)
	c.JSON(http.StatusOK, user)
}

func UpdateUser(db *sql.DB, c *gin.Context) {
	var user User

	if err := c.ShouldBindJSON(&user); err != nil {
		fmt.Println("Failed to bind JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `UPDATE users SET name=$1, email=$2, online=$3, files=$4, updated=EXTRACT(EPOCH FROM now()) WHERE id=$5`
	result, err := db.ExecContext(c, query, user.Name, user.Email, user.Online, user.Files, user.ID)
	if err != nil {
		fmt.Println("Update query failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("Failed to retrieve rows affected:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		fmt.Println("No rows updated for ID:", user.ID)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	fmt.Println("User updated successfully:", user.ID)
	c.JSON(http.StatusOK, gin.H{"message": "User updated!"})
}

func DeleteUserByID(db *sql.DB, c *gin.Context) {
	id := c.Param("id")
	fmt.Println("Deleting user with ID:", id)

	query := `DELETE FROM users WHERE id = $1`
	result, err := db.ExecContext(c, query, id)
	if err != nil {
		fmt.Println("Delete query failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("Failed to retrieve rows affected:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		fmt.Println("No user found to delete with ID:", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	fmt.Println("User deleted successfully:", id)
	c.JSON(http.StatusOK, gin.H{"message": "User deleted!"})
}

type UpdatePasswordRequest struct {
	UserID      string `json:"userId"`
	CurrentPass string `json:"currentPassword"`
	NewPass     string `json:"newPassword"`
}

func UpdatePassword(db *sql.DB, c *gin.Context) {
	var req UpdatePasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println("Failed to bind JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var storedHash string
	query := `SELECT password FROM users WHERE id = $1`
	err := db.QueryRowContext(c, query, req.UserID).Scan(&storedHash)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	} else if err != nil {
		fmt.Println("DB error fetching password:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !CheckPasswordHash(req.CurrentPass, storedHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	newHash, err := HashedPassword(req.NewPass)
	if err != nil {
		fmt.Println("Error hashing new password:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	updateQuery := `UPDATE users SET password=$1, updated=EXTRACT(EPOCH FROM now()) WHERE id=$2`
	_, err = db.ExecContext(c, updateQuery, newHash, req.UserID)
	if err != nil {
		fmt.Println("Error updating password:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	middleware.RevokeTokens(req.UserID)

	fmt.Println("Password updated successfully and tokens revoked for user:", req.UserID)
	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully. Please log in again."})
}

type LogoutRequest struct {
	Id string `json:"id"`
}

func Logout(c *gin.Context) {
	var req LogoutRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println("Failed to bind JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	middleware.RevokeTokens(req.Id)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully, tokens revoked"})
}
