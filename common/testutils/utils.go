package testutils

import (
	"io"
	"os"
	"testing"
)

// ReadFile reads a file and returns the byte contents
func ReadFile(filepath string) ([]byte, error) {
	// Open file
	fileObj, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer fileObj.Close()
	// Read contents
	byteValue, err := io.ReadAll(fileObj)
	if err != nil {
		return nil, err
	}
	return byteValue, nil
}

// TestSetupEnvForGoogleCredentials creates a temporary file containing dummy user account JSON
// then set the environment variable GOOGLE_APPLICATION_CREDENTIALS to point to the file.
// This is useful for tests that assume Google Cloud Client libraries can automatically find
// the user account credentials in any environment.
// At the end of the test, the returned function can be called to perform cleanup.
func TestSetupEnvForGoogleCredentials(t *testing.T) (reset func()) {
	userAccountKey := []byte(`{
		"client_id": "dummyclientid.apps.googleusercontent.com",
		"client_secret": "dummy-secret",
		"quota_project_id": "test-project",
		"refresh_token": "dummy-token",
		"type": "authorized_user"
	}`)

	file, err := os.CreateTemp("", "dummy-user-account")
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(file.Name(), userAccountKey, 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", file.Name())
	if err != nil {
		t.Fatal(err)
	}

	return func() {
		err := os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
		if err != nil {
			t.Log("Cleanup failed", err)
		}
		err = os.Remove(file.Name())
		if err != nil {
			t.Log("Cleanup failed", err)
		}
	}
}
