package controllers

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"

	"firebase.google.com/go/v4/auth"
	"github.com/ThinkBattleground/ThinkBattleground-Backend/config"
	"github.com/ThinkBattleground/ThinkBattleground-Backend/database"
	"github.com/ThinkBattleground/ThinkBattleground-Backend/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"
)

type AuthController struct {
	authClient *auth.Client
}

// NewAuthController creates a new instance of AuthController
func NewAuthController(authClient *auth.Client) *AuthController {
	return &AuthController{
		authClient: authClient,
	}
}

// SignUp registers a new user with email and password
func (ac *AuthController) SignUp(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := (&auth.UserToCreate{}).
		Email(input.Email).
		Password(input.Password).
		EmailVerified(false) // Set email as unverified initially

	user, err := ac.authClient.CreateUser(context.Background(), params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create user in database
	dbUser := models.User{
		FirebaseUID: user.UID,
		Email:       user.Email,
		IsAdmin:     false,
	}

	if result := database.DB.Create(&dbUser); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user in database"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully. Please log in to get your token.",
		"user": gin.H{
			"uid":           user.UID,
			"email":         user.Email,
			"isAdmin":       false,
			"emailVerified": false,
		},
	})
}

// SignIn authenticates an existing user
func (ac *AuthController) SignIn(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call Firebase Auth REST API to verify email/password
	client := &http.Client{}
	reqBody := map[string]interface{}{
		"email":             input.Email,
		"password":          input.Password,
		"returnSecureToken": true,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing request"})
		return
	}

	firebaseAuthURL := "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=" + config.GetEnv("FIREBASE_WEB_API_KEY", "")
	log.Println("Firebase Auth URL:", firebaseAuthURL) // Debugging line
	req, err := http.NewRequest("POST", firebaseAuthURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating request"})
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error communicating with authentication service"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": errorResponse.Error.Message})
		return
	}

	var authResponse struct {
		IDToken      string `json:"idToken"`
		Email        string `json:"email"`
		RefreshToken string `json:"refreshToken"`
		ExpiresIn    string `json:"expiresIn"`
		LocalID      string `json:"localId"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Sign in successful",
		"token":   authResponse.IDToken,
		"user": gin.H{
			"uid":   authResponse.LocalID,
			"email": authResponse.Email,
		},
	})
}

func (ac *AuthController) VerifyToken(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userId":  userId,
		"message": "Token verified successfully",
	})
}

func (ac *AuthController) GetUserProfile(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var user models.User
	if err := database.DB.Where("firebase_uid = ?", userId).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Here you can fetch user profile from your database
	// For now, we'll just return the user ID
	c.JSON(http.StatusOK, gin.H{
		"userId": userId,
		"profile": map[string]interface{}{
			"email":       user.Email,
			"displayName": user.DisplayName,
			"phone":       user.Phone,
			"country":     user.Country,
			"bio":         user.Bio,
			"isAdmin":     user.IsAdmin,
			// Add more profile fields as needed
		},
	})
}

// ForgotPassword sends a password reset email
func (ac *AuthController) ForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call Firebase Auth REST API to send reset password email
	client := &http.Client{}
	reqBody := map[string]interface{}{
		"email":       input.Email,
		"requestType": "PASSWORD_RESET",
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing request"})
		return
	}

	firebaseAuthURL := "https://identitytoolkit.googleapis.com/v1/accounts:sendOobCode?key=AIzaSyDa2Vs7SWYRJbiJOkpPe7Pt_K29mlTaD2A"
	req, err := http.NewRequest("POST", firebaseAuthURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating request"})
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error communicating with authentication service"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": errorResponse.Error.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset email sent successfully",
	})
}

// GetAuthMethods returns available authentication methods
func (ac *AuthController) GetAuthMethods(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"methods": []string{
			"email/password",
			"google",
		},
		"defaultMethod": "email/password",
	})
}

func (ac *AuthController) generateStateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// InitiateGoogleSignIn starts the Google OAuth flow
func (ac *AuthController) InitiateGoogleSignIn(c *gin.Context) {
	// Generate a state token to prevent request forgery
	state := ac.generateStateToken()

	// Configure Google OAuth2 endpoint
	googleOauthConfig := &oauth2.Config{
		ClientID:     config.GetEnv("GOOGLE_CLIENT_ID", ""), // Get from Firebase Console
		ClientSecret: config.GetEnv("GOOGLE_CLIENT_SECRET", ""),
		RedirectURL:  config.GetEnv("HOST", "") + "/api/v1/auth/google/callback", // Your callback URL
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	// Get the auth URL with state
	url := googleOauthConfig.AuthCodeURL(state)

	// Instead of using cookies, we'll return both URL and state
	// The frontend will store the state and include it in the callback
	c.JSON(http.StatusOK, gin.H{
		"url":   url,
		"state": state,
	})
}

// HandleGoogleCallback processes the Google OAuth callback
func (ac *AuthController) HandleGoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code not found"})
		return
	}

	// Get state from query parameters
	state := c.Query("state")
	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "State parameter is missing"})
		return
	}

	code = c.Query("code")
	googleOauthConfig := &oauth2.Config{
		ClientID:     config.GetEnv("GOOGLE_CLIENT_ID", ""), // Get from Firebase Console
		ClientSecret: config.GetEnv("GOOGLE_CLIENT_SECRET", ""),
		RedirectURL:  config.GetEnv("HOST", "") + "/api/v1/auth/google/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	// Exchange auth code for token
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	// Get user info from Google
	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))
	peopleService, err := people.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Google service"})
		return
	}

	person, err := peopleService.People.Get("people/me").PersonFields("emailAddresses,names,photos").Do()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}

	// Get or create Firebase user
	var user *auth.UserRecord
	existingUser, err := ac.authClient.GetUserByEmail(context.Background(), person.EmailAddresses[0].Value)
	if err != nil {
		// User doesn't exist, create new user
		params := (&auth.UserToCreate{}).
			Email(person.EmailAddresses[0].Value).
			EmailVerified(true). // Google OAuth users are pre-verified
			DisplayName(person.Names[0].DisplayName)

		if len(person.Photos) > 0 {
			params = params.PhotoURL(person.Photos[0].Url)
		}

		user, err = ac.authClient.CreateUser(context.Background(), params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
	} else {
		user = existingUser
	}

	// Create or update user in database
	var dbUser models.User
	result := database.DB.Where("firebase_uid = ?", user.UID).First(&dbUser)
	if result.Error != nil {
		// User doesn't exist in database, create new user
		dbUser = models.User{
			FirebaseUID: user.UID,
			Email:       user.Email,
			IsAdmin:     false,
			DisplayName: user.DisplayName,
			PhotoURL:    user.PhotoURL,
		}
		if err := database.DB.Create(&dbUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user in database"})
			return
		}
	} else {
		// Update existing user's information
		dbUser.DisplayName = user.DisplayName
		dbUser.PhotoURL = user.PhotoURL
		if err := database.DB.Save(&dbUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user in database"})
			return
		}
	}

	// Create custom token
	customToken, err := ac.authClient.CustomToken(context.Background(), user.UID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create custom token"})
		return
	}

	// Exchange custom token for ID token using Firebase Auth REST API
	client = &http.Client{}
	reqBody := map[string]interface{}{
		"token":             customToken,
		"returnSecureToken": true,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing request"})
		return
	}

	firebaseAuthURL := "https://identitytoolkit.googleapis.com/v1/accounts:signInWithCustomToken?key=" + config.GetEnv("FIREBASE_WEB_API_KEY", "")
	req, err := http.NewRequest("POST", firebaseAuthURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating request"})
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error exchanging token"})
		return
	}
	defer resp.Body.Close()

	var idTokenResponse struct {
		IDToken      string `json:"idToken"`
		RefreshToken string `json:"refreshToken"`
		ExpiresIn    string `json:"expiresIn"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&idTokenResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing token response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Google sign in successful",
		"token":   idTokenResponse.IDToken,
		"user": gin.H{
			"uid":           user.UID,
			"email":         user.Email,
			"displayName":   user.DisplayName,
			"photoURL":      user.PhotoURL,
			"isAdmin":       dbUser.IsAdmin,
			"emailVerified": true,
		},
	})
}
