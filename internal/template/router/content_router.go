package router

import (
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/constant"
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/middleware"
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/router"
	"github.com/tanveerprottoy/stdlib-go-template/internal/template/module/content"

	"github.com/go-chi/chi"
)

func RegisterContentRoutes(router *router.Router, version string, module *content.Module, authMiddleWare *middleware.AuthMiddleware) {
	router.Mux.Group(
		func(r chi.Router) {
			r.Use(authMiddleWare.AuthUser)
			r.Route(
				constant.ApiPattern+version+constant.ContentsPattern,
				func(r chi.Router) {
					r.Get(constant.RootPattern, module.Handler.ReadMany)
					r.Get(constant.RootPattern+"{id}", module.Handler.ReadOne)
					r.Post(constant.RootPattern, module.Handler.Create)
					r.Patch(constant.RootPattern+"{id}", module.Handler.Update)
					r.Delete(constant.RootPattern+"{id}", module.Handler.Delete)
				},
			)
		},
	)
}
