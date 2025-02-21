package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"bitrix/service"
	"bitrix/storage"

	_ "github.com/lib/pq"
)

var (
	clientID     = os.Getenv("BITRIX_CLIENT_ID")
	clientSecret = os.Getenv("BITRIX_CLIENT_SECRET")
	redirectURI  = os.Getenv("BITRIX_REDIRECT_URI")
)

func main() {
	// 1) Muhit
	if clientID == "" || clientSecret == "" || redirectURI == "" {
		log.Fatal("BITRIX_CLIENT_ID, BITRIX_CLIENT_SECRET, BITRIX_REDIRECT_URI yo'q!")
	}

	// 2) DB ga ulanish
	connStr := "user=godb password=0208 dbname=bitrix sslmode=disable"
	db, err := storage.OpenDatabase(connStr)
	if err != nil {
		log.Fatal("DB connection xatolik:", err)
	}
	defer db.Close()

	// 3) /bitrix/oauth - har bir portal o‘rnatganda code keladi
	http.HandleFunc("/bitrix/oauth", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Missing 'code'", http.StatusBadRequest)
			return
		}

		// code → token
		tokenInfo, err := service.ExchangeCodeForToken(db, code, clientID, clientSecret, redirectURI)
		if err != nil {
			http.Error(w, "Token exchange error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Bu yerda folderIDni qanday aniqlash?
		// Masalan, default folderID= "521316" yoki userdan so'rash
		// Demo uchun statik:
		folderID := "521316"

		// DB'da portals jadvaliga folderID'ni yangilash
		// (InsertOrUpdatePortal) - biz oldin InsertOrUpdateToken qilganmiz,
		// endi folderID ham saqlaymiz. Soddalashtirish uchun:
		if err := storage.UpdatePortalFolderID(db, tokenInfo.MemberID, folderID); err != nil {
			log.Println("FolderID saqlashda xatolik:", err)
		}

		fmt.Fprintf(w, "Token olindi! MemberID: %s, Domain: %s, FolderID: %s\n",
			tokenInfo.MemberID, tokenInfo.PortalDomain, folderID)
	})

	// 4) Ishga tushganda har bir portal uchun goroutine
	go startAutoForAllPortals(db)

	// 5) Server
	port := ":8080"
	log.Println("Server running on", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// startAutoForAllPortals - har 1 soatda BARCHA portalni ketma-ket tekshiradi
func startAutoForAllPortals(db *sql.DB) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		// Har safar ticker bo'lganda, DB dan portals ro'yxatini qayta o'qiymiz
		portals, err := storage.GetAllPortals(db)
		if err != nil {
			log.Println("GetAllPortals xatolik:", err)
			continue
		}
		log.Printf("Portal soni: %d\n", len(portals))

		// Har bir portal uchun audio yuklab olish
		for _, p := range portals {
			checkAndDownloadNewFiles(db, p.MemberID, p.FolderID)
		}

		<-ticker.C
	}
}

// checkAndDownloadNewFiles - har bir portal + folder bo'yicha audio yuklab, DB ga yozish
func checkAndDownloadNewFiles(db *sql.DB, memberID, folderID string) {
	log.Printf("🔍 Portal: %s, folder: %s -> Yangi fayllar tekshirilmoqda...", memberID, folderID)

	audioFiles, err := service.GetAllAudioFiles(db, memberID, folderID, clientID, clientSecret)
	if err != nil {
		log.Println("GetAllAudioFiles xatolik:", err)
		return
	}
	if len(audioFiles) == 0 {
		log.Println("📭 Yangi fayl topilmadi.")
		return
	}
	log.Printf("✅ Topilgan audio fayllar soni: %d\n", len(audioFiles))

	// Barcha faylni qayta ishlash
	for _, audio := range audioFiles {
		// call info
		_, err := service.GetCallInfo(db, memberID, audio.ID, clientID, clientSecret)
		if err != nil {
			log.Println("GetCallInfo xatolik:", err)
			continue
		}
		// DB ga call_info yozish, audio yuklab olish, user.get, va hokazo
		// (Sizning oldingi InsertCallInfo, InsertUser, InsertTotal va h.k.)
		// ...
		log.Printf("⬇️ Yuklab olingan fayl ID: %s (portal: %s)", audio.ID, memberID)
	}
}
