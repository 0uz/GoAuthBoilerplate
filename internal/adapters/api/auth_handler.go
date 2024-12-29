package api

import (
	"encoding/json"
	"net/http"

	resp "github.com/ouz/goauthboilerplate/internal/adapters/api/response"
	"github.com/ouz/goauthboilerplate/internal/adapters/api/util"
	"github.com/ouz/goauthboilerplate/internal/adapters/api/validator"
	authDto "github.com/ouz/goauthboilerplate/internal/application/auth/dto"
	authService "github.com/ouz/goauthboilerplate/internal/domain/auth"
	"github.com/ouz/goauthboilerplate/pkg/errors"
)

type AuthHandler struct {
	authService authService.AuthService
}

func NewAuthHandler(authService authService.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) RefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	clientSecret := util.ExtractClientSecret(r)
	if clientSecret == "" {
		resp.Error(w, errors.UnauthorizedError("x-client-key header is required", nil))
		return
	}

	var request authDto.RefreshAccessTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		resp.Error(w, errors.BadRequestError("Invalid request body"))
		return
	}

	if request.RefreshToken == "" {
		resp.Error(w, errors.BadRequestError("Refresh token is required"))
		return
	}

	tokens, err := h.authService.RefreshAccessToken(r.Context(), request.RefreshToken, clientSecret)
	if err != nil {
		resp.Error(w, err)
		return
	}

	response := authDto.TokenResponse{
		AccessToken:  tokens[0].Token,
		RefreshToken: tokens[1].Token,
	}

	resp.JSON(w, http.StatusCreated, response)
}

func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	clientSecret := util.ExtractClientSecret(r)
	if clientSecret == "" {
		resp.Error(w, errors.UnauthorizedError("x-client-key header is required", nil))
		return
	}
	var request authDto.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		resp.Error(w, errors.BadRequestError("Invalid request body"))
		return
	}

	if err := validator.Validator.Struct(request); err != nil {
		resp.Error(w, errors.BadRequestError(err.Error()))
		return
	}

	tokens, err := h.authService.Login(r.Context(), request.Email, request.Password, clientSecret)
	if err != nil {
		resp.Error(w, err)
		return
	}

	response := authDto.TokenResponse{
		AccessToken:  tokens[0].Token,
		RefreshToken: tokens[1].Token,
	}

	resp.JSON(w, http.StatusCreated, response)
}

func (h *AuthHandler) LoginAnonymousUser(w http.ResponseWriter, r *http.Request) {
	clientSecret := util.ExtractClientSecret(r)
	if clientSecret == "" {
		resp.Error(w, errors.UnauthorizedError("x-client-key header is required", nil))
		return
	}

	var request authDto.AnonymousUserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		resp.Error(w, errors.BadRequestError("Invalid request body"))
		return
	}

	if err := validator.Validator.Struct(request); err != nil {
		resp.Error(w, errors.BadRequestError(err.Error()))
		return
	}

	tokens, err := h.authService.LoginAnonymous(r.Context(), request.Email, clientSecret)
	if err != nil {
		resp.Error(w, err)
		return
	}

	response := authDto.TokenResponse{
		AccessToken:  tokens[0].Token,
		RefreshToken: tokens[1].Token,
	}

	resp.JSON(w, http.StatusCreated, response)
}

func (h *AuthHandler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	user := util.GetAuthenticatedUser(r)
	if user == nil {
		resp.Error(w, errors.UnauthorizedError("User not found", nil))
		return
	}

	if err := h.authService.Logout(r.Context(), user.ID); err != nil {
		resp.Error(w, err)
		return
	}

	resp.JSON(w, http.StatusOK, nil)
}