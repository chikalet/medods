package database

import (
	"database/sql"
	"log"

	"fmt"
	"medods/internal/models"

	"os"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func Connect() (*sql.DB, error) {
	connStr := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Ошибка подключения к БД: %v", err)
		return nil, err
	}
	return db, nil
}

func Auth(guid string) (*sql.Row, error) {
	var user models.User
	db, err := Connect()
	if err != nil {
		return nil, err
	}
	row := db.QueryRow("SELECT id FROM users_tokens WHERE guid = $1", guid)

	newError := row.Scan(&user.ID)
	if newError == sql.ErrNoRows {
		return nil, newError
	} else if newError != nil {
		return nil, newError
	}
	defer db.Close()
	return row, nil
}
func AuthRefresh(rawRefreshToken string, ip string) (*models.User, error) {
	db, err := Connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query(`
        SELECT guid, refresh_token, used, ip_address 
        FROM users_tokens 
        WHERE used = false`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var user models.User
	found := false
	for rows.Next() {
		var hashedToken []byte
		if err := rows.Scan(&user.Guid, &hashedToken, &user.Used, &user.IpAddress); err != nil {
			continue
		}
		if bcrypt.CompareHashAndPassword(hashedToken, []byte(rawRefreshToken)) == nil {
			found = true
			break
		}
	}

	if !found {
		return nil, sql.ErrNoRows
	}

	if user.IpAddress != ip {
		return nil, fmt.Errorf("IP mismatch")
	}

	return &user, nil
}

func GetDates(guid string) (*models.User, error) {
	db, err := Connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var user models.User
	err = db.QueryRow("SELECT guid, refresh_token_id, ip_address FROM users_tokens WHERE guid = $1", guid).
		Scan(&user.Guid, &user.RefreshTokenID, &user.IpAddress)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func SaveRefreshToken(refreshTokenID string, hashedRefresh []byte, ip string, guid string) error {
	db, err := Connect()
	if err != nil {
		log.Printf("Ошибка подключения к БД: %v", err)
		return err
	}
	_, err = db.Exec(`UPDATE users_tokens SET refresh_token_id=$1, refresh_token=$2, ip_address=$3 WHERE guid=$4`, refreshTokenID, hashedRefresh, ip, guid)
	return err
}

func ChangeUsed(guid string) error {
	db, err := Connect()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`UPDATE users_tokens SET used=true WHERE guid=$1`, guid)
	return err
}
