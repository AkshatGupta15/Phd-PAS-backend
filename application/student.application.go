package application

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spo-iitk/ras-backend/mail"
	"github.com/spo-iitk/ras-backend/rc"
	"github.com/spo-iitk/ras-backend/util"
)

type getApplicationResponse struct {
	ApplicationQuestion
	Answer string `json:"answer"`
}

func getApplicationHandler(ctx *gin.Context) {
	pid, err := util.ParseUint(ctx.Param("pid"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sid := getStudentRCID(ctx)
	if sid == 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var questions []getApplicationResponse
	err = fetchApplicationQuestionsAnswers(ctx, pid, sid, &questions)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, questions)
}

type postApplicationRequest struct {
	ResumeID uint                        `json:"resume_id" binding:"required"`
	Answers  []ApplicationQuestionAnswer `json:"answers" binding:"required"`
}

func postApplicationHandler(mail_channel chan mail.Mail) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		pid, err := util.ParseUint(ctx.Param("pid"))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		eid, err := fetchApplicationEventID(ctx, pid)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		sid := getStudentRCID(ctx)
		if sid == 0 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "SRCID not found"})
			return
		}

		proformaEligibility, cpiEligibility, cid, deadline, err := getEligibility(ctx, pid)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		eligible, err := rc.GetStudentEligible(ctx, sid, proformaEligibility, cpiEligibility)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !eligible {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Not eligible to apply"})
			return
		}

		applicationCount, err := getCurrentApplicationCount(ctx, sid)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		applicationMaxCount, err := rc.GetMaxCountfromRC(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if applicationCount >= int(applicationMaxCount) {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Application count maxed out"})
			return
		}

		if time.Now().UnixMilli() > int64(deadline) {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Application deadline passed"})
			return
		}

		var application = EventStudent{
			ProformaEventID:           eid,
			StudentRecruitmentCycleID: sid,
			CompanyRecruitmentCycleID: cid,
			Present:                   true,
		}

		var request postApplicationRequest
		err = ctx.ShouldBindJSON(&request)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var answers []ApplicationQuestionAnswer

		for _, ans := range request.Answers {
			answer := ApplicationQuestionAnswer{
				ApplicationQuestionID:     ans.ApplicationQuestionID,
				StudentRecruitmentCycleID: sid,
				Answer:                    ans.Answer,
			}
			answers = append(answers, answer)
		}

		resumeLink, err := rc.FetchResume(ctx, request.ResumeID, sid)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resume := ApplicationResume{
			StudentRecruitmentCycleID: sid,
			ProformaID:                pid,
			ResumeID:                  request.ResumeID,
			Resume:                    resumeLink,
		}

		err = createApplication(ctx, &application, &answers, &resume)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// {
		// 	// Mail to SPO adminisration
		// 	var proforma Proforma
		// 	err = fetchProforma(ctx, pid, &proforma)
		// 	if err != nil {
		// 		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		// 		return
		// 	}

		// 	if(len(resumeLink)-37 <= 14) {
		// 		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid resume link"})
		// 		return
		// 	}
		// 	student := resumeLink[14 : len(resumeLink)-37]

		// 	loc, _ := time.LoadLocation("Asia/Kolkata")

		// 	msg := "Dear user,\n\n" + "This is to inform you that " + student +
		// 			" has applied on the " + proforma.Profile + " in " + proforma.CompanyName + " at "
		// 	msg += time.Now().In(loc).Format("2006-01-02 15:04") + ".\n"

		// 	recipeints := []string{"spochair@iitk.ac.in", "spovc@iitk.ac.in", "ksnayak@iitk.ac.in", "spooffice@iitk.ac.in"}

		// 	mail_channel <- mail.GenerateMails(recipeints, "New Application by student", msg)
		// }

		logrus.Infof("Application for %d submitted against Performa %d with application ID %v", sid, pid, application.ID)
		ctx.JSON(http.StatusOK, gin.H{"status": "application submitted with id: " + fmt.Sprint(application.ID)})
	}
}

func deleteApplicationHandler(ctx *gin.Context) {
	pid, err := util.ParseUint(ctx.Param("pid"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, _, _, deadline, err := getEligibility(ctx, pid)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if time.Now().UnixMilli() > int64(deadline) {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Application deadline passed"})
		return
	}

	sid := getStudentRCID(ctx)
	if sid == 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "SRCID not found"})
		return
	}

	err = deleteApplication(ctx, pid, sid)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logrus.Infof("Application for %d deleted against Performa %d", sid, pid)
	ctx.JSON(http.StatusOK, gin.H{"status": "application deleted"})
}

func getEventHandler(ctx *gin.Context) {
	eid, err := util.ParseUint(ctx.Param("eid"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var event ProformaEvent
	err = fetchEvent(ctx, eid, &event)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, event)
}

func viewApplicationsHandler(ctx *gin.Context) {
	sid := getStudentRCID(ctx)
	if sid == 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "SRCID not found"})
		return
	}

	var response []ViewApplicationsResponse
	err := fetchApplications(ctx, sid, &response)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}
