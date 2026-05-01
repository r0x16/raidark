# UUIDv7 — Identificadores ordenables por tiempo

Paquete: `github.com/r0x16/Raidark/shared/ids`

## Por qué UUIDv7

UUID v7 (RFC 9562) embebe un timestamp Unix en milisegundos en los 48 bits más significativos. Esto produce identificadores que:

- Son monotónicamente crecientes dentro del mismo milisegundo.
- Mantienen los índices BTree densos (sin fragmentación por inserciones aleatorias).
- Permiten inferir el orden cronológico directamente del ID, sin columna `created_at` adicional.
- Son globalmente únicos sin coordinación central.

## API

### `ids.NewV7() string`

Genera un nuevo UUID v7 en forma canónica (`xxxxxxxx-xxxx-7xxx-yxxx-xxxxxxxxxxxx`).

```go
import "github.com/r0x16/Raidark/shared/ids"

id := ids.NewV7()
// "0196b3a2-1c4f-7e3d-a5f2-0123456789ab"
```

### `ids.IsValidV7(s string) bool`

Valida que una cadena sea un UUID v7 bien formado (forma y bits de versión/variante).

```go
ids.IsValidV7("0196b3a2-1c4f-7e3d-a5f2-0123456789ab") // true
ids.IsValidV7("550e8400-e29b-41d4-a716-446655440000") // false (v4)
ids.IsValidV7("not-a-uuid")                           // false
```

### `ids.UUIDv7` (tipo)

Tipo `string` con soporte nativo de GORM (`GormDataType`, `Scan`, `Value`). Usable como campo PK en cualquier modelo.

```go
type MyEntity struct {
    ID   ids.UUIDv7 `gorm:"primarykey;type:varchar(36)"`
    Name string
}
```

Si lo usas de forma manual, debes poblar `ID` antes de crear el registro, o usar `ids.BaseModel` (ver abajo).

## Uso como PK GORM — `ids.BaseModel`

`ids.BaseModel` es un struct embebible que reemplaza a `gorm.Model` cuando se requiere PK UUID v7. Incluye hook `BeforeCreate` que auto-genera el ID si está vacío.

```go
import (
    "github.com/r0x16/Raidark/shared/ids"
)

type Post struct {
    ids.BaseModel          // ID UUIDv7 + CreatedAt + UpdatedAt + DeletedAt
    Title   string
    Content string
}

// Crear un registro — el ID se genera automáticamente:
post := Post{Title: "Hola mundo"}
db.Create(&post)
fmt.Println(post.ID) // "0196b3a2-1c4f-7e3d-a5f2-0123456789ab"

// Leer por ID:
var found Post
db.First(&found, "id = ?", post.ID)
```

### Diferencias con `shared/datastore/domain/BaseModel`

| Característica | `datastore/domain.BaseModel` | `ids.BaseModel` |
|----------------|------------------------------|-----------------|
| Tipo de PK     | `uint` (autoincrement)       | `UUIDv7` (varchar 36) |
| Ordenación     | Numérica secuencial          | Temporal (RFC 9562) |
| Distribución   | Local a la BD                | Global, sin coordinación |

Usa `ids.BaseModel` para entidades nuevas que participen en eventos de dominio, mensajes o necesiten IDs intercambiables entre servicios.

## Variables de entorno

Ninguna. La generación usa el reloj del sistema y entropía del OS.

## Dependencia

```
github.com/google/uuid v1.6+
```

Ya declarada en `go.mod`. No requiere configuración adicional.
