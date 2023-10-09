package company

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/spo-iitk/ras-backend/util"
)

func getAllCompaniesHandler(ctx *gin.Context) {
	var companies []Company

	err := getAllCompanies(ctx, &companies)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return

	}
	ctx.JSON(http.StatusOK, companies)
}

func getCompanyHandler(ctx *gin.Context) {
	var company Company

	cid, err := strconv.ParseUint(ctx.Param("cid"), 10, 32)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = getCompany(ctx, &company, uint(cid))

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, company)
}

func getLimitedCompaniesHandler(ctx *gin.Context) {
	var companies []Company

	pageSize := ctx.DefaultQuery("pageSize", "100")
	lastFetchedId := ctx.Query("lastFetchedId")
	pageSizeInt, err := util.ParseUint(pageSize)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	lastFetchedIdInt, err := util.ParseUint(lastFetchedId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = getLimitedCompanies(ctx, &companies, uint(lastFetchedIdInt), int(pageSizeInt))

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, companies)
}
