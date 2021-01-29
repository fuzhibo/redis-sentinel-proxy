redis-sentinel-proxy-service
====================

Small command utility that:

* Given a redis sentinel server listening on `SENTINEL_PORT`, keeps asking it for the address of a master named `NAME`

* Proxies all tcp requests that it receives on `PORT` to that master


Usage:

`./redis-sentinel-proxy-service -listen IP:PORT -sentinel IP1:SENTINEL_PORT,IP2:SENTINEL_PORT -master NAME`

## Usage

Edit `kubernetes/redis-sentinel-proxy-service-deployment.yaml`:

```bash
vim kubernetes/redis-sentinel-proxy-service-deployment.yaml
...
        args:
          - "-master"
          - "primary"
          - "-sentinel"
          - "redis-sentinel.$(NAMESPACE):26379" # change this to the sentinel address
```

Create `redis-sentinel-proxy-service-deployment` that uses `redis-sentinel-proxy-service`:

```bash
kubectl apply -f kubernetes/redis-sentinel-proxy-service-deployment.yaml
deployment "redis-sentinel-proxy-service" configured
```

Check if deployment is running: 

```bash
kubectl get pods
redis-sentinel-proxy-service-2064359825-s4n0k   1/1       Running   0          1d
```

Expose `redis-sentinel-proxy-service-deployment`:

```bash
kubectl apply -f kubernetes/redis-sentinel-proxy-service-service.yaml
```

