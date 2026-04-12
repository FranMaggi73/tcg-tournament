import { auth } from '$lib/services/firebase';
import { __idToken } from '$lib/services/auth-utils'; // Assuming this is a helper or we'll get it from auth
import { getIdToken } from 'firebase/auth';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

/**
 * Base API client that handles authentication and requests
 */
export async function apiRequest<T>(path: string, options: RequestInit = {}): Promise<T> {
	const token = await auth.currentUser?.getIdToken();

	const headers = new Headers(options.headers);
	headers.set('Content-Type', 'application/json');
	if (token) {
		headers.set('Authorization', `Bearer ${token}`);
	}

	const response = await fetch(`${API_BASE_URL}${path}`, {
		...options,
		headers
	});

	if (!response.ok) {
		const errorData = await response.json().catch(() => ({}));
		throw new Error(errorData.error || `API Request failed with status ${response.status}`);
	}

	return response.json();
}

/**
 * Tournament API actions
 */
export const tournamentApi = {
	/**
	 * Submit a match result with BO3 scores
	 */
	async submitMatchResult(tournamentId: string, matchId: string, score1: number, score2: number) {
		return apiRequest<any>(`/tournaments/${tournamentId}/matches/${matchId}/result`, {
			method: 'POST',
			body: JSON.stringify({ score1, score2 })
		});
	},

	/**
	 * Trigger the backend to advance the tournament to the next round
	 */
	async advanceTournament() {
		// We will get the tournamentId from the route params in the page
		// but we'll pass it as a parameter for flexibility
	},

	/**
	 * Mark a participant as dropped
	 */
	async dropParticipant(tournamentId: string, uid: string) {
		return apiRequest<any>(`/tournaments/${tournamentId}/participants/${uid}/drop`, {
			method: 'POST'
		});
	}
};

// Correcting the advanceTournament to accept the ID
export async function advanceTournamentApi(tournamentId: string) {
	return apiRequest<any>(`/tournaments/${tournamentId}/advance`, {
		method: 'POST'
	});
}
