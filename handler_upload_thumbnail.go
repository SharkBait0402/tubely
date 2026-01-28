package main

import (
	"fmt"
	"io"
	"net/http"
	"encoding/base64"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

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

	// TODO: implement the upload here
	//
	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse from file", err)
		return
	}

	defer file.Close()

	contentTypes:=header.Header["Content-Type"]

	readByte, err:=io.ReadAll(file)
	if err!=nil {
		respondWithError(w, http.StatusBadRequest, "Unable to read from file", err)
		return
	}

	videoData, err := cfg.db.GetVideo(videoID)
	if err!=nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to get video metadata", err)
		return
	}

	thumbData:=base64.StdEncoding.EncodeToString(readByte)

	dataURL:=fmt.Sprintf("data:%v;base64,%v", contentTypes[0], thumbData)

	videoData.ThumbnailURL = &dataURL

	err=cfg.db.UpdateVideo(videoData)
	if err!=nil {
		respondWithError(w, http.StatusBadRequest, "Unable to update video data", err)
	}

	respondWithJSON(w, http.StatusOK, videoData)
}
