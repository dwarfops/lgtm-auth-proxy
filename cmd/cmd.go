package cmd

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dwarfops/lgtm-auth-proxy/auth"
	"github.com/dwarfops/lgtm-auth-proxy/http"
)

var (
	Root = &cobra.Command{
		Use:   "proxy",
		Short: "LGTM Proxy",
	}

	ServerCmd = &cobra.Command{
		Use:     "server",
		Aliases: []string{"s"},
		Short:   "Run the Proxy server",
		Run: func(cmd *cobra.Command, args []string) {
			var a auth.Auth
			switch viper.GetString("backend_type") {
			case "secretsmanager":
				smb, err := auth.NewSecretsManagerBackend(
					cmd.Context(),
					viper.GetString("secretsmanager.secret_id"),
					viper.GetDuration("secretsmanager.refresh_interval"),
					viper.GetDuration("secretsmanager.stale_threshold"))
				if err != nil {
					log.Fatal().Err(err).Msgf("Failed to create SecretsManagerBackend: %v", err)
				}
				defer smb.Stop()
				a = smb
			default:
				log.Fatal().Msgf("Invalid backend type: %s", viper.GetString("backend_type"))
			}

			go func() {
				conf, err := loadUpstreamConfig()
				if err != nil {
					log.Fatal().Err(err).Msg("Failed to load upstreams config")
				}
				log.Info().Msgf("Loaded %d upstreams", len(conf.Upstreams))
				for _, u := range conf.Upstreams {
					log.Info().Msgf("Upstream: Match=%s, Upstream=%s, Priority=%d", u.Match, u.Upstream, u.Priority)
				}
				for {
					if err := http.RunProxyServer(conf, a); err != nil {
						log.Error().Err(err).Msg("Failed to run proxy server")
					}
					log.Warn().Msgf("Proxy server exited, restarting in 10s")
					time.Sleep(10 * time.Second)
				}
			}()

			select {}
		},
	}
)

func loadUpstreamConfig() (http.ProxyConfig, error) {
	// Sometime we may want to make these configurable.
	DefaultPriority := 0
	DefaultTimeout := 30 * time.Second
	DefaultMaxIdleConns := 100
	DefaultIdleConnTimeout := 90 * time.Second

	var conf http.ProxyConfig
	conf.Listen = viper.GetString("proxy.listen")

	var upstreams []http.UpstreamConfig
	if err := viper.UnmarshalKey("proxy.upstreams", &upstreams); err != nil {
		return conf, fmt.Errorf("failed to unmarshal upstreams config: %w", err)
	}

	for i := range upstreams {
		// Priority
		if upstreams[i].Priority == 0 {
			upstreams[i].Priority = DefaultPriority
		}

		// Timeout
		if timeout := viper.GetString(fmt.Sprintf("proxy.upstreams.%d.timeout", i)); timeout != "" {
			parsedTimeout, err := time.ParseDuration(timeout)
			if err != nil {
				return conf, fmt.Errorf("invalid timeout for upstream %d: %w", i, err)
			}
			upstreams[i].Timeout = parsedTimeout
		} else {
			upstreams[i].Timeout = DefaultTimeout
		}

		// MaxIdleConns
		if upstreams[i].MaxIdleConns == 0 {
			upstreams[i].MaxIdleConns = DefaultMaxIdleConns
		}

		// IdleConnTimeout
		if idleTimeout := viper.GetString(fmt.Sprintf("proxy.upstreams.%d.idle_conn_timeout", i)); idleTimeout != "" {
			parsedIdleTimeout, err := time.ParseDuration(idleTimeout)
			if err != nil {
				return conf, fmt.Errorf("invalid idle_conn_timeout for upstream %d: %w", i, err)
			}
			upstreams[i].IdleConnTimeout = parsedIdleTimeout
		} else {
			upstreams[i].IdleConnTimeout = DefaultIdleConnTimeout
		}
	}

	conf.Upstreams = upstreams

	return conf, nil
}

func init() {
	Root.AddCommand(ServerCmd)
}
