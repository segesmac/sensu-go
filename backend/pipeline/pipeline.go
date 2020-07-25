package pipeline

import (
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/licensing"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/cache"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/rpc"
	"github.com/sensu/sensu-go/types"
)

// Pipeline takes events as inputs, and treats them in various ways according
// to the event's check configuration.
type Pipeline struct {
	store                  store.Store
	assetGetter            asset.Getter
	backendEntity          *corev2.Entity
	extensionExecutor      ExtensionExecutorGetterFunc
	executor               command.Executor
	storeTimeout           time.Duration
	secretsProviderManager *secrets.ProviderManager
	licenseGetter          licensing.Getter
	handlersCache          resourceCache
}

// Config holds the configuration for a Pipeline.
type Config struct {
	Store                   store.Store
	AssetGetter             asset.Getter
	BackendEntity           *corev2.Entity
	ExtensionExecutorGetter ExtensionExecutorGetterFunc
	StoreTimeout            time.Duration
	SecretsProviderManager  *secrets.ProviderManager
	LicenseGetter           licensing.Getter
	HandlersCache           resourceCache
}

// Option is a functional option used to configure Pipelines.
type Option func(*Pipeline)

// resourceCache abstracts the cache.Resource struct for easier testing
type resourceCache interface {
	Get(namespace string) []cache.Value
}

// New creates a new Pipeline from the provided configuration.
func New(c Config, options ...Option) *Pipeline {
	pipeline := &Pipeline{
		store:                  c.Store,
		assetGetter:            c.AssetGetter,
		backendEntity:          c.BackendEntity,
		extensionExecutor:      c.ExtensionExecutorGetter,
		executor:               command.NewExecutor(),
		storeTimeout:           c.StoreTimeout,
		secretsProviderManager: c.SecretsProviderManager,
		licenseGetter:          c.LicenseGetter,
		handlersCache:          c.HandlersCache,
	}
	for _, o := range options {
		o(pipeline)
	}
	return pipeline
}

const (
	// DefaultSocketTimeout specifies the default socket dial
	// timeout in seconds for TCP and UDP handlers.
	DefaultSocketTimeout uint32 = 60
)

// ExtensionExecutorGetterFunc gets an ExtensionExecutor. Used to decouple
// pipelines from gRPC.
type ExtensionExecutorGetterFunc func(*types.Extension) (rpc.ExtensionExecutor, error)
