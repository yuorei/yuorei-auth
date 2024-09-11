package rest

import (
	"encoding/json"

	"net/http"
	"os"

	"github.com/Nerzal/gocloak/v13"
	"github.com/go-chi/chi"
	"github.com/yuorei/yuorei-auth/app/adapter/infra"
)

type CreateUserInput struct {
	FirstName *string
	LastName  *string
	Email     *string
	Username  string
	Password  string
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 入力構造体の初期化
	var input CreateUserInput

	// 各フィールドを手動で割り当て
	input.FirstName = getStringPointer(r.FormValue("firstName"))
	input.LastName = getStringPointer(r.FormValue("lastName"))
	input.Email = getStringPointer(r.FormValue("email"))
	input.Username = r.FormValue("username")
	input.Password = r.FormValue("password")

	// ファイルを取得
	file, _, err := r.FormFile("profileImage")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer file.Close()

	client := gocloak.NewClient(os.Getenv("KEYCLOAK_URL"))
	ctx := r.Context()
	masterToken, err := client.LoginAdmin(ctx, os.Getenv("KEYCLOAK_ADMIN_USERNAME"), os.Getenv("KEYCLOAK_ADMIN_PASSWORD"), "master")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	imageFile, err := infra.ConvertToWebp(file, input.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	imageURL, err := infra.UploadImageForStorage(imageFile, input.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := gocloak.User{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Enabled:   gocloak.BoolP(true),
		Username:  &input.Username,
		Attributes: &map[string][]string{
			"profileImage": {imageURL},
		},
	}

	userID, err := client.CreateUser(ctx, masterToken.AccessToken, os.Getenv("KEYCLOAK_REALM"), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = client.SetPassword(ctx, masterToken.AccessToken, userID, os.Getenv("KEYCLOAK_REALM"), input.Password, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	client := gocloak.NewClient(os.Getenv("KEYCLOAK_URL"))
	token, err := client.Login(r.Context(), os.Getenv("KEYCLOAK_CLIENTID"), os.Getenv("KEYCLOAK_CLIENT_SECRET"), os.Getenv("KEYCLOAK_REALM"), req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

type ProfileImageURL struct {
	URL string `json:"url"`
}

func GetProfileImage(w http.ResponseWriter, r *http.Request) {
	client := gocloak.NewClient(os.Getenv("KEYCLOAK_URL"))
	ctx := r.Context()
	masterToken, err := client.LoginAdmin(ctx, os.Getenv("KEYCLOAK_ADMIN_USERNAME"), os.Getenv("KEYCLOAK_ADMIN_PASSWORD"), "master")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID := chi.URLParam(r, "user-id")
	user, err := client.GetUserByID(ctx, masterToken.AccessToken, os.Getenv("KEYCLOAK_REALM"), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	var imageUrl string
	if user.Attributes != nil {
		images, ok := (*user.Attributes)["profileImage"]
		if ok && len(images) > 0 {
			imageUrl = images[0]
		}
	}
	profileImageURL := ProfileImageURL{
		URL: imageUrl,
	}
	json.NewEncoder(w).Encode(profileImageURL)
}

func getStringPointer(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
