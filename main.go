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
	bitrixAPIURL = "https://yourdomain.bitrix24.ru/rest/1/"
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

func main() {
	connStr := "user=godb password=0208 dbname=bitrix sslmode=disable"
	db, err := storage.OpenDatabase(connStr)
	if err != nil {
		log.Fatal("❌ Database connection failed: ", err)
	}
	defer db.Close()

	folderID := "521316" // O'zgartiring

	// Dastur ishga tushganda eski fayllarni yuklaydi
	checkAndDownloadNewFiles(db, folderID)

	// Har 1 soatda yangi fayllarni yuklab olish uchun avtomatik ishga tushadi
	go startAutoDownload(db, folderID)

	select {} // Dastur doimiy ishlashda davom etadi
}

// 📥 **Yangi fayllarni yuklab olish**
func checkAndDownloadNewFiles(db *sql.DB, folderID string) {
	fmt.Println("🔍 Yangi fayllar tekshirilmoqda...")

	// Fayllarni olish
	allFiles, err := getAllAudioFiles(folderID)
	if err != nil {
		fmt.Println("❌ Audio fayllarni olishda xatolik:", err)
		return
	}

	// Mavjud fayllar ID'larini olish
	lastFileID, err := storage.GetLastDownloadedFileID(db)
	if err != nil {
		fmt.Println("❌ Mavjud fayllarni olishda xatolik:", err)
		return
	}

	// Agar yangi fayllarni aniqlashda hech qanday fayl topilmasa, mavjud fayllarni bo'sh qilib olish
	existingFiles := make(map[string]bool)
	if lastFileID != "" {
		existingFiles[lastFileID] = true
	}

	// Yangi fayllarni filtrlash
	newFiles := filterNewFiles(allFiles, existingFiles)

	if len(newFiles) == 0 {
		fmt.Println("📭 Yangi fayl topilmadi.")
		return
	}

	fmt.Printf("✅ Yangi yuklanadigan fayllar soni: %d\n", len(newFiles))

	// Yangi fayllarni yuklash va qayta ishlash
	downloadAndProcessFiles(newFiles, db, folderID)
}

// ⏳ **Avtomatik yangi fayllarni yuklab borish**
func startAutoDownload(db *sql.DB, folderID string) {
	ticker := time.NewTicker(1 * time.Hour) // ⏳ Har 1 soatda ishga tushadi
	defer ticker.Stop()

	for {
		<-ticker.C
		checkAndDownloadNewFiles(db, folderID)
	}
}

// 🎯 **Yangi fayllarni aniqlash**
func filterNewFiles(allFiles []AudioFile, existingFiles map[string]bool) []AudioFile {
	var newFiles []AudioFile
	for _, file := range allFiles {
		if _, exists := existingFiles[file.ID]; !exists {
			newFiles = append(newFiles, file)
		}
	}
	return newFiles
}

// 🔽 **Yangi fayllarni yuklab qayta ishlash**
func downloadAndProcessFiles(audioFiles []AudioFile, db *sql.DB, folderID string) {
	for _, audio := range audioFiles {
		fmt.Println("⬇️ Yuklanmoqda:", audio.Name)

		callInfo, err := getCallInfo(audio.ID)
		if err != nil {
			fmt.Println("❌ Qo‘ng‘iroq ma’lumotlarini olishda xatolik:", err)
			continue
		}

		err = storage.InsertCallInfo(callInfo, db)
		if err != nil {
			fmt.Println("❌ CallInfo saqlashda xatolik:", err)
		}

		audioPath, err := downloadAudio(audio.DownloadURL, audio.Name)
		if err != nil {
			fmt.Println("❌ Audio yuklab olishda xatolik:", err)
			continue
		}

		userInfo, err := getUserInfo(callInfo.PortalUserID)
		if err != nil {
			fmt.Println("❌ Foydalanuvchini olishda xatolik:", err)
			userInfo = &models.User{Name: "Noma’lum", LastName: ""}
		}

		err = storage.InsertUser(userInfo, db)
		if err != nil {
			fmt.Println("❌ Foydalanuvchini saqlashda xatolik:", err)
		}

		monthInfo, err := getMonthInfo(folderID, callInfo.ID)
		if err != nil {
			fmt.Println("❌ Oylik ma’lumotlarni olishda xatolik:", err)
		}

		err = storage.InsertMonth(monthInfo, db)
		if err != nil {
			fmt.Println("❌ Oylarni saqlashda xatolik:", err)
		}

		total := models.Total{
			AudioPath: audioPath,
			UserID:    userInfo.ID,
			CallID:    callInfo.ID,
		}

		err = storage.InsertTotal(total, db)
		if err != nil {
			fmt.Println("❌ Total saqlashda xatolik:", err)
		}
	}
}

// 📂 **Bitrix24'dan barcha fayllarni olish**
func getAllAudioFiles(folderID string) ([]AudioFile, error) {
	var allAudioFiles []AudioFile
	offset := 0
	limit := 50

	for {
		url := fmt.Sprintf("%s/disk.folder.getchildren?id=%s&navParams[OFFSET]=%d&navParams[LIMIT]=%d", bitrixAPIURL, folderID, offset, limit)
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var response struct {
			Result []AudioFile `json:"result"`
			Total  int         `json:"total"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, err
		}

		if len(response.Result) == 0 {
			break
		}

		allAudioFiles = append(allAudioFiles, response.Result...)
		offset += limit

		time.Sleep(2 * time.Second)

		if len(allAudioFiles) >= response.Total {
			break
		}
	}

	return allAudioFiles, nil
}

// 📥 **Audiolarni yuklab olish**
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
