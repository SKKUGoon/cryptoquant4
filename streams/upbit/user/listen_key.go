package upbituser

import (
	"encoding/json"
	"os"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func CreateListenKey() ([]byte, string, error) {
	access_key := os.Getenv("UPBIT_API_KEY")
	secret_key := os.Getenv("UPBIT_SECRET_KEY")

	// 1. Same with creating non-parameter JWT
	nonce := uuid.NewString()

	// 2. Create claims
	claims := jwt.MapClaims{
		"access_key": access_key,
		"nonce":      nonce,
	}

	// 3. Sign JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret_key))
	if err != nil {
		return nil, "", err
	}

	// 4. Authorization Token
	authToken := "Bearer " + signedToken

	// 5. Marshal body into JSON
	jsonBody, err := json.Marshal(claims)
	if err != nil {
		return nil, "", err
	}

	return jsonBody, authToken, nil
}
