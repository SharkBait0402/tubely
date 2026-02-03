package main

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"os"
	"strings"

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
	fileType:=contentTypes[0]

	splitType:=strings.Split(fileType, "/")

	ext:=splitType[1]

	// readByte, err:=io.ReadAll(file)
	// if err!=nil {
	// 	respondWithError(w, http.StatusBadRequest, "Unable to read from file", err)
	// 	return
	// }

	videoData, err := cfg.db.GetVideo(videoID)
	if err!=nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to get video metadata", err)
		return
	}

	fileName:=fmt.Sprintf("%s.%s", videoIDString, ext)

	savePath:=filepath.Join(cfg.assetsRoot, fileName)

	createdFile, err:=os.Create(savePath)
	if err!=nil {
		respondWithError(w, http.StatusBadRequest, "Unable to create file", err)
		return
	}

	_, err=io.Copy(createdFile, file)
	if err!=nil {
		respondWithError(w, http.StatusBadRequest, "Unable to write to dest file", err)
		return
	}

	thumbURL:=fmt.Sprintf("http://localhost:8091/%s", savePath)

	videoData.ThumbnailURL = &thumbURL

	err=cfg.db.UpdateVideo(videoData)
	if err!=nil {
		respondWithError(w, http.StatusBadRequest, "Unable to update video data", err)
	}

	respondWithJSON(w, http.StatusOK, videoData)
}
