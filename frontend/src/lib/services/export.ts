/**
 * Generates a CSV string from the tournament standings.
 */
export function generateStandingsCSV(tournament: any, standings: any[]): string {
	const header = 'Rank,Player Name,Points,Wins,Losses,OMW%,GW%,OGW%\n';
	const rows = standings.map(s => {
		const name = s.name || 'Unknown';
		const omwPct = ((s.omw || 0) * 100).toFixed(1);
		const gwPct = ((s.gw || 0) * 100).toFixed(1);
		const ogwPct = ((s.ogw || 0) * 100).toFixed(1);
		return `${s.rank},${name},${s.totalScore},${s.wins || 0},${s.losses || 0},${omwPct}%,${gwPct}%,${ogwPct}%`;
	}).join('\n');

	return `${header}${rows}`;
}

/**
 * Triggers a browser download of a CSV file.
 */
export function downloadCSV(content: string, fileName: string): void {
	const blob = new Blob([content], { type: 'text/csv;charset=utf-8' });
	const url = URL.createObjectURL(blob);
	const link = document.createElement('a');
	link.setAttribute('href', url);
	link.setAttribute('download', fileName);
	link.click();
	URL.revokeObjectURL(url);
}