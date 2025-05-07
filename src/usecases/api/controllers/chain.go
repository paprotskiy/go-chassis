package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
)

const (
	cookieName     = "must-be-preset"         // TODO: requires configuration
	cookieLifeTime = 5 * time.Hour            //
	salt           = "requires configuration" // TODO: create salt strategy
	stepLifetime   = 10 * time.Minute         // TODO: configure if required or reduce
	stepNotBefore  = 0                        // TODO: configure if required or reduce
	expClaim       = "exp"
	nbfClaim       = "nbf"
	stepKey        = "id"
)

func Chain[In any, Out any](api ApiParser[In, Out], operation func(context.Context, In, *string) (result *Out, key *string, err error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, err := api.ParseRequest(r)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, err.Error())
			return
		}

		key, _ := GetSessionKey(r)

		res, newKey, err := operation(r.Context(), *input, key)
		if err != nil {
			api.HandleErr(err, w, r)
			return
		}

		if newKey != nil {
			commitStep(w, *newKey)
		}

		api.HandleReply(res, w, r)
	}
}

func GetSessionKey(r *http.Request) (*string, error) {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get session :: %v")
	}

	key, err := parseIdFromToken(c.Value)
	if err != nil {
		return nil, errors.Wrap(err, `failed to extract key from token`)
	}

	return key, nil
}

func parseIdFromToken(token string) (*string, error) {

	claims := jwt.MapClaims{}
	keyFunc := func(token *jwt.Token) (any, error) {
		return []byte(salt), nil
	}
	if _, err := new(jwt.Parser).ParseWithClaims(token, claims, keyFunc); err != nil {
		return nil, errors.Wrap(err, `failed to parse jwt tokenv`)
	}

	id, ok := claims[stepKey].(string)
	if !ok {
		return nil, errors.New(`failed to parse jwt token payload`)
	}

	return &id, nil
}

func commitStep(w http.ResponseWriter, key string) {
	packedNewKey, _ := createChainJwtToken(key)

	http.SetCookie(w, &http.Cookie{
		Name:       cookieName,
		Value:      *packedNewKey,
		Path:       "",
		Domain:     "",
		Expires:    time.Time{},
		RawExpires: "",
		MaxAge:     int(cookieLifeTime.Seconds()),
		Secure:     false,
		HttpOnly:   true,                    //  XSS countermeasure
		SameSite:   http.SameSiteStrictMode, // CSRF countermeasure
		Raw:        "",
		Unparsed:   nil,
	})
}

func createChainJwtToken(key string) (*string, error) {
	now := time.Now()

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		expClaim: now.Add(stepLifetime).Unix(),
		nbfClaim: now.Add(stepNotBefore).Unix(),
		stepKey:  key,
	})

	token, err := jwtToken.SignedString([]byte(salt))
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate token")
	}

	return &token, nil
}
