package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/traviswt/gke-auth-plugin/pkg/conf"
	v1 "k8s.io/client-go/pkg/apis/clientauthentication/v1"
)

type Credentials struct {
	AccessToken                string    `json:"access_token"`
	TokenExpiry                time.Time `json:"token_expiry"`
	ImpersonatedServiceAccount string    `json:"impersonated_service_account"`
}

func GetCachedCredentials(impersonationAccount string) *v1.ExecCredential {
	cl := cacheLocation()
	if cl == "" {
		return nil
	}

	creds, err := loadFile(cl)
	if err != nil {
		return nil
	}

	// Check if token has expired
	if creds.TokenExpiry.Before(time.Now()) {
		return nil
	}

	// Ensure that requested service account (if any) matches cached credentials
	if impersonationAccount != creds.ImpersonatedServiceAccount {
		return nil
	}

	return newExecCredential(creds.AccessToken, creds.TokenExpiry)
}

func SaveCredentialsToCache(ec *Credentials) {
	doNotCache := os.Getenv("GKE_AUTH_PLUGIN_DO_NOT_CACHE")
	if strings.ToLower(doNotCache) == "true" {
		return
	}
	cl := cacheLocation()
	if cl == "" {
		return
	}
	err := saveFile(cl, ec)
	if err != nil {
		fmt.Printf("error writing credentials to cache: %v\n", err)
	}
}

// cacheLocation returns the file to Cache the exec cred to, if blank, don't Cache.
// KUBECONFIG takes precedence, if not set, the .kube directory will be used, if it exists.
// Otherwise credentials will be cached in the user's home directory, which is not ideal,
// but better than nothing.
func cacheLocation() string {
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return filepath.Join(filepath.Dir(kubeconfig), conf.CacheFileName)
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	kubeDir := filepath.Join(userHomeDir, ".kube")
	if _, err := os.Stat(kubeDir); err == nil {
		return filepath.Join(kubeDir, conf.CacheFileName)
	}

	return filepath.Join(userHomeDir, conf.CacheFileName)
}

func loadFile(file string) (*Credentials, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var cachedCredentials Credentials
	err = json.Unmarshal(data, &cachedCredentials)
	if err != nil {
		return nil, err
	}
	return &cachedCredentials, nil
}

func saveFile(file string, creds *Credentials) error {
	if creds == nil {
		return nil
	}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	return os.WriteFile(file, data, 0600)
}
