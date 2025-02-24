package main

import (
	"bitrix/models"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"bitrix/service"
	"bitrix/storage"

	_ "github.com/lib/pq"
)

var (
	clientID     = "CLIENT_ID"
	clientSecret = "CLIENT_SECRET"
	redirectURI  = "REDIRECT_URI"
)

func main() {
	// 1) Muhitni tekshirish
	if clientID == "" || clientSecret == "" || redirectURI == "" {
		log.Fatal("Iltimos, BITRIX_CLIENT_ID, BITRIX_CLIENT_SECRET, BITRIX_REDIRECT_URI ni environmentda sozlang.")
	}

	// 2) DB ga ulanish
	connStr := "user=godb password=0208 dbname=bitrix sslmode=disable" // misol: "user=godb password=0208 dbname=bitrix sslmode=disable"
	db, err := storage.OpenDatabase(connStr)
	if err != nil {
		log.Fatal("DB connection xatolik:", err)
	}
	defer db.Close()

	// 3) "/" – oddiy sahifa (test uchun)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to My Bitrix24 Install + Call Records App!\n")
		fmt.Fprintf(w, "client_id: %s\nredirect_uri: %s\n", clientID, redirectURI)
	})

	// 4) "/bitrix/oauth" – Bitrix24 ilovasini o‘rnatish (install) paytida code keladi
	http.HandleFunc("/bitrix/oauth", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Missing 'code'", http.StatusBadRequest)
			return
		}

		// code -> token (access_token, refresh_token, ...)
		tokenInfo, err := service.ExchangeCodeForToken(db, code, clientID, clientSecret, redirectURI)
		if err != nil {
			http.Error(w, "Token exchange error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Demo uchun folderID ni statik qilamiz (disk.folder ID)
		folderID := "521316"
		// DB da folderID saqlash (portals jadvalida)
		if err := storage.UpdatePortalFolderID(db, tokenInfo.MemberID, folderID); err != nil {
			log.Println("FolderID saqlashda xatolik:", err)
		}

		// Install muvaffaqiyatli bo'ldi
		fmt.Fprintf(w, "✅ Ilova muvaffaqiyatli o‘rnatildi!\n")
		fmt.Fprintf(w, "MemberID: %s\nDomain: %s\n", tokenInfo.MemberID, tokenInfo.PortalDomain)
		fmt.Fprintf(w, "FolderID: %s\n", folderID)
	})

	// 5) Avtomatik call records yuklab olish (har 1 soatda)
	go startAutoDownload(db)

	// 6) Serverni ishga tushirish
	port := ":8090"
	log.Printf("Server running on %s...", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// startAutoDownload – har 1 soatda call recordlarni yuklab olish
func startAutoDownload(db *sql.DB) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		portals, err := storage.GetAllPortals(db)
		if err != nil {
			log.Println("GetAllPortals xatolik:", err)
			<-ticker.C
			continue
		}
		log.Printf("Portal soni: %d\n", len(portals))

		// Har bir portal uchun
		for _, p := range portals {
			checkAndDownloadRecords(db, p.MemberID, p.FolderID)
		}

		<-ticker.C
	}
}

// checkAndDownloadRecords – call recordlarni yuklab, DB ga yozish
func checkAndDownloadRecords(db *sql.DB, memberID, folderID string) {
	log.Printf("🔍 Portal %s, folder %s – call recordlarni tekshirish...", memberID, folderID)

	// 1) Disk papkadan audio fayllar
	audioFiles, err := service.GetAllAudioFiles(db, memberID, folderID, clientID, clientSecret)
	if err != nil {
		log.Println("GetAllAudioFiles xatolik:", err)
		return
	}
	if len(audioFiles) == 0 {
		log.Println("📭 Yangi audio fayl topilmadi.")
		return
	}
	log.Printf("✅ Topilgan audio fayllar soni: %d\n", len(audioFiles))

	// 2) Har bir audio fayl uchun call info, user info, yuklab olish
	for _, audio := range audioFiles {
		callInfo, err := service.GetCallInfo(db, memberID, audio.ID, clientID, clientSecret)
		if err != nil {
			log.Println("GetCallInfo xatolik:", err)
			continue
		}
		// DB ga call_info yozish
		if err := storage.InsertCallInfo(callInfo, db); err != nil {
			log.Println("InsertCallInfo xatolik:", err)
		}

		// Audio faylni yuklab olish
		audioPath, err := service.DownloadAudio(audio.DownloadURL, audio.Name)
		if err != nil {
			log.Println("DownloadAudio xatolik:", err)
			continue
		}

		// Foydalanuvchini olish
		userInfo, err := service.GetUserInfo(db, memberID, callInfo.PortalUserID, clientID, clientSecret)
		if err != nil {
			log.Println("UserInfo xatolik:", err)
			userInfo = &models.User{ID: callInfo.PortalUserID, Name: "Noma'lum"}
		}
		if err := storage.InsertUser(userInfo, db); err != nil {
			log.Println("InsertUser xatolik:", err)
		}

		// Total jadvaliga yozish
		total := models.Total{
			AudioPath: audioPath,
			UserID:    userInfo.ID,
			CallID:    callInfo.ID,
		}
		if err := storage.InsertTotal(total, db); err != nil {
			log.Println("InsertTotal xatolik:", err)
		}

		log.Printf("⬇️ Yuklab olingan fayl: %s, CallID: %s, UserID: %s\n",
			audioPath, callInfo.ID, userInfo.ID)
	}
}
