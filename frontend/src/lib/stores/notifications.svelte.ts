import { authStore } from './auth.svelte';
import { friendshipApi } from '$lib/services/api';
import { notificationService } from '$lib/services/notifications';
import { onAuthStateChanged } from 'firebase/auth';
import { auth } from '$lib/services/firebase';

interface InvitationCount {
	count: number;
	isLoading: boolean;
}

export const invitationCountStore = $state<InvitationCount>({
	count: 0,
	isLoading: false
});

let unsubscribe: (() => void) | null = null;

export async function refreshInvitationCount() {
	if (!authStore.user) {
		invitationCountStore.count = 0;
		return;
	}

	invitationCountStore.isLoading = true;
	try {
		const uid = authStore.user.uid;

		const [notifications, pendingRequests] = await Promise.all([
			notificationService.getNotifications(uid),
			friendshipApi.getPendingRequests().catch(() => [])
		]);

		const unreadTournamentInvites = notifications.filter(n => !n.read).length;
		const unreadFriendRequests = pendingRequests.length;

		invitationCountStore.count = unreadTournamentInvites + unreadFriendRequests;
	} catch (e) {
		console.error('Error loading invitation count:', e);
		invitationCountStore.count = 0;
	} finally {
		invitationCountStore.isLoading = false;
	}
}

export function startInvitationCountObserver() {
	if (unsubscribe) return;

	unsubscribe = onAuthStateChanged(auth, (user) => {
		if (user) {
			refreshInvitationCount();
		} else {
			invitationCountStore.count = 0;
		}
	});
}

export function stopInvitationCountObserver() {
	if (unsubscribe) {
		unsubscribe();
		unsubscribe = null;
	}
}
