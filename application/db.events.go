package application

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

func fetchEventsByRC(ctx *gin.Context, rid uint, events *[]getAllEventsByRCResponse) error {
	tx := db.WithContext(ctx).Model(&ProformaEvent{}).
		Joins("JOIN proformas ON proformas.id = proforma_events.proforma_id").
		Where("proformas.deleted_at IS NULL AND proformas.recruitment_cycle_id = ?", rid).
		Order("start_time DESC, proforma_id, sequence").
		Select("proforma_events.*, proformas.company_name, proformas.role, proformas.profile").
		Find(events)
	return tx.Error
}

func fetchEvent(ctx *gin.Context, id uint, event *ProformaEvent) error {
	tx := db.WithContext(ctx).Where("id = ?", id).Order("sequence").First(event)
	return tx.Error
}

func fetchEventsByProforma(ctx *gin.Context, pid uint, events *[]ProformaEvent) error {
	tx := db.WithContext(ctx).Where("proforma_id = ?", pid).Order("sequence").Find(events)
	return tx.Error
}

func createEvent(ctx *gin.Context, event *ProformaEvent) error {
	tx := db.WithContext(ctx).Where("proforma_id = ? AND name = ?", event.ProformaID, event.Name).FirstOrCreate(event)
	return tx.Error
}

func updateEvent(ctx *gin.Context, event *ProformaEvent) error {
	tx := db.WithContext(ctx).Clauses(clause.Returning{}).Where("id = ?", event.ID).Updates(event)
	return tx.Error
}

func updateEventCalID(event *ProformaEvent) error {
	tx := db.Clauses(clause.Returning{}).Where("id = ?", event.ID).Updates(event)
	return tx.Error
}

func deleteEvent(ctx *gin.Context, id uint) error {
	tx := db.WithContext(ctx).Where("id = ?", id).Delete(&ProformaEvent{})
	return tx.Error
}

func fetchEventsByStudent(ctx *gin.Context, rid uint, events *[]proformaEventStudentResponse) error {
	tx := db.WithContext(ctx).
		Model(&ProformaEvent{}).
		Joins("JOIN proformas ON proformas.id=proforma_events.proforma_id AND proformas.recruitment_cycle_id = ?", rid).
		Where("proforma_events.start_time > 0").
		Order("start_time DESC, proforma_id, sequence").
		Select("proforma_events.*, proformas.company_name, proformas.role, proformas.profile").
		Find(events)
	return tx.Error
}
