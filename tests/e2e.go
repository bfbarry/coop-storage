package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	METASERVERBASE = "http://localhost:7678"
	TESTDATADIR    = "./data"

	FILENAME  = "test.txt"
	FILEPATH  = fmt.Sprintf("%s/%s", TESTDATADIR, FILENAME)
	IMAGENAME = "sample_logo.png"
	IMAGEPATH = fmt.Sprintf("%s/%s", TESTDATADIR, IMAGENAME)
)

func main() {
	secretKey := loadClerkSecretKey()

	token, err := fetchTestingToken(secretKey)
	if err != nil {
		log.Printf("Failed to fetch Clerk testing token: %v\n", err)
		os.Exit(1)
	}
	log.Printf("Got Clerk testing token")

	// Upload test.txt
	objectKey, err := uploadFile(token, FILENAME, FILEPATH)
	if err != nil {
		log.Printf("Upload failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Upload completed.")

	if err := downloadFile(objectKey); err != nil {
		log.Printf("Download failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Download completed.")

	accountID, err := createAccount("testuser@example.com")
	if err != nil {
		log.Printf("Create account failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Account created (id=%d).\n", accountID)

	// Metadata for test.txt at root
	if err := writeMetadata(objectKey, FILENAME, "text/plain", accountID, nil); err != nil {
		log.Printf("Write metadata failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Metadata written for test.txt.")

	// Upload image into a subfolder
	imageKey, err := uploadFile(token, IMAGENAME, IMAGEPATH)
	if err != nil {
		log.Printf("Image upload failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Image upload completed.")

	folderID, err := createFolder("images", accountID)
	if err != nil {
		log.Printf("Create folder failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Folder created (id=%d).\n", folderID)

	if err := writeMetadata(imageKey, IMAGENAME, "image/png", accountID, &folderID); err != nil {
		log.Printf("Write image metadata failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Metadata written for image in subfolder.")
}

func loadClerkSecretKey() string {
	if key := os.Getenv("CLERK_SECRET_KEY"); key != "" {
		return key
	}
	f, err := os.Open("../metadata-server/.env")
	if err != nil {
		log.Fatal("CLERK_SECRET_KEY not set and could not open ../metadata-server/.env:", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "CLERK_SECRET_KEY=") {
			val := strings.TrimPrefix(line, "CLERK_SECRET_KEY=")
			return strings.Trim(val, `"'`)
		}
	}
	log.Fatal("CLERK_SECRET_KEY not found in ../metadata-server/.env")
	return ""
}

const TEST_USER_ID = "user_3H7897hFhliU007YPd4WRve8Gxf"

// fetchTestingToken creates a Clerk session for the test user and returns a JWT.
func fetchTestingToken(secretKey string) (string, error) {
	// 1. Create a session for the test user
	body, _ := json.Marshal(map[string]string{"user_id": TEST_USER_ID})
	req, err := http.NewRequest(http.MethodPost, "https://api.clerk.com/v1/sessions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create session request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+secretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("create session failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create session returned %d: %s", resp.StatusCode, b)
	}

	var session struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return "", fmt.Errorf("decode session response: %w", err)
	}
	if session.ID == "" {
		return "", fmt.Errorf("clerk returned empty session ID")
	}

	// 2. Get a JWT for that session
	tokenBody, _ := json.Marshal(map[string]int{"expires_in_seconds": 300})
	tokenReq, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://api.clerk.com/v1/sessions/%s/tokens", session.ID),
		bytes.NewReader(tokenBody),
	)
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}
	tokenReq.Header.Set("Authorization", "Bearer "+secretKey)
	tokenReq.Header.Set("Content-Type", "application/json")

	tokenResp, err := http.DefaultClient.Do(tokenReq)
	if err != nil {
		return "", fmt.Errorf("create token failed: %w", err)
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(tokenResp.Body)
		return "", fmt.Errorf("create token returned %d: %s", tokenResp.StatusCode, b)
	}

	var result struct {
		JWT string `json:"jwt"`
	}
	if err := json.NewDecoder(tokenResp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}
	if result.JWT == "" {
		return "", fmt.Errorf("clerk returned empty JWT")
	}
	return result.JWT, nil
}

func uploadFile(token, filename, filepath string) (string, error) {
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return "", fmt.Errorf("file '%s' not found: %w", filepath, err)
	}

	// Step 1: get presigned upload URL from metadata server
	presignBody, _ := json.Marshal(map[string]any{
		"filename":       filename,
		"content_type":   "application/octet-stream",
		"content_length": fileInfo.Size(),
	})

	req, err := http.NewRequest(http.MethodPost, METASERVERBASE+"/upload/presign", bytes.NewReader(presignBody))
	if err != nil {
		return "", fmt.Errorf("create presign request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("presign request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("presign returned %d: %s", resp.StatusCode, body)
	}

	var presign struct {
		UploadURL string `json:"upload_url"`
		ObjectKey string `json:"object_key"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&presign); err != nil {
		return "", fmt.Errorf("decode presign response: %w", err)
	}
	log.Printf("Got presign URL for object_key: %s", presign.ObjectKey)

	// Step 2: PUT file bytes directly to RustFS presigned URL
	file, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	putReq, err := http.NewRequest(http.MethodPut, presign.UploadURL, file)
	if err != nil {
		return "", fmt.Errorf("create PUT request: %w", err)
	}
	putReq.Header.Set("Content-Type", "application/octet-stream")
	putReq.ContentLength = fileInfo.Size()

	putResp, err := http.DefaultClient.Do(putReq)
	if err != nil {
		return "", fmt.Errorf("PUT to RustFS failed: %w", err)
	}
	defer putResp.Body.Close()

	if putResp.StatusCode != http.StatusOK && putResp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(putResp.Body)
		return "", fmt.Errorf("RustFS PUT returned %d: %s", putResp.StatusCode, body)
	}

	log.Printf("File uploaded to object key: %s", presign.ObjectKey)
	return presign.ObjectKey, nil
}

func createAccount(email string) (int, error) {
	body, _ := json.Marshal(map[string]string{"email": email})
	resp, err := http.Post(METASERVERBASE+"/accounts", "application/json", bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("create account request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("create account returned %d: %s", resp.StatusCode, b)
	}

	var account struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return 0, fmt.Errorf("decode account response: %w", err)
	}
	return account.ID, nil
}

// createFolder creates a directory entry in metadata (no object_key, is_file=false).
func createFolder(name string, ownerID int) (int, error) {
	body, _ := json.Marshal(map[string]any{
		"owner_id": ownerID,
		"name":     name,
		"file_type": "inode/directory",
		"is_file":  false,
		"version":  1,
	})
	resp, err := http.Post(METASERVERBASE+"/metadata", "application/json", bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("create folder request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("create folder returned %d: %s", resp.StatusCode, b)
	}

	var m struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return 0, fmt.Errorf("decode folder response: %w", err)
	}
	return m.ID, nil
}

// writeMetadata records file metadata. Pass parentID=nil for root-level files.
func writeMetadata(objectKey, name, fileType string, ownerID int, parentID *int) error {
	payload := map[string]any{
		"owner_id":   ownerID,
		"object_key": objectKey,
		"file_type":  fileType,
		"is_file":    true,
		"name":       name,
		"version":    1,
	}
	if parentID != nil {
		payload["parent_id"] = *parentID
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(METASERVERBASE+"/metadata", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("write metadata request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("write metadata returned %d: %s", resp.StatusCode, b)
	}

	var m map[string]any
	json.NewDecoder(resp.Body).Decode(&m)
	log.Printf("Metadata entry: %v", m)
	return nil
}

func downloadFile(objectKey string) error {
	// Step 1: get presigned download URL from metadata server
	// objectKey contains slashes (user/fileID/name) so it maps directly to the wildcard path
	resp, err := http.Get(fmt.Sprintf("%s/download/presign/%s", METASERVERBASE, objectKey))
	if err != nil {
		return fmt.Errorf("download presign request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download presign returned %d: %s", resp.StatusCode, body)
	}

	var presign struct {
		DownloadURL string `json:"download_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&presign); err != nil {
		return fmt.Errorf("decode presign response: %w", err)
	}
	log.Printf("Got download URL: %s", presign.DownloadURL)

	// Step 2: fetch the file from RustFS
	getResp, err := http.Get(presign.DownloadURL)
	if err != nil {
		return fmt.Errorf("GET from RustFS failed: %w", err)
	}
	defer getResp.Body.Close()

	outPath := fmt.Sprintf("%s/downloaded_%s", TESTDATADIR, FILENAME)
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, getResp.Body); err != nil {
		return fmt.Errorf("write downloaded file: %w", err)
	}

	log.Printf("File downloaded to %s", outPath)
	return nil
}
