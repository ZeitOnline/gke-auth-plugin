package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/impersonate"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientauthv1 "k8s.io/client-go/pkg/apis/clientauthentication/v1"
)

var (
	gcpScopes = []string{
		"https://www.googleapis.com/auth/cloud-platform",
		"https://www.googleapis.com/auth/userinfo.email",
	}
)

func getTokenSource(ctx context.Context, impersonationAccount string) (oauth2.TokenSource, error) {
	if impersonationAccount != "" {
		return impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
			TargetPrincipal: impersonationAccount,
			Scopes:          gcpScopes,
		})
	}

	cred, err := google.FindDefaultCredentials(ctx, gcpScopes...)
	if err != nil {
		return nil, err
	}

	if cred == nil {
		return nil, errors.New("failed finding default credentials")
	}

	return cred.TokenSource, nil
}

func Gcp(ctx context.Context, impersonationAccount string) error {
	// Use cached exec credential
	if ec := GetCachedCredentials(impersonationAccount); ec != nil {
		credString := formatJSON(ec)
		fmt.Print(credString)
		return nil
	}

	ts, err := getTokenSource(ctx, impersonationAccount)
	if err != nil {
		return err
	}

	token, err := ts.Token()
	if err != nil {
		return err
	}

	// Create ExecCredential from token
	ec := newExecCredential(token.AccessToken, token.Expiry)
	creds := Credentials{token.AccessToken, token.Expiry, impersonationAccount}

	// Cache exec credential
	SaveCredentialsToCache(&creds)
	credString := formatJSON(ec)
	fmt.Print(credString)
	return nil
}

func formatJSON(ec *clientauthv1.ExecCredential) string {
	//pretty print
	enc, _ := json.MarshalIndent(ec, "", "  ")
	return string(enc)
}

func newExecCredential(token string, exp time.Time) *clientauthv1.ExecCredential {
	expiryTime := metav1.NewTime(exp)
	//the google token sometimes contains trailing periods,
	//they cause problems with various tools, thus right trim
	token = strings.TrimSuffix(token, ".")
	ec := &clientauthv1.ExecCredential{
		TypeMeta: metav1.TypeMeta{
			APIVersion: clientauthv1.SchemeGroupVersion.Identifier(),
			Kind:       "ExecCredential",
		},
		Status: &clientauthv1.ExecCredentialStatus{
			ExpirationTimestamp: &expiryTime,
			Token:               token,
		},
	}
	return ec
}
