package router

import (
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/constant"
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/middleware"
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/router"
	"github.com/tanveerprottoy/stdlib-go-template/internal/template/module/user"

	"github.com/go-chi/chi"
)

func RegisterUserRoutes(router *router.Router, version string, module *user.Module, authMiddleWare *middleware.Auth) {
	router.Mux.Route(
		constant.ApiPattern+version+constant.UsersPattern,
		func(r chi.Router) {
			// public routes
			r.Get(constant.RootPattern+"public", module.Handler.Public)
			r.Group(func(r chi.Router) {
				// r.Use(authMiddleWare.AuthUser)
				r.Get(constant.RootPattern, module.Handler.ReadMany)
				r.Get(constant.RootPattern+"{id}", module.Handler.ReadOne)
				r.Post(constant.RootPattern, module.Handler.Create)
				r.Patch(constant.RootPattern+"{id}", module.Handler.Update)
				r.Delete(constant.RootPattern+"{id}", module.Handler.Delete)
			})
		},
	)
}
