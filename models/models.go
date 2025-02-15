package models

type User struct {
	ID               string   `json:"id"`
	XML_ID           string   `json:"xml_id"`
	Active           bool     `json:"active"`
	Name             string   `json:"name"`
	LastName         string   `json:"last_name"`
	SecondName       string   `json:"second_name"`
	Email            string   `json:"email"`
	LastLogin        string   `json:"last_login"`
	TimeZone         string   `json:"time_zone"`
	TimeZoneOffset   string   `json:"time_zone_offset"`
	PersonalPhoto    string   `json:"personal_photo"`
	PersonalGender   string   `json:"personal_gender"`
	PersonalWWW      string   `json:"personal_www"`
	PersonalBirthday string   `json:"personal_birthday"`
	PersonalMobile   string   `json:"personal_mobile"`
	PersonalCity     string   `json:"personal_city"`
	WorkPhone        string   `json:"work_phone"`
	WorkPosition     string   `json:"work_position"`
	EmploymentDate   string   `json:"employment_date"`
	UserType         string   `json:"user_type"`
	Department       []string `json:"department"`
}

// months jadvali
type Month struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Code                 string `json:"code"`
	StorageID            string `json:"storage_id"`
	Type                 string `json:"type"`
	ParentID             string `json:"parent_id"`
	DeletedType          string `json:"deleted_type"`
	GlobalContentVersion string `json:"global_content_version"`
	FileID               string `json:"file_id"`
	Size                 string `json:"size"`
	CreateTime           string `json:"create_time"`
	UpdateTime           string `json:"update_time"`
	DeleteTime           string `json:"delete_time"`
	CreatedBy            string `json:"created_by"`
	UpdatedBy            string `json:"updated_by"`
	DeletedBy            string `json:"deleted_by"`
	DownloadURL          string `json:"download_url"`
	DetailURL            string `json:"detail_url"`
}

type CallInfo struct {
	ID                string `json:"id"`
	PortalUserID      string `json:"portal_user_id"`
	PortalNumber      string `json:"portal_number"`
	PhoneNumber       string `json:"phone_number"`
	CallID            string `json:"call_id"`
	ExternalCallID    string `json:"external_call_id"`
	CallCategory      string `json:"call_category"`
	CallDuration      string `json:"call_duration"`
	CallStartDate     string `json:"call_start_date"`
	CallRecordURL     string `json:"call_record_url"`
	CallVote          string `json:"call_vote"`
	Cost              string `json:"cost"`
	CostCurrency      string `json:"cost_currency"`
	CallFailedCode    string `json:"call_failed_code"`
	CallFailedReason  string `json:"call_failed_reason"`
	CRMEntityType     string `json:"crm_entity_type"`
	CRMEntityID       string `json:"crm_entity_id"`
	CRMActivityID     string `json:"crm_activity_id"`
	RestAppID         string `json:"rest_app_id"`
	RestAppName       string `json:"rest_app_name"`
	TranscriptID      string `json:"transcript_id"`
	TranscriptPending string `json:"transcript_pending"`
	SessionID         string `json:"session_id"`
	RedialAttempt     string `json:"redial_attempt"`
	Comment           string `json:"comment"`
	RecordDuration    string `json:"record_duration"`
	RecordFileID      string `json:"record_file_id"`
	CallType          string `json:"call_type"`
}
type Total struct {
	AudioPath string `json:"audio_path"`
	CallID    string `json:"call_id"`
	UserID    string `json:"user_id"`
}
