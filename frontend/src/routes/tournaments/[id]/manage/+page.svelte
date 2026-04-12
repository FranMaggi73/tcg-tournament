<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { subscribeToTournament, updateTournament, subscribeToMatches } from '$lib/services/tournament';
	import { advanceTournamentApi } from '$lib/services/api';
	import type { Tournament, Match } from '$lib/types/firebase';
	import ParticipantManager from '$lib/components/tournaments/ParticipantManager.svelte';
	import MatchResultEditor from '$lib/components/tournaments/MatchResultEditor.svelte';

	let { data } = $props<{ data: { tournamentId: string } }>();
	let tournament = $state<Tournament | null>(null);
	let matches = $state<Match[]>([]);
	let unsubscribeTournament: () => void;
	let unsubscribeMatches: () => void;
	let activeTab = $state('settings');

	onMount(() => {
		unsubscribeTournament = subscribeToTournament(data.tournamentId, (updatedT) => {
			tournament = updatedT;
			// Re-subscribe to matches if the round changed
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

	async function handleUpdateName(newName: string) {
		if (!tournament) return;
		await updateTournament(tournament.id, { name: newName });
	}

	async function handleAdvanceRound() {
		if (!tournament) return;
		try {
			await advanceTournamentApi(tournament.id);
		} catch (e: any) {
			alert(`Error al avanzar ronda: ${e.message}`);
		}
	}
</script>

<div class="p-8 max-w-6xl mx-auto">
	<header class="flex justify-between items-center mb-8">
		<div>
			<h1 class="text-3xl font-bold text-primary">Panel de Gestión</h1>
			<p class="text-base-content/60">Administrando: <span class="text-primary font-medium">{tournament?.name || 'Cargando...'}</span></p>
		</div>

		<a href="/tournaments/manage" class="btn btn-outline btn-sm">Volver al Panel</a>
	</header>

	<div class="flex flex-col md:flex-row gap-8">
		<!-- Navigation Tabs -->
		<div class="flex flex-col gap-2 w-full md:w-48">
			<div class="tabs tabs-vertical bg-base-200 p-2 rounded-box border border-base-300">
				<button
					type="button"
					class="tab {activeTab === 'settings' ? 'tab-active' : ''}"
					onclick={() => activeTab = 'settings'}
				>
					Configuración
				</button>
				<button
					type="button"
					class="tab {activeTab === 'participants' ? 'tab-active' : ''}"
					onclick={() => activeTab = 'participants'}
				>
					Participantes
				</button>
				<button
					type="button"
					class="tab {activeTab === 'matches' ? 'tab-active' : ''}"
					onclick={() => activeTab = 'matches'}
				>
					Partidas
				</button>
			</div>
		</div>

		<!-- Tab Content -->
		<div class="flex-1 bg-base-200 p-6 rounded-box border border-base-300 shadow-inner">
			{#if !tournament}
				<div class="flex justify-center py-20">
					<span class="loading loading-ring loading-lg text-primary"></span>
				</div>
			{:else}
				{#if activeTab === 'settings'}
					<div class="space-y-6">
						<div class="form-control w-full max-w-md">
							<label class="label" for="tournament-name">
								<span class="label-text">Nombre del Torneo</span>
							</label>
							<div class="flex gap-2">
								<input
									id="tournament-name"
									type="text"
									bind:value={tournament.name}
									class="input input-bordered w-full"
								/>
								<button
									class="btn btn-primary"
									onclick={() => { const t = tournament; if (t) handleUpdateName(t.name); }}
								>
									Guardar
								</button>
							</div>
						</div>
					</div>
				{:else if activeTab === 'participants'}
					<ParticipantManager tournament={tournament} />
				{:else if activeTab === 'matches'}
					<div class="space-y-6">
						<MatchResultEditor
							tournamentId={tournament.id}
							round={tournament.currentRound}
							matches={matches}
						/>

						<div class="flex justify-center pt-6 border-t border-base-300">
							<button
								class="btn btn-primary btn-wide"
								onclick={handleAdvanceRound}
								disabled={matches.length === 0 || matches.some(m => m.status === 'pending')}
							>
								{#if matches.length === 0}
									Esperando Pairings...
								{:else if matches.some(m => m.status === 'pending')}
									Finalizar Partidas Pendientes
								{:else}
									Avanzar a Ronda {tournament.currentRound + 1}
								{/if}
							</button>
						</div>
					</div>
				{/if}
			{/if}
		</div>
	</div>
</div>
