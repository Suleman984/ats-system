package controllers

import (
	"errors"
	"net/http"
	"strings"

	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"
	"ats-backend/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	CompanyName  string `json:"company_name" binding:"required,max=255"`
	Email        string `json:"email" binding:"required,email,max=255"`
	Password     string `json:"password" binding:"required,min=8,max=72"`
	Name         string `json:"name" binding:"required,max=255"`
	EmbeddedMode bool   `json:"embedded_mode"`
	EmbedDomain  string `json:"embed_domain" binding:"omitempty,max=255"`
}

// dummyHash keeps login timing the same whether or not the email exists.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("timing-equalizer"), bcrypt.DefaultCost)

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}

	var admin models.Admin
	err := config.DB.Where("lower(email) = ?", strings.ToLower(req.Email)).First(&admin).Error
	hash := dummyHash
	if err == nil {
		hash = []byte(admin.PasswordHash)
	}
	if bcrypt.CompareHashAndPassword(hash, []byte(req.Password)) != nil || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	respondWithAdminToken(c, http.StatusOK, &admin, "")
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		serverError(c, "Failed to hash password")
		return
	}

	email := strings.ToLower(req.Email)
	company := models.Company{CompanyName: req.CompanyName, Email: email, EmbeddedMode: req.EmbeddedMode}
	if req.EmbedDomain != "" {
		company.EmbedDomain = &req.EmbedDomain
	}
	admin := models.Admin{Name: req.Name, Email: email, PasswordHash: string(hashed), Role: "admin"}

	err = config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&company).Error; err != nil {
			return err
		}
		admin.CompanyID = company.ID
		return tx.Create(&admin).Error
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		c.JSON(http.StatusConflict, gin.H{"error": "An account with this email already exists"})
		return
	}
	if err != nil {
		serverError(c, "Failed to create account")
		return
	}

	services.LogCompanyRegistered(company.ID, admin.ID, company.CompanyName, admin.Email)
	respondWithAdminToken(c, http.StatusCreated, &admin, "Registration successful")
}

func respondWithAdminToken(c *gin.Context, status int, admin *models.Admin, message string) {
	token, err := utils.GenerateJWT(admin.ID, admin.CompanyID)
	if err != nil {
		serverError(c, "Failed to generate token")
		return
	}
	body := gin.H{
		"token": token,
		"admin": gin.H{"id": admin.ID, "name": admin.Name, "email": admin.Email, "company_id": admin.CompanyID},
	}
	if message != "" {
		body["message"] = message
	}
	c.JSON(status, body)
}
