package main

import (
	"net/http"
	"github.com/google/uuid"
	"io"
	"os"
	"context"
	"encoding/hex"
	"crypto/rand"
	"mime"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	r.Body=http.MaxBytesReader(w, r.Body, 1 << 30)

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

	videoData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't get video", err)
		return
	} else if videoData.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "User does not own video", err)
		return
	}


	file, _, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse from file", err)
		return
	}

	defer file.Close()


	mediaType, _, err:=mime.ParseMediaType("video/mp4")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to get file type", err)
		return
	}

	tmp, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to create temp file", err)
		return
	}

	defer os.Remove(tmp.Name())
	defer tmp.Close()

	_, err = io.Copy(tmp, file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to copy to temp file", err)
		return
	}

	_, err = tmp.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not reset file pointer", err)
		return
	}


	key:=make([]byte, 32)
	rand.Read(key)

	encKey := hex.EncodeToString(key)

	objKey:= encKey + ".mp4"

	objParams:=&s3.PutObjectInput{
		Bucket: aws.String(cfg.s3Bucket),
		Key: aws.String(objKey),
		Body: tmp,
		ContentType: aws.String(mediaType),
	}

	_, err = cfg.s3Client.PutObject(context.Background(), objParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to put obj to s3", err)
		return
	}

	newurl:= fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, objKey)

	videoData.VideoURL = &newurl

	err = cfg.db.UpdateVideo(videoData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update video", err)
		return
	}

}
