# SmartThings metrics
![ci](https://github.com/moikot/smartthings-metrics/workflows/ci/badge.svg)

A micro-service that provides SmartThings metrics to Prometheus.

## Run

For this service to have access to SmartThings API, you need to provide it with OAuth credentials. To get OAuth credentials, do the following:

1. Go to the [SmartThings Developer Workspace](https://account.smartthings.com/workspace)
2. Create a new app or select an existing one
3. Under "App Credentials", you'll find your Client ID and Client Secret
4. Make sure your app has the following OAuth scopes:
   - `devices:read`

### Run as a standalone app

**Prerequisites:**
  * [Golang >=1.14](https://golang.org/doc/install)

```bash
$ go get github.com/moikot/smartthings-metrics
$ smartthings-metrics -client-id YOUR_CLIENT_ID -client-secret YOUR_CLIENT_SECRET
$ curl localhost:9153/metrics
```

**Note:** Using `-interval` you can define the refresh interval in seconds. The default value of the refresh interval is 60 seconds.

### Generating Self-Signed Certificates

To generate a self-signed certificate for testing or development purposes, you can use OpenSSL:

```bash
# Generate a private key
openssl genrsa -out key.pem 2048

# Generate a self-signed certificate
openssl req -x509 -new -nodes \
  -key key.pem \
  -sha256 -days 365 \
  -out cert.pem \
  -subj "/CN=localhost"
```

**Note:** Self-signed certificates are not recommended for production use. For production environments, use certificates from a trusted Certificate Authority (CA).

### Using Let's Encrypt Certificates

For production environments, you can use free SSL certificates from Let's Encrypt. Here's how to set it up:

1. Install Certbot (Let's Encrypt client):
```bash
# On Ubuntu/Debian
sudo apt-get update
sudo apt-get install certbot

# On macOS with Homebrew
brew install certbot
```

2. Obtain a certificate:
```bash
# For a single domain
sudo certbot certonly --standalone -d your-domain.com

# For multiple domains
sudo certbot certonly --standalone -d your-domain.com -d www.your-domain.com
```

3. The certificates will be stored in:
   - Certificate: `/etc/letsencrypt/live/your-domain.com/fullchain.pem`
   - Private key: `/etc/letsencrypt/live/your-domain.com/privkey.pem`

4. Use the certificates with the service:
```bash
$ smartthings-metrics -client-id YOUR_CLIENT_ID -client-secret YOUR_CLIENT_SECRET \
  -cert-file /etc/letsencrypt/live/your-domain.com/fullchain.pem \
  -key-file /etc/letsencrypt/live/your-domain.com/privkey.pem
```

**Note:** Let's Encrypt certificates expire after 90 days. Set up automatic renewal:
```bash
# Test renewal
sudo certbot renew --dry-run

# Add to crontab (runs twice daily)
echo "0 0,12 * * * root python -c 'import random; import time; time.sleep(random.random() * 3600)' && certbot renew -q" | sudo tee -a /etc/crontab > /dev/null
```

#### HTTPS Support

To enable HTTPS, provide SSL certificate and private key files:

```bash
$ smartthings-metrics -client-id YOUR_CLIENT_ID -client-secret YOUR_CLIENT_SECRET \
  -cert-file /path/to/cert.pem -key-file /path/to/key.pem
$ curl -k https://localhost:9153/metrics
```

You can also set the certificate files using environment variables:
```bash
export SSL_CERT_FILE=/path/to/cert.pem
export SSL_KEY_FILE=/path/to/key.pem
```

### Run as a Docker container

**Prerequisites:**
  * [Docker](https://docs.docker.com/get-docker/)

```bash
$ docker run -d --rm -p 9153:9153 moikot/smartthings-metrics -client-id YOUR_CLIENT_ID -client-secret YOUR_CLIENT_SECRET
$ curl localhost:9153/metrics
```

#### HTTPS Support

To enable HTTPS in Docker, mount your SSL certificate files:

```bash
$ docker run -d --rm -p 9153:9153 \
  -v /path/to/cert.pem:/etc/ssl/certs/cert.pem \
  -v /path/to/key.pem:/etc/ssl/private/key.pem \
  moikot/smartthings-metrics \
  -client-id YOUR_CLIENT_ID \
  -client-secret YOUR_CLIENT_SECRET \
  -cert-file /etc/ssl/certs/cert.pem \
  -key-file /etc/ssl/private/key.pem
$ curl -k https://localhost:9153/metrics
```

### Deploy to a Kubernetes cluster

**Prerequisites:**
  * [Kuberentes](https://kubernetes.io/)
  * [Helm 3](https://helm.sh)

SmartThing metrics service is installed to Kubernetes via its [Helm chart](https://github.com/moikot/helm-charts/tree/master/charts/smartthings-metrics).

```
$ helm repo add moikot https://moikot.github.io/helm-charts
$ helm install smartthings-metrics moikot/smartthings-metrics --create-namespace --namespace smartthings \
  --set clientId=YOUR_CLIENT_ID \
  --set clientSecret=YOUR_CLIENT_SECRET
```

#### HTTPS Support

To enable HTTPS in Kubernetes, create a secret with your SSL certificate and key:

```bash
$ kubectl create secret tls smartthings-metrics-tls \
  --cert=/path/to/cert.pem \
  --key=/path/to/key.pem \
  -n smartthings

$ helm install smartthings-metrics moikot/smartthings-metrics \
  --create-namespace \
  --namespace smartthings \
  --set clientId=YOUR_CLIENT_ID \
  --set clientSecret=YOUR_CLIENT_SECRET \
  --set ssl.enabled=true \
  --set ssl.certSecret=smartthings-metrics-tls
```

## How it works

The service uses [SmartThings API](https://smartthings.developer.samsung.com/docs/api-ref/st-api.html) to obtain the current status of all connected devices periodically. It exposes the metrics received at `localhost:9153/metrics` so that [Prometheus](https://prometheus.io/) could scrape them.

### Gauges

Metric are exposed as Prometheus [gauges](https://prometheus.io/docs/concepts/metric_types/#gauge) their names are formed using pattern `smartthings_[component_name]_[capability_name]_[attribute_name]_[measurement_unit]` with all individual names converted to the snake case.

**Examples:**
 * `smartthings_motion_sensor_motion`
 * `smartthings_battery_battery_percent`
 * `smartthings_power_meter_power_watt`

**Note:**  
  * The component name is used unless it the `main` component.
  * The measurement unit is added for the values with units only.
  * Gauge with the name  `smartthings_health_state` is used for the health probe.

### Values

All the attributes of type `number` and `integer`