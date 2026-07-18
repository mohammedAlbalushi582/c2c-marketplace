package handler

import (
	"encoding/json"
	"net/http"

	"github.com/alamjad/marketplace/internal/delivery/http/middleware"
	"github.com/alamjad/marketplace/internal/platform/validator"
	authuc "github.com/alamjad/marketplace/internal/usecase/auth"
)

type AuthHandler struct {
	svc *authuc.Service
}

func NewAuthHandler(svc *authuc.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type registerRequest struct {
	Email          string  `json:"email" validate:"required,email"`
	Username       *string `json:"username"`
	Password       string  `json:"password" validate:"required,min=6"`
	FullName       string  `json:"full_name" validate:"required"`
	Phone          *string `json:"phone"`
	WhatsappNumber *string `json:"whatsapp_number"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "invalid json")
		return
	}
	if err := validator.Struct(req); err != nil {
		BadRequest(w, err.Error())
		return
	}
	res, err := h.svc.Register(r.Context(), authuc.RegisterInput{
		Email: req.Email, Username: req.Username, Password: req.Password,
		FullName: req.FullName, Phone: req.Phone, WhatsappNumber: req.WhatsappNumber,
	})
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusCreated, authResp(res))
}

type loginRequest struct {
	Identifier string `json:"identifier" validate:"required"` // email or username
	Password   string `json:"password" validate:"required"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "invalid json")
		return
	}
	if err := validator.Struct(req); err != nil {
		BadRequest(w, err.Error())
		return
	}
	res, err := h.svc.Login(r.Context(), req.Identifier, req.Password)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, authResp(res))
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "invalid json")
		return
	}
	res, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, authResp(res))
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserFromContext(r.Context())
	user, err := h.svc.Me(r.Context(), u.ID)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, toUserDTO(user))
}

func authResp(res *authuc.Result) AuthResponse {
	return AuthResponse{
		User:         toUserDTO(res.User),
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		ExpiresIn:    res.ExpiresIn,
	}
}
