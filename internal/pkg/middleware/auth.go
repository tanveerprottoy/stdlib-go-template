package middleware

import (
	"net/http"

	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/constant"
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/response"
	"github.com/tanveerprottoy/stdlib-go-template/internal/template/module/auth"
)

type Auth struct {
	Service *auth.Service
}

func NewAuth(s *auth.Service) *Auth {
	m := new(Auth)
	m.Service = s
	return m
}

// AuthUserMiddleWare auth user
func (m *Auth) AuthUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := m.Service.Authorize(r)
		if err != nil {
			response.RespondError(http.StatusForbidden, constant.Error, err, w)
			return
		}
		next.ServeHTTP(w, r)
	})
}
