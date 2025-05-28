package v1

import (
	"errors"
	"fmt"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"

	"auth-service/internal/services"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

const (
	localID = "localID"
)

func (h *Handler) initAuthRoutes(api fiber.Router) {
	servicesRoute := api.Group("/auth")
	servicesRoute.Post("/register", h.register)
	servicesRoute.Get("/sign_in", h.authorize)

	servicesRoute = api.Group("/users", h.deserializeMW)
	servicesRoute.Get("/read", h.getUsers)
	servicesRoute.Put("/update", h.update)
	servicesRoute.Delete("/delete", h.delete)

}

type (
	registerReq struct {
		Name     string `form:"name"         json:"name"         validate:"required"       xml:"name"`
		Email    string `form:"email"        json:"email"        validate:"required,email" xml:"email"`
		Password string `form:"password"     json:"password"     validate:"required"       xml:"password"`
	}

	registerResp struct {
		Error bool                      `json:"error"`
		Data  services.CreateUserResult `json:"data"`
	}
)

func (h *Handler) register(ctx fiber.Ctx) error {
	body := registerReq{
		Name:     "",
		Email:    "",
		Password: "",
	}

	if err := ctx.Bind().Body(&body); err != nil {
		return fiber.NewError(
			fiber.StatusUnprocessableEntity,
			fmt.Errorf("ctx.Bind().Body: %w", err).Error(),
		)
	}

	res, err := h.authService.Create(ctx.Context(),
		services.CreateUserParams{
			Name:     body.Name,
			Email:    body.Email,
			Password: body.Password,
		})
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("h.userService.Register: %w", err).Error(),
		)
	}

	jsonResponse, err := jsoniter.Marshal(&registerResp{
		Error: false,
		Data:  res,
	})
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("json.Marshal: %w", err).Error(),
		)
	}

	if err = ctx.Status(fiber.StatusOK).Send(jsonResponse); err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("ctx.Send: %w", err).Error(),
		)
	}

	return nil
}

type (
	authorizeReq struct {
		Email    string `form:"email"    json:"email"    validate:"required,email" xml:"email"`
		Password string `form:"password" json:"password" validate:"required"       xml:"password"`
	}

	authorizeResp struct {
		Error bool   `json:"error"`
		JWT   string `json:"jwt"`
	}
)

func (h *Handler) authorize(ctx fiber.Ctx) error {
	body := authorizeReq{
		Email:    "",
		Password: "",
	}

	if err := ctx.Bind().Body(&body); err != nil {
		return fiber.NewError(
			fiber.StatusUnprocessableEntity,
			fmt.Errorf("ctx.Bind().Body: %w", err).Error(),
		)
	}

	userID, err := h.authService.Authorize(ctx.Context(),
		services.AuthorizeParams{
			Email:    body.Email,
			Password: body.Password,
		})
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("h.userService.Authorize: %w", err).Error(),
		)
	}

	jwt := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		localID: userID,
		"exp":   time.Now().Add(h.tokenLifeTime).Unix(),
	})

	token, err := jwt.SignedString([]byte(h.jwtKey))
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("jwt.SignedString([]byte(h.key)): %w", err).Error(),
		)
	}

	jsonResponse, err := jsoniter.Marshal(&authorizeResp{
		Error: false,
		JWT:   token,
	})
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("json.Marshal: %w", err).Error(),
		)
	}

	if err = ctx.Status(fiber.StatusOK).Send(jsonResponse); err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("ctx.Send: %w", err).Error(),
		)
	}

	return nil
}

type (
	getUsers struct {
		Error bool                `json:"error"`
		Data  services.ReadResult `json:"data"`
	}
)

func (h *Handler) getUsers(ctx fiber.Ctx) error {
	idLocals := ctx.Locals(localID)
	userID, ok := idLocals.(int) //nolint:varnamelen
	if !ok {
		return fiber.NewError(
			fiber.StatusUnauthorized,
			errors.New("idLocals.(int): invalid token").Error(),
		)
	}

	res, err := h.authService.Read(ctx.Context(), userID)
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("h.userService.Read: %w", err).Error(),
		)
	}

	jsonResponse, err := jsoniter.Marshal(&getUsers{
		Error: false,
		Data:  res,
	})
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("json.Marshal: %w", err).Error(),
		)
	}

	if err = ctx.Status(fiber.StatusOK).Send(jsonResponse); err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("ctx.Send: %w", err).Error(),
		)
	}
	return nil
}

type (
	updateReq struct {
		NewName string `form:"new_name"  json:"new_name"  validate:"omitempty"      xml:"new_name"`
	}

	updateResp struct {
		Error bool `json:"error"`
		Data  int  `json:"data"`
	}
)

func (h *Handler) update(ctx fiber.Ctx) error {
	idLocals := ctx.Locals(localID)
	userID, ok := idLocals.(int) //nolint:varnamelen
	if !ok {
		return fiber.NewError(
			fiber.StatusUnauthorized,
			errors.New("idLocals.(int): invalid token").Error(),
		)
	}

	body := updateReq{
		NewName: "",
	}

	if err := ctx.Bind().Body(&body); err != nil {
		return fiber.NewError(
			fiber.StatusUnprocessableEntity,
			fmt.Errorf("ctx.Bind().Body: %w", err).Error(),
		)
	}

	updateParams := services.UpdateUsersParams{
		ID:   userID,
		Name: &body.NewName,
	}

	if updateParams.Name == nil {
		return fiber.NewError(
			fiber.StatusBadRequest,
			errors.New("updOpt.Name == nil").Error(),
		)
	}

	err := h.authService.Update(ctx.Context(), updateParams)
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("h.userService.Update: %w", err).Error(),
		)
	}

	jsonResponse, err := jsoniter.Marshal(&updateResp{
		Error: false,
		Data:  fiber.StatusOK,
	})
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("json.Marshal: %w", err).Error(),
		)
	}

	if err = ctx.Status(fiber.StatusOK).Send(jsonResponse); err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("ctx.Send: %w", err).Error(),
		)
	}
	return nil
}

type (
	deleteResp struct {
		Error bool `json:"error"`
		Data  int  `json:"data"`
	}
)

func (h *Handler) delete(ctx fiber.Ctx) error {
	idLocals := ctx.Locals(localID)
	userID, ok := idLocals.(int) //nolint:varnamelen
	if !ok {
		return fiber.NewError(
			fiber.StatusUnauthorized,
			errors.New("idLocals.(int): invalid token").Error(),
		)
	}

	err := h.authService.Delete(ctx.Context(), userID)
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("h.userService.Delete: %w", err).Error(),
		)
	}

	jsonResponse, err := jsoniter.Marshal(&deleteResp{
		Error: false,
		Data:  fiber.StatusOK,
	})
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("json.Marshal: %w", err).Error(),
		)
	}

	if err = ctx.Status(fiber.StatusOK).Send(jsonResponse); err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("ctx.Send: %w", err).Error(),
		)
	}

	return nil
}

func (h *Handler) deserializeMW(ctx fiber.Ctx) error {
	tokenString := ctx.Get("Authorization")

	if tokenString == "" {
		return fiber.NewError(
			fiber.StatusUnauthorized,
			errors.New("tokenString is empty").Error(),
		)
	}

	tokenString = strings.ReplaceAll(tokenString, "Bearer ", "")
	token, err := jwt.Parse(tokenString, func(_ *jwt.Token) (interface{}, error) {
		return []byte(h.jwtKey), nil
	})
	if err != nil {
		return fiber.NewError(
			fiber.StatusUnauthorized,
			fmt.Errorf("jwt.Parse: %w", err).Error(),
		)
	}

	claims, ok := token.Claims.(jwt.MapClaims) //nolint:varnamelen
	if !ok {
		return fiber.NewError(
			fiber.StatusUnauthorized,
			errors.New("token.Claims.(jwt.MapClaims): invalid token").Error(),
		)
	}

	userID, ok := claims[localID].(float64) //nolint:varnamelen
	if !ok {
		return fiber.NewError(
			fiber.StatusUnauthorized,
			errors.New("claims["+localID+"].(float64): invalid token").Error(),
		)
	}

	ctx.Locals(localID, int(userID))

	return ctx.Next() //nolint:wrapcheck
}
