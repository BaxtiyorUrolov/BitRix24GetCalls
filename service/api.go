package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"bitrix/models"
	"bitrix/storage"
)

// callBitrixMethod - universal funksiyamiz
func callBitrixMethod(db *sql.DB, memberID, method string, params url.Values, clientID, clientSecret string) (map[string]interface{}, error) {
	// 1) DB dan tokenni olish
	tokenInfo, err := storage.GetTokenByMemberID(db, memberID)
	if err != nil {
		return nil, fmt.Errorf("Token topilmadi yoki DB xatolik: %v", err)
	}

	// 2) Token eskirgan bo‘lsa, yangilash
	if IsTokenExpired(tokenInfo) {
		if err := RefreshToken(db, tokenInfo, clientID, clientSecret); err != nil {
			return nil, fmt.Errorf("Tokenni yangilashda xatolik: %v", err)
		}
	}

	// 3) Endi token yaroqli, so‘rovni yuboramiz
	endpoint := tokenInfo.ClientEndpoint // masalan: https://yourdomain.bitrix24.ru/rest/
	fullURL := fmt.Sprintf("%s%s?auth=%s", endpoint, method, tokenInfo.AccessToken)

	resp, err := http.PostForm(fullURL, params)
	if err != nil {
		return nil, fmt.Errorf("Bitrix API so'rovda xatolik: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Bitrix API status: %d, body: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("JSON parse xatolik: %v", err)
	}

	return result, nil
}

// GetUserInfo - user.get
func GetUserInfo(db *sql.DB, memberID, userID, clientID, clientSecret string) (*models.User, error) {
	params := url.Values{}
	params.Set("id", userID)

	res, err := callBitrixMethod(db, memberID, "user.get", params, clientID, clientSecret)
	if err != nil {
		return nil, err
	}

	rawArr, ok := res["result"].([]interface{})
	if !ok || len(rawArr) == 0 {
		return nil, fmt.Errorf("Foydalanuvchi topilmadi, ID: %s", userID)
	}

	data, err := json.Marshal(rawArr[0])
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// GetCallInfo - voximplant.statistic.get
func GetCallInfo(db *sql.DB, memberID, callID, clientID, clientSecret string) (*models.CallInfo, error) {
	params := url.Values{}
	params.Set("FILTER[ID]", callID)

	res, err := callBitrixMethod(db, memberID, "voximplant.statistic.get", params, clientID, clientSecret)
	if err != nil {
		return nil, err
	}

	rawArr, ok := res["result"].([]interface{})
	if !ok || len(rawArr) == 0 {
		return nil, fmt.Errorf("Qo‘ng‘iroq ma’lumotlari topilmadi, Call ID: %s", callID)
	}

	data, err := json.Marshal(rawArr[0])
	if err != nil {
		return nil, err
	}

	var callInfo models.CallInfo
	if err := json.Unmarshal(data, &callInfo); err != nil {
		return nil, err
	}
	return &callInfo, nil
}
