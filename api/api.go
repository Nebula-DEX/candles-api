package api

import (
	"candles-api/store"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

const Port = 8889

type ErrorResponse struct {
	Error string `json:"error"`
}

type Api struct {
	store *store.Store
}

func NewApi(
	store *store.Store,
) *Api {
	return &Api{
		store: store,
	}
}

func (a *Api) getCandles(
	c *gin.Context,
	marketId string,
	intervalStr string,
	fromTimestampStr string,
	toTimestampStr string,
) {
	if len(marketId) == 0 {
		c.JSON(http.StatusBadRequest, &ErrorResponse{Error: "marketId required"})
	} else if len(intervalStr) == 0 {
		c.JSON(http.StatusBadRequest, &ErrorResponse{Error: "interval required"})
	} else if len(fromTimestampStr) == 0 {
		c.JSON(http.StatusBadRequest, &ErrorResponse{Error: "fromTimestamp required"})
	} else {
		interval, err1 := strconv.ParseUint(intervalStr, 10, 0)
		fromTimestamp, err2 := strconv.ParseUint(fromTimestampStr, 10, 0)
		toTimestamp := uint64(0)
		var err3 error
		if len(toTimestampStr) > 0 {
			toTimestamp, err3 = strconv.ParseUint(toTimestampStr, 10, 0)
		}
		if err1 != nil {
			c.JSON(http.StatusBadRequest, &ErrorResponse{Error: "interval format invalid"})
		} else if err2 != nil {
			c.JSON(http.StatusBadRequest, &ErrorResponse{Error: "fromTimestamp format invalid"})
		} else if err3 != nil {
			c.JSON(http.StatusBadRequest, &ErrorResponse{Error: "toTimestamp format invalid"})
		} else {
			candles := a.store.GetCandles(marketId, interval, fromTimestamp, toTimestamp)
			c.JSON(http.StatusOK, candles)
		}
	}
}

func (a *Api) Start() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/data/:marketId/:interval/:fromTimestamp/:toTimestamp", func(c *gin.Context) {
		marketId := c.Param("marketId")
		intervalStr := c.Param("interval")
		fromTimestampStr := c.Param("fromTimestamp")
		toTimestampStr := c.Param("toTimestamp")
		a.getCandles(c, marketId, intervalStr, fromTimestampStr, toTimestampStr)
	})
	r.GET("/data/:marketId/:interval/:fromTimestamp", func(c *gin.Context) {
		marketId := c.Param("marketId")
		intervalStr := c.Param("interval")
		fromTimestampStr := c.Param("fromTimestamp")
		a.getCandles(c, marketId, intervalStr, fromTimestampStr, "")

	})
	log.Infof("listening on 0.0.0.0:%d", Port)
	err := r.Run(fmt.Sprintf(":%d", Port))
	if err != nil {
		log.Fatal(err)
	}
}
