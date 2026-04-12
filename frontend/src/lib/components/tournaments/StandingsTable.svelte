<script lang="ts">
	import { tournamentApi } from '$lib/services/api';
	import { resolveUserProfiles } from '$lib/services/user';
	import { generateStandingsCSV, downloadCSV } from '$lib/services/export';
	import type { StandingEntry } from '$lib/types/firebase';

	let { tournamentId } = $props<{ tournamentId: string }>();
	let standings = $state<StandingEntry[]>([]);
	let profiles = $state<Record<string, any>>({});
	let isLoading = $state(true);
	let errorMessage = $state('');

	async function fetchStandings() {
		isLoading = true;
		errorMessage = '';
		try {
			const data = await tournamentApi.getStandings(tournamentId);
			standings = data;

			// Resolve profiles for the standings list
			const uids = standings.map(s => s.uid);
			const resolved = await resolveUserProfiles(uids);
			profiles = resolved;
		} catch (e: any) {
			errorMessage = e.message;
		} finally {
			isLoading = false;
		}
	}

	// Initial fetch and update when tournamentId changes
	$effect(() => {
		fetchStandings();
	});

	function handleExport() {
		// We need the tournament name for the filename,
		// so we'll fetch it or pass it in.
		// For now, we'll use the ID.
		const csv = generateStandingsCSV({ id: tournamentId, name: 'Tournament' } as any, standings);
		downloadCSV(csv, `standings_${tournamentId}.csv`);
	}
</script>

<div class="space-y-6">
	<div class="flex justify-between items-center mb-4">
		<h3 class="text-xl font-bold text-primary">Clasificación Profesional</h3>
		<button class="btn btn-outline btn-sm" onclick={handleExport}>
			<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
			</svg>
			Exportar CSV
		</button>
	</div>

	{#if isLoading}
		<div class="space-y-4">
			{#each Array(5) as _}
				<div class="flex items-center gap-4 p-4 bg-base-300 rounded-lg border border-base-300">
					<div class="skeleton h-8 w-8 rounded-full"></div>
					<div class="skeleton h-4 w-32"></div>
					<div class="skeleton h-4 w-12 ml-auto"></div>
					<div class="skeleton h-4 w-12"></div>
					<div class="skeleton h-4 w-12"></div>
					<div class="skeleton h-4 w-12"></div>
				</div>
			{/each}
		</div>
	{:else if errorMessage}
		<div class="alert alert-error">
			<span>{errorMessage}</span>
		</div>
	{:else}
		<div class="overflow-x-auto">
			<table class="table table-zebra w-full">
				<thead>
					<tr class="text-base-content/60">
						<th>Rango</th>
						<th>Jugador</th>
						<th class="text-center">Pts</th>
						<th class="text-center">OMW%</th>
						<th class="text-center">GW%</th>
						<th class="text-center">OGW%</th>
					</tr>
				</thead>
				<tbody>
					{#each standings as entry, index}
						<tr class="hover transition-colors">
							<td class="font-bold text-center">
								{#if index === 0}
									<div class="badge badge-primary gap-2">🥇 {entry.rank}</div>
								{:else if index === 1}
									<div class="badge badge-secondary gap-2">🥈 {entry.rank}</div>
								{:else if index === 2}
									<div class="badge badge-accent gap-2">🥉 {entry.rank}</div>
								{:else}
									{entry.rank}
								{/if}
							</td>
							<td>
								<div class="flex items-center gap-3">
									<div class="avatar">
										<div class="w-8 h-8 rounded-full overflow-hidden bg-base-300 ring-1 ring-primary">
											{#if profiles[entry.uid]?.photoURL}
												<img src={profiles[entry.uid]?.photoURL} alt="avatar" />
											{:else}
												<div class="w-full h-full flex items-center justify-center text-xs font-bold">
													{profiles[entry.uid]?.displayName?.charAt(0).toUpperCase() || 'U'}
 l la la
										</div>
									</div>
								</div>
							</div>
							<span class="font-medium">{profiles[entry.uid]?.displayName || entry.uid}</span>
							</td>
							<td class="text-center font-bold">{entry.points}</td>
							<td class="text-center">{entry.omw}%</td>
							<td class="text-center">{entry.gw}%</td>
							<td class="text-center">{entry.ogw}%</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
