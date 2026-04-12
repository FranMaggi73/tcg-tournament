export interface FirebaseUser {
	uid: string;
	email: string | null;
	displayName: string | null;
	photoURL: string | null;
}

export interface Tournament {
	id: string;
	name: string;
	createdBy: string; // UID of the judge
	createdAt: Date;
	status: 'pending' | 'ongoing' | 'finished';
	currentRound: number;
	participants: string[]; // Array of user UIDs
}

export interface Match {
	id: string;
	player1: string; // UID
	player2: string; // UID
	winner: string | null; // UID of the winner
	status: 'pending' | 'completed';
	updatedAt: Date;
}

export interface Round {
	number: number;
	matches: Match[];
}

export interface UserProfile {
	uid: string;
	displayName: string;
	photoURL: string | null;
	bio?: string;
	updatedAt: Date;
}

export interface StandingEntry {
	rank: number;
	uid: string;
	points: number;
	omw: number; // Opponent Match Win %
	gw: number;  // Game Win %
	ogw: number; // Opponent Game Win %
}
