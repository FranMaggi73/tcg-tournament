import { auth } from '$lib/services/firebase';
import { onAuthStateChanged, type User } from 'firebase/auth';
import type { FirebaseUser } from '$lib/types/firebase';

// Define the reactive state for authentication
interface AuthState {
	user: FirebaseUser | null;
	isLoading: boolean;
}

export const authStore = $state<AuthState>({
	user: null,
	isLoading: true
});

/**
 * Initializes the auth observer to keep the authStore in sync with Firebase Auth
 */
export function initAuthObserver() {
	onAuthStateChanged(auth, (firebaseUser: User | null) => {
		if (firebaseUser) {
			authStore.user = {
				uid: firebaseUser.uid,
				email: firebaseUser.email,
				displayName: firebaseUser.displayName,
				photoURL: firebaseUser.photoURL
			};
		} else {
			authStore.user = null;
		}
		authStore.isLoading = false;
	});
}

/**
 * Returns a promise that resolves when auth state is ready (not loading).
 * Useful in load functions to wait for auth before checking user state.
 */
export function waitForAuth(): Promise<void> {
	return new Promise((resolve) => {
		if (!authStore.isLoading) {
			resolve();
			return;
		}
		// Poll until auth is ready
		const interval = setInterval(() => {
			if (!authStore.isLoading) {
				clearInterval(interval);
				resolve();
			}
		}, 50);
	});
}
