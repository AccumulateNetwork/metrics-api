package api

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/AccumulateNetwork/metrics-api/schema"
	"github.com/AccumulateNetwork/metrics-api/store"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const DefaultPaginationStart = 0
const DefaultPaginationCount = 10
const ErrorCodeBadPagination = -1000
const ErrorCodeNothingFound = -1001
const ErrorMessageNothingFound = "searched item not found"

type API struct {
	HTTP     *echo.Echo
	Validate *validator.Validate
}

type PaginationParams struct {
	Start int    `json:"start" validate:"min=0"`
	Count int    `json:"count" validate:"min=0"`
	Order string `json:"order"`
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
	schema.ACME
	Staked            int64      `json:"staked"`
	Circulating       int64      `json:"circulating"`
	TotalTokens       float64    `json:"totalTokens"`
	MaxTokens         float64    `json:"maxTokens"`
	StakedTokens      float64    `json:"stakedTokens"`
	CirculatingTokens float64    `json:"circulatingTokens"`
	UpdatedAt         *time.Time `json:"updatedAt"`
}

type StakingResponse struct {
	schema.ValidatorsNumber
}
type StakersResponse struct {
	Result []*schema.StakingRecord `json:"result"`
	PaginationResponse
}

type ValidatorsResponse struct {
	Result []*schema.Validator `json:"result"`
	PaginationResponse
}

type TokensResponse struct {
	Result []*schema.Token `json:"result"`
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
	publicAPI.GET("/supply/:filter", api.getSupply)
	publicAPI.GET("/staking", api.getStaking)
	publicAPI.GET("/staking/stakers", api.getStakers)
	publicAPI.POST("/staking/stakers/search", api.searchStaker)
	publicAPI.GET("/validators", api.getValidators)
	publicAPI.POST("/validators/search", api.searchValidator)
	publicAPI.GET("/tokens", api.getTokens)
	publicAPI.POST("/tokens/search", api.searchToken)

	api.HTTP.Logger.Fatal(api.HTTP.Start(":" + strconv.Itoa(port)))

	return nil

}

// GetPaginationParams parses and validates pagination params
func (api *API) GetPaginationParams(c echo.Context) (*PaginationParams, error) {

	params := &PaginationParams{Start: DefaultPaginationStart, Count: DefaultPaginationCount, Order: schema.DefaultOrder}

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

	if c.QueryParam("order") == schema.AlternativeOrder {
		params.Order = schema.AlternativeOrder
	}

	if err := api.Validate.Struct(params); err != nil {
		return nil, err
	}

	return params, nil

}

// getSupply returns ACME supply
func (api *API) getSupply(c echo.Context) error {

	res := &SupplyResponse{ACME: *store.ACME}

	res.Staked = store.GetTotalStake()
	res.Circulating = res.Total - store.FoundationTotalBalance

	res.TotalTokens = math.Round(float64(res.Total) * math.Pow10(-1*int(res.Precision)))
	res.MaxTokens = math.Round(float64(res.Max) * math.Pow10(-1*int(res.Precision)))
	res.CirculatingTokens = math.Round(float64(res.Circulating) * math.Pow10(-1*int(res.Precision)))
	res.StakedTokens = math.Round(float64(res.Staked) * math.Pow10(-1*int(res.Precision)))

	res.UpdatedAt = store.UpdatedAt

	switch c.Param("filter") {
	case "total":
		return c.String(http.StatusOK, fmt.Sprintf("%.f", res.TotalTokens))
	case "max":
		return c.String(http.StatusOK, fmt.Sprintf("%.f", res.MaxTokens))
	case "circulating":
		return c.String(http.StatusOK, fmt.Sprintf("%.f", res.CirculatingTokens))
	case "staked":
		return c.String(http.StatusOK, fmt.Sprintf("%.f", res.StakedTokens))
	}

	return c.JSON(http.StatusOK, res)

}

// getStaking returns staking metrics
func (api *API) getStaking(c echo.Context) error {

	validators := store.GetValidatorsNumber()

	res := &StakingResponse{ValidatorsNumber: *validators}

	return c.JSON(http.StatusOK, res)

}

// getStakers returns stakers
func (api *API) getStakers(c echo.Context) error {

	params, err := api.GetPaginationParams(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &ErrorResponse{Code: ErrorCodeBadPagination, Error: err.Error()})
	}

	res := &StakersResponse{}

	stakers := &schema.StakingRecords{}
	copier.Copy(&stakers.Items, store.StakingRecords.Items)

	if c.QueryParam("sort") != "" {
		stakers.Sort(c.QueryParam("sort"), params.Order)
	}

	lastElementIndex := params.Start + params.Count
	if lastElementIndex > len(stakers.Items) {
		lastElementIndex = len(stakers.Items)
	}

	res.Result = stakers.Items[params.Start:lastElementIndex]
	res.Start = params.Start
	res.Count = params.Count
	res.Total = len(stakers.Items)

	return c.JSON(http.StatusOK, res)

}

// searchStaker returns tokens
func (api *API) searchStaker(c echo.Context) error {

	res := &schema.StakingRecord{}

	if c.FormValue("stake") != "" {
		res = store.SearchStakingRecordByAccount(c.FormValue("stake"), store.StakingRecords.Items)
	}

	if res == nil || res.Identity == "" {
		return c.JSON(http.StatusOK, &ErrorResponse{Code: ErrorCodeNothingFound, Error: ErrorMessageNothingFound})
	}

	return c.JSON(http.StatusOK, res)

}

// getValidators returns validators
func (api *API) getValidators(c echo.Context) error {

	params, err := api.GetPaginationParams(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &ErrorResponse{Code: ErrorCodeBadPagination, Error: err.Error()})
	}

	res := &ValidatorsResponse{}

	validators := store.GetValidators()

	if c.QueryParam("sort") != "" {
		validators.Sort(c.QueryParam("sort"), params.Order)
	}

	lastElementIndex := params.Start + params.Count
	if lastElementIndex > len(validators.Items) {
		lastElementIndex = len(validators.Items)
	}

	res.Result = validators.Items[params.Start:lastElementIndex]
	res.Start = params.Start
	res.Count = params.Count
	res.Total = len(validators.Items)

	return c.JSON(http.StatusOK, res)

}

// searchValidator returns tokens
func (api *API) searchValidator(c echo.Context) error {

	res := &schema.Validator{}

	if c.FormValue("identity") != "" {
		res = store.SearchValidatorByIdentity(c.FormValue("identity"), store.GetValidators().Items)
	}

	if res == nil || res.Identity == "" {
		return c.JSON(http.StatusOK, &ErrorResponse{Code: ErrorCodeNothingFound, Error: ErrorMessageNothingFound})
	}

	return c.JSON(http.StatusOK, res)

}

// getTokens returns tokens
func (api *API) getTokens(c echo.Context) error {

	params, err := api.GetPaginationParams(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &ErrorResponse{Code: ErrorCodeBadPagination, Error: err.Error()})
	}

	res := &TokensResponse{}

	tokens := &schema.Tokens{}
	copier.Copy(&tokens.Items, store.Tokens.Items)

	lastElementIndex := params.Start + params.Count
	if lastElementIndex > len(tokens.Items) {
		lastElementIndex = len(tokens.Items)
	}

	res.Result = tokens.Items[params.Start:lastElementIndex]
	res.Start = params.Start
	res.Count = params.Count
	res.Total = len(tokens.Items)

	return c.JSON(http.StatusOK, res)

}

// searchToken returns tokens
func (api *API) searchToken(c echo.Context) error {

	res := &schema.Token{}

	if c.FormValue("symbol") != "" {
		res = store.SearchTokenBySymbol(c.FormValue("symbol"), store.Tokens.Items)
	}

	if res == nil || res.TokenIssuer == "" {
		return c.JSON(http.StatusOK, &ErrorResponse{Code: ErrorCodeNothingFound, Error: ErrorMessageNothingFound})
	}

	return c.JSON(http.StatusOK, res)

}
