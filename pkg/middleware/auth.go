package middleware

import (
	"net/http"
	"strings"
	"time"

	"simple-kanban/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims JWTクレーム構造体
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

// AuthMiddleware JWT認証ミドルウェア
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Authorizationヘッダーからトークンを取得
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "認証トークンが必要です",
			})
			c.Abort()
			return
		}

		// Bearer tokenの形式チェック
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "不正な認証ヘッダー形式です",
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// JWTトークンを検証
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// 署名方法の確認
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWT.SecretKey), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "無効な認証トークンです",
			})
			c.Abort()
			return
		}

		// ユーザー情報をコンテキストに設定
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)

		c.Next()
	}
}

// GenerateToken JWTトークンを生成します
func GenerateToken(userID uuid.UUID, email string, cfg *config.Config) (string, error) {
	expirationTime := time.Now().Add(time.Duration(cfg.JWT.ExpireHours) * time.Hour)

	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "simple-kanban",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWT.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken JWTトークンを検証します
func ValidateToken(tokenString string, cfg *config.Config) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(cfg.JWT.SecretKey), nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return claims, nil
}

// GetUserIDFromContext コンテキストからユーザーIDを取得します
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, gin.Error{Err: jwt.ErrTokenInvalidClaims}
	}

	id, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.Nil, gin.Error{Err: jwt.ErrTokenInvalidClaims}
	}

	return id, nil
}

// GetUserEmailFromContext コンテキストからユーザーメールを取得します
func GetUserEmailFromContext(c *gin.Context) (string, error) {
	userEmail, exists := c.Get("user_email")
	if !exists {
		return "", gin.Error{Err: jwt.ErrTokenInvalidClaims}
	}

	email, ok := userEmail.(string)
	if !ok {
		return "", gin.Error{Err: jwt.ErrTokenInvalidClaims}
	}

	return email, nil
}
