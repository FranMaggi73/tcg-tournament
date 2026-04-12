import { getTournament } from '$lib/services/tournament';
import { resolveUserProfiles } from '$lib/services/user';

/**
 * Generates a CSV string from the tournament standings.
 * @param tournament The tournament data.
 * @param standings The standings data provided by the backend.
 */
export function generateStandingsCSV(tournament: any, standings: any[]) {
	const header = 'Rank,Player Name,Points,OMW%,GW%,OGW%\n';
	const rows = standings.map(s => {
		// Resolve the player name if possible, otherwise use UID
		const name = s.displayName || s.userId || 'Unknown';
		return `${s.rank},${name},${s.points},${s.omw || 0}%,${s.gw || 0}%,${s.ogw || 0}%`;
	}).join('\n');

	return `${header}${rows}`;
}

/**
 * Triggers a browser download of a CSV file.
 */
export function downloadCSV(content: string, fileName: string) {
	const blob = new Blob([content], { type: 'text/csv;charset=utf-8;' });
	const url = URL.createObjectURL(blob);
	const link = document.createElement('a');
	link.setAttribute('href', url);
	link.setAttribute('download', fileName);
	link.click();
	URL.revokeObjectURL(url);
}
