import { db } from '$lib/services/firebase';
import {
	collection,
	collectionGroup,
	doc,
	updateDoc,
	getDoc,
	getDocs,
	onSnapshot,
	query,
	where
} from 'firebase/firestore';
import type { Tournament, Match, Round, Player } from '$lib/types/firebase';

const TOURNAMENTS_COLLECTION = 'tournaments';

/**
 * Fetches all tournaments created by a specific judge
 */
export async function getTournamentsByJudge(userId: string): Promise<Tournament[]> {
	const q = query(collection(db, TOURNAMENTS_COLLECTION), where('createdBy', '==', userId));
	const querySnapshot = await getDocs(q);

	return querySnapshot.docs.map(doc => ({
		id: doc.id,
		...doc.data()
	} as Tournament));
}

/**
 * Fetches all tournaments where the user is a registered player.
 * Uses a collection group query on 'players' matching by email,
 * then fetches the parent tournament for each.
 */
export async function getTournamentsByPlayer(email: string, userId: string): Promise<Tournament[]> {
	const playersGroup = collectionGroup(db, 'players');
	const q = query(playersGroup, where('email', '==', email));
	const playerSnapshot = await getDocs(q);

	const tournamentIds: string[] = [];
	for (const playerDoc of playerSnapshot.docs) {
		// Extract tournament ID from the path: tournaments/{tournamentId}/players/{playerId}
		const pathSegments = playerDoc.ref.path.split('/');
		const tournamentId = pathSegments[1];
		if (!tournamentIds.includes(tournamentId)) {
			tournamentIds.push(tournamentId);
		}
	}

	// Fetch each tournament document in parallel, excluding own tournaments
	const tournaments = await Promise.all(
		tournamentIds.map(async (tid) => {
			const tDoc = await getDoc(doc(db, TOURNAMENTS_COLLECTION, tid));
			if (tDoc.exists() && tDoc.data().createdBy !== userId) {
				return { id: tDoc.id, ...tDoc.data() } as Tournament;
			}
			return null;
		})
	);

	return tournaments.filter((t): t is Tournament => t !== null);
}

/**
 * Updates tournament data
 */
export async function updateTournament(id: string, data: Partial<Tournament>): Promise<void> {
	const docRef = doc(db, TOURNAMENTS_COLLECTION, id);
	await updateDoc(docRef, data);
}

/**
 * Fetches a tournament document once
 */
export async function getTournament(id: string): Promise<Tournament | null> {
	const docRef = doc(db, TOURNAMENTS_COLLECTION, id);
	const docSnap = await getDoc(docRef);

	if (docSnap.exists()) {
		return { id, ...docSnap.data() } as Tournament;
	}
	return null;
}

/**
 * Subscribes to real-time updates of a tournament.
 * Returns an unsubscribe function to be called in onDestroy.
 */
export function subscribeToTournament(id: string, callback: (tournament: Tournament) => void) {
	const docRef = doc(db, TOURNAMENTS_COLLECTION, id);

	return onSnapshot(docRef, (snapshot) => {
		if (snapshot.exists()) {
			callback({ id, ...snapshot.data() } as Tournament);
		}
	});
}

/**
 * Fetches the rounds for a given tournament from Firestore.
 */
export async function getRounds(tournamentId: string): Promise<Round[]> {
	const roundsCol = collection(db, TOURNAMENTS_COLLECTION, tournamentId, 'rounds');
	const snapshot = await getDocs(roundsCol);
	return snapshot.docs.map(doc => ({
		id: doc.id,
		...doc.data()
	} as Round));
}

/**
 * Finds the round document for a given round number.
 */
export async function findRound(tournamentId: string, roundNumber: number): Promise<Round | null> {
	const roundsCol = collection(db, TOURNAMENTS_COLLECTION, tournamentId, 'rounds');
	const q = query(roundsCol, where('roundNumber', '==', roundNumber));
	const snapshot = await getDocs(q);
	if (snapshot.empty) return null;
	const d = snapshot.docs[0];
	return { id: d.id, ...d.data() } as Round;
}

/**
 * Subscribes to real-time updates of matches for a specific round in a tournament.
 * Uses the round's Firestore document ID (not the round number).
 */
export function subscribeToMatches(tournamentId: string, roundDocId: string, callback: (matches: Match[]) => void) {
	const matchesCol = collection(db, TOURNAMENTS_COLLECTION, tournamentId, 'rounds', roundDocId, 'matches');

	return onSnapshot(matchesCol, (snapshot) => {
		const matches = snapshot.docs.map(doc => ({
			id: doc.id,
			...doc.data()
		} as Match));
		callback(matches);
	});
}

/**
 * Subscribes to real-time updates of players for a tournament.
 */
export function subscribeToPlayers(tournamentId: string, callback: (players: Player[]) => void) {
	const playersCol = collection(db, TOURNAMENTS_COLLECTION, tournamentId, 'players');

	return onSnapshot(playersCol, (snapshot) => {
		const players = snapshot.docs.map(doc => ({
			id: doc.id,
			...doc.data()
		} as Player));
		callback(players);
	});
}