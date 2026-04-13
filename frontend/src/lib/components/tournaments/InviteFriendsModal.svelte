<script lang="ts">
	import { onMount } from 'svelte';
	import { friendshipApi } from '$lib/services/api';
	import { notificationService } from '$lib/services/notifications';
	import { authStore } from '$lib/stores/auth.svelte';
	import { getCachedProfile } from '$lib/stores/users.svelte';
	import type { Friendship } from '$lib/types/firebase';

	let { tournament, onClose } = $props<{
		tournament: any,
		onClose: () => void
	}>();

	let friends = $state<Friendship[]>([]);
	let selectedFriends = $state<Set<string>>(new Set());
	let isLoading = $state(true);
	let isSending = $state(false);
	let errorMessage = $state('');

	async function loadFriends() {
		try {
			isLoading = true;
			friends = await friendshipApi.getFriends();
		} catch (e: any) {
			errorMessage = 'Error cargando amigos: ' + e.message;
		} finally {
			isLoading = false;
		}
	}

	onMount(loadFriends);

	async function handleSend() {
		if (selectedFriends.size === 0) return;

		isSending = true;
		errorMessage = '';

		try {
			const senderId = authStore.user?.uid;
			if (!senderId) throw new Error('Usuario no autenticado');

			const sendPromises = Array.from(selectedFriends).map(friendId => {
				return notificationService.sendInvite(
					friendId,
					senderId,
					tournament.id,
					tournament.inviteCode,
					tournament.name
				);
			});

			await Promise.all(sendPromises);
			alert(`¡Invitaciones enviadas a ${selectedFriends.size} amigos!`);
			onClose();
		} catch (e: any) {
			errorMessage = 'Error al enviar invitaciones: ' + e.message;
		} finally {
			isSending = false;
		}
	}

	function toggleFriend(uid: string) {
		if (selectedFriends.has(uid)) {
			selectedFriends.delete(uid);
		} else {
			selectedFriends.add(uid);
		}
	}

	function getFriendUid(friendship: Friendship) {
		return friendship.user1Id === authStore.user?.uid ? friendship.user2Id : friendship.user1Id;
	}
</script>

<div class="modal modal-open">
	<div class="modal-box max-w-md">
		<h3 class="text-xl font-bold text-primary mb-4">Invitar Amigos</h3>
		<p class="text-sm opacity-70 mb-6">Envía el código de invitación directamente a tus amigos aceptados.</p>

		{#if isLoading}
			<div class="flex justify-center py-12">
				<span class="loading loading-ring loading-lg text-primary"></span>
			</div>
		{:else if errorMessage}
			<div class="alert alert-error mb-4 py-2 text-xs">
				<span>{errorMessage}</span>
			</div>
		{/if}

		<div class="space-y-2 max-h-64 overflow-y-auto mb-6">
			{#each friends as friend}
				{@const friendUid = getFriendUid(friend)}
				{@const profile = getCachedProfile(friendUid)}
				<label class="flex items-center gap-3 p-3 bg-base-300 rounded-box cursor-pointer hover:bg-base-100 transition-colors border border-transparent {selectedFriends.has(friendUid) ? 'border-primary bg-base-100' : ''}">
					<input
						type="checkbox"
						class="checkbox checkbox-primary checkbox-sm"
						checked={selectedFriends.has(friendUid)}
						onchange={() => toggleFriend(friendUid)}
					/>
					<div class="flex items-center gap-3">
						<div class="avatar placeholder avatar-xs">
							<div class="bg-neutral text-neutral-content rounded-full w-8">
								<span>{profile?.displayName?.charAt(0).toUpperCase() || 'U'}</span>
							</div>
						</div>
						<span class="text-sm font-medium">{profile?.displayName || friendUid}</span>
					</div>
				</label>
			{/each}
			{#if friends.length === 0}
				<p class="text-center py-8 opacity-50 italic text-sm">No tienes amigos aceptados.</p>
			{/if}
		</div>

		<div class="modal-action">
			<button class="btn btn-ghost" onclick={onClose} disabled={isSending}>Cancelar</button>
			<button
				class="btn btn-primary"
				onclick={handleSend}
				disabled={isSending || selectedFriends.size === 0}
			>
				{#if isSending}
					<span class="loading loading-spinner"></span>
				{/if}
				Enviar Invitaciones
			</button>
		</div>
	</div>
	<div class="modal-backdrop" role="button" tabindex="0" onclick={onClose} onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') onClose(); }}></div>
</div>

<style>
	.modal-backdrop {
		position: fixed;
		inset: 0;
		background-color: rgba(0, 0, 0, 0.5);
		z-index: 40;
	}
	.modal-box {
		z-index: 50;
	}
</style>
