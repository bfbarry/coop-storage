package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	// "github.com/bfbarry/coop-storage/metadata-server"
)

type App struct {
	repo IMetadataRepo
}

func NewApp(repo IMetadataRepo) *App {
	return &App{repo: repo}
}

func (this *App) PostMetadata(w http.ResponseWriter, r *http.Request) {
	var params CreateMetadataParams

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	gm, err := this.repo.CreateMetadata(r.Context(), &params)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create metadata %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(gm)
}

func (this *App) GetMetadata(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	m, err := this.repo.GetMetadata(r.Context(), int32(id))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get metadata: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (this *App) GetChildren(w http.ResponseWriter, r *http.Request) {
	var parentID *int32
	if s := r.URL.Query().Get("parent_id"); s != "" {
		id, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			http.Error(w, "Invalid parent_id", http.StatusBadRequest)
			return
		}
		id32 := int32(id)
		parentID = &id32
	}
	items, err := this.repo.GetChildren(r.Context(), parentID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get children: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (this *App) GetChildrenHome(w http.ResponseWriter, r *http.Request) {
	ownerID, err := strconv.ParseInt(r.URL.Query().Get("owner_id"), 10, 32)
	if err != nil {
		http.Error(w, "Invalid owner_id", http.StatusBadRequest)
		return
	}
	items, err := this.repo.GetChildrenHome(r.Context(), int32(ownerID))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get home children: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (this *App) PutMetadata(w http.ResponseWriter, r *http.Request) {
	var params UpdateMetaParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	m, err := this.repo.UpdateMeta(r.Context(), &params)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update metadata: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (this *App) PatchMetadataParent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	var body struct {
		ParentID *int32 `json:"parent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	m, err := this.repo.MoveMetadata(r.Context(), &MoveMetadataParams{
		ID:       int32(id),
		ParentID: body.ParentID,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to move metadata: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (this *App) DeleteMetadata(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	if err := this.repo.DeleteMetadata(r.Context(), int32(id)); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete metadata: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Accounts

func (this *App) PostAccount(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ClerkID string `json:"clerk_id"`
		Email   string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	a, err := this.repo.CreateAccount(r.Context(), &CreateAccountParams{
		ClerkID: &body.ClerkID,
		Email:   body.Email,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create account: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, a)
}

func (this *App) GetAccount(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	a, err := this.repo.GetAccount(r.Context(), int32(id))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get account: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (this *App) GetAccountByClerkID(w http.ResponseWriter, r *http.Request) {
	clerkID := r.URL.Query().Get("clerk_id")
	if clerkID == "" {
		http.Error(w, "clerk_id query param required", http.StatusBadRequest)
		return
	}
	a, err := this.repo.GetAccountByClerkID(r.Context(), clerkID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Account not found: %v", err), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (this *App) PutAccount(w http.ResponseWriter, r *http.Request) {
	var params UpdateAccountParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	a, err := this.repo.UpdateAccount(r.Context(), &params)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update account: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (this *App) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	if err := this.repo.DeleteAccount(r.Context(), int32(id)); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete account: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Permissions

func (this *App) PostPermission(w http.ResponseWriter, r *http.Request) {
	var params CreatePermissionParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	p, err := this.repo.CreatePermission(r.Context(), &params)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create permission: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (this *App) GetPermission(w http.ResponseWriter, r *http.Request) {
	metadataID, err := strconv.ParseInt(r.URL.Query().Get("metadata_id"), 10, 32)
	if err != nil {
		http.Error(w, "Invalid metadata_id", http.StatusBadRequest)
		return
	}
	accountID, err := strconv.ParseInt(r.URL.Query().Get("account_id"), 10, 32)
	if err != nil {
		http.Error(w, "Invalid account_id", http.StatusBadRequest)
		return
	}
	params := GetPermissionParams{MetadataID: int32(metadataID), AccountID: int32(accountID)}
	p, err := this.repo.GetPermission(r.Context(), &params)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get permission: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (this *App) DeletePermission(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	if err := this.repo.DeletePermission(r.Context(), int32(id)); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete permission: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
