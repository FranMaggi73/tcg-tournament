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
- lint: (Eliminado para evitar conflictos de formato)

## Convenciones
- Componentes en `src/lib/components/`
- Stores en `src/lib/stores/`
- Tipos en `src/lib/types/`
- Servicios Firebase en `src/lib/services/`
- Servicios API (Go Backend) en `src/lib/services/api.ts`
- Rutas en `src/routes/`

## Roles & Seguridad
- `judge`: Crea torneos, gestiona participantes (incluyendo Drops), registra resultados BO3 y avanza rondas.
- `viewer`: Consulta el torneo, pairings y standings en tiempo real (read-only).
- El rol se deriva del campo `createdBy` en el documento del torneo en Firestore.
- **Acceso**: Las rutas de gestión `/manage` están protegidas mediante validación de UID en el servidor/cliente.

## Lógica de Torneo (Sincronización Backend Go)
- **Real-time**: Se utiliza Firestore `onSnapshot` para reflejar cambios instantáneos en pairings y standings.
- **Mutaciones**: Toda acción que afecte la lógica del torneo (resultados de partidas, avance de ronda, drops) DEBE pasar por la API de Go para asegurar la atomicidad y el cálculo correcto de tiebreakers.
- **BO3**: Los resultados se registran como puntuaciones numéricas (0, 1, 2), donde un jugador debe alcanzar 2 puntos para ganar.
- **Sistemas Suizo**: El frontend muestra el ranking basado en Puntos $\rightarrow$ OMW% $\rightarrow$ GW% $\rightarrow$ OGW%.

## Tiempo real
- NO usar WebSockets.
- Usar Firestore `onSnapshot` para actualizaciones en vivo.
- Desuscribirse siempre en `onDestroy` para evitar memory leaks.

## Notas
- El sistema suizo calcula pairings y standings en el backend (Go).
- El frontend resuelve UIDs a perfiles de usuario mediante una caché reactiva en `src/lib/stores/users.svelte.ts`.
- Los pairings y standings son la fuente de verdad proveniente del backend.
