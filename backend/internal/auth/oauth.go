package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Sosokker/todolist-backend/internal/config"
	"golang.org/x/oauth2"
	googleOAuth "golang.org/x/oauth2/google"
)

// GoogleUserInfo holds user details fetched from Google.
type GoogleUserInfo struct {
	ID            string `json:"id"`             // The unique Google ID
	Email         string `json:"email"`          // The user's email address
	VerifiedEmail bool   `json:"verified_email"` // Whether Google has verified the email
	Name          string `json:"name"`           // User's full name
	GivenName     string `json:"given_name"`     // First name
	FamilyName    string `json:"family_name"`    // Last name
	Picture       string `json:"picture"`        // URL to profile picture
	Locale        string `json:"locale"`         // User's locale (e.g., "en")
}

// OAuthProvider defines the interface for OAuth operations.
type OAuthProvider interface {
	GetAuthCodeURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error)
	FetchUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error)
	GetOAuth2Config() *oauth2.Config // Expose underlying config if needed
}

// googleOAuthProvider implements OAuthProvider for Google.
type googleOAuthProvider struct {
	cfg *oauth2.Config
}

// NewGoogleOAuthProvider creates a new provider instance configured for Google.
func NewGoogleOAuthProvider(appCfg *config.Config) OAuthProvider {
	return &googleOAuthProvider{
		cfg: &oauth2.Config{
			ClientID:     appCfg.OAuth.Google.ClientID,
			ClientSecret: appCfg.OAuth.Google.ClientSecret,
			RedirectURL:  appCfg.OAuth.Google.RedirectURL,
			Scopes:       appCfg.OAuth.Google.Scopes,
			Endpoint:     googleOAuth.Endpoint,
		},
	}
}

// GetAuthCodeURL generates the URL for Google's consent page.
func (g *googleOAuthProvider) GetAuthCodeURL(state string) string {
	// Add options like AccessTypeOffline to get a refresh token,
	authURL := g.cfg.AuthCodeURL(state, oauth2.AccessTypeOffline /*, oauth2.ApprovalForce, oauth2.SetAuthURLParam("prompt", "select_account") */)
	return authURL
}

// ExchangeCode exchanges the authorization code for an access token and refresh token.
func (g *googleOAuthProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := g.cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange google auth code '%s': %w", code, err)
	}
	if !token.Valid() {
		return nil, fmt.Errorf("exchanged token is invalid")
	}
	return token, nil
}

// FetchUserInfo uses the access token to get user details from Google's UserInfo endpoint.
func (g *googleOAuthProvider) FetchUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := g.cfg.Client(ctx, token)

	userInfoURL := "https://www.googleapis.com/oauth2/v3/userinfo" // v3 is common
	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to request google user info from %s: %w", userInfoURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google user info request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode google user info response: %w", err)
	}

	if userInfo.ID == "" || userInfo.Email == "" {
		return nil, fmt.Errorf("invalid user info received from google (missing ID or Email)")
	}

	return &userInfo, nil
}

// GetOAuth2Config returns the underlying oauth2.Config object.
func (g *googleOAuthProvider) GetOAuth2Config() *oauth2.Config {
	return g.cfg
}
