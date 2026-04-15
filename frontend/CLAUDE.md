# TCG Tournament — Frontend

## Stack
- SvelteKit 2 + TypeScript
- Svelte 5 Runes (`$state`, `$derived`, `$derived.by`, `$effect`, `$props`, `$bindable`)
- Tailwind CSS 4 + DaisyUI 5 (tema oscuro custom TCG)
- Heroicons
- Firebase SDK 12 (Firestore real-time + Auth)
- Vite 8 + Vitest 4

## Comandos clave
- dev: `npm run dev`
- build: `npm run build`
- typecheck: `npm run check`
- test: `npm run test`

## Estructura
```
src/
├── lib/
│   ├── components/
│   │   ├── Sidebar.svelte                  # Navegación global colapsable + badge de invitaciones
│   │   ├── auth/
│   │   │   └── LoginForm.svelte            # Login con Google (signInWithPopup)
│   │   └── tournaments/
│   │       ├── CreateTournamentModal.svelte # Modal para crear torneo
│   │       ├── InviteFriendsModal.svelte    # Modal para enviar invites a amigos
│   │       ├── MatchResultEditor.svelte     # Editor de resultados de matches
│   │       ├── ParticipantManager.svelte    # Gestión de jugadores (drop/restore/remove/join)
│   │       └── StandingsTable.svelte        # Tabla de standings con export CSV
│   ├── services/
│   │   ├── api.ts            # REST API client (Go backend) — tournamentApi + friendshipApi + advanceTournamentApi
│   │   ├── auth-utils.ts     # logout helper
│   │   ├── export.ts         # CSV export (generateStandingsCSV, downloadCSV)
│   │   ├── firebase.ts       # Firebase init (auth, db)
│   │   ├── notifications.ts  # Firestore notifications (sendInvite, getNotifications, markAsRead...)
│   │   ├── roles.ts          # judge/viewer role derivation (getTournamentRole)
│   │   ├── tournament.ts     # Firestore tournament CRUD + subscriptions
│   │   └── user.ts           # User profile resolution + cache (ensureUserProfile, resolveUserProfiles...)
│   ├── stores/
│   │   ├── auth.svelte.ts         # Auth state reactivo + initAuthObserver + waitForAuth
│   │   ├── notifications.svelte.ts # Badge counter de invitaciones (invitationCountStore, refreshInvitationCount)
│   │   └── users.svelte.ts        # User profile cache reactivo (userCache, setCachedProfile, getCachedProfile)
│   └── types/
│       └── firebase.ts       # TypeScript interfaces
└── routes/
    ├── +layout.svelte        # Auth guard + Sidebar condicional
    ├── +layout.ts            # ssr = false (app CSR-only)
    ├── +page.svelte          # Landing: login o "unirse con código"
    ├── profile/+page.svelte  # Perfil, amigos, solicitudes pendientes, notificaciones
    └── tournaments/
        ├── manage/+page.svelte          # Lista de torneos del judge + torneos donde participa
        └── [id]/
            ├── +page.svelte            # Vista pública: pairings en vivo
            ├── +page.ts                # Load function (sin guard — cualquiera puede ver)
            └── manage/
                ├── +page.svelte        # Panel del judge: settings, participants, matches
                └── +page.ts           # Load + guard: redirige si no es judge
```

## Tipos clave (`src/lib/types/firebase.ts`)
- **Tournament**: `id`, `name`, `createdBy`, `status`, `currentRound`, `totalRounds`, `format`, `inviteCode`
- **Player**: `id`, `name`, `email`, `totalScore`, `wins`, `losses`, `draws`, `omw`, `gw`, `ogw`, `status`, `hadBye`
- **Match**: `id`, `roundId`, `player1Id`, `player2Id`, `player1Score`, `player2Score`, `winnerId`, `status`
- **Round**: `id`, `tournamentId`, `roundNumber`, `status`, `createdAt`
- **UserProfile**: `uid`, `displayName`, `photoURL`, `bio`, `updatedAt`
- **Friendship**: `id`, `user1Id`, `user2Id`, `status`, `createdAt`
- **Notification**: `id`, `type`, `recipientId`, `senderId`, `tournamentId`, `inviteCode`, `tournamentName`, `message`, `read`, `createdAt`, `expiresAt`

## Roles & Seguridad
- `judge`: Crea torneos, gestiona participantes (remove/drop/restore), registra resultados, avanza rondas, rollback, finaliza, elimina torneos.
- `viewer`: Consulta el torneo, pairings y standings en tiempo real (read-only).
- El rol se deriva del campo `createdBy` del torneo (`getTournamentRole(uid, tournament)`).
- **Acceso**: La ruta `/tournaments/[id]/manage` está protegida por `+page.ts` que valida `UID == createdBy`. Redirige a `/tournaments/[id]` si no es judge.
- **Auth guard global**: `+layout.svelte` redirige a `/` si el usuario no está autenticado y la ruta no es `/`.

## Flujo de Datos
- **Lectura**: Firestore `onSnapshot` directo (torneo, players, matches, rounds)
- **Mutaciones**: API Go (`api.ts`) — crear torneo, unirse, resultados, drops, avance de ronda, rollback, delete, complete
- **Players**: Se leen de la subcolección `/tournaments/{id}/players` via `subscribeToPlayers()`
- **Matches**: Se leen de `/tournaments/{id}/rounds/{roundDocId}/matches` via `subscribeToMatches()` — el `roundDocId` se obtiene con `findRound(tournamentId, roundNumber)` que devuelve el documento de Firestore
- **Notificaciones**: Se leen directamente de Firestore en `/profile` y para el badge en `Sidebar`

## Convención de nombres frontend ↔ backend
| Frontend (TS) | Backend (Go) | Nota |
|---------------|---------------|------|
| `player1Id` | `Player1ID` | Match |
| `player2Id` | `Player2ID` | Match, "BYE" para byes |
| `winnerId` | `WinnerID` | Match, `""` para empate |
| `roundId` | `RoundID` | Match — se envía en el body al llamar UpdateMatchResult |
| `player1Score` | `Player1Score` | Match |
| `player2Score` | `Player2Score` | Match |
| `totalScore` | `TotalScore` | Player |
| `omw` / `gw` / `ogw` | `OMW` / `GW` / `OGW` | Player |
| `hadBye` | `HadBye` | Player |
| `createdBy` | `CreatedBy` | Tournament |
| `inviteCode` | `InviteCode` | Tournament |
| `currentRound` | `CurrentRound` | Tournament |

## Tiempo real
- NO usar WebSockets.
- Usar Firestore `onSnapshot` para actualizaciones en vivo.
- Desuscribirse siempre en `onDestroy` para evitar memory leaks.
- `subscribeToTournament` → documento del torneo
- `subscribeToPlayers` → subcolección de jugadores
- `subscribeToMatches(tournamentId, roundDocId, cb)` → subcolección de matches (requiere `roundDocId` del documento Firestore, NO el número de ronda)
- `findRound(tournamentId, roundNumber)` → resuelve el `roundDocId` a partir del número de ronda

## API Client (`api.ts`)

### `tournamentApi`
| Método | Descripción |
|--------|-------------|
| `createTournament(name, format)` | POST `/tournaments` |
| `submitMatchResult(tId, matchId, roundId, s1, s2)` | PATCH `/tournaments/:id/matches/:matchId` — roundId va en el body |
| `joinByCode(code, email, name)` | POST `/tournaments/join` |
| `getStandings(tId)` | GET `/tournaments/:id/standings` |
| `dropParticipant(tId, playerId)` | PATCH `.../players/:id/status` con `{status: "dropped"}` |
| `restoreParticipant(tId, playerId)` | PATCH `.../players/:id/status` con `{status: "active"}` |
| `removeParticipant(tId, playerId)` | DELETE `.../players/:id` — solo en `registration` |
| `deleteTournament(tId)` | DELETE `/tournaments/:id` |
| `completeTournament(tId)` | PATCH `/tournaments/:id/complete` |
| `rollbackRound(tId)` | POST `/tournaments/:id/rollback` |

### `friendshipApi`
| Método | Descripción |
|--------|-------------|
| `addFriend(friendId)` | POST `/friends` |
| `getFriends()` | GET `/friends` — solo aceptados |
| `getPendingRequests()` | GET `/friends/pending` |
| `updateStatus(friendshipId, status)` | PATCH `/friends/:id` |

### `advanceTournamentApi(tournamentId)`
- POST `/tournaments/:id/rounds/next`
- Función standalone (no en `tournamentApi`)

## Sidebar (`Sidebar.svelte`)
- Colapsable (estado `collapsed` como `$bindable`)
- Muestra badge numérico en el ítem "Perfil" con el total de invitaciones no leídas + solicitudes de amistad pendientes
- El contador se actualiza en `onMount` via `refreshInvitationCount()` del store `notifications.svelte.ts`
- Resuelve el perfil del usuario actual desde `userCache`

## Sistema de Notificaciones
- Las notificaciones se escriben **directamente en Firestore** desde el frontend (no pasan por el backend)
- `notificationService.sendInvite(...)` verifica duplicados antes de crear
- Expiración automática a los 7 días (`expiresAt`)
- Al aceptar/rechazar, se marcan como leídas y se eliminan de Firestore con un delay de 500ms
- El store `invitationCountStore` en `notifications.svelte.ts` suma invites no leídas + pending friend requests para el badge

## BO3 — Empates
- El frontend permite seleccionar `1-1` como resultado válido en BO3
- `isValidScore()` en `MatchResultEditor` acepta: victorias (2-0, 2-1, 0-2, 1-2) y empate (1-1)
- El backend deriva `winnerId = ""` cuando los scores son iguales (empate)

## Notas de implementación importantes
- La app es **CSR-only** (`ssr = false` en `+layout.ts`) — todo depende de Firebase Auth cliente
- `waitForAuth()` en los `load()` functions garantiza que el estado de auth esté listo antes de verificar roles
- `ensureUserProfile()` se llama automáticamente en el auth observer al primer login, creando el documento en `/users/{uid}` sin sobreescribir datos existentes
- `resolveUserProfiles(uids)` usa caché en memoria (`userCache`) para evitar N+1 queries a Firestore
- El perfil de usuario **no guarda** el `photoURL` de Google por defecto — el avatar se configura manualmente