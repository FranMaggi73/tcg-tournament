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
	status: 'registration' | 'playing' | 'completed';
	currentRound: number;
	totalRounds: number;
	participants: string[]; // Array of user UIDs
	format: 'BO1' | 'BO3';
	inviteCode: string;
}

export interface Match {
	id: string;
	roundId: string;
	player1: string; // UID
	player2: string; // UID
	player1Score: number;
	player2Score: number;
	winner: string | null; // UID of the winner
	status: 'scheduled' | 'completed';
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

export interface Friendship {
	id: string;
	user1Id: string;
	user2Id: string;
	status: 'pending' | 'accepted' | 'declined';
	createdAt: Date;
}

export interface Notification {
	id: string;
	recipientId: string;
	senderId: string;
	tournamentId: string;
	inviteCode: string;
	tournamentName: string;
	message: string;
	read: boolean;
	createdAt: Date;
}
