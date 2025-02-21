package storage

import (
	"bitrix/models"
	"database/sql"
	"time"
)

func InsertOrUpdateToken(db *sql.DB, t *models.TokenInfo) error {
	// Misol uchun: unique bo'ladigan maydon: MemberID yoki PortalDomain
	query := `
		INSERT INTO tokens (portal_domain, member_id, access_token, refresh_token, expires_in, scope, last_update, client_endpoint)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (member_id) 
		DO UPDATE SET
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			expires_in = EXCLUDED.expires_in,
			scope = EXCLUDED.scope,
			last_update = EXCLUDED.last_update,
			client_endpoint = EXCLUDED.client_endpoint
	`
	_, err := db.Exec(query,
		t.PortalDomain,
		t.MemberID,
		t.AccessToken,
		t.RefreshToken,
		t.ExpiresIn,
		t.Scope,
		time.Now(), // last_update
		t.ClientEndpoint,
	)
	return err
}

// GetTokenByMemberID - member_id orqali tokenni olish
func GetTokenByMemberID(db *sql.DB, memberID string) (*models.TokenInfo, error) {
	query := `SELECT portal_domain, member_id, access_token, refresh_token, expires_in, scope, last_update, client_endpoint 
			  FROM tokens WHERE member_id = $1`
	row := db.QueryRow(query, memberID)

	var t models.TokenInfo
	var lastUpdate time.Time

	err := row.Scan(&t.PortalDomain, &t.MemberID, &t.AccessToken, &t.RefreshToken, &t.ExpiresIn, &t.Scope, &lastUpdate, &t.ClientEndpoint)
	if err != nil {
		return nil, err
	}
	t.LastUpdate = lastUpdate
	return &t, nil
}
