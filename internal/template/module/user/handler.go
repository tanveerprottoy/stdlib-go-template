package user

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/constant"
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/response"
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/validatorext"
	"github.com/tanveerprottoy/stdlib-go-template/internal/template/module/user/dto"
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/adapter"
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/httpext"
)

// Hanlder is responsible for extracting data
// from request body and building and seding response
type Handler struct {
	service  *Service
	validate *validator.Validate
}

func NewHandler(s *Service, v *validator.Validate) *Handler {
	h := new(Handler)
	h.service = s
	h.validate = v
	return h
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var v dto.CreateUpdateUserDTO
	// parse the request body
	err := httpext.ParseRequestBody(r.Body, &v)
	if err != nil {
		response.RespondError(http.StatusBadRequest, constant.Errors, []string{constant.InvalidRequestBody}, w)
		return
	}
	// validate the request body
	validationErrs := validatorext.ValidateStruct(&v, h.validate)
	if validationErrs != nil {
		response.RespondError(http.StatusBadRequest, constant.Errors, validationErrs, w)
		return
	}
	ctx := r.Context()
	e, httpErr := h.service.Create(&v, ctx)
	if httpErr.Err != nil {
		response.RespondError(httpErr.Code, constant.Error, httpErr.Err.Error(), w)
		return
	}
	response.Respond(http.StatusCreated, e, w)
}

func (h *Handler) ReadMany(w http.ResponseWriter, r *http.Request) {
	limit := 10
	page := 1
	var err error
	limitStr := httpext.GetQueryParam(r, constant.KeyLimit)
	if limitStr != "" {
		limit, err = adapter.StringToInt(limitStr)
		if err != nil {
			response.RespondError(http.StatusBadRequest, constant.Error, err.Error(), w)
			return
		}
	}
	pageStr := httpext.GetQueryParam(r, constant.KeyPage)
	if pageStr != "" {
		page, err = adapter.StringToInt(pageStr)
		if err != nil {
			response.RespondError(http.StatusBadRequest, constant.Error, err.Error(), w)
			return
		}
	}
	e, httpErr := h.service.ReadMany(limit, page, nil)
	if httpErr.Err != nil {
		response.RespondError(httpErr.Code, constant.Error, httpErr.Err.Error(), w)
		return
	}
	response.Respond(http.StatusOK, e, w)
}

func (h *Handler) ReadOne(w http.ResponseWriter, r *http.Request) {
	id := httpext.GetURLParam(r, constant.KeyId)
	e, httpErr := h.service.ReadOne(id, nil)
	if httpErr.Err != nil {
		response.RespondError(httpErr.Code, constant.Error, httpErr.Err.Error(), w)
		return
	}
	response.Respond(http.StatusOK, e, w)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := httpext.GetURLParam(r, constant.KeyId)
	var v dto.CreateUpdateUserDTO
	// parse the request body
	err := httpext.ParseRequestBody(r.Body, &v)
	if err != nil {
		response.RespondError(http.StatusBadRequest, constant.Errors, []string{constant.InvalidRequestBody}, w)
		return
	}
	// validate the request body
	validationErrs := validatorext.ValidateStruct(&v, h.validate)
	if validationErrs != nil {
		response.RespondError(http.StatusBadRequest, constant.Errors, validationErrs, w)
		return
	}
	e, httpErr := h.service.Update(id, &v, nil)
	if httpErr.Err != nil {
		response.RespondError(httpErr.Code, constant.Error, httpErr.Err.Error(), w)
		return
	}
	response.Respond(http.StatusOK, e, w)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := httpext.GetURLParam(r, constant.KeyId)
	e, httpErr := h.service.Delete(id, nil)
	if httpErr.Err != nil {
		response.RespondError(httpErr.Code, constant.Error, httpErr.Err.Error(), w)
		return
	}
	response.Respond(http.StatusOK, e, w)
}

func (h *Handler) Public(w http.ResponseWriter, r *http.Request) {
	response.Respond(http.StatusOK, map[string]string{"message": "public api"}, w)
}
