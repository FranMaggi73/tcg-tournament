<script lang="ts">
	import { onMount } from 'svelte';
	import { authStore } from '$lib/stores/auth.svelte';
	import { getUserProfile, updateUserProfile } from '$lib/services/user';
	import { friendshipApi, tournamentApi } from '$lib/services/api';
	import { notificationService } from '$lib/services/notifications';
	import type { UserProfile, Friendship, Notification } from '$lib/types/firebase';

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

	// Notifications state
	let notifications = $state<Notification[]>([]);
	let isNotifLoading = $state(false);

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
			await loadPendingRequests();
			await loadNotifications();
		} catch (e: any) {
			errorMessage = e.message;
		} finally {
			isLoading = false;
		}
	});

	async function loadFriends() {
		try {
			const all = await friendshipApi.getFriends();
			friends = all ?? [];
		} catch (e: any) {
			console.error('Error loading friends:', e);
		}
	}

	async function loadPendingRequests() {
		try {
			const result = await friendshipApi.getPendingRequests();
			pendingRequests = result ?? [];
		} catch (e: any) {
			console.error('Error loading pending requests:', e);
		}
	}

	async function loadNotifications() {
		if (!authStore.user) return;
		isNotifLoading = true;
		try {
			notifications = await notificationService.getNotifications(authStore.user.uid);
		} catch (e: any) {
			console.error('Error loading notifications:', e);
		} finally {
			isNotifLoading = false;
		}
	}

	async function handleJoinTournament(notification: Notification) {
		try {
			const playerName = profile?.displayName || 'Jugador';
			await tournamentApi.joinByCode(notification.inviteCode, authStore.user?.email || '', playerName);
			await notificationService.markAsRead(notification.id);
			alert('¡Te has unido al torneo exitosamente!');
			await loadNotifications();
		} catch (e: any) {
			alert(`Error al unirse: ${e.message}`);
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

	async function handleAcceptRequest(friendshipId: string) {
		try {
			await friendshipApi.updateStatus(friendshipId, 'accepted');
			await loadFriends();
			await loadPendingRequests();
		} catch (e: any) {
			alert(`Error al aceptar solicitud: ${e.message}`);
		}
	}

	async function handleDeclineRequest(friendshipId: string) {
		try {
			await friendshipApi.updateStatus(friendshipId, 'declined');
			await loadPendingRequests();
		} catch (e: any) {
			alert(`Error al rechazar solicitud: ${e.message}`);
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
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
			<!-- Profile Settings -->
			<div class="lg:col-span-1 card bg-base-200 shadow-xl border border-base-300 p-8">
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
			<div class="lg:col-span-2 space-y-6">
				<!-- Notifications Card -->
				<div class="card bg-base-200 shadow-xl border border-base-300 p-8">
					<h2 class="text-xl font-bold text-primary mb-6">Invitaciones Recibidas</h2>

					{#if isNotifLoading}
						<div class="flex justify-center py-8">
							<span class="loading loading-spinner loading-md text-primary"></span>
						</div>
					{:else if notifications.length === 0}
						<p class="text-center py-8 opacity-50 italic">No tienes invitaciones pendientes.</p>
					{:else}
						<div class="space-y-4">
							{#each notifications as notif}
								<div class="flex items-center justify-between p-4 bg-base-300 rounded-box border border-base-300 {notif.read ? 'opacity-50' : 'ring-1 ring-primary' }">
									<div class="flex flex-col">
										<p class="font-bold text-sm">{notif.message}</p>
										<p class="text-xs opacity-60">{notif.tournamentName}</p>
									</div>
									<button
										class="btn btn-primary btn-xs"
										onclick={() => handleJoinTournament(notif)}
										disabled={notif.read}
									>
										Unirse
									</button>
								</div>
							{/each}
						</div>
					{/if}
				</div>

				<!-- Pending Friend Requests Card -->
				<div class="card bg-base-200 shadow-xl border border-base-300 p-8">
					<h2 class="text-xl font-bold text-primary mb-6">Solicitudes de Amistad</h2>

					{#if pendingRequests.length === 0}
						<p class="text-center py-8 opacity-50 italic">No tienes solicitudes de amistad pendientes.</p>
					{:else}
						<div class="space-y-3">
							{#each pendingRequests as request}
								<div class="flex items-center justify-between p-4 bg-base-300 rounded-box border border-base-300">
									<div class="flex items-center gap-3">
										<div class="avatar placeholder">
											<div class="bg-neutral text-neutral-content rounded-full w-10">
												<span class="text-xs">{request.user1Id.charAt(0).toUpperCase()}</span>
											</div>
										</div>
										<div>
											<p class="text-sm font-bold">{request.user1Id}</p>
											<p class="text-xs opacity-50">Quiere ser tu amigo</p>
										</div>
									</div>
									<div class="flex gap-2">
										<button
											class="btn btn-success btn-xs"
											onclick={() => handleAcceptRequest(request.id)}
										>
											Aceptar
										</button>
										<button
											class="btn btn-error btn-xs btn-outline"
											onclick={() => handleDeclineRequest(request.id)}
										>
											Rechazar
										</button>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>

				<!-- Friends List Card -->
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
							{@const friendUid = friend.user1Id === authStore.user?.uid ? friend.user2Id : friend.user1Id}
							<div class="flex items-center gap-3 p-3 bg-base-300 rounded-box border border-base-300">
								<div class="avatar placeholder">
									<div class="bg-neutral text-neutral-content rounded-full w-10">
										<span class="text-xs">{friendUid.charAt(0).toUpperCase()}</span>
									</div>
								</div>
								<div class="flex-1 overflow-hidden">
									<p class="text-sm font-bold truncate">{friendUid}</p>
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
	{#if profile}
		<div class="flex justify-center py-4 opacity-50 text-xs">
			Actualizado el {(profile.updatedAt as any)?.toDate ? (profile.updatedAt as any).toDate().toLocaleString() : profile.updatedAt?.toLocaleString() || 'Recientemente'}
		</div>
	{/if}
</div>