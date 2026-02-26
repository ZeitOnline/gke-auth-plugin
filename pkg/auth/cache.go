package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/traviswt/gke-auth-plugin/pkg/conf"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/pkg/apis/clientauthentication/v1"
)

func GetExecCredential() *v1.ExecCredential {
	cl := cacheLocation()
	if cl == "" {
		return nil
	}
	ec, err := loadFile(cl)
	if err != nil {
		return nil
	}
	if ec.Status != nil && ec.Status.ExpirationTimestamp != nil &&
		ec.Status.ExpirationTimestamp.Before(&metav1.Time{Time: time.Now()}) {
		os.Remove(cl)
		return nil
	}
	return ec
}

func SaveExecCredential(ec *v1.ExecCredential) {
	doNotCache := os.Getenv("GKE_AUTH_PLUGIN_DO_NOT_CACHE")
	if strings.ToLower(doNotCache) == "true" {
		return
	}
	cl := cacheLocation()
	if cl == "" {
		return
	}
	_ = saveFile(cl, ec)
}

// cacheLocation returns the file to Cache the exec cred to, if blank, don't Cache
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

func loadFile(file string) (*v1.ExecCredential, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var ec v1.ExecCredential
	err = json.Unmarshal(data, &ec)
	if err != nil {
		return nil, err
	}
	return &ec, nil
}

func saveFile(file string, ec *v1.ExecCredential) error {
	if ec == nil {
		return nil
	}
	data, err := json.MarshalIndent(ec, "", "  ")
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	return os.WriteFile(file, data, 0600)
}
