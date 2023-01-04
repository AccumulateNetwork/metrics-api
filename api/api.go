package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/AccumulateNetwork/metrics-api/schema"
	"github.com/AccumulateNetwork/metrics-api/store"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const DefaultPaginationStart = 0
const DefaultPaginationCount = 10

type API struct {
	HTTP     *echo.Echo
	Validate *validator.Validate
}

type PaginationParams struct {
	Start int `json:"start" validate:"min=0"`
	Count int `json:"count" validate:"min=0"`
}

type PaginationResponse struct {
	PaginationParams
	Total int `json:"total"`
}

type ErrorResponse struct {
	Result bool   `json:"result"`
	Code   int    `json:"code"`
	Error  string `json:"error"`
}
type SupplyResponse struct {
	SupplyLimit int64 `json:"supplyLimit"`
	Issued      int64 `json:"issued"`
	Staked      int64 `json:"staked"`
}

type StakingResponse struct {
	APR              float64 `json:"apr"`
	CoreValidator    int64   `json:"coreValidator"`
	CoreFollower     int64   `json:"coreFollower"`
	StakingValidator int64   `json:"stakingValidator"`
	Delegates        int64   `json:"delegates"`
	PureStakers      int64   `json:"pureStakers"`
}
type StakersResponse struct {
	Result []*schema.StakingRecord `json:"result"`
	PaginationResponse
}

// StartAPI configures and starts REST API server
func StartAPI(port int) error {

	api := &API{}

	api.HTTP = echo.New()
	api.HTTP.HideBanner = true

	// init validator v10
	api.Validate = validator.New()

	// remove trailing slash middleware
	// https://echo.labstack.com/middleware/trailing-slash/
	api.HTTP.Pre(middleware.RemoveTrailingSlash())

	// recover middleware
	// https://echo.labstack.com/middleware/recover/
	api.HTTP.Use(middleware.Recover())

	// logger middleware
	// https://echo.labstack.com/middleware/logger/
	api.HTTP.Use(middleware.Logger())

	// v1 public metrics API
	api.HTTP.GET("/v1", func(c echo.Context) error {
		return c.String(http.StatusOK, "Accumulate Metrics API")
	})
	publicAPI := api.HTTP.Group("/v1")

	publicAPI.GET("/supply", api.getSupply)
	publicAPI.GET("/staking", api.getStaking)
	publicAPI.GET("/staking/stakers", api.getStakers)

	api.HTTP.Logger.Fatal(api.HTTP.Start(":" + strconv.Itoa(port)))

	return nil

}

// GetPaginationParams parses and validates pagination params
func (api *API) GetPaginationParams(c echo.Context) (*PaginationParams, error) {

	params := &PaginationParams{Start: DefaultPaginationStart, Count: DefaultPaginationCount}

	if c.QueryParam("start") != "" {
		start, err := strconv.Atoi(c.QueryParam("start"))
		if err != nil {
			err = fmt.Errorf("'start' expected to be an integer, '%s' received", c.QueryParam("start"))
			log.Error(err)
			return nil, err
		}
		params.Start = start
	}

	if c.QueryParam("count") != "" {
		count, err := strconv.Atoi(c.QueryParam("count"))
		if err != nil {
			err = fmt.Errorf("'limit' expected to be an integer, '%s' received", c.QueryParam("limit"))
			log.Error(err)
			return nil, err
		}
		params.Count = count
	}

	if err := api.Validate.Struct(params); err != nil {
		return nil, err
	}

	return params, nil

}

// getStaking returns staking metrics
func (api *API) getSupply(c echo.Context) error {

	res := &SupplyResponse{}

	return c.JSON(http.StatusOK, res)

}

// getStaking returns staking metrics
func (api *API) getStaking(c echo.Context) error {

	res := &StakingResponse{}

	return c.JSON(http.StatusOK, res)

}

// getStakers returns stakers
func (api *API) getStakers(c echo.Context) error {

	params, err := api.GetPaginationParams(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &ErrorResponse{Code: http.StatusBadGateway, Error: err.Error()})
	}

	res := &StakersResponse{}
	res.Result = store.StakingRecords.Items[params.Start : params.Start+params.Count]
	res.Start = params.Start
	res.Count = params.Count
	res.Total = len(store.StakingRecords.Items)

	return c.JSON(http.StatusOK, res)

}
