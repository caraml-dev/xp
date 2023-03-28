package services

import (
	"context"
	"net/http"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/caraml-dev/xp/common/testutils"
	"github.com/gojek/mlp/api/pkg/auth"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2/google"

	mlp "github.com/gojek/mlp/api/client"
)

func TestNewMLPService(t *testing.T) {
	reset := testutils.TestSetupEnvForGoogleCredentials(t)
	defer reset()

	// Create test Google client
	gc, err := auth.InitGoogleClient(context.Background())
	require.NoError(t, err)
	// Create test projects and environments
	projects := []mlp.Project{{ID: 1}}

	// Patch new MLP Client method
	defer monkey.UnpatchAll()
	monkey.Patch(newMLPClient,
		func(googleClient *http.Client, basePath string) *mlpClient {
			assert.Equal(t, gc, googleClient)
			assert.Equal(t, "mlp-base-path", basePath)
			// Create test client
			mlpClient := &mlpClient{
				api: &mlp.APIClient{
					ProjectApi: &mlp.ProjectApiService{},
				},
			}
			// Patch Get Projects
			monkey.PatchInstanceMethod(reflect.TypeOf(mlpClient.api.ProjectApi), "ProjectsGet",
				func(svc *mlp.ProjectApiService, ctx context.Context, localVarOptionals *mlp.ProjectApiProjectsGetOpts,
				) ([]mlp.Project, *http.Response, error) {
					return projects, nil, nil
				})
			return mlpClient
		},
	)

	svc, err := NewMLPService("mlp-base-path")
	assert.NoError(t, err)
	assert.NotNil(t, svc)
	// Test side effects
	proj, err := svc.GetProject(1)
	require.NotNil(t, proj)
	assert.Equal(t, projects[0], *proj)
	assert.NoError(t, err)
}

func TestNewMLPClient(t *testing.T) {
	reset := testutils.TestSetupEnvForGoogleCredentials(t)
	defer reset()

	// Create test Google client
	gc, err := google.DefaultClient(context.Background(), "https://www.googleapis.com/auth/userinfo.email")
	require.NoError(t, err)
	// Create expected MLP config
	cfg := mlp.NewConfiguration()
	cfg.BasePath = "base-path"
	cfg.HTTPClient = gc

	// Test
	resultClient := newMLPClient(gc, "base-path")
	require.NotNil(t, resultClient)
	assert.Equal(t, mlp.NewAPIClient(cfg), resultClient.api)
}

func TestMLPServiceGetProject(t *testing.T) {
	defer monkey.UnpatchAll()
	projects := []mlp.Project{
		{
			ID: 1,
		},
		{
			ID: 2,
		},
	}

	svc := newTestMLPService()
	monkey.PatchInstanceMethod(reflect.TypeOf(svc.mlpClient.api.ProjectApi), "ProjectsGet",
		func(svc *mlp.ProjectApiService, ctx context.Context, localVarOptionals *mlp.ProjectApiProjectsGetOpts,
		) ([]mlp.Project, *http.Response, error) {
			return projects, nil, nil
		})

	// getting valid project should refresh cache and return the project
	project, err := svc.GetProject(1)
	assert.NoError(t, err)
	assert.Equal(t, *project, projects[0])

	// getting invalid project should return error
	_, err = svc.GetProject(3)
	assert.Error(t, err)
}

func newTestMLPService() *mlpService {
	svc := &mlpService{
		mlpClient: &mlpClient{
			api: &mlp.APIClient{
				ProjectApi: &mlp.ProjectApiService{},
			},
		},
		cache: cache.New(time.Second*2, time.Second*2),
	}
	return svc
}
