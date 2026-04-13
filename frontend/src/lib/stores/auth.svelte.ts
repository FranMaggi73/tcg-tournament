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

// Auth ready promise — resolves once onAuthStateChanged fires for the first time
let authReadyResolve: () => void;
const authReadyPromise = new Promise<void>((resolve) => {
	authReadyResolve = resolve;
});

/**
 * Initializes the auth observer to keep the authStore in sync with Firebase Auth.
 * Called eagerly at module load time so auth state is available for load functions.
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