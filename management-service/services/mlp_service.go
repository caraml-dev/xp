package services

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	mlp "github.com/caraml-dev/mlp/api/client"
	"github.com/caraml-dev/mlp/api/pkg/auth"
	"github.com/caraml-dev/xp/management-service/errors"
	"github.com/patrickmn/go-cache"
)

const (
	mlpCacheExpirySeconds  = 600
	mlpCacheCleanUpSeconds = 900
	mlpQueryTimeoutSeconds = 30
)

// MLPService provides a set of methods to interact with the MLP APIs
type MLPService interface {
	// GetProject gets the project matching the provided id
	GetProject(id int64) (*mlp.Project, error)
}

type mlpService struct {
	mlpClient *mlpClient
	cache     *cache.Cache
}

type mlpClient struct {
	api *mlp.APIClient
}

func newMLPClient(googleClient *http.Client, basePath string) *mlpClient {
	cfg := mlp.NewConfiguration()
	cfg.BasePath = basePath
	cfg.HTTPClient = googleClient

	return &mlpClient{mlp.NewAPIClient(cfg)}
}

// NewMLPService returns a service that retrieves information that is shared across MLP projects.
func NewMLPService(mlpBasePath string) (MLPService, error) {
	httpClient := http.DefaultClient
	googleClient, err := auth.InitGoogleClient(context.Background())
	if err == nil {
		httpClient = googleClient
	} else {
		log.Println("Google default credential not found. Fallback to HTTP default client")
	}

	svc := &mlpService{
		mlpClient: newMLPClient(httpClient, mlpBasePath),
		cache:     cache.New(mlpCacheExpirySeconds*time.Second, mlpCacheCleanUpSeconds*time.Second),
	}

	err = svc.refreshProjects()
	if err != nil {
		return nil, err
	}
	return svc, nil
}

// GetProject gets the project matching the provided id. This method will hit the cache first,
// and if not found, will call MLP API once to get the updated list of projects and refresh the cache,
// then try to get the value again. If still not found, will return a freecache NotFound error.
func (service mlpService) GetProject(id int64) (*mlp.Project, error) {
	project, err := service.getProject(id)
	if err != nil {
		err = service.refreshProjects()
		if err != nil {
			return nil, err
		}
		return service.getProject(id)
	}
	return project, nil
}

func (service mlpService) getProject(id int64) (*mlp.Project, error) {
	key := buildProjectKey(id)
	cachedValue, found := service.cache.Get(key)
	if !found {
		return nil, errors.Newf(errors.NotFound, "MLP Project info for id %d not found in the cache", id)
	}
	// Cast the data
	project, ok := cachedValue.(mlp.Project)
	if !ok {
		return nil, errors.Newf(errors.NotFound, "Malformed project info found in the cache for id %d", id)
	}
	return &project, nil
}

func (service mlpService) refreshProjects() error {
	ctx, cancel := context.WithTimeout(context.Background(), mlpQueryTimeoutSeconds*time.Second)
	defer cancel()

	projects, resp, err := service.mlpClient.api.ProjectApi.ProjectsGet(ctx, nil)
	if err != nil {
		return err
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	for _, project := range projects {
		key := buildProjectKey(int64(project.ID))
		service.cache.Set(key, project, cache.DefaultExpiration)
	}
	return nil
}

func buildProjectKey(id int64) string {
	return fmt.Sprintf("proj:%d", id)
}
