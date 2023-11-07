package http

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"reflect"

	"github.com/golang-jwt/jwt"
	"github.com/levelsoftware/echoip/config"
)

func ParseJWT(runConfig *config.Config, tokenString string) error {
	if _, err := jwt.Parse(tokenString, GetTokenKey(runConfig)); err != nil {
		if runConfig.Debug {
			log.Printf("Error validating token ( %s ): %s \n", tokenString, err)
		}

		return new(InvalidTokenError)
	}

	return nil
}

func GetTokenKey(runConfig *config.Config) func(token *jwt.Token) (interface{}, error) {
	signingMethod := jwt.GetSigningMethod(runConfig.Jwt.SigningMethod)

	var key interface{}

	switch signingMethod.Alg() {
	case "ES256", "ES384", "ES512":
		pubKey, _ := GetECDSAKey(runConfig.Jwt.PublicKeyData)
		key = pubKey
	case "RS256", "RS384", "RS512":
		pubKey, _ := GetRSAKey(runConfig.Jwt.PublicKeyData)
		key = pubKey
	default:
		key = []byte(runConfig.Jwt.Secret)
	}

	return func(token *jwt.Token) (interface{}, error) {
		expected := reflect.TypeOf(signingMethod)
		got := reflect.TypeOf(token.Method)
		if expected != got {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return key, nil
	}
}

func GetECDSAKey(data []byte) (*ecdsa.PublicKey, error) {
	pkiBlock, _ := pem.Decode(data)
	var publicKey *ecdsa.PublicKey
	pubInterface, _ := x509.ParsePKIXPublicKey(pkiBlock.Bytes)
	publicKey = pubInterface.(*ecdsa.PublicKey)
	return publicKey, nil
}

func GetRSAKey(data []byte) (*rsa.PublicKey, error) {
	pkiBlock, _ := pem.Decode(data)
	var publicKey *rsa.PublicKey
	pubInterface, _ := x509.ParsePKIXPublicKey(pkiBlock.Bytes)
	publicKey = pubInterface.(*rsa.PublicKey)
	return publicKey, nil
}
