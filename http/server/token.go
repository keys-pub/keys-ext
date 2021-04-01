package server

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

func (s *Server) GenerateToken() (string, error) {
	jt := jwt.New(jwt.GetSigningMethod("HS256"))
	token, err := jt.SignedString(s.tokenKey)
	if err != nil {
		return "", errors.Wrapf(err, "failed to generate access token")
	}
	return token, nil
}

func (s *Server) ValidateToken(token string) error {
	t, err := jwt.Parse(token, s.jwtToken)
	if err != nil {
		return err
	}
	return t.Claims.Valid()
}

func (s *Server) jwtToken(t *jwt.Token) (interface{}, error) {
	return s.tokenKey, nil
}
