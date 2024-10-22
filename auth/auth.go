package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/rs/zerolog/log"
)

// Auth is used by the HTTP server to authorize requests for a given tenant.
type Auth interface {
	ValidateTokenForTenant(ctx context.Context, token string, tenantID string) error
	Ready() bool
}

// SecretsManagerBackend is an implementation of the Auth interface that uses
// AWS Secrets Manager.
type SecretsManagerBackend struct {
	RefreshInterval    time.Duration
	MaxStaleDuration   time.Duration
	client             *secretsmanager.Client
	secretID           string
	cache              map[string]string
	mu                 sync.RWMutex
	stopChan           chan struct{}
	lastRefreshTime    atomic.Value // stores time.Time
	lastRefreshSuccess atomic.Bool
}

// NewSecretsManagerBackend creates a new SecretsManagerBackend for AWS
// SecretsManager.
func NewSecretsManagerBackend(ctx context.Context, secretId string, refreshInterval time.Duration, maxStaleDuration time.Duration) (*SecretsManagerBackend, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	client := secretsmanager.NewFromConfig(cfg)

	smb := &SecretsManagerBackend{
		RefreshInterval:  refreshInterval,
		MaxStaleDuration: maxStaleDuration,
		client:           client,
		secretID:         secretId,
		cache:            make(map[string]string),
		stopChan:         make(chan struct{}),
	}
	smb.lastRefreshTime.Store(time.Time{})

	// Initial fetch of secrets
	if err := smb.refreshSecrets(ctx); err != nil {
		return nil, fmt.Errorf("initial secrets fetch failed: %w", err)
	}

	// Start background refresh process
	go smb.periodicRefresh(ctx)

	return smb, nil
}

func (s *SecretsManagerBackend) refreshSecrets(ctx context.Context) error {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: &s.secretID,
	}

	result, err := s.client.GetSecretValue(ctx, input)
	if err != nil {
		s.lastRefreshSuccess.Store(false)
		return fmt.Errorf("failed to get secret value: %w", err)
	}

	var secretData map[string]string
	if err := json.Unmarshal([]byte(*result.SecretString), &secretData); err != nil {
		s.lastRefreshSuccess.Store(false)
		return fmt.Errorf("failed to unmarshal secret data: %w", err)
	}

	s.mu.Lock()
	s.cache = secretData
	s.mu.Unlock()

	s.lastRefreshTime.Store(time.Now())
	s.lastRefreshSuccess.Store(true)

	return nil
}

func (s *SecretsManagerBackend) periodicRefresh(ctx context.Context) {
	ticker := time.NewTicker(s.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.refreshSecrets(ctx); err != nil {
				log.Error().Err(err).Msg("Failed to refresh secrets")
			} else {
				log.Debug().Msg("Refreshed secrets")
			}
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (s *SecretsManagerBackend) Stop() {
	close(s.stopChan)
}

func (s *SecretsManagerBackend) ValidateTokenForTenant(ctx context.Context, token string, tenantID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	expectedTenantID, exists := s.cache[token]
	if !exists {
		return fmt.Errorf("no tenant found for token")
	}
	if expectedTenantID != tenantID {
		return fmt.Errorf("token is not valid for the provided tenant ID")
	}

	return nil
}

func (s *SecretsManagerBackend) Ready() bool {
	if !s.lastRefreshSuccess.Load() {
		return false
	}

	lastRefresh := s.lastRefreshTime.Load().(time.Time)
	if time.Since(lastRefresh) > s.MaxStaleDuration {
		return false
	}

	return true
}
