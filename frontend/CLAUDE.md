# TCG Tournament — Frontend

## Stack
- SvelteKit + TypeScript
- Tailwind CSS + DaisyUI (tema custom oscuro TCG)
- Heroicons
- Firebase SDK (Firestore real-time + Auth)
- Svelte 5 Runes ($state, $derived, $effect)

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
│   │   ├── auth/LoginForm.svelte
│   │   └── tournaments/
│   │       ├── CreateTournamentModal.svelte
│   │       ├── InviteFriendsModal.svelte
│   │       ├── MatchResultEditor.svelte
│   │       ├── ParticipantManager.svelte
│   │       └── StandingsTable.svelte
│   ├── services/
│   │   ├── api.ts            # REST API client (Go backend)
│   │   ├── auth-utils.ts     # logout helper
│   │   ├── export.ts         # CSV export
│   │   ├── firebase.ts       # Firebase init
│   │   ├── notifications.ts  # Firestore notifications
│   │   ├── roles.ts          # judge/viewer role derivation
│   │   ├── tournament.ts     # Firestore tournament CRUD + subscriptions
│   │   └── user.ts           # User profile resolution + cache
│   ├── stores/
│   │   ├── auth.svelte.ts    # Auth state (reactive)
│   │   └── users.svelte.ts   # User profile cache (reactive)
│   └── types/
│       └── firebase.ts       # TypeScript interfaces (Tournament, Player, Match, Round, etc.)
└── routes/
    ├── +layout.svelte        # Init auth observer
    ├── +page.svelte          # Landing (login + join by code)
    ├── profile/+page.svelte  # User profile, friends, notifications
    └── tournaments/
        ├── manage/+page.svelte          # Judge's tournament list
        └── [id]/
            ├── +page.svelte            # Viewer: live pairings + standings
            ├── +page.ts                # Load + role guard
            └── manage/+page.svelte     # Judge: settings, participants, matches
```

## Tipos clave (`src/lib/types/firebase.ts`)
- **Tournament**: `id`, `name`, `createdBy`, `status`, `currentRound`, `totalRounds`, `format`, `inviteCode` (SIN `participants`)
- **Player**: `id`, `name`, `email`, `totalScore`, `wins`, `losses`, `draws`, `omw`, `gw`, `ogw`, `status`, `hadBye`
- **Match**: `id`, `roundId`, `player1Id`, `player2Id`, `player1Score`, `player2Score`, `winnerId`, `status`
- **Round**: `id`, `tournamentId`, `roundNumber`, `status`, `createdAt`
- **UserProfile**: `uid`, `displayName`, `photoURL`, `bio`, `updatedAt`
- **Friendship**: `id`, `user1Id`, `user2Id`, `status`, `createdAt`
- **Notification**: `id`, `recipientId`, `senderId`, `tournamentId`, `inviteCode`, `tournamentName`, `message`, `read`, `createdAt`

## Roles & Seguridad
- `judge`: Crea torneos, gestiona participantes (incluyendo Drops), registra resultados, avanza rondas, elimina torneos.
- `viewer`: Consulta el torneo, pairings y standings en tiempo real (read-only).
- El rol se deriva del campo `createdBy` en el documento del torneo en Firestore (`getTournamentRole()`).
- **Acceso**: La ruta `/manage` está protegida por `+page.ts` que valida UID == createdBy.

## Flujo de Datos
- **Lectura**: Firestore `onSnapshot` directo (torneo, players, matches)
- **Mutaciones**: API Go (`api.ts`) — crear torneo, unirse, resultados, drops, avance de ronda, rollback, delete
- **Players**: Se leen de la subcolección `/tournaments/{id}/players` via `subscribeToPlayers()`
- **Matches**: Se leen de `/tournaments/{id}/rounds/{roundDocId}/matches` via `subscribeToMatches()` — el `roundDocId` se obtiene con `findRound()`

## Convención de nombres frontend ↔ backend
| Frontend (TS) | Backend (Go) | Nota |
|---------------|---------------|------|
| `player1Id` | `Player1ID` | Match |
| `player2Id` | `Player2ID` | Match, "BYE" para byes |
| `winnerId` | `WinnerID` | Match, `""` para empate |
| `roundId` | `RoundID` | Match |
| `player1Score` | `Player1Score` | Match |
| `player2Score` | `Player2Score` | Match |
| `totalScore` | `TotalScore` | Player |
| `omw` / `gw` / `ogw` | `OMW` / `GW` / `OGW` | Player |
| `hadBye` | `HadBye` | Player |

## Tiempo real
- NO usar WebSockets.
- Usar Firestore `onSnapshot` para actualizaciones en vivo.
- Desuscribirse siempre en `onDestroy` para evitar memory leaks.
- `subscribeToTournament` → documento del torneo
- `subscribeToPlayers` → subcolección de jugadores
- `subscribeToMatches` → subcolección de matches (requiere `roundDocId`, NO número de ronda)

## API Client (`api.ts`)
- `submitMatchResult(tournamentId, matchId, roundId, score1, score2)` — envía scores + roundId
- `joinByCode(code, email, name)` — POST `/tournaments/join`
- `getStandings(tournamentId)` — GET `/tournaments/:id/standings`
- `dropParticipant(tournamentId, playerId)` — PATCH `/tournaments/:id/players/:playerId/status`
- `deleteTournament(tournamentId)` — DELETE `/tournaments/:id`
- `advanceTournamentApi(tournamentId)` — POST `/tournaments/:id/rounds/next`
- `friendshipApi.addFriend/getFriends/updateStatus`

## BO3 — Empates
- El frontend permite seleccionar `1-1` como resultado válido en BO3
- La validación `isValidScore()` acepta empates: `(s1 === 1 && s2 === 1)`
- El backend deriva `winnerId = ""` cuando los scores son iguales