package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"medods/internal/database"
	"medods/internal/models"
	"net"
	"net/http"
	"time"
)

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Метод не найден", http.StatusMethodNotAllowed)
		return
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	refreshToken := r.FormValue("refresh")

	user, err := database.AuthRefresh(refreshToken, ip)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Рефреш токен не опознан", http.StatusUnauthorized)
		} else {
			http.Error(w, "ошибка сервераr", http.StatusInternalServerError)
		}
		log.Printf("Ошибка рефреш токена: %v", err)
		return
	}

	ChangeAccess(r.FormValue("access"), user, r, w)
}

func ChangeAccess(accessToken string, user *models.User, r *http.Request, w http.ResponseWriter) {
	guid := user.Guid
	dates, err := database.GetDates(guid)
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(models.SecretKey), nil
	})

	if err != nil {
		log.Printf("Ошибка при парсинге токена: %v\n", err)
		return
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		log.Println("Payload токена:")
		for key, value := range claims {
			log.Printf("%s: %v\n", key, value)
		}
		if claims["refresh_token_id"] != dates.RefreshTokenID {
			log.Printf("refresh token id не подходит, неверный access токен")
			return
		} else {
			createAccess := jwt.MapClaims{
				"user_id":          user.ID,
				"ip":               user.IpAddress,
				"refresh_token_id": user.RefreshTokenID,
				"exp":              time.Now().Add(15 * time.Minute).Unix(),
			}
			token := jwt.NewWithClaims(jwt.SigningMethodHS512, createAccess)
			accessToken, _ := token.SignedString([]byte(models.SecretKey))
			err := database.ChangeUsed(dates.Guid)
			if err != nil {
				log.Printf("Ошибка при запросе к бд для изменения статуса refresh токена: %v", err)
			}
			payload := map[string]string{
				"access_token": accessToken,
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
	} else {
		log.Printf("Невалидный токен или ошибка в claims")
		return
	}
}
