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

## Roles de usuario
- `judge`: crea torneo, avanza rondas, edita resultados (requiere pasar por el backend)
- `viewer`: ve el torneo en tiempo real (solo Firestore onSnapshot, sin backend)
- El rol se deriva del campo `createdBy` en Firestore

## Contrato entre frontend y backend
- El frontend envía Firebase ID token en el header `Authorization: Bearer <token>`
- El backend verifica el token y extrae el UID
- Si `UID == torneo.createdBy` → es judge, puede operar
- El backend devuelve JSON, el frontend nunca calcula pairings

## Algoritmo suizo (solo backend)
- Rondas = ceil(log2(jugadores))
- Puntaje: victoria=3, empate=1, derrota=0 (BO3); victoria=1 (BO1)
- Tiebreakers: OMW% → GW% → OGW%
- Editar rondas pasadas recalcula standings pero NO re-parea rondas futuras

## Comandos
- Frontend: `cd frontend && npm run dev`
- Backend: `cd backend && go run cmd/main.go`
- Tests backend: `cd backend && go test ./...`