package main

import (
	"bitrix/models"
	"bitrix/storage"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Bitrix24 va Telegram konfiguratsiyasi
var (
	bitrixAPIURL = "https://visainfo.bitrix24.ru/rest/1/cnrbvh682ozxjlx6"
)

// Audio fayl strukturalari
type AudioFile struct {
	ID          string `json:"ID"`
	Name        string `json:"NAME"`
	DownloadURL string `json:"DOWNLOAD_URL"`
}

type BitrixResponse struct {
	Result []AudioFile `json:"result"`
	Total  int         `json:"total"`
}

// Foydalanuvchini olish (PORTAL_USER_ID orqali)
func getUserInfo(userID string) (*models.User, error) {
	url := fmt.Sprintf("%s/user.get?id=%s", bitrixAPIURL, userID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("❌ Bitrix24 foydalanuvchini olishda xatolik: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Result []models.User `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("❌ JSON parsingda xatolik: %v", err)
	}

	if len(result.Result) == 0 {
		return nil, fmt.Errorf("❌ Foydalanuvchi topilmadi, ID: %s", userID)
	}

	return &result.Result[0], nil
}

func getMonthInfo(folderID, searchID string) (*models.Month, error) {
	url := fmt.Sprintf("%s/disk.folder.getchildren?id=%s", bitrixAPIURL, folderID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("❌ API so‘rovida xatolik: %v", err)
	}
	defer resp.Body.Close()

	// JSON javobini parse qilish
	var result struct {
		Result []models.Month `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("❌ JSON parsingda xatolik: %v", err)
	}

	// 🔍 ID bo‘yicha qidirish
	for _, item := range result.Result {
		if item.ID == searchID {
			return &item, nil
		}
	}

	return nil, fmt.Errorf("❌ ID %s bo‘yicha ma'lumot topilmadi", searchID)
}

func getCallInfo(callID string) (*models.CallInfo, error) {

	time.Sleep(2 * time.Second)

	url := fmt.Sprintf("%s/voximplant.statistic.get?FILTER[ID]=%s", bitrixAPIURL, callID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("❌ Bitrix24 API so‘rovda xatolik: %v", err)
	}
	defer resp.Body.Close()

	// API dan kelgan javobni ekranga chiqarish
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("❌ Javobni o‘qishda xatolik: %v", err)
	}

	// JSON parsing
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return nil, fmt.Errorf("❌ JSON parsingda xatolik: %v", err)
	}

	// `result` mavjudligini tekshirish
	rawResult, ok := rawResponse["result"]
	if !ok {
		return nil, fmt.Errorf("❌ API javobida `result` mavjud emas, Call ID: %s", callID)
	}

	// `result` massiv ekanligini tekshirish
	resultArray, ok := rawResult.([]interface{})
	if !ok || len(resultArray) == 0 {
		return nil, fmt.Errorf("❌ Qo‘ng‘iroq ma’lumotlari topilmadi, Call ID: %s", callID)
	}

	// JSON-ni `models.CallInfo` ga deserialize qilish
	callInfoBytes, err := json.Marshal(resultArray[0])
	if err != nil {
		return nil, fmt.Errorf("❌ JSON serializationda xatolik: %v", err)
	}

	var callInfo models.CallInfo
	if err := json.Unmarshal(callInfoBytes, &callInfo); err != nil {
		return nil, fmt.Errorf("❌ JSON parsingda xatolik: %v", err)
	}

	return &callInfo, nil
}

// Audiolarni yuklab olish
func downloadAudio(url, fileName string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	filePath := filepath.Join("downloads", fileName)
	os.MkdirAll("downloads", os.ModePerm)

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

// Asosiy funksiya
func main() {

	connStr := "user=godb password=0208 dbname=bitrix sslmode=disable"
	db, err := storage.OpenDatabase(connStr)
	if err != nil {
		log.Fatal("❌ Database connection failed: ", err)
	}
	defer db.Close()

	folderID := "521316" // O'zgartiring

	// 1️⃣ Bitrix24'dan barcha audio fayllarni olish
	audioFiles, err := getAllAudioFiles(folderID)
	if err != nil {
		fmt.Println("❌ Audio fayllarni olishda xatolik:", err)
		return
	}

	for _, audio := range audioFiles {
		fmt.Println("⬇️ Yuklanmoqda:", audio.Name)

		// ✅ Qo‘ng‘iroq ID ni olish
		callInfo, err := getCallInfo(audio.ID)
		if err != nil {
			fmt.Println("❌ Qo‘ng‘iroq ma’lumotlarini olishda xatolik:", err)
			continue
		}

		err = storage.InsertCallInfo(callInfo, db)
		if err != nil {
			fmt.Println("❌ CallInfo saqlashda xatolik:", err)
		}

		// 🔽 Audio yuklab olish
		audioPath, err := downloadAudio(audio.DownloadURL, audio.Name)
		if err != nil {
			fmt.Println("❌ Audio yuklab olishda xatolik:", err)
			continue
		}

		//go startAutoDownload(db, folderID)

		// 👤 Foydalanuvchi ma’lumotlarini olish
		userInfo, err := getUserInfo(callInfo.PortalUserID)
		if err != nil {
			fmt.Println("❌ Foydalanuvchini olishda xatolik:", err)
			userInfo = &models.User{Name: "Noma’lum", LastName: ""}
		}

		err = storage.InsertUser(userInfo, db)
		if err != nil {
			fmt.Println("Foydalanuvchini saqlashda xatolik:", err)
		}

		monthInfo, err := getMonthInfo(folderID, callInfo.ID)
		if err != nil {
			fmt.Println("oylar ruyxatini olishda xatolik:", err)
		}

		err = storage.InsertMonth(monthInfo, db)
		if err != nil {
			fmt.Println("oylarni saqlashda xatolik: ", err)
		}

		total := models.Total{
			AudioPath: audioPath,
			UserID:    userInfo.ID,
			CallID:    callInfo.ID,
		}

		err = storage.InsertTotal(total, db)
		if err != nil {
			fmt.Println("total saqlashda xatolik: ", err)
		}

	}
}

func getAllAudioFiles(folderID string) ([]AudioFile, error) {
	var allAudioFiles []AudioFile
	offset := 0
	limit := 50 // Har safar 50 ta fayl yuklaymiz

	for {
		url := fmt.Sprintf("%s/disk.folder.getchildren?id=%s&navParams[OFFSET]=%d&navParams[LIMIT]=%d", bitrixAPIURL, folderID, offset, limit)
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var response BitrixResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, err
		}

		// Agar yangi fayllar bo'lmasa, to'xtaymiz
		if len(response.Result) == 0 {
			break
		}

		allAudioFiles = append(allAudioFiles, response.Result...)

		// Keyingi sahifaga o'tish
		offset += limit
		fmt.Printf("✅ Yuklangan fayllar soni: %d\n", len(allAudioFiles))

		time.Sleep(2 * time.Second)

		// Agar yuklangan fayllar umumiy soniga yetib kelsa, to‘xtaymiz
		if len(allAudioFiles) >= response.Total {
			break
		}
	}

	return allAudioFiles, nil
}

func getNewAudioFiles(db *sql.DB, folderID string) ([]AudioFile, error) {
	lastFileID, err := storage.GetLastDownloadedFileID(db)
	if err != nil {
		return nil, err
	}

	allFiles, err := getAllAudioFiles(folderID) // Bitrixdan barcha fayllarni olamiz
	if err != nil {
		return nil, err
	}

	var newFiles []AudioFile
	for _, file := range allFiles {
		if file.ID > lastFileID { // Faqat oxirgi yuklanganidan keyin kelganlarini olamiz
			newFiles = append(newFiles, file)
		}
	}

	// Agar yangi fayllar bo‘lsa, oxirgi yuklangan faylni yangilaymiz
	if len(newFiles) > 0 {
		newLastFileID := newFiles[len(newFiles)-1].ID
		err := storage.UpdateLastDownloadedFileID(db, newLastFileID)
		if err != nil {
			return nil, err
		}
	}

	return newFiles, nil
}

func startAutoDownload(db *sql.DB, folderID string) {
	ticker := time.NewTicker(1 * time.Hour) // ⏳ Har 1 soatda ishga tushadi
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("🔄 Yangi fayllarni tekshiryapman...")
			newFiles, err := getNewAudioFiles(db, folderID)
			if err != nil {
				fmt.Println("⚠️ Xatolik:", err)
				continue
			}

			if len(newFiles) > 0 {
				fmt.Printf("✅ Yangi yuklangan fayllar soni: %d\n", len(newFiles))
			} else {
				fmt.Println("📭 Yangi fayl topilmadi.")
			}
		}
	}
}
