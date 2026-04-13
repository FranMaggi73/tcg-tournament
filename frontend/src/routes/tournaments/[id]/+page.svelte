<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { subscribeToTournament, subscribeToMatches, findRound, subscribeToPlayers } from '$lib/services/tournament';
	import { resolveUserProfiles } from '$lib/services/user';
	import { authStore } from '$lib/stores/auth.svelte';
	import { getTournamentRole } from '$lib/services/roles';
	import type { Tournament, Match, Player } from '$lib/types/firebase';

	let { data } = $props<{ data: { tournamentId: string } }>();
	let tournament = $state<Tournament | null>(null);
	let matches = $state<Match[]>([]);
	let players = $state<Player[]>([]);
	let profiles = $state<Record<string, any>>({});
	let role = $state<'judge' | 'viewer'>('viewer');
	let unsubscribeTournament: () => void;
	let unsubscribeMatches: () => void;
	let unsubscribePlayers: () => void;

	onMount(() => {
		unsubscribeTournament = subscribeToTournament(data.tournamentId, async (updatedT) => {
			tournament = updatedT;
			role = getTournamentRole(authStore.user?.uid ?? null, updatedT);

			// Resolve the round document ID
			if (updatedT.currentRound > 0) {
				const round = await findRound(data.tournamentId, updatedT.currentRound);
				if (round) {
					if (unsubscribeMatches) unsubscribeMatches();
					unsubscribeMatches = subscribeToMatches(data.tournamentId, round.id, (updatedM) => {
						matches = updatedM;
						// Resolve profiles for match players
						const uids = updatedM.flatMap((m: Match) => {
							const ids = [m.player1Id];
							if (m.player2Id !== 'BYE') ids.push(m.player2Id);
							return ids;
						});
						resolveUserProfiles(uids).then(resolved => {
							profiles = resolved;
						});
					});
				}
			} else {
				matches = [];
			}
		});

		unsubscribePlayers = subscribeToPlayers(data.tournamentId, (updatedP) => {
			players = updatedP;
			const uids = updatedP.map(p => p.id);
			resolveUserProfiles(uids).then(resolved => {
				profiles = { ...profiles, ...resolved };
			});
		});
	});

	onDestroy(() => {
		if (unsubscribeTournament) unsubscribeTournament();
		if (unsubscribeMatches) unsubscribeMatches();
		if (unsubscribePlayers) unsubscribePlayers();
	});

	let activePlayerCount = $derived(players.filter(p => p.status === 'active').length);
</script>

<div class="p-8 max-w-5xl mx-auto">
	<header class="flex justify-between items-center mb-12">
		<div>
			<h1 class="text-4xl font-bold text-primary mb-2">{tournament?.name || 'Cargando Torneo...'}</h1>
			<div class="flex items-center gap-3">
				<span class="badge badge-secondary py-3 px-4">Ronda {tournament?.currentRound || 0}</span>
				<span class="text-base-content/60">• {activePlayerCount} Participantes</span>
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

					{#if match.player2Id === 'BYE'}
						<div class="text-center py-2">
							<span class="font-bold">{profiles[match.player1Id]?.displayName || match.player1Id}</span>
							<span class="badge badge-info ml-2">BYE</span>
						</div>
					{:else}
						<div class="flex flex-col gap-3">
							<div class="flex items-center justify-between p-3 rounded-lg {match.winnerId === match.player1Id ? 'bg-success/20 ring-1 ring-success' : 'bg-base-300'}">
								<span class="font-bold">{profiles[match.player1Id]?.displayName || match.player1Id}</span>
								{#if match.status === 'completed'}
									<span class="text-sm opacity-60">{match.player1Score}</span>
								{/if}
								{#if match.winnerId === match.player1Id}
									<span class="badge badge-success text-xs">Ganador</span>
								{/if}
							</div>

							<div class="text-center text-xs font-bold opacity-30">VS</div>

							<div class="flex items-center justify-between p-3 rounded-lg {match.winnerId === match.player2Id ? 'bg-success/20 ring-1 ring-success' : 'bg-base-300'}">
								<span class="font-bold">{profiles[match.player2Id]?.displayName || match.player2Id}</span>
								{#if match.status === 'completed'}
									<span class="text-sm opacity-60">{match.player2Score}</span>
								{/if}
								{#if match.winnerId === match.player2Id}
									<span class="badge badge-success text-xs">Ganador</span>
								{/if}
							</div>
						</div>
					{/if}
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