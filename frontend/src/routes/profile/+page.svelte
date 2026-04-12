<script lang="ts">
	import { onMount } from 'svelte';
	import { authStore } from '$lib/stores/auth.svelte';
	import { getUserProfile, updateUserProfile } from '$lib/services/user';
	import type { UserProfile } from '$lib/types/firebase';

	let { data } = $props();
	let profile = $state<UserProfile | null>(null);
	let isLoading = $state(true);
	let isSaving = $state(false);
	let errorMessage = $state('');

	onMount(async () => {
		if (!authStore.user) {
			// Redirect logic would usually go in +page.ts
			return;
		}

		try {
			profile = await getUserProfile(authStore.user.uid);
			if (!profile) {
				// Initialize a default profile if none exists
				profile = {
					uid: authStore.user.uid,
					displayName: authStore.user.displayName || 'Nuevo Jugador',
					photoURL: null,
					updatedAt: new Date()
				};
			}
		} catch (e: any) {
			errorMessage = e.message;
		} finally {
			isLoading = false;
		}
	});

	async function handleSave() {
		if (!profile) return;

		isSaving = true;
		errorMessage = '';
		try {
			await updateUserProfile(profile.uid, {
				displayName: profile.displayName,
				photoURL: profile.photoURL,
				bio: profile.bio
			});
			alert('Perfil actualizado con éxito');
		} catch (e: any) {
			errorMessage = e.message;
		} finally {
			isSaving = false;
		}
	}
</script>

<div class="p-8 max-w-2xl mx-auto">
	<header class="flex flex-col items-center text-center mb-12">
		<div class="avatar mb-4">
			<div class="w-24 h-24 rounded-full ring ring-primary ring-offset-base-100 ring-offset-2 overflow-hidden">
				{#if profile?.photoURL}
					<img src={profile.photoURL} alt="Avatar" />
				{:else}
					<div class="bg-base-300 w-full h-full flex items-center justify-center text-2xl font-bold text-primary">
						{profile?.displayName?.charAt(0).toUpperCase() || 'U'}
					</div>
				{/if}
			</div>
		</div>
		<h1 class="text-3xl font-bold text-primary">Mi Perfil TCG</h1>
		<p class="text-base-content/60">Personaliza tu identidad en el torneo</p>
	</header>

	{#if isLoading}
		<div class="flex justify-center py-20">
			<span class="loading loading-ring loading-lg text-primary"></span>
		</div>
	{:else if profile}
		<div class="card bg-base-200 shadow-xl border border-base-300 p-8">
			<div class="space-y-6">
				<div class="form-control">
					<label class="label" for="profile-nickname">
						<span class="label-text font-bold">Nickname TCG</span>
					</label>
					<input
						id="profile-nickname"
						type="text"
						bind:value={profile.displayName}
						placeholder="Ej. ShadowMaster99"
						class="input input-bordered w-full"
					/>
				</div>

				<div class="form-control">
					<label class="label" for="profile-avatar">
						<span class="label-text font-bold">URL de Avatar</span>
					</label>
					<input
						id="profile-avatar"
						type="text"
						bind:value={profile.photoURL}
						placeholder="https://..."
						class="input input-bordered w-full"
					/>
				</div>

				<div class="form-control">
					<label class="label" for="profile-bio">
						<span class="label-text font-bold">Biografía</span>
					</label>
					<textarea
						id="profile-bio"
						bind:value={profile.bio}
						placeholder="Cuéntanos sobre tu mazo favorito..."
						class="textarea textarea-bordered w-full h-24"
					></textarea>
				</div>

				{#if errorMessage}
					<div class="alert alert-error py-2 text-sm">
						<span>{errorMessage}</span>
					</div>
				{/if}

				<div class="flex justify-center mt-8">
					<button
						class="btn btn-primary w-full max-w-xs"
						onclick={handleSave}
						disabled={isSaving}
					>
						{#if isSaving}
							<span class="loading loading-spinner"></span>
						{/if}
						Guardar Cambios
					</button>
				</div>
			</div>
		</div>
	{/if}
</div>
