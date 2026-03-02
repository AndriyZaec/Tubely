package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
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

	fmt.Println("uploading video", videoID, "by user", userID)

	videoMetadata, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Video not found", err)
		return
	}

	if videoMetadata.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	uploadFile, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusNotFound, "No file in form", err)
		return
	}
	defer uploadFile.Close()

	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if mediaType != "video/mp4" || err != nil {
		respondWithError(w, http.StatusBadRequest, "Bad file format", err)
		return
	}

	tempFileName := "tubely-upload.mp4"
	tempFile, err := os.CreateTemp("", tempFileName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Can't add temp file", err)
		return
	}
	defer os.Remove(tempFileName)
	defer tempFile.Close()

	_, err = io.Copy(tempFile, uploadFile)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't save file", err)
		return
	}

	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Seek start failure", err)
		return
	}

	ratio, err := getVideoAspectRatio(tempFile.Name())
	if err != nil {
		fmt.Println("error getting ratio:", err)
	}

	generatedAssetName := fmt.Sprintf("%s/%s.%s", ratio, generateAssetName(), strings.Split(mediaType, "/")[1])

	_, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         &generatedAssetName,
		Body:        tempFile,
		ContentType: &mediaType,
	})
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "Couldnt save file", err)
		return
	}

	s3URL := cfg.getS3URL(generatedAssetName)
	videoMetadata.VideoURL = &s3URL
	videoMetadata.UpdatedAt = time.Now()

	err = cfg.db.UpdateVideo(videoMetadata)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, videoMetadata)
}
