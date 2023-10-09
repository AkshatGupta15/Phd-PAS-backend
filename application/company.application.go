package application

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spo-iitk/ras-backend/rc"
	"github.com/spo-iitk/ras-backend/util"
)

type studentCompanysideResponse struct {
	ID                           uint    `json:"id"`
	Name                         string  `json:"name"`
	Email                        string  `json:"email"`
	RollNo                       string  `json:"roll_no"`
	CPI                          float64 `json:"cpi"`
	ProgramDepartmentID          uint    `json:"program_department_id"`
	SecondaryProgramDepartmentID uint    `json:"secondary_program_department_id"`
	Resume                       string  `json:"resume"`
	StatusName                   string  `json:"status_name"`
	Frozen                       bool    `json:"frozen"`
}

func getStudentsForCompanyByRole(ctx *gin.Context) {
	pid, err := util.ParseUint(ctx.Param("pid"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var applied []ApplicantsByRole
	err = fetchApplicantDetails(ctx, pid, &applied)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var srids []uint
	for _, applicant := range applied {
		srids = append(srids, applicant.StudentRCID)
	}

	var allStudentsRC []rc.StudentRecruitmentCycle
	err = rc.FetchStudentBySRID(ctx, srids, &allStudentsRC)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var allStudentsRCMap = make(map[uint]*rc.StudentRecruitmentCycle)
	for i := range allStudentsRC {
		allStudentsRCMap[allStudentsRC[i].ID] = &allStudentsRC[i]
	}

	var validApplicants []studentCompanysideResponse
	for _, s := range applied {
		if allStudentsRCMap[s.StudentRCID].IsFrozen {
			continue
		}

		applicant_details := studentCompanysideResponse{}
		applicant_details.ID = s.StudentRCID
		applicant_details.Resume = s.ResumeLink
		applicant_details.StatusName = s.Name

		studentRC := allStudentsRCMap[s.StudentRCID]

		applicant_details.Name = studentRC.Name
		applicant_details.Email = studentRC.Email
		applicant_details.RollNo = studentRC.RollNo
		applicant_details.CPI = studentRC.CPI
		applicant_details.ProgramDepartmentID = studentRC.ProgramDepartmentID
		applicant_details.SecondaryProgramDepartmentID = studentRC.SecondaryProgramDepartmentID

		validApplicants = append(validApplicants, applicant_details)
	}

	ctx.JSON(http.StatusOK, validApplicants)
}
