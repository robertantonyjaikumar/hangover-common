package middlewares

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/robertantonyjaikumar/hangover-common/config"
	"github.com/robertantonyjaikumar/hangover-common/logger"
	"go.uber.org/zap"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		bearerToken := c.Request.Header.Get("Authorization")
		bearerToken = strings.ReplaceAll(bearerToken, "Bearer ", "")
		// splitToken := strings.Split(bearerToken, "Bearer ")
		// bearerToken = splitToken[1]
		if bearerToken == "" {
			c.AbortWithStatus(401)
			c.Next()
			return
		}

		claimPayload, err := ValidateToken(bearerToken)
		if err != nil {
			c.AbortWithStatus(401)
			return
		}

		c.Set("x-claim-payload", claimPayload)
		c.Set("x-token", bearerToken)

		c.Next()
	}
}

// JWK represents the JSON Web Key format
type JWK struct {
	Kid     string `json:"kid"`
	Alg     string `json:"alg"`
	Kty     string `json:"kty"`
	Use     string `json:"use"`
	N       string `json:"n"`
	E       string `json:"e"`
	X5c     string `json:"x5c"`
	X5t     string `json:"x5t"`
	X5tS256 string `json:"x5t#S256"`
}

type Jwks struct {
	Keys []JsonWebKeys `json:"keys"`
}

type JsonWebKeys struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKSet represents the JSON Web Key Set format
type JWKSet struct {
	Keys []JWK `json:"keys"`
}

func fetchJwks(url string) (*Jwks, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jwks Jwks
	err = json.NewDecoder(resp.Body).Decode(&jwks)
	if err != nil {
		return nil, err
	}

	return &jwks, nil
}

func jwksToPublicKey(jwks *Jwks, kid string) (*rsa.PublicKey, error) {
	for _, key := range jwks.Keys {
		if key.Kid == kid && key.Kty == "RSA" {
			modulus, err := base64.RawURLEncoding.DecodeString(key.N)
			if err != nil {
				return nil, err
			}

			exponent, err := base64.RawURLEncoding.DecodeString(key.E)
			if err != nil {
				return nil, err
			}

			return &rsa.PublicKey{
				N: big.NewInt(0).SetBytes(modulus),
				E: int(big.NewInt(0).SetBytes(exponent).Int64()),
			}, nil
		}
	}

	return nil, fmt.Errorf("public key not found")
}

func ClientMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqKey := c.Request.Header.Get("X-Auth-Key")
		reqSecret := c.Request.Header.Get("X-Auth-Secret")

		var key string
		var secret string
		if key = "mg_dev_key_001"; len(strings.TrimSpace(key)) == 0 {
			c.AbortWithStatus(500)
		}
		if secret = "FMvHd45WBXnkxKYJQtFfzNRuqAnPfP3L"; len(strings.TrimSpace(secret)) == 0 {
			c.AbortWithStatus(401)
		}
		if key != reqKey || secret != reqSecret {
			c.AbortWithStatus(401)
			return
		}
		c.Next()
	}
}

func ValidateToken(tokenString string) (JwtAuthPayload, error) {
	cfg := config.LoadAuthConfig()
	var jwtKey = []byte(cfg.AccessSecret)
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		},
	)

	if err != nil {
		return JwtAuthPayload{}, err
	}

	claims = token.Claims.(jwt.MapClaims)

	payload := claims["Payload"]

	payloadMap := payload.(map[string]interface{})
	jsonString, _ := json.Marshal(payloadMap)

	jwtPayload := JwtAuthPayload{}

	json.Unmarshal(jsonString, &jwtPayload)

	return jwtPayload, err

}

func ValidateSessionToken(tokenString string) (JwtSessionPayload, error) {
	cfg := config.LoadAuthConfig()
	var jwtKey = []byte(cfg.AccessSecret)
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		},
	)

	if err != nil {
		logger.Error("Validate Session Token failed", zap.Error(err))
		return JwtSessionPayload{}, err
	}

	claims = token.Claims.(jwt.MapClaims)

	payload := claims["Payload"]

	payloadMap := payload.(map[string]interface{})
	jsonString, _ := json.Marshal(payloadMap)

	jwtPayload := JwtSessionPayload{}

	json.Unmarshal(jsonString, &jwtPayload)

	return jwtPayload, err

}
