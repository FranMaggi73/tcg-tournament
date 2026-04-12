<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { subscribeToTournament, subscribeToMatches } from '$lib/services/tournament';
	import { authStore } from '$lib/stores/auth.svelte';
	import { getTournamentRole } from '$lib/services/roles';
	import type { Tournament, Match } from '$lib/types/firebase';

	let { data } = $props<{ data: { tournamentId: string } }>();
	let tournament = $state<Tournament | null>(null);
	let matches = $state<Match[]>([]);
	let role = $state<'judge' | 'viewer'>('viewer');
	let unsubscribeTournament: () => void;
	let unsubscribeMatches: () => void;

	onMount(() => {
		unsubscribeTournament = subscribeToTournament(data.tournamentId, (updatedT) => {
			tournament = updatedT;
			role = getTournamentRole(authStore.user?.uid ?? null, updatedT);

			if (unsubscribeMatches) unsubscribeMatches();
			unsubscribeMatches = subscribeToMatches(data.tournamentId, updatedT.currentRound, (updatedM) => {
				matches = updatedM;
			});
		});
	});

	onDestroy(() => {
		if (unsubscribeTournament) unsubscribeTournament();
		if (unsubscribeMatches) unsubscribeMatches();
	});
</script>

<div class="p-8 max-w-5xl mx-auto">
	<header class="flex justify-between items-center mb-12">
		<div>
			<h1 class="text-4xl font-bold text-primary mb-2">{tournament?.name || 'Cargando Torneo...'}</h1>
			<div class="flex items-center gap-3">
				<span class="badge badge-secondary py-3 px-4">Ronda {tournament?.currentRound || 1}</span>
				<span class="text-base-content/60">• {tournament?.participants.length || 0} Participantes</span>
			</div>
		</div>

		{#if role === 'judge'}
			<a href="/tournaments/{data.tournamentId}/manage" class="btn btn-primary btn-sm">Administrar Torneo</a>
		{/if}
	</header>

	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
		{#each matches as match}
			<div class="card bg-base-200 shadow-md border border-base-300 overflow-hidden">
				<div class="p-4 flex flex-col gap-4">
					<div class="flex justify-between items-center opacity-50 text-xs font-mono">
						<span>Match {match.id.substring(0, 5)}</span>
						{#if match.status === 'completed'}
							<span class="text-success font-bold uppercase">Finalizado</span>
						{:else}
							<span class="text-warning font-bold uppercase">En Curso</span>
						{/if}
					</div>

					<div class="flex flex-col gap-3">
						<div class="flex items-center justify-between p-3 rounded-lg {match.winner === match.player1 ? 'bg-success/20 ring-1 ring-success' : 'bg-base-300'}">
							<span class="font-bold">{match.player1}</span>
							{#if match.winner === match.player1}
								<span class="badge badge-success text-xs">Ganador</span>
							{/if}
						</div>

						<div class="text-center text-xs font-bold opacity-30">VS</div>

						<div class="flex items-center justify-between p-3 rounded-lg {match.winner === match.player2 ? 'bg-success/20 ring-1 ring-success' : 'bg-base-300'}">
							<span class="font-bold">{match.player2}</span>
							{#if match.winner === match.player2}
								<span class="badge badge-success text-xs">Ganador</span>
							{/if}
						</div>
					</div>
				</div>
			</div>
		{/each}

		{#if matches.length === 0}
			<div class="col-span-full flex flex-col items-center justify-center py-20 text-center">
				<div class="text-6xl mb-4">⏳</div>
				<h2 class="text-2xl font-bold">Esperando Pairings</h2>
				<p class="text-base-content/60">El juez está preparando los enfrentamientos de esta ronda.</p>
			</div>
		{/if}
	</div>
</div>
