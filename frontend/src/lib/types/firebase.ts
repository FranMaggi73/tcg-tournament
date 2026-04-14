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
	date: Date;
	status: 'registration' | 'playing' | 'completed';
	currentRound: number;
	totalRounds: number;
	format: 'BO1' | 'BO3';
	inviteCode: string;
}

export interface Player {
	id: string;
	name: string;
	email: string;
	totalScore: number;
	wins: number;
	losses: number;
	draws: number;
	omw: number; // Opponent Match Win %
	gw: number;  // Game Win %
	ogw: number; // Opponent Game Win %
	status: 'active' | 'dropped';
	hadBye: boolean;
}

export interface Round {
	id: string;
	tournamentId: string;
	roundNumber: number;
	status: 'pairing' | 'playing' | 'completed';
	createdAt: Date;
}

export interface Match {
	id: string;
	roundId: string;
	player1Id: string; // UID
	player2Id: string; // UID, "BYE" for byes
	player1Score: number;
	player2Score: number;
	winnerId: string | null; // UID of the winner, empty string for draw
	status: 'scheduled' | 'completed';
}

export interface UserProfile {
	uid: string;
	displayName: string;
	photoURL: string | null;
	bio?: string;
	updatedAt: Date;
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
	type: 'tournament_invite';
	recipientId: string;
	senderId: string;
	tournamentId: string;
	inviteCode: string;
	tournamentName: string;
	message: string;
	read: boolean;
	createdAt: Date;
	expiresAt: Date; // Auto-expires after 7 days
}