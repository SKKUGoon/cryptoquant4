package upbittrade

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"net/url"

	upbitrest "cryptoquant.com/m/internal/upbit/rest"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func (t *Trader) GetSignature(orderSheet upbitrest.OrderSheet) ([]byte, string, error) {
	// 1. Build query string
	queryMap := orderSheet.ToParamsMap()
	query := url.Values{}
	for k, v := range queryMap {
		query.Add(k, v)
	}
	decodedQueryString, err := url.QueryUnescape(query.Encode()) // Upbit does unquote
	if err != nil {
		return nil, "", err
	}

	// 2. Hash it with SHA512
	hash := sha512.New()
	hash.Write([]byte(decodedQueryString))
	queryHash := hex.EncodeToString(hash.Sum(nil))

	// 3. Generate nonce
	nonce := uuid.NewString()

	// 4. Create claims
	claims := jwt.MapClaims{
		"access_key":     t.pubkey,
		"nonce":          nonce,
		"query_hash":     queryHash,
		"query_hash_alg": "SHA512",
	}

	// 5. Sign JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(t.prikey))
	if err != nil {
		return nil, "", err
	}

	// 6. Authorization Token
	authToken := "Bearer " + signedToken

	// 7. Marshal body into JSON
	jsonBody, err := json.Marshal(queryMap)
	if err != nil {
		return nil, "", err
	}

	return jsonBody, authToken, nil
}
