<script lang="ts">
	import { onMount } from 'svelte';
	import { authStore, waitForAuth } from '$lib/stores/auth.svelte';
	import { getUserProfile, updateUserProfile, resolveUserProfiles } from '$lib/services/user';
	import { friendshipApi, tournamentApi } from '$lib/services/api';
	import { notificationService } from '$lib/services/notifications';
	import { refreshInvitationCount } from '$lib/stores/notifications.svelte';
	import type { UserProfile, Friendship, Notification } from '$lib/types/firebase';

	let { data } = $props();
	let profile = $state<UserProfile | null>(null);
	let isLoading = $state(true);
	let isSaving = $state(false);
	let errorMessage = $state('');
	let copiedUid = $state(false);

	// Friends state
	let friends = $state<Friendship[]>([]);
	let pendingRequests = $state<Friendship[]>([]);
	let friendSearch = $state('');
	let isAddingFriend = $state(false);
	let friendProfiles = $state<Record<string, UserProfile | null>>({});

	// Notifications state
	let notifications = $state<Notification[]>([]);
	let isNotifLoading = $state(false);

	async function resolveFriendUids() {
		const uids: string[] = [];
		for (const f of friends) {
			if (!friendProfiles[f.user1Id]) uids.push(f.user1Id);
			if (!friendProfiles[f.user2Id]) uids.push(f.user2Id);
		}
		for (const r of pendingRequests) {
			if (!friendProfiles[r.user1Id]) uids.push(r.user1Id);
			if (!friendProfiles[r.user2Id]) uids.push(r.user2Id);
		}
		if (uids.length > 0) {
			const resolved = await resolveUserProfiles(uids);
			friendProfiles = { ...friendProfiles, ...resolved };
		}
	}

	onMount(async () => {
		await waitForAuth();

		if (!authStore.user) {
			isLoading = false;
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
			await Promise.all([
				loadFriends(),
				loadPendingRequests(),
				loadNotifications(),
				refreshInvitationCount()
			]);
		} catch (e: any) {
			errorMessage = e.message;
		} finally {
			isLoading = false;
		}
	});

	async function handleCopyUid() {
		if (!profile?.uid) return;
		try {
			await navigator.clipboard.writeText(profile.uid);
			copiedUid = true;
			setTimeout(() => { copiedUid = false; }, 2000);
		} catch {
			const textArea = document.createElement('textarea');
			textArea.value = profile.uid;
			document.body.appendChild(textArea);
			textArea.select();
			document.execCommand('copy');
			document.body.removeChild(textArea);
			copiedUid = true;
			setTimeout(() => { copiedUid = false; }, 2000);
		}
	}

	async function loadFriends() {
		try {
			const all = await friendshipApi.getFriends();
			friends = all ?? [];
			await resolveFriendUids();
		} catch (e: any) {
			console.error('Error loading friends:', e);
		}
	}

	async function loadPendingRequests() {
		try {
			const result = await friendshipApi.getPendingRequests();
			pendingRequests = result ?? [];
			await resolveFriendUids();
		} catch (e: any) {
			console.error('Error loading pending requests:', e);
		}
	}

	async function loadNotifications() {
		if (!authStore.user) return;
		isNotifLoading = true;
		try {
			const uid = authStore.user.uid;
			notifications = await notificationService.getNotifications(uid);
		} catch (e: any) {
			console.error('[profile] Error loading notifications:', e);
		} finally {
			isNotifLoading = false;
		}
	}

	async function handleJoinTournament(notification: Notification) {
		try {
			const playerName = profile?.displayName || authStore.user?.displayName || 'Jugador';
			await tournamentApi.joinByCode(notification.inviteCode, authStore.user?.email || '', playerName);
			await notificationService.markAsReadAndDelete(notification.id);
			alert('¡Te has unido al torneo exitosamente!');
			await loadNotifications();
			await refreshInvitationCount();
		} catch (e: any) {
			alert(`Error al unirse: ${e.message}`);
		}
	}

	async function handleRejectInvite(notificationId: string) {
		try {
			await notificationService.markAsReadAndDelete(notificationId);
			await loadNotifications();
			await refreshInvitationCount();
		} catch (e: any) {
			alert(`Error al rechazar invitación: ${e.message}`);
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
			await refreshInvitationCount();
		} catch (e: any) {
			alert(`Error al aceptar solicitud: ${e.message}`);
		}
	}

	async function handleDeclineRequest(friendshipId: string) {
		try {
			await friendshipApi.updateStatus(friendshipId, 'declined');
			await loadPendingRequests();
			await refreshInvitationCount();
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
				photoURL: profile.photoURL,
				bio: profile.bio
			});
			alert('Perfil actualizado con exito');
		} catch (e: any) {
			errorMessage = e.message;
		} finally {
			isSaving = false;
		}
	}

	function getDisplayName(uid: string): string {
		return friendProfiles[uid]?.displayName || uid.substring(0, 8) + '...';
	}

	function getPhotoURL(uid: string): string | null {
		return friendProfiles[uid]?.photoURL || null;
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
		<h1 class="text-3xl font-bold text-primary">{profile?.displayName || authStore.user?.displayName || 'Mi Perfil'}</h1>
		<p class="text-base-content/60">{authStore.user?.email}</p>
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
					<!-- Mi UID - Compartible -->
					<div class="form-control">
						<label class="label" for="profile-uid">
							<span class="label-text font-bold">Mi ID</span>
						</label>
						<div class="flex gap-2">
							<input
								id="profile-uid"
								type="text"
								value={profile.uid}
								readonly
								class="input input-bordered w-full text-xs font-mono bg-base-300"
							/>
							<button
								class="btn btn-outline btn-sm {copiedUid ? 'btn-success' : ''}"
								onclick={handleCopyUid}
							>
								{#if copiedUid}
									Copiado
								{:else}
									Copiar
								{/if}
							</button>
						</div>
					</div>

					<div class="divider my-0"></div>

					<!-- Nombre (readonly, viene de Google) -->
					<div class="form-control">
						<label class="label" for="profile-name">
							<span class="label-text font-bold">Nombre</span>
						</label>
						<input
							id="profile-name"
							type="text"
							value={profile.displayName}
							readonly
							class="input input-bordered w-full bg-base-300"
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
							<span class="label-text font-bold">Biografia</span>
						</label>
						<textarea
							id="profile-bio"
							bind:value={profile.bio}
							placeholder="Cuentanos sobre tu mazo favorito..."
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
								{@const isExpired = notif.expiresAt && new Date(notif.expiresAt) <= new Date()}
								{@const expiresLabel = notif.expiresAt
									? (() => {
										const diff = new Date(notif.expiresAt).getTime() - Date.now();
										if (diff <= 0) return 'Expirada';
										const days = Math.floor(diff / (1000 * 60 * 60 * 24));
										if (days > 0) return `Expira en ${days}d`;
										const hours = Math.floor(diff / (1000 * 60 * 60));
										if (hours > 0) return `Expira en ${hours}h`;
										const mins = Math.floor(diff / (1000 * 60));
										return `Expira en ${mins}m`;
									})()
									: ''}
								<div class="flex items-center justify-between p-4 bg-base-300 rounded-box border border-base-300 {notif.read || isExpired ? 'opacity-50' : 'ring-1 ring-primary' }">
									<div class="flex flex-col flex-1 min-w-0">
										<p class="font-bold text-sm truncate">{notif.message}</p>
										<p class="text-xs opacity-60">{notif.tournamentName}</p>
										<p class="text-xs opacity-40 mt-1">{expiresLabel}</p>
									</div>
									{#if !isExpired}
										<div class="flex gap-2 ml-4 shrink-0">
											<button
												class="btn btn-error btn-outline btn-xs"
												onclick={() => handleRejectInvite(notif.id)}
											>
												Rechazar
											</button>
											<button
												class="btn btn-primary btn-xs"
												onclick={() => handleJoinTournament(notif)}
											>
												Unirse
											</button>
										</div>
									{:else}
										<span class="badge badge-ghost badge-xs ml-4">Expirada</span>
									{/if}
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
								{@const senderUid = request.user1Id}
								{@const senderName = getDisplayName(senderUid)}
								{@const senderPhoto = getPhotoURL(senderUid)}
								<div class="flex items-center justify-between p-4 bg-base-300 rounded-box border border-base-300">
									<div class="flex items-center gap-3">
										<div class="avatar placeholder">
											<div class="bg-neutral text-neutral-content rounded-full w-10 overflow-hidden">
												{#if senderPhoto}
													<img src={senderPhoto} alt="avatar" class="w-full h-full object-cover" />
												{:else}
													<div class="w-full h-full flex items-center justify-center text-xs font-bold">
														{senderName.charAt(0).toUpperCase()}
													</div>
												{/if}
											</div>
										</div>
										<div>
											<p class="text-sm font-bold">{senderName}</p>
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

					<p class="text-sm opacity-60 mb-4">Pega el ID de tu amigo para enviarle una solicitud. Tu amigo puede copiar su ID desde su perfil.</p>

					<div class="flex gap-2 mb-6">
						<input
							type="text"
							bind:value={friendSearch}
							placeholder="Pega el ID de tu amigo aqui..."
							class="input input-bordered flex-1 font-mono text-sm"
						/>
						<button
							class="btn btn-primary"
							onclick={handleAddFriend}
							disabled={isAddingFriend}
						>
							{#if isAddingFriend}
								<span class="loading loading-spinner"></span>
							{:else}
								Anadir
							{/if}
						</button>
					</div>

					<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
						{#each friends as friend}
							{@const friendUid = friend.user1Id === authStore.user?.uid ? friend.user2Id : friend.user1Id}
							{@const friendName = getDisplayName(friendUid)}
							{@const friendPhoto = getPhotoURL(friendUid)}
							<div class="flex items-center gap-3 p-3 bg-base-300 rounded-box border border-base-300">
								<div class="avatar placeholder">
									<div class="bg-neutral text-neutral-content rounded-full w-10 overflow-hidden">
										{#if friendPhoto}
											<img src={friendPhoto} alt="avatar" class="w-full h-full object-cover" />
										{:else}
											<div class="w-full h-full flex items-center justify-center text-xs font-bold">
												{friendName.charAt(0).toUpperCase()}
											</div>
										{/if}
									</div>
								</div>
								<div class="flex-1 overflow-hidden">
									<p class="text-sm font-bold truncate">{friendName}</p>
								</div>
							</div>
						{/each}
						{#if friends.length === 0}
							<p class="text-center col-span-2 py-8 opacity-50 italic">Aun no tienes amigos agregados.</p>
						{/if}
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>