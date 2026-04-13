import { redirect } from '@sveltejs/kit';
import { authStore, waitForAuth } from '$lib/stores/auth.svelte';
import { getTournament } from '$lib/services/tournament';
import { getTournamentRole } from '$lib/services/roles';

export async function load({ params }) {
	const { id } = params;

	// Wait for auth state to be ready
	await waitForAuth();

	// Role check: must be logged in
	if (!authStore.user) {
		throw redirect(302, '/');
	}

	const tournament = await getTournament(id);
	if (!tournament) {
		throw redirect(302, '/tournaments');
	}

	// Role check: must be the judge
	const role = getTournamentRole(authStore.user.uid, tournament);
	if (role !== 'judge') {
		throw redirect(302, `/tournaments/${id}`);
	}

	return {
		tournamentId: id
	};
}
