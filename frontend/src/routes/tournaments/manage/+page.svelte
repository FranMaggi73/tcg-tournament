<script lang="ts">
	import { onMount } from 'svelte';
	import { getTournamentsByJudge, getTournamentsByPlayer } from '$lib/services/tournament';
	import { authStore, waitForAuth } from '$lib/stores/auth.svelte';
	import CreateTournamentModal from '$lib/components/tournaments/CreateTournamentModal.svelte';
	import type { Tournament } from '$lib/types/firebase';

	let myTournaments = $state<Tournament[]>([]);
	let joinedTournaments = $state<Tournament[]>([]);
	let isModalOpen = $state(false);
	let isLoading = $state(true);

	async function loadTournaments() {
		isLoading = true;
		try {
			await waitForAuth();
			if (!authStore.user) return;

			const uid = authStore.user.uid;

			// Load independently so one failure doesn't block the other
			getTournamentsByJudge(uid)
				.then(result => { myTournaments = result ?? []; })
				.catch(e => { console.error('Error loading my tournaments:', e); });

			getTournamentsByPlayer(authStore.user.email ?? '', uid)
				.then(result => { joinedTournaments = result ?? []; })
				.catch(e => { console.error('Error loading joined tournaments:', e); });
		} catch (e) {
			console.error('Error loading tournaments:', e);
		} finally {
			isLoading = false;
		}
	}

	onMount(loadTournaments);
</script>

<div class="p-8 max-w-5xl mx-auto">
	<header class="flex justify-between items-center mb-8">
		<div>
			<h1 class="text-3xl font-bold text-primary">Torneos</h1>
			<p class="text-base-content/60">Gestiona los eventos que has creado y los que participas</p>
		</div>

		<button class="btn btn-primary" onclick={() => isModalOpen = true}>
			<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" viewBox="0 0 20 20" fill="currentColor">
				<path fill-rule="evenodd" d="M10 5a1 1 0 011 1v3h3a1 1 0 010 2h-3v4a1 1 0 01-1 1V12H6a1 1 0 010-2h3V6a1 1 0 011-1z" clip-rule="evenodd" />
			</svg>
			Nuevo Torneo
		</button>
	</header>

	{#if isLoading}
		<div class="flex justify-center py-20">
			<span class="loading loading-ring loading-lg text-primary"></span>
		</div>
	{:else}
		<!-- Mis Torneos (creados por mi) -->
		<section class="mb-10">
			<h2 class="text-xl font-bold text-primary mb-4">Mis Torneos</h2>
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
				{#each myTournaments as t}
					<div class="card bg-base-200 shadow-md border border-base-300 hover:border-primary transition-colors group">
						<div class="card-body">
							<div class="flex justify-between items-start">
								<h3 class="card-title text-xl group-hover:text-primary transition-colors">{t.name}</h3>
								<div class="badge badge-secondary">{t.status}</div>
							</div>
							<div class="flex items-center gap-4 mt-2 text-sm text-base-content/60">
								<span>Formato: {t.format}</span>
								<span>Ronda: {t.currentRound}</span>
							</div>
							{#if t.inviteCode && t.status === 'registration'}
								<div class="mt-2 text-xs opacity-60">Codigo: <span class="font-mono font-bold text-primary">{t.inviteCode}</span></div>
							{/if}
							<div class="card-actions justify-end mt-4">
								<a href="/tournaments/{t.id}/manage" class="btn btn-outline btn-sm">Administrar</a>
							</div>
						</div>
					</div>
				{/each}
				{#if myTournaments.length === 0}
					<div class="flex flex-col items-center justify-center py-16 text-center col-span-full">
						<p class="text-base-content/70">No has creado ningun torneo todavia.</p>
						<p class="text-sm text-base-content/50">Crea uno con el boton de arriba.</p>
					</div>
				{/if}
			</div>
		</section>

		<!-- Torneos Unidos (donde participo) -->
		<section>
			<h2 class="text-xl font-bold text-primary mb-4">Torneos que Participas</h2>
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
				{#each joinedTournaments as t}
					<div class="card bg-base-200 shadow-md border border-base-300 hover:border-accent transition-colors group">
						<div class="card-body">
							<div class="flex justify-between items-start">
								<h3 class="card-title text-xl group-hover:text-accent transition-colors">{t.name}</h3>
								<div class="badge badge-accent">{t.status}</div>
							</div>
							<div class="flex items-center gap-4 mt-2 text-sm text-base-content/60">
								<span>Formato: {t.format}</span>
								<span>Ronda: {t.currentRound}</span>
							</div>
							<div class="card-actions justify-end mt-4">
								<a href="/tournaments/{t.id}" class="btn btn-outline btn-sm">Ver Torneo</a>
							</div>
						</div>
					</div>
				{/each}
				{#if joinedTournaments.length === 0}
					<div class="flex flex-col items-center justify-center py-16 text-center col-span-full">
						<p class="text-base-content/70">No te has unido a ningun torneo todavia.</p>
						<p class="text-sm text-base-content/50">Usa un codigo de invitacion desde Inicio.</p>
					</div>
				{/if}
			</div>
		</section>
	{/if}

	{#if isModalOpen}
		<CreateTournamentModal
			onClose={() => {
				isModalOpen = false;
				loadTournaments();
			}}
		/>
	{/if}
</div>