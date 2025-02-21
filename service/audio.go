package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type AudioFile struct {
	ID          string `json:"ID"`
	Name        string `json:"NAME"`
	DownloadURL string `json:"DOWNLOAD_URL"`
}

// GetAllAudioFiles - disk.folder.getchildren
func GetAllAudioFiles(db *sql.DB, memberID, folderID, clientID, clientSecret string) ([]AudioFile, error) {
	var allAudioFiles []AudioFile
	offset := 0
	limit := 50

	for {
		params := url.Values{}
		params.Set("id", folderID)
		params.Set("navParams[OFFSET]", fmt.Sprintf("%d", offset))
		params.Set("navParams[LIMIT]", fmt.Sprintf("%d", limit))
		// Fayl linklari chiqishi uchun:
		params.Add("select[]", "ID")
		params.Add("select[]", "NAME")
		params.Add("select[]", "DOWNLOAD_URL")

		res, err := callBitrixMethod(db, memberID, "disk.folder.getchildren", params, clientID, clientSecret)
		if err != nil {
			return nil, err
		}

		var response struct {
			Result []AudioFile `json:"result"`
			Total  int         `json:"total"`
		}
		bytesRes, _ := json.Marshal(res)
		if err := json.Unmarshal(bytesRes, &response); err != nil {
			return nil, err
		}

		if len(response.Result) == 0 {
			break
		}
		allAudioFiles = append(allAudioFiles, response.Result...)
		offset += limit

		if len(allAudioFiles) >= response.Total {
			break
		}
		time.Sleep(1 * time.Second)
	}

	return allAudioFiles, nil
}

// DownloadAudio - oddiy GET bilan faylni yuklab olish
func DownloadAudio(downloadURL, fileName string) (string, error) {
	resp, err := http.Get(downloadURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	os.MkdirAll("downloads", os.ModePerm)
	filePath := filepath.Join("downloads", fileName)

	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return filePath, nil
}
