# TCG Tournament — Proyecto General

## Arquitectura
Monorepo con frontend y backend separados que se comunican vía REST API.

- `frontend/` — SvelteKit + TypeScript + Tailwind + DaisyUI
- `backend/` — Go 1.22+ + Gin + Firebase Admin SDK

## Firebase — roles por capa
- **Frontend**: Firebase client SDK → Auth (login/logout) + Firestore `onSnapshot` (real-time read-only)
- **Backend**: Firebase Admin SDK → solo verifica tokens JWT en cada request

El frontend NUNCA llama al backend para leer datos — lee Firestore directo.
El backend NUNCA escribe en Firestore directamente desde el cliente — solo el judge puede modificar vía API.

## Modelo de Datos

### Torneos (colección `tournaments`)
- Documento raíz con campos: `id`, `name`, `createdBy`, `status`, `currentRound`, `totalRounds`, `format`, `inviteCode`
- **Sin campo `participants`** — los jugadores viven en la subcolección `/tournaments/{id}/players`
- Subcolección `players`: documentos individuales por jugador
- Subcolección `rounds`: cada round tiene ID único (UUID, NO `round_N`)
  - Subcolección `matches` dentro de cada round

### Jugadores (subcolección `tournaments/{id}/players`)
Campos: `id`, `name`, `email`, `totalScore`, `wins`, `losses`, `draws`, `omw`, `gw`, `ogw`, `status`, `hadBye`

### Matches (subcolección `tournaments/{id}/rounds/{roundId}/matches`)
Campos: `id`, `roundId`, `player1Id`, `player2Id`, `player1Score`, `player2Score`, `winnerId`, `status`
- `player2Id = "BYE"` para byes
- `winnerId = ""` para empates

### Amistades (colección `friendships`)
Campos: `id`, `user1Id`, `user2Id`, `status`, `createdAt`

## Roles de usuario
- `judge`: crea torneo, avanza rondas, edita resultados, dropea jugadores (requiere pasar por el backend)
- `viewer`: ve el torneo en tiempo real (solo Firestore onSnapshot, sin backend)
- El rol se deriva del campo `createdBy` en Firestore

## Contrato entre frontend y backend
- El frontend envía Firebase ID token en el header `Authorization: Bearer <token>`
- El backend verifica el token y extrae el UID
- Si `UID == torneo.createdBy` → es judge, puede operar
- El backend devuelve JSON, el frontend nunca calcula pairings

## API Routes

### Públicas
| Método | Ruta | Handler |
|--------|------|---------|
| GET | `/tournaments/:id` | GetTournament |
| GET | `/tournaments/:id/standings` | GetStandings |
| GET | `/tournaments/:id/export` | ExportStandings |
| POST | `/tournaments/:id/players` | RegisterPlayer |
| POST | `/tournaments/join` | JoinTournamentByCode |

### Protegidas (requieren Auth)
| Método | Ruta | Handler |
|--------|------|---------|
| POST | `/tournaments` | CreateTournament |
| DELETE | `/tournaments/:id` | DeleteTournament |
| POST | `/tournaments/:id/rounds/next` | NextRound |
| PATCH | `/tournaments/:id/matches/:matchId` | UpdateMatchResult |
| PATCH | `/tournaments/:id/players/:playerId/status` | UpdatePlayerStatus |
| POST | `/tournaments/:id/rollback` | RollbackRound |
| POST | `/friends` | AddFriend |
| GET | `/friends` | GetFriends |
| PATCH | `/friends/:id` | UpdateFriendshipStatus |

## Ciclo de Vida del Torneo
- **Estado `registration`**: Los jugadores pueden unirse mediante un `InviteCode` único. El judge puede configurar el formato.
- **Estado `playing`**: Se activa al generar la primera ronda.
  - El `InviteCode` queda invalidado.
  - Ya no se pueden unir nuevos jugadores.
  - El formato (BO1/BO3) queda bloqueado y no puede cambiarse.
- **Estado `completed`**: Torneo finalizado.
- **Rollback**: Elimina la ronda actual y sus matches de Firestore. Si se vuelve a ronda 0, el status regresa a `registration`.

## Formatos y Validación de Resultados
- **BO1**: Solo se permite resultado 1-0 o 0-1. Puntaje: victoria=1, derrota=0.
- **BO3**: Solo se permiten resultados 2-0, 2-1, 0-2, 1-2 o 1-1 (empate).
  - Puntaje: victoria=3, empate=1, derrota=0.
  - El `winnerId` se deriva de los scores en el backend (vacío = empate).

## Algoritmo suizo (solo backend)
- Rondas = ceil(log2(jugadores))
- Tiebreakers: OMW% → GW% → OGW%
- Editar rondas pasadas recalcula standings pero NO re-parea rondas futuras

## Sistema de Amigos e Invitaciones
- Los usuarios pueden añadir amigos mediante solicitudes de amistad (`pending` → `accepted`).
- El judge puede listar sus amigos aceptados para facilitar el envío del código de invitación del torneo.
- Las invitaciones se envían como notificaciones en la colección `notifications` de Firestore.

## Comandos
- Frontend: `cd frontend && npm run dev`
- Backend: `cd backend && go run cmd/main.go`
- Tests backend: `cd backend && go test ./...`
- Typecheck frontend: `cd frontend && npm run check`

## Issues conocidos
- `ProcessMatchAtomic` usa `+=` para scores — editar un resultado existente duplica puntos (requiere full recalc)
- `DeleteTournament` no elimina subcolecciones de Firestore (players, rounds, matches)
- No hay tests en el backend