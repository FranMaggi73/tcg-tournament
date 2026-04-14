<script lang="ts">
	import { authStore } from '$lib/stores/auth.svelte';
	import LoginForm from '$lib/components/auth/LoginForm.svelte';
	import { tournamentApi } from '$lib/services/api';

	// Join Tournament state
	let inviteCode = $state('');
	let playerName = $state('');
	let isJoining = $state(false);
	let joinError = $state('');
	let joinSuccess = $state(false);

	async function handleJoin() {
		if (!inviteCode || !playerName) {
			joinError = 'Todos los campos son obligatorios';
			return;
		}

		isJoining = true;
		joinError = '';
		joinSuccess = false;

		try {
			console.log('[Join] Sending request:', { code: inviteCode, email: authStore.user?.email, name: playerName });
			const result = await tournamentApi.joinByCode(inviteCode, authStore.user?.email || '', playerName);
			console.log('[Join] Success:', result);
			joinSuccess = true;
		} catch (e: any) {
			console.error('[Join] Error:', e);
			joinError = e.message;
		} finally {
			isJoining = false;
		}
	}
</script>

<div class="p-8 flex flex-col items-center justify-center min-h-screen gap-8">
	{#if authStore.isLoading}
		<div class="flex flex-col items-center gap-4">
			<span class="loading loading-ring loading-lg text-primary"></span>
			<p class="text-base-content/60">Cargando sesión...</p>
		</div>
	{:else if !authStore.user}
		<div class="flex flex-col items-center gap-6 w-full">
			<div class="text-center">
				<h1 class="text-4xl font-bold text-primary mb-2">TCG Tournament</h1>
				<p class="text-base-content/70">Gestiona tus torneos de cartas en tiempo real</p>
			</div>

			<LoginForm />
		</div>
	{:else}
		<div class="w-full max-w-lg">
			<h1 class="text-3xl font-bold text-primary mb-6 text-center">Unirse a un Torneo</h1>

			<div class="card bg-base-200 shadow-xl border border-base-300">
				<div class="card-body">
					{#if joinSuccess}
						<div class="alert alert-success py-4">
							<span>Te has unido exitosamente al torneo!</span>
						</div>
					{:else}
						<div class="space-y-4">
							<div class="form-control">
								<label class="label" for="invite-code">
									<span class="label-text">Codigo de Invitacion</span>
								</label>
								<input
									id="invite-code"
									type="text"
									bind:value={inviteCode}
									placeholder="Ej. a1b2c3d4"
									class="input input-bordered w-full"
								/>
							</div>
							<div class="form-control">
								<label class="label" for="player-name">
									<span class="label-text">Nombre de Jugador</span>
								</label>
								<input
									id="player-name"
									type="text"
									bind:value={playerName}
									placeholder="Tu nickname"
									class="input input-bordered w-full"
								/>
							</div>

							{#if joinError}
								<p class="text-error text-sm">{joinError}</p>
							{/if}

							<button
								class="btn btn-primary w-full"
								onclick={handleJoin}
								disabled={isJoining}
							>
								{#if isJoining}
									<span class="loading loading-spinner"></span>
								{/if}
								Unirse al Torneo
							</button>
						</div>
					{/if}
				</div>
			</div>
		</div>
	{/if}
</div>