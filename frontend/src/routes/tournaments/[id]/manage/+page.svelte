<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { subscribeToTournament, updateTournament, subscribeToMatches, findRound, subscribeToPlayers } from '$lib/services/tournament';
	import { advanceTournamentApi, tournamentApi } from '$lib/services/api';
	import { authStore } from '$lib/stores/auth.svelte';
	import type { Tournament, Match, Player } from '$lib/types/firebase';
	import ParticipantManager from '$lib/components/tournaments/ParticipantManager.svelte';
	import MatchResultEditor from '$lib/components/tournaments/MatchResultEditor.svelte';
	import InviteFriendsModal from '$lib/components/tournaments/InviteFriendsModal.svelte';

	let { data } = $props<{ data: { tournamentId: string } }>();
	let tournament = $state<Tournament | null>(null);
	let matches = $state<Match[]>([]);
	let players = $state<Player[]>([]);
	let currentRoundDocId = $state<string>('');
	let unsubscribeTournament: () => void;
	let unsubscribeMatches: () => void;
	let unsubscribePlayers: () => void;
	let activeTab = $state('settings');
	let isInviteModalOpen = $state(false);
	let isAdvancing = $state(false);
	let isDeleting = $state(false);
	let isFinalizing = $state(false);

	onMount(() => {
		unsubscribeTournament = subscribeToTournament(data.tournamentId, async (updatedT) => {
			tournament = updatedT;

			// Resolve the round document ID for the current round
			if (updatedT.currentRound > 0) {
				const round = await findRound(data.tournamentId, updatedT.currentRound);
				if (round) {
					// Re-subscribe to matches if the round changed
					if (unsubscribeMatches) unsubscribeMatches();
					currentRoundDocId = round.id;
					unsubscribeMatches = subscribeToMatches(data.tournamentId, round.id, (updatedM) => {
						matches = updatedM;
					});
				}
			} else {
				matches = [];
				currentRoundDocId = '';
			}
		});

		unsubscribePlayers = subscribeToPlayers(data.tournamentId, (updatedP) => {
			players = updatedP;
		});
	});

	onDestroy(() => {
		if (unsubscribeTournament) unsubscribeTournament();
		if (unsubscribeMatches) unsubscribeMatches();
		if (unsubscribePlayers) unsubscribePlayers();
	});

	async function handleUpdateName(newName: string) {
		if (!tournament) return;
		await updateTournament(tournament.id, { name: newName });
	}

	async function handleAdvanceRound() {
		if (!tournament) return;
		isAdvancing = true;
		try {
			await advanceTournamentApi(tournament.id);
		} catch (e: any) {
			alert(`Error al avanzar ronda: ${e.message}`);
		} finally {
			isAdvancing = false;
		}
	}

	async function handleDeleteTournament() {
		if (!tournament) return;
		if (!confirm(`¿Estás seguro de que quieres eliminar el torneo "${tournament.name}"? Esta acción no se puede deshacer.`)) return;
		isDeleting = true;
		try {
			await tournamentApi.deleteTournament(tournament.id);
			goto('/tournaments/manage');
		} catch (e: any) {
			alert(`Error al eliminar torneo: ${e.message}`);
		} finally {
			isDeleting = false;
		}
	}

	async function handleFinalizeTournament() {
		if (!tournament) return;
		if (!confirm('¿Estás seguro de que quieres finalizar el torneo? No se podrán hacer más cambios.')) return;
		isFinalizing = true;
		try {
			await tournamentApi.completeTournament(tournament.id);
		} catch (e: any) {
			alert(`Error al finalizar torneo: ${e.message}`);
		} finally {
			isFinalizing = false;
		}
	}

	// Count active players
	let activePlayerCount = $derived(players.filter(p => p.status === 'active').length);
</script>

<div class="p-8 max-w-6xl mx-auto">
	<header class="flex justify-between items-center mb-8">
		<div class="flex flex-col gap-1">
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
						{#if tournament.status === 'registration'}
							<div class="alert alert-info shadow-sm">
								<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
								<div>
									<span>El torneo está en fase de registro.</span>
								</div>
							</div>
						{:else if tournament.status === 'playing'}
							<div class="alert alert-warning shadow-sm">
								<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.88m-13.88 0a4.9999999999999996-4.9999999999999996-4.9999999999999996 4.9999999999999996-4.9999999999999996"></path></svg>
								<div>
									<span>El torneo está en curso. El registro está cerrado y la configuración bloqueada.</span>
								</div>
							</div>
						{:else if tournament.status === 'completed'}
							<div class="alert alert-success shadow-sm">
								<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
								<div>
									<span>El torneo ha finalizado.</span>
								</div>
							</div>
						{/if}

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
									disabled={tournament.status !== 'registration'}
								/>
								<button
									class="btn btn-primary"
									onclick={() => { const t = tournament; if (t) handleUpdateName(t.name); }}
									disabled={tournament.status !== 'registration'}
								>
									Guardar
								</button>
							</div>
						</div>

						<div class="form-control w-full max-w-md">
							<label class="label" for="tournament-format">
								<span class="label-text">Formato</span>
							</label>
							<div class="flex gap-2">
								<select
									id="tournament-format"
									bind:value={tournament.format}
									class="select select-bordered w-full"
									disabled={tournament.status !== 'registration'}
								>
									<option value="BO1">Best of 1</option>
									<option value="BO3">Best of 3</option>
								</select>
								<button
									class="btn btn-primary"
									onclick={() => { const t = tournament; if (t) updateTournament(t.id, { format: t.format }); }}
									disabled={tournament.status !== 'registration'}
								>
									Guardar
								</button>
							</div>
						</div>

						{#if tournament.status === 'registration'}
							<div class="card bg-base-300 p-4 rounded-box border border-base-300">
								<div class="flex flex-col sm:flex-row items-center justify-between gap-4">
									<div class="text-center sm:text-left">
										<p class="text-sm opacity-70">Código de Invitación:</p>
										<p class="text-2xl font-mono font-bold text-primary">{tournament.inviteCode || 'Cargando...'}</p>
									</div>
									<div class="flex gap-2">
										<button class="btn btn-sm btn-outline" onclick={() => { const code = tournament?.inviteCode; if(code) { navigator.clipboard.writeText(code); alert('Código copiado!'); } }}>
											Copiar Código
										</button>
										<button class="btn btn-sm btn-primary" onclick={() => isInviteModalOpen = true}>
											Enviar a Amigos
										</button>
									</div>
								</div>
							</div>
						{/if}

						{#if tournament.status === 'playing'}
							<div class="card bg-warning/10 p-4 rounded-box border border-warning/30">
								<div class="flex flex-col sm:flex-row items-center justify-between gap-4">
									<div class="text-center sm:text-left">
										<p class="font-medium">Torneo en curso</p>
										<p class="text-sm opacity-70">Puedes finalizar el torneo en cualquier momento.</p>
									</div>
									<button
										class="btn btn-sm btn-warning"
										onclick={handleFinalizeTournament}
										disabled={isFinalizing}
									>
										{#if isFinalizing}
											<span class="loading loading-spinner"></span>
										{/if}
										Finalizar Torneo
									</button>
								</div>
							</div>
						{/if}

						{#if tournament.status === 'completed'}
							<div class="card bg-success/10 p-4 rounded-box border border-success/30">
								<div class="flex flex-col sm:flex-row items-center justify-between gap-4">
									<div class="text-center sm:text-left">
										<p class="font-medium text-success">Torneo finalizado</p>
										<p class="text-sm opacity-70">El torneo ha concluido.</p>
									</div>
								</div>
							</div>
						{/if}

						{#if tournament.status !== 'playing'}
							<div class="divider"></div>
							<div class="flex justify-end">
								<button
									class="btn btn-sm btn-error btn-outline"
									onclick={handleDeleteTournament}
									disabled={isDeleting}
								>
									{#if isDeleting}
										<span class="loading loading-spinner"></span>
									{/if}
									Eliminar Torneo
								</button>
							</div>
						{/if}
					</div>
				{:else if activeTab === 'participants'}
					<ParticipantManager {tournament} {players} />
				{:else if activeTab === 'matches'}
					<div class="space-y-6">
						{#if currentRoundDocId}
							<MatchResultEditor
								tournamentId={tournament.id}
								roundDocId={currentRoundDocId}
								matches={matches}
								format={tournament.format}
							/>
						{:else}
							<div class="text-center py-10 text-base-content/50">
								No hay rondas generadas aún. Avanza a la siguiente ronda para generar pairings.
							</div>
						{/if}

						<div class="flex justify-center pt-6 border-t border-base-300">
							<button
								class="btn btn-primary btn-wide"
								onclick={handleAdvanceRound}
								disabled={isAdvancing || tournament.status === 'completed'}
							>
								{#if isAdvancing}
									<span class="loading loading-spinner"></span>
								{/if}
								{#if tournament.currentRound === 0}
									Iniciar Torneo (Ronda 1)
								{:else if matches.length > 0 && matches.some(m => m.status === 'scheduled')}
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

	{#if isInviteModalOpen && tournament}
		<InviteFriendsModal
			{tournament}
			onClose={() => isInviteModalOpen = false}
		/>
	{/if}
</div>