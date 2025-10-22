package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/joekings2k/logistics-eta/db/sqlc"
	"github.com/joekings2k/logistics-eta/token"
	"github.com/joekings2k/logistics-eta/util"
	"github.com/lib/pq"
)


type CreateVehicleRequest struct {
	LicensePlate string `json:"license_plate" binding:"required"`
	Model string `json:"model" binding:"required"`
	ImageUrl string `json:"image_url"`
	Capacity int32 `json:"capacity"`
}

type CreateVehicleResponse struct {
	ID uuid.UUID `json:"id"`
	DriverID uuid.UUID `json:"driver_id"`
	LicensePlate string `json:"license_plate"`
	Model string `json:"model"`
	ImageUrl string `json:"image_url"`
	Capacity int32 `json:"capacity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (server *Server)	CreateVehicle(ctx *gin.Context) {
	var req CreateVehicleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil{
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	user, err := server.store.GetUserByID(ctx, authPayload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if user.Role != string(util.RoleDriver) {
		err := errors.New("user is not a driver")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	arg:= db.CreateVehicleParams{
		ID:uuid.New(),
		DriverID: user.ID,
		LicensePlate: req.LicensePlate,
		Model: sql.NullString{String: req.Model, Valid: true},
		ImageUrl: sql.NullString{String: req.ImageUrl, Valid: true},
		Capacity: sql.NullInt32{Int32: req.Capacity, Valid: true},
	}

	vehicle, err := server.store.CreateVehicle(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error);ok{
			switch pqErr.Code.Name(){
			case "unique_violation":
				ctx.JSON(http.StatusConflict, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	response := newVehicleResponse(vehicle)
	ctx.JSON(http.StatusOK, response)

}


func newVehicleResponse (vehicle db.Vehicle) CreateVehicleResponse {
	return CreateVehicleResponse{
		ID: vehicle.ID,
		DriverID: vehicle.DriverID,
		LicensePlate: vehicle.LicensePlate,
		Model: vehicle.Model.String,
		ImageUrl: vehicle.ImageUrl.String,
		Capacity: vehicle.Capacity.Int32,
		CreatedAt: vehicle.CreatedAt.Time,
		UpdatedAt: vehicle.UpdatedAt.Time,
	}
}

type GetVehicleRequest struct {
	ID string `uri:"id" binding:"required"`
}
func (server *Server)GetVehicle(ctx*gin.Context) {
	var req GetVehicleRequest
	if err := ctx.ShouldBindUri(&req); err != nil{
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	vehicleID, err := uuid.Parse(req.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	vehicle, err := server.store.GetVehicleByID(ctx, vehicleID)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	fmt.Print(
		"error hrer")
	if vehicle.DriverID != authPayload.UserID {
		err := errors.New("vehicle does not belong to user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, newVehicleResponse(vehicle))
}


type GetVehiclesByDriverIDParams struct {
	PageID int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server)GetVehiclesByDriverID(ctx*gin.Context) {
	var req GetVehiclesByDriverIDParams
	if err := ctx.ShouldBindQuery(&req); err != nil{
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.GetVehiclesByDriverIDParams{
		DriverID: authPayload.UserID,
		Limit: req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}
	vehicles, err := server.store.GetVehiclesByDriverID(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, vehicles)
}
