# TCG Tournament — Backend

## Stack
- Go 1.22+
- Gin (HTTP router)
- Firebase Admin SDK
- Firestore como base de datos

## Comandos clave
- run: `go run cmd/main.go`
- test: `go test ./...`
- build: `go build ./cmd/main.go`

## Estructura
```
backend/
├── cmd/main.go                        # Entry point + router setup
└── internal/
    ├── handlers/
    │   ├── handlers.go                 # Tournament handlers
    │   └── friendship_handler.go       # Friendship handlers
    ├── middleware/
    │   └── middleware.go               # Firebase Auth middleware + CORS
    ├── models/
    │   └── models.go                   # Structs (Tournament, Player, Round, Match, Friendship)
    └── tournament/
        ├── repository.go               # Firestore CRUD operations
        └── swiss.go                    # Swiss pairing algorithm + standings
```

## API Routes

### Públicas (sin auth)
| Método | Ruta | Handler | Descripción |
|--------|------|---------|-------------|
| GET | `/tournaments/:id` | GetTournament | Obtener datos del torneo |
| GET | `/tournaments/:id/standings` | GetStandings | Obtener jugadores con stats |
| GET | `/tournaments/:id/export` | ExportStandings | Exportar standings ordenados |
| POST | `/tournaments/:id/players` | RegisterPlayer | Registrar jugador por tournament ID |
| POST | `/tournaments/join` | JoinTournamentByCode | Unirse con inviteCode |

### Protegidas (requieren Firebase ID token)
| Método | Ruta | Handler | Descripción |
|--------|------|---------|-------------|
| POST | `/tournaments` | CreateTournament | Crear torneo (genera inviteCode) |
| DELETE | `/tournaments/:id` | DeleteTournament | Eliminar torneo (no permitido en `playing`) |
| PATCH | `/tournaments/:id/complete` | CompleteTournament | Finalizar torneo (solo desde `playing`) |
| POST | `/tournaments/:id/rounds/next` | NextRound | Generar pairings de siguiente ronda |
| PATCH | `/tournaments/:id/matches/:matchId` | UpdateMatchResult | Registrar/actualizar resultado |
| PATCH | `/tournaments/:id/players/:playerId/status` | UpdatePlayerStatus | Dropear o restaurar jugador |
| DELETE | `/tournaments/:id/players/:playerId` | RemovePlayer | Eliminar jugador (solo en `registration`) |
| POST | `/tournaments/:id/rollback` | RollbackRound | Eliminar ronda actual + recalcular |
| POST | `/friends` | AddFriend | Enviar solicitud de amistad |
| GET | `/friends` | GetFriends | Listar amigos aceptados |
| GET | `/friends/pending` | GetPendingRequests | Listar solicitudes de amistad pendientes |
| PATCH | `/friends/:id` | UpdateFriendshipStatus | Aceptar/rechazar solicitud |

## Modelos (Go Structs)

### Tournament
```go
type Tournament struct {
    ID           string    `json:"id" firestore:"id"`
    Name         string    `json:"name" firestore:"name"`
    Date         time.Time `json:"date" firestore:"date"`
    MaxPlayers   int       `json:"maxPlayers" firestore:"maxPlayers"`
    CurrentRound int       `json:"currentRound" firestore:"currentRound"`
    TotalRounds  int       `json:"totalRounds" firestore:"totalRounds"`
    CreatedBy    string    `json:"createdBy" firestore:"createdBy"`
    Status       string    `json:"status" firestore:"status"`       // registration, playing, completed
    Format       string    `json:"format" firestore:"format"`       // BO1, BO3
    InviteCode   string    `json:"inviteCode" firestore:"inviteCode"`
}
```

### Player (subcolección `tournaments/{id}/players`)
```go
type Player struct {
    ID         string  `json:"id" firestore:"id"`
    Name       string  `json:"name" firestore:"name"`
    Email      string  `json:"email" firestore:"email"`
    TotalScore int     `json:"totalScore" firestore:"totalScore"`
    Wins       int     `json:"wins" firestore:"wins"`
    Losses     int     `json:"losses" firestore:"losses"`
    Draws      int     `json:"draws" firestore:"draws"`
    OMW        float64 `json:"omw" firestore:"omw"`
    GW         float64 `json:"gw" firestore:"gw"`
    OGW        float64 `json:"ogw" firestore:"ogw"`
    Status     string  `json:"status" firestore:"status"` // "active", "dropped"
    HadBye     bool    `json:"hadBye" firestore:"hadBye"`
}
```

### Round (subcolección `tournaments/{id}/rounds`)
```go
type Round struct {
    ID           string    `json:"id" firestore:"id"`
    TournamentID string    `json:"tournamentId" firestore:"tournamentId"`
    RoundNumber  int       `json:"roundNumber" firestore:"roundNumber"`
    Status       string    `json:"status" firestore:"status"` // "pairing", "playing", "completed"
    CreatedAt    time.Time `json:"createdAt" firestore:"createdAt"`
}
```

### Match (subcolección `tournaments/{id}/rounds/{roundId}/matches`)
```go
type Match struct {
    ID           string `json:"id" firestore:"id"`
    RoundID      string `json:"roundId" firestore:"roundId"`
    Player1ID    string `json:"player1Id" firestore:"player1Id"`
    Player2ID    string `json:"player2Id" firestore:"player2Id"` // "BYE" para byes
    Player1Score int    `json:"player1Score" firestore:"player1Score"`
    Player2Score int    `json:"player2Score" firestore:"player2Score"`
    WinnerID     string `json:"winnerId" firestore:"winnerId"`  // "" para empate
    Status       string `json:"status" firestore:"status"`      // "scheduled", "completed"
}
```

### Friendship (colección `friendships`)
```go
type Friendship struct {
    ID        string    `json:"id" firestore:"id"`
    User1ID   string    `json:"user1Id" firestore:"user1Id"`
    User2ID   string    `json:"user2Id" firestore:"user2Id"`
    Status    string    `json:"status" firestore:"status"` // "pending", "accepted"
    CreatedAt time.Time `json:"createdAt" firestore:"createdAt"`
}
```

## UpdateMatchResult — Flujo
1. Recibe `player1Score`, `player2Score`, `roundId` en el body
2. Valida scores según formato:
   - **BO3**: 2-0, 2-1, 0-2, 1-2 (victoria) o 1-1 (empate)
   - **BO1**: 1-0 o 0-1 (no hay empates)
3. Busca el match existente con `GetMatch(tournamentID, roundID, matchID)`
4. Deriva `winnerId` a partir de los scores:
   - `player1Score > player2Score` → `winnerId = player1Id`
   - `player2Score > player1Score` → `winnerId = player2Id`
   - `player1Score == player2Score` → `winnerId = ""` (empate)
5. Ejecuta `ProcessMatchResult` → `ProcessMatchAtomic` en transacción de Firestore
6. Ejecuta `UpdateStandings` para recalcular tiebreakers (OMW, GW, OGW)

## Scoring real (ProcessMatchResult en swiss.go)
**IMPORTANTE**: El código actual otorga los mismos puntos independientemente del formato:
- Victoria (cualquier formato): **3 puntos**
- Empate (BO3 1-1): **1 punto** para cada jugador
- Derrota: **0 puntos**
- Bye: **3 puntos** al jugador (cuenta como victoria)

La validación de scores SÍ es diferente por formato (BO1 no admite empates), pero el puntaje otorgado es siempre 3/1/0.

## Ciclo de vida del torneo

### CompleteTournament
- Solo funciona si `status == "playing"`
- Cambia `status` a `"completed"`
- No genera ninguna ronda adicional

### DeleteTournament
- **Prohibido** si `status == "playing"` (debe finalizarse primero)
- Permitido en `registration` y `completed`
- **Bug conocido**: no elimina subcolecciones (players, rounds, matches quedan huérfanos)

### RemovePlayer
- Solo permitido durante `status == "registration"`
- Elimina físicamente el documento del jugador
- Durante `playing` usar `UpdatePlayerStatus` con `status: "dropped"` en su lugar

## Algoritmo Suizo Profesional
- **Rondas**: $\lceil\log_2(\text{jugadores activos})\rceil$
- **Pairings**:
  - Ordenados por score descendente
  - Ronda 1: shuffle aleatorio (todos en 0)
  - **No-Repeat**: Dos jugadores NUNCA se enfrentan más de una vez
  - **Byes**: Si el número de jugadores activos es impar, el jugador con menor puntaje que NO haya tenido bye lo recibe. Cuenta como victoria automática (`player2Id = "BYE"`, `status = "completed"`)
- **Standings & Tiebreakers** (en orden):
  1. TotalScore (descendente)
  2. OMW% — Opponent Match Win %
  3. GW% — Game Win %
  4. OGW% — Opponent Game Win %
- **Drops**: Jugadores con `status == "dropped"` son excluidos de pairings futuros

## RollbackRound — Flujo
1. Valida `currentRound > 0` y que el usuario es el judge
2. Obtiene todos los rounds del torneo
3. Para el round con `roundNumber == currentRound`:
   - Elimina todos sus matches de Firestore
   - Elimina el documento del round
4. Resetea todos los stats de todos los jugadores a 0 (`totalScore`, `wins`, `losses`, `draws`, `omw`, `gw`, `ogw`)
5. Decrementa `currentRound`
6. Si `currentRound == 0` → cambia `status` a `"registration"`
7. Si quedan rondas previas → recalcula standings con `UpdateStandings`

**Nota**: El reset de stats es total (no incremental), por lo que si quedan rondas anteriores, `UpdateStandings` recalcula solo los tiebreakers (OMW/GW/OGW). Los campos `wins/losses/draws/totalScore` quedan en 0 hasta la próxima edición de resultado.

## Auth & Seguridad
- Verificar Firebase ID token en cada request protegida (`middleware.AuthMiddleware`)
- **RBAC**: El UID del token debe coincidir con `createdBy` del torneo para acciones de Judge
- **Atomicidad**: `ProcessMatchAtomic` usa Transacciones de Firestore para actualizar match + player stats simultáneamente
- **Validación de scores**: El handler rechaza scores inválidos antes de procesar

## Firestore — Estructura de colecciones
```
tournaments/{tournamentId}
├── (campos del documento)
├── players/{playerId}
└── rounds/{roundId}
    └── matches/{matchId}

friendships/{friendshipId}

notifications/{notificationId}

users/{uid}
```

## Issues conocidos
- `ProcessMatchAtomic` usa `+=` para scores — editar un resultado existente duplica `wins/losses/draws/totalScore`. `UpdateStandings` solo recalcula tiebreakers, no corrige los contadores acumulativos.
- `DeleteTournament` no elimina subcolecciones (players, rounds, matches quedan huérfanos en Firestore).
- El scoring siempre otorga 3 puntos por victoria sin importar el formato (BO1 y BO3 reciben el mismo puntaje).
- No hay tests.