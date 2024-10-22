# LGTM Auth Proxy

LGTM Auth Proxy is a robust, configurable reverse proxy designed to authenticate and route requests to various backend services which utilize the X-Scope-OrgID header. It's specifically built for securing and managing access to Mimir, Loki, and Tempo.

## Features

- Token-based authentication with tenant isolation
- Configurable upstream routing based on hostname patterns
- Dynamic secret management using AWS Secrets Manager
- Health check endpoint for Kubernetes readiness probes
- Efficient routing with priority-based matching
- Customizable HTTP client settings for each upstream

## Installation


### Local

```bash
go get github.com/dwarfops/lgtm-auth-proxy
```

### Kubernetes

```bash
helm repo add dwarfops public.ecr.aws/l6l5o3s2
helm repo update
helm show values dwarfops/helm/lgtm-auth-proxy
helm install lgtm-auth-proxy dwarfops/helm/lgtm-auth-proxy
```

Examples values

```yaml
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
config: |
  backend_type = "secretsmanager"
  log_level = "info"

  [secretsmanager]
  secret_id = "testing"
  refresh_interval = "1m"
  # If secret has not been updated in this time, it will be considered stale and /ready will return 503.
  stale_threshold = "15m"

  [proxy]
  listen = ':{{ $.Values.service.port }}'

  [[proxy.upstreams]]
  match = "loki\\.dwarfops\\.com"
  upstream = "http://loki-gateway.loki.svc.cluster.local"
  priority = 100
  timeout = "30s"
  max_idle_conns = 100
  idle_conn_timeout = "90s"

  [[proxy.upstreams]]
  match = ".*"
  upstream = "http://mimir-gateway.mimir.svc.cluster.local"
  priority = 0
resources:
  requests:
    cpu: 200m
    memory: 64Mi
  limits:
    cpu: 500m
    memory: 64Mi
serviceAccount:
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::1234567890:role/lgtm-auth-proxy
```

Required AWS Permissions
```hcl
data "aws_iam_policy_document" "lgtm_auth_proxy" {
  statement {
    effect = "Allow"

    actions = [
      "secretsmanager:GetSecretValue",
    ]

    resources = [
      "arn:aws:secretsmanager:us-east-1:1234567890:secret:testing*"
    ]
  }
}
```

## Configuration

LGTM Auth Proxy uses TOML for configuration. Here's an example configuration file:

```toml
backend_type = "secretsmanager" # Currently only supported option.
log_level = "info"

[secretsmanager]
secret_id = "testing"
refresh_interval = "1m"
# If secret has not been updated in this time, it will be considered stale and /ready will return 503.
stale_threshold = "15m"

[proxy]
listen = ":8000"

[[proxy.upstreams]]
match = "loki\\.dwarfops\\.com"
upstream = "http://loki.svc.local"
priority = 100
timeout = "30s"
max_idle_conns = 100
idle_conn_timeout = "90s"

[[proxy.upstreams]]
match = ".*"
upstream = "http://mimir.svc.local"
priority = 0
```

### Proxy Configuration

- `listen`: The address and port on which the proxy server will listen.
- `upstreams`: An array of upstream configurations.
  - `match`: A regex pattern to match against incoming request hostnames.
  - `upstream`: The URL of the backend service.
  - `priority`: Priority for route matching (higher numbers are checked first). [Default 0]
  - `timeout`: Request timeout for this upstream. [Default 30s]
  - `max_idle_conns`: Maximum number of idle connections. [Default 100]
  - `idle_conn_timeout`: How long an idle connection is kept in the pool. [Default 90s]

### Secrets Manager Configuration

- `secret_id`: The AWS Secrets Manager secret ID containing authentication tokens.
- `refresh_interval`: How often to refresh the secrets from AWS.
- `max_stale_duration`: Maximum allowed staleness of secrets before considering the service unhealthy.

## Usage

1. Set up your configuration file as described above.
2. Run the proxy:

   ```bash
   ./lgtm-auth-proxy --config path/to/your/config.toml
   ```

3. The proxy will start and listen on the configured address.

## Authentication

The proxy expects incoming requests to include:

- An `Authorization` header with a Bearer token
- An `X-Scope-OrgID` header specifying the tenant

Example:
```
Authorization: Bearer your-secret-token
X-Scope-OrgID: tenant1
```

## AWS Secrets Manager

The proxy uses AWS Secrets Manager to store and retrieve authentication tokens. The secret should be a JSON string mapping tokens to tenant IDs:

```json
{
  "token1": "tenant1",
  "token2": "tenant1",
  "token3": "tenant2"
}
```

Ensure that your AWS credentials are properly configured to access the specified secret.

## Health Check

The `/ready` endpoint can be used for Kubernetes readiness probes. It returns:

- 200 OK if the proxy is ready to serve traffic
- 503 Service Unavailable if the secrets are stale or unavailable

## Development

To contribute to LGTM Auth Proxy:

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

Distributed under the MIT License. See `LICENSE` for more information.

## Contact

[@dwarfops](https://twitter.com/dwarfops) - foss@dwarfops.com
