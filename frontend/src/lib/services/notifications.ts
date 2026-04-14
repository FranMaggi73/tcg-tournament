import { db } from './firebase';
import { collection, addDoc, query, where, getDocs, updateDoc, deleteDoc, doc, limit } from 'firebase/firestore';
import type { Notification } from '../types/firebase';

export const notificationService = {
	/**
	 * Sends a tournament invitation notification to a specific user.
	 * Checks for duplicate invites (same recipient + tournament) before sending.
	 */
	async sendInvite(recipientId: string, senderId: string, tournamentId: string, inviteCode: string, tournamentName: string) {
		try {
			// Check if there's already a non-expired invite for this recipient + tournament
			const existing = await this.hasExistingInvite(recipientId, tournamentId);
			if (existing) {
				throw new Error('Ya enviaste una invitación a este amigo para este torneo.');
			}

			const now = new Date();
			const expiresAt = new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000); // 7 days

			const notificationsRef = collection(db, 'notifications');
			const message = `Has sido invitado al torneo ${tournamentName}`;

			return await addDoc(notificationsRef, {
				type: 'tournament_invite',
				recipientId,
				senderId,
				tournamentId,
				inviteCode,
				tournamentName,
				message,
				read: false,
				createdAt: now,
				expiresAt
			});
		} catch (e) {
			console.error('[notifications] Error sending invite to Firestore:', e);
			throw e;
		}
	},

	/**
	 * Checks if a non-expired tournament invite already exists for this recipient + tournament.
	 */
	async hasExistingInvite(recipientId: string, tournamentId: string): Promise<boolean> {
		try {
			const notificationsRef = collection(db, 'notifications');
			const q = query(
				notificationsRef,
				where('recipientId', '==', recipientId),
				where('tournamentId', '==', tournamentId),
				limit(10)
			);
			const snapshot = await getDocs(q);
			const now = new Date();
			for (const doc of snapshot.docs) {
				const data = doc.data();
				const expiresAt = data.expiresAt instanceof Date ? data.expiresAt : new Date(data.expiresAt?.seconds ? data.expiresAt.toDate() : 0);
				if (!data.read && expiresAt > now) {
					return true;
				}
			}
			return false;
		} catch (e) {
			console.error('[notifications] Error checking existing invite:', e);
			return false;
		}
	},

	/**
	 * Fetches notifications for a specific user, sorted by newest first.
	 * Expired notifications are automatically filtered out.
	 */
	async getNotifications(userId: string, maxItems = 20) {
		try {
			const notificationsRef = collection(db, 'notifications');
			const q = query(
				notificationsRef,
				where('recipientId', '==', userId),
				limit(maxItems)
			);

			const querySnapshot = await getDocs(q);
			const now = new Date();
			const results: Notification[] = [];

			for (const docSnap of querySnapshot.docs) {
				const data = docSnap.data();
				const expiresAt = data.expiresAt instanceof Date
					? data.expiresAt
					: data.expiresAt?.seconds
						? data.expiresAt.toDate()
						: new Date(0);

				// Skip expired notifications
				if (expiresAt <= now) continue;

				results.push({
					id: docSnap.id,
					type: data.type ?? 'tournament_invite',
					recipientId: data.recipientId,
					senderId: data.senderId,
					tournamentId: data.tournamentId,
					inviteCode: data.inviteCode,
					tournamentName: data.tournamentName,
					message: data.message,
					read: data.read,
					createdAt: data.createdAt instanceof Date ? data.createdAt : new Date(),
					expiresAt
				});
			}

			// Sort by createdAt descending in memory (avoids composite index requirement)
			results.sort((a, b) => {
				const aTime = a.createdAt instanceof Date ? a.createdAt.getTime() : 0;
				const bTime = b.createdAt instanceof Date ? b.createdAt.getTime() : 0;
				return bTime - aTime;
			});

			return results;
		} catch (e) {
			console.error('[notifications] Error fetching notifications from Firestore:', e);
			return [];
		}
	},

	/**
	 * Marks a notification as read and deletes it after a short delay.
	 */
	async markAsReadAndDelete(notificationId: string) {
		const notificationRef = doc(db, 'notifications', notificationId);
		await updateDoc(notificationRef, { read: true });
		// Delete after 500ms so the UI can show the "accepted" state briefly
		setTimeout(async () => {
			try {
				await deleteDoc(notificationRef);
			} catch (e) {
				console.error('[notifications] Error deleting notification:', e);
			}
		}, 500);
	},

	/**
	 * Marks a notification as read.
	 */
	async markAsRead(notificationId: string) {
		const notificationRef = doc(db, 'notifications', notificationId);
		return await updateDoc(notificationRef, {
			read: true
		});
	}
};
