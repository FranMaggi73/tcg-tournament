<script lang="ts">
	import { onMount } from 'svelte';
	import { authStore } from '$lib/stores/auth.svelte';
	import { getUserProfile, updateUserProfile } from '$lib/services/user';
	import { friendshipApi } from '$lib/services/api';
	import type { UserProfile, Friendship } from '$lib/types/firebase';

	let { data } = $props();
	let profile = $state<UserProfile | null>(null);
	let isLoading = $state(true);
	let isSaving = $state(false);
	let errorMessage = $state('');

	// Friends state
	let friends = $state<Friendship[]>([]);
	let pendingRequests = $state<Friendship[]>([]);
	let friendSearch = $state('');
	let isAddingFriend = $state(false);

	onMount(async () => {
		if (!authStore.user) {
			return;
		}

		try {
			profile = await getUserProfile(authStore.user.uid);
			if (!profile) {
				profile = {
					uid: authStore.user.uid,
					displayName: authStore.user.displayName || 'Nuevo Jugador',
					photoURL: null,
					updatedAt: new Date()
				};
			}
			await loadFriends();
		} catch (e: any) {
			errorMessage = e.message;
		} finally {
			isLoading = false;
		}
	});

	async function loadFriends() {
		try {
			const all = await friendshipApi.getFriends();
			// The backend currently returns only accepted friends.
			// In a real app, we'd have a separate endpoint for pending requests.
			friends = all;
		} catch (e: any) {
			console.error('Error loading friends:', e);
		}
	}

	async function handleAddFriend() {
		if (!friendSearch) return;
		isAddingFriend = true;
		try {
			await friendshipApi.addFriend(friendSearch);
			alert('Solicitud de amistad enviada');
			friendSearch = '';
		} catch (e: any) {
			alert(`Error: ${e.message}`);
		} finally {
			isAddingFriend = false;
		}
	}

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

<div class="p-8 max-w-4xl mx-auto">
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
		<div class="grid grid-cols-1 md:grid-cols-3 gap-8">
			<!-- Profile Settings -->
			<div class="md:col-span-1 card bg-base-200 shadow-xl border border-base-300 p-8">
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

			<!-- Friends List -->
			<div class="md:col-span-2 space-y-6">
				<div class="card bg-base-200 shadow-xl border border-base-300 p-8">
					<h2 class="text-xl font-bold text-primary mb-6">Mis Amigos</h2>

					<div class="flex gap-2 mb-6">
						<input
							type="text"
							bind:value={friendSearch}
							placeholder="ID del amigo..."
							class="input input-bordered flex-1"
						/>
						<button
							class="btn btn-primary"
							onclick={handleAddFriend}
							disabled={isAddingFriend}
						>
							{#if isAddingFriend}
								<span class="loading loading-spinner"></span>
							{:else}
								Añadir
							{/if}
						</button>
					</div>

					<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
						{#each friends as friend}
							<div class="flex items-center gap-3 p-3 bg-base-300 rounded-box border border-base-300">
								<div class="avatar placeholder">
									<div class="bg-neutral text-neutral-content rounded-full w-10">
										<span>U</span>
									</div>
								</div>
								<div class="flex-1 overflow-hidden">
									<p class="text-sm font-bold truncate">{friend.user1Id === authStore.user?.uid ? friend.user2Id : friend.user1Id}</p>
									<p class="text-xs opacity-50">Amigo aceptado</p>
								</div>
							</div>
						{/each}
						{#if friends.length === 0}
							<p class="text-center col-span-2 py-8 opacity-50 italic">Aún no tienes amigos agregados.</p>
						{/if}
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>
