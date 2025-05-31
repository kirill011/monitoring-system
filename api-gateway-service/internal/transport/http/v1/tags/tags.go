package tags

import (
	"api-gateway-service/internal/transport/natshandlers/tags"
	pbtags "api-gateway-service/proto/api-gateway/tags"
	"errors"
	"fmt"
	"strings"

	jsoniter "github.com/json-iterator/go"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

const (
	localID = "localID"
)

type tagsHandler struct {
	natsHandlers *tags.TagsHandler
	jwtKey       string
}

type Config struct {
	JWTKey       string
	NatsHandlers *tags.TagsHandler
}

func NewTagsHandler(cfg *Config) *tagsHandler {
	return &tagsHandler{
		jwtKey:       cfg.JWTKey,
		natsHandlers: cfg.NatsHandlers,
	}
}

func (h *tagsHandler) InitTagsRoutes(api fiber.Router) {
	servicesRoute := api.Group("/tags", h.deserializeMW)
	servicesRoute.Post("/create", h.create)
	servicesRoute.Get("/read", h.read)
	servicesRoute.Put("/update", h.update)
	servicesRoute.Delete("/delete", h.delete)

}

type (
	createReq struct {
		Name           string `form:"name"         		json:"name"         	validate:"required"       				xml:"name"`
		DeviceID       int32  `form:"device_id"    		json:"device_id"    	validate:"required"       				xml:"device_id"`
		Regexp         string `form:"regexp"       		json:"regexp"       	validate:"required"          			xml:"regexp"`
		CompareType    string `form:"compare_type" 		json:"compare_type" 	validate:"required,oneof='<' '>' '='"   xml:"compare_type"`
		Value          string `form:"value"        		json:"value"        	validate:"required"       				xml:"value"`
		ArrayIndex     int32  `form:"array_index"  		json:"array_index"  	validate:"required"       				xml:"array_index"`
		Subject        string `form:"subject"      		json:"subject"      	validate:"required"	                    xml:"subject"`
		ServinityLevel string `form:"servinity_level" 	json:"servinity_level" 	validate:"omitempty"					xml:"servinity_level"`
	}

	createResp struct {
		Data pbtags.CreateResp `json:"data"`
	}
)

func (h *tagsHandler) create(ctx fiber.Ctx) error {
	body := createReq{
		Name:     "",
		DeviceID: 0,
		Regexp:   "",
	}

	if err := ctx.Bind().Body(&body); err != nil {
		return fiber.NewError(
			fiber.StatusUnprocessableEntity,
			fmt.Errorf("ctx.Bind().Body: %w", err).Error(),
		)
	}

	res, err := h.natsHandlers.PublishCreate(
		pbtags.CreateReq{
			Tag: &pbtags.Tag{
				Name:           body.Name,
				DeviceID:       body.DeviceID,
				Regexp:         body.Regexp,
				CompareType:    body.CompareType,
				Value:          body.Value,
				ArrayIndex:     body.ArrayIndex,
				Subject:        body.Subject,
				ServinityLevel: body.ServinityLevel,
			},
		},
	)
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("h.nats.PublishCreate: %w", err).Error(),
		)
	}

	jsonResponse, err := jsoniter.Marshal(
		&createResp{
			Data: res,
		},
	)
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
	read struct {
		Data pbtags.ReadResp `json:"data"`
	}
)

func (h *tagsHandler) read(ctx fiber.Ctx) error {
	idLocals := ctx.Locals(localID)
	_, ok := idLocals.(int) //nolint:varnamelen
	if !ok {
		return fiber.NewError(
			fiber.StatusUnauthorized,
			errors.New("idLocals.(int): invalid token").Error(),
		)
	}

	res, err := h.natsHandlers.PublishRead()
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("h.natsHandlers.PublishRead: %w", err).Error(),
		)
	}

	jsonResponse, err := jsoniter.Marshal(
		&read{
			Data: res,
		},
	)
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
		ID             int32  `form:"id"           		json:"id"           	validate:"required"        						xml:"id"`
		Name           string `form:"name"         		json:"name"         	validate:"omitempty"       						xml:"name"`
		DeviceID       int32  `form:"device_id"    		json:"device_id"    	validate:"omitempty"       						xml:"device_id"`
		Regexp         string `form:"regexp"       		json:"regexp"       	validate:"omitempty"       						xml:"regexp"`
		CompareType    string `form:"compare_type" 		json:"compare_type" 	validate:"omitempty,oneof='<' '>' '='"       	xml:"compare_type"`
		Value          string `form:"value"        		json:"value"        	validate:"omitempty"       						xml:"value"`
		ArrayIndex     int32  `form:"array_index"  		json:"array_index"  	validate:"omitempty"       						xml:"array_index"`
		Subject        string `form:"subject"      		json:"subject"      	validate:"omitempty"							xml:"subject"`
		ServinityLevel string `form:"servinity_level" 	json:"servinity_level" 	validate:"omitempty"							xml:"servinity_level"`
	}

	updateResp struct {
		Data int `json:"data"`
	}
)

func (h *tagsHandler) update(ctx fiber.Ctx) error {
	idLocals := ctx.Locals(localID)
	_, ok := idLocals.(int) //nolint:varnamelen
	if !ok {
		return fiber.NewError(
			fiber.StatusUnauthorized,
			errors.New("idLocals.(int): invalid token").Error(),
		)
	}

	body := updateReq{
		ID:       0,
		Name:     "",
		DeviceID: 0,
		Regexp:   "",
	}

	if err := ctx.Bind().Body(&body); err != nil {
		return fiber.NewError(
			fiber.StatusUnprocessableEntity,
			fmt.Errorf("ctx.Bind().Body: %w", err).Error(),
		)
	}

	if body.Name == "" && body.DeviceID == 0 &&
		body.Regexp == "" && body.CompareType == "" && body.Value == "" &&
		body.ArrayIndex == 0 && body.Subject == "" && body.ServinityLevel == "" {
		return fiber.NewError(
			fiber.StatusBadRequest,
			errors.New(`body.Name == "" && body.DeviceID == 0 && body.Regexp == "" && 
			body.CompareType == "" && body.Value == "" && body.ArrayIndex == 0 && 
			body.Subject == "" && body.ServinityLevel == ""`).Error(),
		)
	}

	err := h.natsHandlers.PublishUpdate(
		pbtags.UpdateReq{
			Tag: &pbtags.Tag{
				ID:             body.ID,
				Name:           body.Name,
				DeviceID:       body.DeviceID,
				Regexp:         body.Regexp,
				CompareType:    body.CompareType,
				Value:          body.Value,
				ArrayIndex:     body.ArrayIndex,
				Subject:        body.Subject,
				ServinityLevel: body.ServinityLevel,
			},
		},
	)
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("h.userService.Update: %w", err).Error(),
		)
	}

	jsonResponse, err := jsoniter.Marshal(
		&updateResp{
			Data: fiber.StatusOK,
		},
	)
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
	deleteReq struct {
		ID int `form:"id"     json:"id"     validate:"required"            xml:"id"`
	}

	deleteResp struct {
		Data int `json:"data"`
	}
)

func (h *tagsHandler) delete(ctx fiber.Ctx) error {
	idLocals := ctx.Locals(localID)
	_, ok := idLocals.(int) //nolint:varnamelen
	if !ok {
		return fiber.NewError(
			fiber.StatusUnauthorized,
			errors.New("idLocals.(int): invalid token").Error(),
		)
	}

	body := deleteReq{
		ID: 0,
	}

	if err := ctx.Bind().Body(&body); err != nil {
		return fiber.NewError(
			fiber.StatusUnprocessableEntity,
			fmt.Errorf("ctx.Bind().Body: %w", err).Error(),
		)
	}

	err := h.natsHandlers.PublishDelete(
		pbtags.DeleteReq{
			ID: int32(body.ID),
		},
	)
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("h.natsHandlers.PublishDelete: %w", err).Error(),
		)
	}

	jsonResponse, err := jsoniter.Marshal(
		&deleteResp{
			Data: fiber.StatusOK,
		},
	)
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

func (h *tagsHandler) deserializeMW(ctx fiber.Ctx) error {
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
