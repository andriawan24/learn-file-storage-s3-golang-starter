package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

const maxMemory = 10 << 20

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't parse multipart form", err)
		return
	}

	file, fileHeader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse multipart form", err)
		return
	}

	contentType := fileHeader.Header.Get("Content-Type")

	imageData, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't read file", err)
		return
	}

	encodedImage := base64.StdEncoding.EncodeToString(imageData)
	dataUrl := fmt.Sprintf("data:%s;base64,%s", contentType, encodedImage)

	metadata, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't read video metadata", err)
		return
	}

	if metadata.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Access denied", err)
		return
	}

	metadata.ThumbnailURL = &dataUrl

	cfg.db.UpdateVideo(metadata)

	respondWithJSON(w, http.StatusOK, struct{}{})
}
