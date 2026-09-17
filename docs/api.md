# Movie Maven API

Base URL (local): `http://localhost:8001`

All endpoints are unauthenticated `GET`s and return JSON. Share [`../openapi.yaml`](../openapi.yaml) with other apps — it is the machine-readable contract (Postman, Insomnia, OpenAPI generators).

## Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/movies` | List movies already in the local catalog |
| `GET` | `/populate-movies` | Ingest one TMDB Discover page for 2025 English-language titles |
| `GET` | `/populate-movies-since` | Ingest TMDB pages from a start year through six months ago |
| `GET` | `/populate-genres` | Ingest TMDB's full movie and TV genre lists |

Duplicates are keyed by `tmdb_id`. A second ingest updates the existing row instead of inserting another.

---

### `GET /movies`

Returns a page of non-deleted catalog rows, ordered by `release_date` descending (`NULLS LAST`).

**Query**

| Name | Required | Description |
|------|----------|-------------|
| `page` | no | 1-based page number. Defaults to `1`. |
| `page_size` | no | Movies per page. Defaults to `20`, maximum `100`. |

A page past the last page returns an empty `movies` array.

```bash
curl "http://localhost:8001/movies"
curl "http://localhost:8001/movies?page=2&page_size=20"
```

**200**

```json
{
  "page": 1,
  "page_size": 20,
  "total_pages": 12,
  "total_results": 240,
  "movies": []
}
```

**400** — invalid `page` or `page_size`

**500** — `{ "error": "failed to load movies" }`

---

### `GET /populate-movies`

Fetches one TMDB Discover page (`primary_release_year=2025`, `with_original_language=en`, `vote_count.gte=50`) and upserts the results.

```bash
curl http://localhost:8001/populate-movies
```

**200**

```json
{
  "saved": 20,
  "movies": []
}
```

**502** — TMDB request failed (`failed to fetch movies from TMDB`)

**500** — parse or database save failed

Non-200 TMDB responses are forwarded with TMDB’s status code and body.

---

### `GET /populate-movies-since`

**Query**

| Name | Required | Description |
|------|----------|-------------|
| `year` | yes | Integer from 1870 through current year + 1. Inclusive lower bound (`{year}-01-01`). |

TMDB filters applied by the server:

- `primary_release_date.gte` = `{year}-01-01`
- `primary_release_date.lte` = six months before today
- `vote_count.gte` = `200`
- `with_original_language` = `en`
- `sort_by` = `primary_release_date.desc`

Pages are fetched sequentially up to TMDB’s `total_pages`, capped by `tmdbMaxDiscoverPages` (TMDB Discover itself tops out at 500 pages). This call can take a long time.

```bash
curl "http://localhost:8001/populate-movies-since?year=2020"
```

**200**

```json
{
  "year": 2020,
  "pages_fetched": 10,
  "total_pages": 10,
  "saved": 187
}
```

`saved` counts upserts (inserts and updates), not distinct new rows.

**400** — missing or invalid `year`

**502** — TMDB fetch failed; body also includes `year`, `pages_fetched`, and `saved`

**500** — parse or save failed

---

### `GET /populate-genres`

Fetches TMDB's movie genre list (`/genre/movie/list`) and TV genre list (`/genre/tv/list`) and upserts both into the local catalog.

```bash
curl http://localhost:8001/populate-genres
```

**200**

```json
{
  "saved": 39,
  "genres": []
}
```

**502** — TMDB request failed (`failed to fetch movie genres from TMDB` or `failed to fetch tv genres from TMDB`)

**500** — parse or database save failed

Non-200 TMDB responses are forwarded with TMDB's status code and body.

---

## Movie object

| Field | Type | Notes |
|-------|------|--------|
| `id` | integer | Local primary key |
| `tmdb_id` | integer | Unique TMDB id |
| `title` | string | |
| `original_title` | string | |
| `overview` | string | |
| `release_date` | string \| null | RFC 3339; null if TMDB had no parseable date |
| `poster_path` | string | TMDB path, not a full URL |
| `backdrop_path` | string | TMDB path, not a full URL |
| `original_language` | string | ISO 639-1, e.g. `en` |
| `vote_average` | number | |
| `vote_count` | integer | |
| `popularity` | number | TMDB popularity score |
| `adult` | boolean | |
| `genre_ids` | integer[] | TMDB genre ids |
| `created_at` | string | RFC 3339 |
| `updated_at` | string | RFC 3339 |

Soft-deleted rows are omitted from `/movies`.

## Genre object

| Field | Type | Notes |
|-------|------|--------|
| `id` | integer | TMDB genre id |
| `media_type` | string | `movie` or `tv`; part of the composite primary key alongside `id` |
| `name` | string | |
