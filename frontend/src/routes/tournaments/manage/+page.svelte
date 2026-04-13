<script lang="ts">
	import { onMount } from 'svelte';
	import { getTournamentsByJudge } from '$lib/services/tournament';
	import { authStore, waitForAuth } from '$lib/stores/auth.svelte';
	import CreateTournamentModal from '$lib/components/tournaments/CreateTournamentModal.svelte';
	import type { Tournament } from '$lib/types/firebase';

	let tournaments = $state<Tournament[]>([]);
	let isModalOpen = $state(false);
	let isLoading = $state(true);

	async function loadTournaments() {
		isLoading = true;
		try {
			await waitForAuth();
			if (!authStore.user) return;
			const data = await getTournamentsByJudge(authStore.user.uid);
			tournaments = data ?? [];
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
			<h1 class="text-3xl font-bold text-primary">Mis Torneos</h1>
			<p class="text-base-content/60">Gestiona los eventos que has creado</p>
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
	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
		{#each tournaments as t}
			<div class="card bg-base-200 shadow-md border border-base-300 hover:border-primary transition-colors group">
				<div class="card-body">
					<div class="flex justify-between items-start">
						<h2 class="card-title text-xl group-hover:text-primary transition-colors">{t.name}</h2>
						<div class="badge badge-secondary">{t.status}</div>
					</div>
					<div class="flex items-center gap-4 mt-2 text-sm text-base-content/60">
						<span>Formato: {t.format}</span>
						<span>Ronda: {t.currentRound}</span>
					</div>
					{#if t.inviteCode && t.status === 'registration'}
						<div class="mt-2 text-xs opacity-60">Código: <span class="font-mono font-bold text-primary">{t.inviteCode}</span></div>
					{/if}
					<div class="card-actions justify-end mt-4">
						<a href="/tournaments/{t.id}/manage" class="btn btn-outline btn-sm">Administrar</a>
					</div>
				</div>
			</div>
		{/each}
		{#if tournaments.length === 0}
			<div class="flex flex-col items-center justify-center py-20 text-center col-span-full">
				<div class="text-5xl mb-4">🏆</div>
				<p class="text-lg text-base-content/70">Aún no has creado ningún torneo.</p>
				<p class="text-sm text-base-content/50">Empieza creando uno ahora mismo.</p>
			</div>
		{/if}
	</div>
	{/if}

	{#if isModalOpen}
		<CreateTournamentModal
			onClose={() => {
				isModalOpen = false;
			}}
		/>
	{/if}
</div>