# REST Pagination

Package: `github.com/r0x16/Raidark/shared/api/rest`

## Why keyset / cursor pagination

Offset-based pagination (`LIMIT N OFFSET M`) degrades at large offsets because the database must scan and discard M rows before returning N results. It also drifts when rows are inserted or deleted between pages.

Keyset pagination uses an opaque cursor that encodes the last-seen sort key. The query becomes `WHERE sort_key > cursor_value LIMIT N`, which runs in O(log N) regardless of page depth and is stable under concurrent writes.

## Wire shape

```json
{
  "items": [ /* array of T */ ],
  "pagination": {
    "next_cursor": "eyJpZCI6...",
    "limit": 20
  }
}
```

`next_cursor` is omitted when there is no next page. Clients must treat the cursor as opaque — its internal format may change between releases.

## API

### `rest.Page[T any]`

Generic response envelope for list endpoints.

```go
import "github.com/r0x16/Raidark/shared/api/rest"

type TopicSummary struct { ID string; Title string }

page := rest.Page[TopicSummary]{
    Items: topics,
    Pagination: rest.PageMeta{
        NextCursor: cursor,
        Limit:      rest.DefaultLimit,
    },
}
return c.JSON(http.StatusOK, page)
```

### `rest.EncodeCursor(v any) (string, error)`

Serializes any JSON-marshallable value to a URL-safe base64 string (no padding). Use this to encode the sort fields of the last item in the result set.

```go
type topicCursor struct {
    CreatedAt time.Time `json:"created_at"`
    ID        string    `json:"id"`
}

cursor, err := rest.EncodeCursor(topicCursor{CreatedAt: last.CreatedAt, ID: last.ID})
```

### `rest.DecodeCursor(cursor string, dst any) error`

Reverses `EncodeCursor`. Returns an error if the cursor string was tampered with or is malformed — treat this as an invalid request (400).

```go
var c topicCursor
if err := rest.DecodeCursor(rawCursor, &c); err != nil {
    return rest.RenderError(ctx, http.StatusBadRequest, &rest.RESTError{
        Code:    "common.invalid_cursor",
        Message: "The pagination cursor is invalid or expired.",
    })
}
```

### `rest.ClampLimit(limit int) int`

Validates and bounds the client-supplied limit:

- `<= 0` → `DefaultLimit` (20)
- `> MaxLimit` → `MaxLimit` (100)
- otherwise → unchanged

```go
limit := rest.ClampLimit(queryLimit)
```

### Constants

| Constant | Value | Description |
|----------|-------|-------------|
| `DefaultLimit` | 20 | Page size when the caller omits `limit`. |
| `MaxLimit` | 100 | Hard cap on page size. |

## Cursor security note

The cursor is base64-encoded JSON — it is **not encrypted**. Do not embed sensitive data (e.g. user IDs of other users, internal row versions) unless the endpoint already exposes that data. If tampering the cursor could grant unauthorized access, validate that the decoded fields are in range before using them in a query.
