package http

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/dwarfops/lgtm-auth-proxy/auth"
)

type ProxyServer struct {
	auth    auth.Auth
	proxies map[string]*httputil.ReverseProxy
	routes  []Route
}

type Route struct {
	Match    *regexp.Regexp
	Proxy    *httputil.ReverseProxy
	Priority int
}

type UpstreamConfig struct {
	Match           string
	Upstream        string
	Priority        int
	Timeout         time.Duration
	MaxIdleConns    int
	IdleConnTimeout time.Duration
}

type ProxyConfig struct {
	Listen    string
	Upstreams []UpstreamConfig
}

func NewProxyServer(config ProxyConfig, a auth.Auth) (*ProxyServer, error) {
	server := &ProxyServer{
		auth:    a,
		proxies: make(map[string]*httputil.ReverseProxy), // Initialize the map
		routes:  make([]Route, 0, len(config.Upstreams)),
	}

	for _, upstream := range config.Upstreams {
		u, err := url.Parse(upstream.Upstream)
		if err != nil {
			return nil, fmt.Errorf("failed to parse upstream URL %s: %w", upstream.Upstream, err)
		}

		proxy := httputil.NewSingleHostReverseProxy(u)
		proxy.Transport = &http.Transport{
			MaxIdleConns:       upstream.MaxIdleConns,
			IdleConnTimeout:    upstream.IdleConnTimeout,
			DisableCompression: true,
		}

		// Modify the Director function to add headers to the request
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.Header.Set("X-Proxy", "lgtm-auth-proxy")
		}
		proxy.ModifyResponse = func(resp *http.Response) error {
			resp.Header.Set("X-Proxy", "lgtm-auth-proxy")
			return nil
		}

		server.proxies[upstream.Upstream] = proxy

		re, err := regexp.Compile(upstream.Match)
		if err != nil {
			return nil, fmt.Errorf("failed to compile regex %s: %w", upstream.Match, err)
		}

		route := Route{
			Match:    re,
			Proxy:    proxy,
			Priority: upstream.Priority,
		}
		server.routes = append(server.routes, route)
	}

	// Sort routes by priority (highest first)
	sort.Slice(server.routes, func(i, j int) bool {
		return server.routes[i].Priority > server.routes[j].Priority
	})

	return server, nil
}

func (s *ProxyServer) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	if s.auth.Ready() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Not Ready"))
	}
}

func (s *ProxyServer) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}

func (s *ProxyServer) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	tenantID := r.Header.Get("X-Scope-OrgID")

	if !strings.HasPrefix(token, "Bearer ") {
		http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)
		return
	}
	token = strings.TrimPrefix(token, "Bearer ")

	if token == "" || tenantID == "" {
		http.Error(w, "Missing Authorization or X-Tenant-ID header", http.StatusUnauthorized)
		return
	}

	if err := s.auth.ValidateTokenForTenant(r.Context(), token, tenantID); err != nil {
		log.Info().Err(err).Msg("ValidateToken returned error")
		http.Error(w, "Invalid token for tenant", http.StatusUnauthorized)
		return
	}

	i := sort.Search(len(s.routes), func(i int) bool {
		return s.routes[i].Match.MatchString(r.Host)
	})

	var targetProxy *httputil.ReverseProxy
	if i < len(s.routes) {
		targetProxy = s.routes[i].Proxy
	}

	if targetProxy == nil {
		http.Error(w, "No matching upstream found", http.StatusNotFound)
		return
	}

	targetProxy.ServeHTTP(w, r)
}

func RunProxyServer(config ProxyConfig, a auth.Auth) error {
	server, err := NewProxyServer(config, a)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ready", server.ReadinessHandler)
	mux.HandleFunc("/alive", server.LivenessHandler)
	mux.HandleFunc("/", server.ProxyHandler)

	log.Info().Msgf("Starting proxy server on %s", config.Listen)
	if err := http.ListenAndServe(config.Listen, mux); err != nil {
		return fmt.Errorf("failed to start proxy server: %w", err)
	}
	return nil
}
