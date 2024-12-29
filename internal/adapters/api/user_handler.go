package api

import (
	"encoding/json"
	"net/http"

	resp "github.com/ouz/goauthboilerplate/internal/adapters/api/response"
	"github.com/ouz/goauthboilerplate/internal/adapters/api/validator"
	authDto "github.com/ouz/goauthboilerplate/internal/application/auth/dto"
	"github.com/ouz/goauthboilerplate/internal/domain/user"
	"github.com/ouz/goauthboilerplate/pkg/errors"
)

type UserHandler struct {
	userService user.UserService
}

func NewUserHandler(userService user.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var request authDto.UserRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		resp.Error(w, errors.BadRequestError("Invalid request body"))
		return
	}

	if err := validator.Validator.Struct(request); err != nil {
		resp.Error(w, errors.BadRequestError("Invalid request body"))
		return
	}

	if err := h.userService.Register(r.Context(), request); err != nil {
		resp.Error(w, err)
		return
	}

	resp.JSON(w, http.StatusCreated, nil)
}

func (h *UserHandler) RegisterAnonymousUser(w http.ResponseWriter, r *http.Request) {
	user, err := h.userService.RegisterAnonymousUser(r.Context())
	if err != nil {
		resp.Error(w, err)
		return
	}

	response := authDto.AnonymousUserResponse{
		Email: user.Email,
	}

	resp.JSON(w, http.StatusCreated, response)
}

func (h *UserHandler) ConfirmUser(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	confirmation := vars.Get("confirmation")
	if confirmation == "" {
		returnNotFound(w, r)
		return
	}

	if err := h.userService.ConfirmUser(r.Context(), confirmation); err != nil {
		returnNotFound(w, r)
		return
	}

	http.ServeFile(w, r, "internal/ports/api/template/email_confirmation_response.html")
}

func returnNotFound(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "internal/ports/api/template/not_found.html")
}