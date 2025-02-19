package storage

import (
	"bitrix/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
)

// CallInfo ma'lumotlarini saqlash
func InsertCallInfo(call *models.CallInfo, db *sql.DB) error {
	query := `
	INSERT INTO CallInfo (
		id, portal_user_id, portal_number, phone_number, call_id, external_call_id,
		call_category, call_duration, call_start_date, call_record_url, call_vote, cost,
		cost_currency, call_failed_code, call_failed_reason, crm_entity_type, crm_entity_id,
		crm_activity_id, rest_app_id, rest_app_name, transcript_id, transcript_pending,
		session_id, redial_attempt, comment, record_duration, record_file_id, call_type
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
		$18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28
	) ON CONFLICT (id) DO NOTHING;`

	result, err := db.Exec(
		query,
		call.ID, call.PortalUserID, call.PortalNumber, call.PhoneNumber, call.CallID,
		call.ExternalCallID, call.CallCategory, call.CallDuration, call.CallStartDate,
		call.CallRecordURL, call.CallVote, call.Cost, call.CostCurrency, call.CallFailedCode,
		call.CallFailedReason, call.CRMEntityType, call.CRMEntityID, call.CRMActivityID,
		call.RestAppID, call.RestAppName, call.TranscriptID, call.TranscriptPending,
		call.SessionID, call.RedialAttempt, call.Comment, call.RecordDuration,
		call.RecordFileID, call.CallType,
	)
	if err != nil {
		log.Printf("❌ CallInfo saqlashda xatolik (ID: %s): %v", call.ID, err)
		return fmt.Errorf("CallInfo saqlab bo‘lmadi: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("⚠️ CallInfo (ID: %s) allaqachon mavjud, qo‘shilmadi", call.ID)
	}

	return nil
}

func InsertUser(user *models.User, db *sql.DB) error {
	// Departmentni JSON formatiga o'tkazamiz
	departmentJSON, err := json.Marshal(user.Department)
	if err != nil {
		return fmt.Errorf("❌ Department JSON serialize qilishda xatolik: %v", err)
	}

	query := `
		INSERT INTO users (
			id, xml_id, active, name, last_name, second_name, email, last_login,
			time_zone, time_zone_offset, personal_photo, personal_gender, personal_www,
			personal_birthday, personal_mobile, personal_city, work_phone, work_position,
			uf_employment_date, user_type, department_ids
		) VALUES (
			$1, $2, $3::VARCHAR, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21
		) ON CONFLICT (id) DO NOTHING;`

	result, err := db.Exec(query, user.ID, user.XML_ID, fmt.Sprintf("%t", user.Active), user.Name, user.LastName, user.SecondName,
		user.Email, user.LastLogin, user.TimeZone, user.TimeZoneOffset, user.PersonalPhoto,
		user.PersonalGender, user.PersonalWWW, user.PersonalBirthday, user.PersonalMobile,
		user.PersonalCity, user.WorkPhone, user.WorkPosition, user.EmploymentDate,
		user.UserType, string(departmentJSON))

	if err != nil {
		return fmt.Errorf("❌ User saqlashda xatolik (ID: %s): %v", user.ID, err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("⚠️ User (ID: %s) allaqachon mavjud, qo‘shilmadi", user.ID)
	}

	return nil
}

func InsertMonth(month *models.Month, db *sql.DB) error {
	query := `
		INSERT INTO months (
			id, name, code, storage_id, type, parent_id, deleted_type, 
			global_content_version, file_id, size, create_time, update_time, delete_time, 
			created_by, updated_by, deleted_by, download_url, detail_url
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 
			$11, $12, $13, $14, $15, $16, $17, $18
		) ON CONFLICT (id) DO NOTHING;`

	result, err := db.Exec(query, month.ID, month.Name, month.Code, month.StorageID, month.Type, month.ParentID,
		month.DeletedType, month.GlobalContentVersion, month.FileID, month.Size,
		month.CreateTime, month.UpdateTime, month.DeleteTime,
		month.CreatedBy, month.UpdatedBy, month.DeletedBy,
		month.DownloadURL, month.DetailURL)

	if err != nil {
		return fmt.Errorf("❌ Month saqlashda xatolik: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("⚠️ Month (ID: %s) allaqachon mavjud, qo‘shilmadi", month.ID)
	}

	return nil
}

func InsertTotal(total models.Total, db *sql.DB) error {

	query := `
		INSERT INTO total (audio_path, call_id, user_id)
		VALUES ($1, $2, $3)`

	result, err := db.Exec(query, total.AudioPath, total.CallID, total.UserID)

	if err != nil {
		return fmt.Errorf("❌ Total ma'lumotini qo'shishda xatolik: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Println("⚠️ Total ma'lumoti allaqachon mavjud, qo‘shilmadi")
	} else {
		log.Println("✅ Total ma'lumoti bazaga muvaffaqiyatli qo‘shildi!")
	}
	return nil
}

func GetLastDownloadedFileID(db *sql.DB) (string, error) {
	var lastFileID string
	err := db.QueryRow("SELECT call_id FROM total ORDER BY call_id DESC LIMIT 1").Scan(&lastFileID)
	if err == sql.ErrNoRows {
		return "", nil // Agar hech narsa topilmasa, bo'sh string qaytarish
	} else if err != nil {
		return "", err
	}
	return lastFileID, nil
}
