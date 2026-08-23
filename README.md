# Autumn Grain Resilience

Autumn Grain Resilience coordinates crop monitoring, field operations, disaster response, drying capacity, recovery projects, and farmer support for agricultural authorities.

## Run

```bash
GOTOOLCHAIN=local go run ./cmd/server
```

The service applies versioned SQLite migrations at startup and exposes `/healthz` and `/readyz`.
