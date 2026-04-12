import type { Tournament } from '$lib/types/firebase';

export type UserRole = 'judge' | 'viewer';

/**
 * Determines the user's role for a specific tournament.
 * @param userId The current user's UID
 * @param tournament The tournament document
 * @returns 'judge' if the user created the tournament, otherwise 'viewer'
 */
export function getTournamentRole(userId: string | null, tournament: Tournament): UserRole {
	if (userId && tournament.createdBy === userId) {
		return 'judge';
	}
	return 'viewer';
}
