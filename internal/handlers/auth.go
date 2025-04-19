package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"log"
	"medods/internal/database"
	"medods/internal/models"
	"net"
	"net/http"
	"time"
)

func AuthGuid(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		guid := r.FormValue("guid")
		_, err := database.Auth(guid)
		if err != nil {
			http.Error(w, "Такого guid не существует", http.StatusInternalServerError)
			log.Printf("Такого guid не существует: %v", err)
		} else {
			RefreshTocken(guid, r, w)
		}
	}
}

func RefreshTocken(guid string, r *http.Request, w http.ResponseWriter) {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)

	refreshTokenID := uuid.New().String()
	refreshTokenBytes := make([]byte, 32)
	_, err := rand.Read(refreshTokenBytes)
	if err != nil {
		http.Error(w, "токен не генерируется", http.StatusInternalServerError)
		log.Printf("токен не генерируется: %v", err)
		return
	}
	refreshToken := base64.StdEncoding.EncodeToString(refreshTokenBytes)
	hashedRefresh, _ := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)

	err = database.SaveRefreshToken(refreshTokenID, hashedRefresh, ip, guid)
	if err != nil {
		http.Error(w, "ошибка при сохранении refresh токена", http.StatusInternalServerError)
		log.Printf("ошибка при сохранении refresh токена: %v", err)
		return
	}

	row, err := database.GetDates(guid)
	if err != nil {
		http.Error(w, "ошибка при получении данных для access токена", http.StatusInternalServerError)
		log.Printf("ошибка при получении данных: %v", err)
		return
	}

	result := AccessToken(row, r, w)

	payload := map[string]string{
		"refresh_token": refreshToken,
		"access_token":  result,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "ошибка при кодировании JSON", http.StatusInternalServerError)
		log.Printf("ошибка JSON: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(payloadJSON)
}

func AccessToken(user *models.User, r *http.Request, w http.ResponseWriter) string {
	createAccess := jwt.MapClaims{
		"user_id":          user.ID,
		"ip":               user.IpAddress,
		"refresh_token_id": user.RefreshTokenID,
		"exp":              time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, createAccess)
	accessToken, _ := token.SignedString([]byte(models.SecretKey))

	return accessToken
}
