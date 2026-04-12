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
- `cmd/` — entry point
- `internal/tournament/` — lógica suiza, pairings y standings
- `internal/handlers/` — HTTP handlers
- `internal/middleware/` — auth JWT/Firebase
- `internal/models/` — structs compartidos

## Algoritmo Suizo Profesional
- **Rondas**: $\lceil\log_2(\text{jugadores})\rceil$
- **Puntaje (BO3)**: victoria=3, empate=1, derrota=0.
- **Pairings**: 
  - Emparejamiento por rango de puntaje (descendente).
  - **No-Repeat**: Dos jugadores NUNCA se enfrentan más de una vez por torneo.
  - **Byes**: Si el número de jugadores es impar, el jugador con menor puntaje que NO haya tenido un bye lo recibe (cuenta como victoria).
- **Standings & Tiebreakers**:
  1. Puntaje Total (Descendente)
  2. OMW% (Opponent Match Win %): Fuerza de los oponentes.
  3. GW% (Game Win %): Juegos ganados / Juegos jugados.
  4. OGW% (Opponent Game Win %): Promedio de GW% de los oponentes.
- **Drops**: Jugadores con status `dropped` son excluidos de pairings y recálculos de standings.

## Auth & Seguridad
- Verificar Firebase ID token en cada request.
- **RBAC**: El UID del token debe coincidir con el campo `createdBy` del torneo para realizar acciones de Judge (pairings, resultados).
- **Atomicidad**: Actualizaciones de resultados y puntajes se ejecutan mediante Transacciones de Firestore para evitar inconsistencias.
