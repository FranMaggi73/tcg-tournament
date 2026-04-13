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
    │   └── middleware.go               # Firebase Auth middleware
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
| DELETE | `/tournaments/:id` | DeleteTournament | Eliminar torneo (solo si status=registration) |
| POST | `/tournaments/:id/rounds/next` | NextRound | Generar pairings de siguiente ronda |
| PATCH | `/tournaments/:id/matches/:matchId` | UpdateMatchResult | Registrar/actualizar resultado |
| PATCH | `/tournaments/:id/players/:playerId/status` | UpdatePlayerStatus | Dropear jugador |
| POST | `/tournaments/:id/rollback` | RollbackRound | Eliminar ronda actual + recalcular |
| POST | `/friends` | AddFriend | Enviar solicitud de amistad |
| GET | `/friends` | GetFriends | Listar amigos aceptados |
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
    Format       string    `json:"format" firestore:"format"`      // BO1, BO3
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
    Status     string  `json:"status" firestore:"status"` // active, dropped
    HadBye     bool    `json:"hadBye" firestore:"hadBye"`
}
```

### Match (subcolección `tournaments/{id}/rounds/{roundId}/matches`)
```go
type Match struct {
    ID           string `json:"id" firestore:"id"`
    RoundID       string `json:"roundId" firestore:"roundId"`
    Player1ID    string `json:"player1Id" firestore:"player1Id"`
    Player2ID    string `json:"player2Id" firestore:"player2Id"` // "BYE" para byes
    Player1Score int    `json:"player1Score" firestore:"player1Score"`
    Player2Score int    `json:"player2Score" firestore:"player2Score"`
    WinnerID     string `json:"winnerId" firestore:"winnerId"`  // "" para empate
    Status       string `json:"status" firestore:"status"`      // scheduled, completed
}
```

## UpdateMatchResult — Flujo
1. Recibe `player1Score`, `player2Score`, `roundId` en el body
2. Valida scores según formato (BO1: 1-0/0-1, BO3: 2-0/2-1/0-2/1-2/1-1)
3. Busca el match existente con `GetMatch(tournamentID, roundID, matchID)`
4. Deriva `winnerId` a partir de los scores:
   - `player1Score > player2Score` → `winnerId = player1Id`
   - `player2Score > player1Score` → `winnerId = player2Id`
   - `player1Score == player2Score` → `winnerId = ""` (empate)
5. Ejecuta `ProcessMatchAtomic` en transacción de Firestore
6. Ejecuta `UpdateStandings` para recalcular tiebreakers

## Algoritmo Suizo Profesional
- **Rondas**: $\lceil\log_2(\text{jugadores activos})\rceil$
- **Puntaje**:
  - BO1: victoria=1, derrota=0
  - BO3: victoria=3, empate=1, derrota=0
- **Pairings**:
  - Emparejamiento por rango de puntaje (descendente).
  - Ronda 1: shuffle aleatorio (todos tienen score 0).
  - **No-Repeat**: Dos jugadores NUNCA se enfrentan más de una vez por torneo.
  - **Byes**: Si el número de jugadores activos es impar, el jugador con menor puntaje que NO haya tenido un bye lo recibe (cuenta como victoria, `player2Id = "BYE"`).
- **Standings & Tiebreakers**:
  1. Puntaje Total (Descendente)
  2. OMW% (Opponent Match Win %): Fuerza de los oponentes.
  3. GW% (Game Win %): Juegos ganados / Juegos jugados.
  4. OGW% (Opponent Game Win %): Promedio de GW% de los oponentes.
- **Drops**: Jugadores con status `dropped` son excluidos de pairings y standings.
- **Empates BO3**: `1-1` es válido → ambos jugadores reciben 1 punto (`draws++`).

## RollbackRound — Flujo
1. Valida que `currentRound > 0` y que el usuario es el judge
2. Obtiene todos los rounds del torneo
3. Para el round con `roundNumber == currentRound`:
   - Elimina todos sus matches de Firestore
   - Elimina el documento del round
4. Resetea todos los stats de jugadores a 0 (totalScore, wins, losses, draws, omw, gw, ogw)
5. Decrementa `currentRound`
6. Si `currentRound == 0`, cambia status a `registration`
7. Si quedan rounds, recalcula standings con `UpdateStandings`

## Auth & Seguridad
- Verificar Firebase ID token en cada request protegida (`middleware.AuthMiddleware`).
- **RBAC**: El UID del token debe coincidir con `createdBy` del torneo para acciones de Judge.
- **Atomicidad**: `ProcessMatchAtomic` usa Transacciones de Firestore para actualizar match + player stats.
- **Validación de scores**: El handler rechaza scores inválidos antes de procesar.

## Firestore — Estructura de colecciones
```
tournaments/{tournamentId}
├── (campos del documento)
├── players/{playerId}
└── rounds/{roundId}
    └── matches/{matchId}

friendships/{friendshipId}

notifications/{notificationId}
```

## Issues conocidos
- `ProcessMatchAtomic` usa `+=` para scores — editar un resultado existente duplica puntos. Workaround: siempre llamar `UpdateStandings` después, pero los stats incrementales se acumulan incorrectamente.
- `DeleteTournament` no elimina subcolecciones (players, rounds, matches) — quedan huérfanos en Firestore.
- No hay tests.