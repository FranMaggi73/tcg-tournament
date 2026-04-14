<script lang="ts">
	import { tournamentApi } from '$lib/services/api';
	import type { Tournament, Player } from '$lib/types/firebase';

	let { tournament, players } = $props<{
		tournament: Tournament;
		players: Player[];
	}>();

	let isDropping = $state<Record<string, boolean>>({});

	async function handleDropPlayer(playerId: string) {
		const playerName = players.find((p: Player) => p.id === playerId)?.name || playerId;
		if (!confirm(`Estas seguro de que quieres eliminar a ${playerName} del torneo?`)) {
			return;
		}

		isDropping[playerId] = true;
		try {
			await tournamentApi.dropParticipant(tournament.id, playerId);
		} catch (e: any) {
			alert(`Error: ${e.message}`);
		} finally {
			isDropping[playerId] = false;
		}
	}

	let activePlayers = $derived(players.filter((p: Player) => p.status === 'active'));
	let droppedPlayers = $derived(players.filter((p: Player) => p.status === 'dropped'));
</script>

<div class="space-y-4">
	<div class="flex justify-between items-center">
		<h3 class="text-lg font-bold">Participantes ({activePlayers.length} activos{droppedPlayers.length > 0 ? `, ${droppedPlayers.length} eliminados` : ''})</h3>
	</div>

	<div class="grid grid-cols-1 gap-3">
		{#each players as p}
			<div class="card bg-base-300 p-3 rounded-box border border-base-300 flex justify-between items-center {p.status === 'dropped' ? 'opacity-50' : ''}">
				<div class="flex items-center gap-3">
					<div class="avatar">
						<div class="w-10 rounded-full bg-base-200 ring-1 ring-primary flex items-center justify-center text-xs font-bold text-primary">
							{p.name.charAt(0).toUpperCase()}
						</div>
					</div>
					<div>
						<p class="font-bold">{p.name}</p>
						<p class="text-xs opacity-60">{p.email}</p>
					</div>
				</div>
				<div class="flex items-center gap-2">
					{#if p.status === 'active'}
						<button
							class="btn btn-sm btn-error text-white"
							disabled={!!isDropping[p.id]}
							onclick={() => handleDropPlayer(p.id)}
						>
							{#if isDropping[p.id]}
								<span class="loading loading-spinner loading-xs"></span>
							{/if}
							Eliminar
						</button>
					{:else}
						<div class="badge badge-ghost">Eliminado</div>
					{/if}
				</div>
			</div>
		{/each}
		{#if players.length === 0}
			<div class="text-center py-8 opacity-50 italic text-sm">No hay participantes registrados.</div>
		{/if}
	</div>
</div>