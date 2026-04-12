import { db } from '$lib/services/firebase';
import { auth } from '$lib/services/firebase';
import { doc, getDoc, setDoc, updateDoc } from 'firebase/firestore';
import { updateProfile } from 'firebase/auth';
import type { UserProfile } from '$lib/types/firebase';
import { setCachedProfile, getCachedProfile, userCache } from '$lib/stores/users.svelte';

const USERS_COLLECTION = 'users';

/**
 * Resolves a list of UIDs to UserProfiles, using a cache to avoid N+1 queries.
 */
export async function resolveUserProfiles(uids: string[]): Promise<Record<string, UserProfile | null>> {
	const results: Record<string, UserProfile | null> = {};
	const missingUids: string[] = [];

	// 1. Check cache first
	for (const uid of uids) {
		const cached = getCachedProfile(uid);
		if (cached) {
			results[uid] = cached;
		} else {
			missingUids.push(uid);
		}
	}

	if (missingUids.length === 0) return results;

	// 2. Fetch missing profiles in parallel
	userCache.isLoading = true;
	try {
		const fetchPromises = missingUids.map(async (uid) => {
			const docRef = doc(db, USERS_COLLECTION, uid);
			const docSnap = await getDoc(docRef);

			let profile: UserProfile | null = null;
			if (docSnap.exists()) {
				profile = docSnap.data() as UserProfile;
			} else {
				// Fallback to Auth profile if Firestore doc doesn't exist
				const authUser = auth.currentUser;
				if (authUser && authUser.uid === uid) {
					profile = {
						uid: authUser.uid,
						displayName: authUser.displayName || 'Usuario Anónimo',
						photoURL: authUser.photoURL,
						updatedAt: new Date()
					};
				} else {
					profile = {
						uid,
						displayName: `Jugador ${uid.substring(0, 5)}`,
						photoURL: null,
						updatedAt: new Date()
					};
				}
			}

			setCachedProfile(uid, profile);
			return { uid, profile };
		});

		const fetched = await Promise.all(fetchPromises);
		fetched.forEach(({ uid, profile }) => {
			results[uid] = profile;
		});
	} finally {
		userCache.isLoading = false;
	}

	return results;
}

/**
 * Updates the user's profile in both Firestore and Firebase Auth
 */
export async function updateUserProfile(uid: string, updates: Partial<UserProfile>) {
	const user = auth.currentUser;
	if (!user) throw new Error('No authenticated user found');

	// 1. Update Firestore
	const userRef = doc(db, USERS_COLLECTION, uid);
	await setDoc(userRef, {
		...updates,
		uid,
		updatedAt: new Date()
	}, { merge: true });

	// 2. Update Firebase Auth Profile
	if (updates.displayName || updates.photoURL) {
		await updateProfile(user, {
			displayName: updates.displayName || user.displayName,
			photoURL: updates.photoURL || user.photoURL
		});
	}

	// 3. Update Local Cache
	const current = getCachedProfile(uid) || { uid, displayName: '', photoURL: null, updatedAt: new Date() };
	setCachedProfile(uid, { ...current, ...updates });
}

/**
 * Retrieves a single user profile
 */
export async function getUserProfile(uid: string): Promise<UserProfile | null> {
	const cached = getCachedProfile(uid);
	if (cached) return cached;

	const docRef = doc(db, USERS_COLLECTION, uid);
	const docSnap = await getDoc(docRef);

	if (docSnap.exists()) {
		const profile = docSnap.data() as UserProfile;
		setCachedProfile(uid, profile);
		return profile;
	}
	return null;
}
