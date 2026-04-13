import { db } from './firebase';
import { collection, addDoc, query, where, getDocs, updateDoc, doc, orderBy, limit } from 'firebase/firestore';
import type { Notification } from '../types/firebase';

export const notificationService = {
	/**
	 * Sends a tournament invitation notification to a specific user.
	 */
	async sendInvite(recipientId: string, senderId: string, tournamentId: string, inviteCode: string, tournamentName: string) {
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
	},

	/**
	 * Fetches notifications for a specific user, sorted by newest first.
	 */
	async getNotifications(userId: string, maxItems = 20) {
		const notificationsRef = collection(db, 'notifications');
		const q = query(
			notificationsRef,
			where('recipientId', '==', userId),
			orderBy('createdAt', 'desc'),
			limit(maxItems)
		);

		const querySnapshot = await getDocs(q);
		return querySnapshot.docs.map(doc => ({
			id: doc.id,
			...doc.data()
		})) as Notification[];
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
