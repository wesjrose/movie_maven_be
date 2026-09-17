# Movie Maven Backend

### Tech Stack
- Go
- Gin
- PostgreSQL (GORM)

### API

- Human-readable docs: [docs/api.md](docs/api.md)
- Integration spec (import into Postman, Insomnia, or OpenAPI codegen): [openapi.yaml](openapi.yaml), also served live at `GET /openapi.yaml` so consumers (e.g. a frontend codegen step) always get the current spec instead of a stale copy

Local server: `http://localhost:8001`

### Run with auto reload

```
go tool air
```
