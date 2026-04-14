import { auth } from '$lib/services/firebase';
import { onAuthStateChanged, type User } from 'firebase/auth';
import type { FirebaseUser } from '$lib/types/firebase';
import { ensureUserProfile } from '$lib/services/user';

// Define the reactive state for authentication
interface AuthState {
	user: FirebaseUser | null;
	isLoading: boolean;
}

export const authStore = $state<AuthState>({
	user: null,
	isLoading: true
});

// Auth ready promise — resolves once onAuthStateChanged fires for the first time
let authReadyResolve: () => void;
const authReadyPromise = new Promise<void>((resolve) => {
	authReadyResolve = resolve;
});

/**
 * Initializes the auth observer to keep the authStore in sync with Firebase Auth.
 * Also auto-creates the user's profile in Firestore on first login.
 */
export function initAuthObserver() {
	onAuthStateChanged(auth, async (firebaseUser: User | null) => {
		if (firebaseUser) {
			authStore.user = {
				uid: firebaseUser.uid,
				email: firebaseUser.email,
				displayName: firebaseUser.displayName,
				photoURL: firebaseUser.photoURL
			};

			// Auto-create profile in Firestore so other users can see the name
			try {
				await ensureUserProfile(
					firebaseUser.uid,
					firebaseUser.displayName,
					firebaseUser.email
				);
			} catch (e) {
				console.error('Error ensuring user profile:', e);
			}
		} else {
			authStore.user = null;
		}
		authStore.isLoading = false;
		authReadyResolve();
	});
}

/**
 * Returns a promise that resolves when auth state is ready (not loading).
 * Useful in load functions to wait for auth before checking user state.
 */
export function waitForAuth(): Promise<void> {
	if (!authStore.isLoading) return Promise.resolve();
	return authReadyPromise;
}

// Eagerly initialize auth observer when this module is imported
initAuthObserver();