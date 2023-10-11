package rc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/spo-iitk/ras-backend/util"
)

type ResumeRequest struct {
	Resume string `json:"resume"`
}

type ResumeDeleteRequest struct {
	Filename string `json:"filename"`
	Secret   string `json:"secret"`
}

func postStudentResumeHandler(ctx *gin.Context) {
	rid, err := util.ParseUint(ctx.Param("rid"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var request ResumeRequest
	err = ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sid := getStudentRCID(ctx)
	if sid == 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "SRCID not found"})
		return
	}

	// var resumes []StudentRecruitmentCycleResume
	// err = fetchStudentResume(ctx, sid, &resumes)
	// if err != nil {
	// 	ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }

	// if len(resumes) > 0 {
	// 	ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Delete current resume before adding new one"})
	// }

	err = addStudentResume(ctx, request.Resume, sid, rid)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "Resume Added Successfully"})
}

func getStudentResumeHandler(ctx *gin.Context) {
	sid := getStudentRCID(ctx)
	if sid == 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "SRCID not found"})
		return
	}

	var resumes []StudentRecruitmentCycleResume
	err := fetchStudentResume(ctx, sid, &resumes)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resumes)
}

func deleteStudentResumeHandler(ctx *gin.Context) {
	rid, err := util.ParseUint(ctx.Param("rid"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sid := getStudentRCID(ctx)
	if sid == 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "SRCID not found"})
		return
	}

	var resumes []StudentRecruitmentCycleResume
	err = fetchStudentResume(ctx, sid, &resumes)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	errorChan := make(chan error)

	go func() {
		url := viper.GetString("CDN.URL")+"/delete"
		reqBodyString := fmt.Sprintf(`{"Filename":"%s","Secret":"%s"}`,resumes[0].Resume,viper.GetString("CDN.secret"))
		reqBody := []byte(reqBodyString)

		timeout := 8 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		httpClient := &http.Client{}

		req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(reqBody))
		if err != nil {
			errorChan <- err
			return
		}
		req = req.WithContext(ctx)

		req.Header.Set("Content-Type", "application/json")
		log.Print("request => ",req)
		response, err := httpClient.Do(req)
		log.Print(response,err)
		if err != nil {
			errorChan <- err
			return
		}
		defer response.Body.Close()

		_, err = io.ReadAll(response.Body)
		if err != nil {
			errorChan <- err
			return
		}

		errorChan <- nil
	}()

	err = <-errorChan
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = deleteStudentResume(ctx, sid, rid)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "Resume Deleted Successfully"})
}
