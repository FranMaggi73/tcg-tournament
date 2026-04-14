import { db } from './firebase';
import { collection, addDoc, query, where, getDocs, updateDoc, doc, limit } from 'firebase/firestore';
import type { Notification } from '../types/firebase';

export const notificationService = {
	/**
	 * Sends a tournament invitation notification to a specific user.
	 */
	async sendInvite(recipientId: string, senderId: string, tournamentId: string, inviteCode: string, tournamentName: string) {
		try {
			const notificationsRef = collection(db, 'notifications');

			const message = `Has sido invitado al torneo ${tournamentName}`;

			return await addDoc(notificationsRef, {
				recipientId,
				senderId,
				tournamentId,
				inviteCode,
				tournamentName,
				message,
				read: false,
				createdAt: new Date()
			});
		} catch (e) {
			console.error('[notifications] Error sending invite to Firestore:', e);
			throw e;
		}
	},

	/**
	 * Fetches notifications for a specific user, sorted by newest first.
	 * Note: Uses a single-field query (recipientId) and sorts in memory to avoid
	 * requiring a composite Firestore index. For large datasets, add the composite
	 * index (recipientId ASC, createdAt DESC) in Firebase Console for better performance.
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
			const results = querySnapshot.docs.map(doc => ({
				id: doc.id,
				...doc.data()
			})) as Notification[];

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
	 * Marks a notification as read.
	 */
	async markAsRead(notificationId: string) {
		const notificationRef = doc(db, 'notifications', notificationId);
		return await updateDoc(notificationRef, {
			read: true
		});
	}
};
