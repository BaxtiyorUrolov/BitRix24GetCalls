package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Bitrix24 API bazaviy URL manzili
const (
	bitrixURL = "https://visainfo.bitrix24.ru/rest/1/cnrbvh682ozxjlx6/"
)

// Barcha CRM qo‘ng‘iroqlarini olish
func getAllCRMCalls() error {
	url := bitrixURL + "crm.activity.list.json"
	fmt.Println("GET so‘rov yuborilmoqda:", url)

	// API ga so‘rov yuborish
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// API javobini o‘qish
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Javob:", string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Bitrix24 API xatosi: %d, javob: %s", resp.StatusCode, string(body))
	}

	// JSON parsing uchun struct
	var responseData struct {
		Result []struct {
			ID          string `json:"ID"`
			Subject     string `json:"SUBJECT"`
			CallType    string `json:"PROVIDER_TYPE_ID"`
			Description string `json:"DESCRIPTION"`
		} `json:"result"`
	}

	// JSON ma'lumotlarini deserializatsiya qilish
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return err
	}

	// Natijalarni ekranga chiqarish
	fmt.Println("\n📞 CRMdagi Qo‘ng‘iroqlar ro‘yxati:")
	for _, call := range responseData.Result {
		fmt.Printf("- Qo‘ng‘iroq ID: %s | Mavzu: %s | Turi: %s | Tavsif: %s\n",
			call.ID, call.Subject, call.CallType, call.Description)
	}

	return nil
}

func main() {
	err := getAllCRMCalls()
	if err != nil {
		fmt.Println("Xatolik:", err)
	}
}
