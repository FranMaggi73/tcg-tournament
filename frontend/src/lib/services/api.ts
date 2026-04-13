import { auth } from '$lib/services/firebase';

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
	 * Submit a match result
	 */
	async submitMatchResult(tournamentId: string, matchId: string, roundId: string, score1: number, score2: number) {
		return apiRequest<any>(`/tournaments/${tournamentId}/matches/${matchId}`, {
			method: 'PATCH',
			body: JSON.stringify({
				player1Score: score1,
				player2Score: score2,
				roundId
			})
		});
	},

	/**
	 * Join a tournament using an invite code
	 */
	async joinByCode(code: string, email: string, name: string) {
		return apiRequest<any>('/tournaments/join', {
			method: 'POST',
			body: JSON.stringify({ code, email, name })
		});
	},

	/**
	 * Get standings for a tournament
	 */
	async getStandings(tournamentId: string) {
		return apiRequest<any[]>(`/tournaments/${tournamentId}/standings`);
	},

	/**
	 * Drop a participant from a tournament
	 */
	async dropParticipant(tournamentId: string, playerId: string) {
		return apiRequest<any>(`/tournaments/${tournamentId}/players/${playerId}/status`, {
			method: 'PATCH',
			body: JSON.stringify({ status: 'dropped' })
		});
	},

	/**
	 * Delete a tournament (only allowed in registration status)
	 */
	async deleteTournament(tournamentId: string) {
		return apiRequest<any>(`/tournaments/${tournamentId}`, {
			method: 'DELETE'
		});
	}
};

/**
 * Trigger the backend to advance the tournament to the next round
 */
export async function advanceTournamentApi(tournamentId: string) {
	return apiRequest<any>(`/tournaments/${tournamentId}/rounds/next`, {
		method: 'POST'
	});
}

/**
 * Friendship API actions
 */
export const friendshipApi = {
	async addFriend(friendId: string) {
		return apiRequest<any>('/friends', {
			method: 'POST',
			body: JSON.stringify({ friendId })
		});
	},

	async getFriends() {
		return apiRequest<any[]>('/friends', {
			method: 'GET'
		});
	},

	async updateStatus(friendshipId: string, status: 'accepted' | 'declined') {
		return apiRequest<any>(`/friends/${friendshipId}`, {
			method: 'PATCH',
			body: JSON.stringify({ status })
		});
	}
};