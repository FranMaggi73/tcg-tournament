<script lang="ts">
	import { tournamentApi } from '$lib/services/api';
	import { resolveUserProfiles } from '$lib/services/user';
	import { generateStandingsCSV, downloadCSV } from '$lib/services/export';
	import type { Player } from '$lib/types/firebase';

	let { tournamentId, tournamentName = 'Tournament' } = $props<{
		tournamentId: string;
		tournamentName?: string;
	}>();
	let players = $state<Player[]>([]);
	let profiles = $state<Record<string, any>>({});
	let isLoading = $state(true);
	let errorMessage = $state('');

	async function fetchStandings() {
		isLoading = true;
		errorMessage = '';
		try {
			const data = await tournamentApi.getStandings(tournamentId);
			players = data;
			const uids = data.map((p: Player) => p.id);
			const resolved = await resolveUserProfiles(uids);
			profiles = resolved;
		} catch (e: any) {
			errorMessage = e.message;
		} finally {
			isLoading = false;
		}
	}

	$effect(() => {
		fetchStandings();
	});

	function handleExport() {
		const csv = generateStandingsCSV({ id: tournamentId, name: tournamentName } as any, players.map((p, i) => ({
			rank: i + 1,
			name: profiles[p.id]?.displayName || p.name,
			totalScore: p.totalScore,
			omw: p.omw,
			gw: p.gw,
			ogw: p.ogw
		})));
		downloadCSV(csv, `standings_${tournamentId}.csv`);
	}

	// Sort players by score > OMW > GW > OGW
	let sortedPlayers = $derived([...players].sort((a, b) => {
		if (a.totalScore !== b.totalScore) return b.totalScore - a.totalScore;
		if (a.omw !== b.omw) return b.omw - a.omw;
		if (a.gw !== b.gw) return b.gw - a.gw;
		return b.ogw - a.ogw;
	}));
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
						<th>#</th>
						<th>Jugador</th>
						<th class="text-center">Pts</th>
						<th class="text-center">W</th>
						<th class="text-center">L</th>
						<th class="text-center">OMW%</th>
						<th class="text-center">GW%</th>
						<th class="text-center">OGW%</th>
					</tr>
				</thead>
				<tbody>
					{#each sortedPlayers as p, index}
						<tr class="hover transition-colors {p.status === 'dropped' ? 'opacity-50' : ''}">
							<td class="font-bold text-center">
								{#if index === 0}
									<div class="badge badge-primary gap-2">1</div>
								{:else if index === 1}
									<div class="badge badge-secondary gap-2">2</div>
								{:else if index === 2}
									<div class="badge badge-accent gap-2">3</div>
								{:else}
									{index + 1}
								{/if}
							</td>
							<td>
								<div class="flex items-center gap-3">
									<div class="avatar">
										<div class="w-8 h-8 rounded-full overflow-hidden bg-base-300 ring-1 ring-primary">
											{#if profiles[p.id]?.photoURL}
												<img src={profiles[p.id]?.photoURL} alt="avatar" />
											{:else}
												<div class="w-full h-full flex items-center justify-center text-xs font-bold">
													{profiles[p.id]?.displayName?.charAt(0).toUpperCase() || p.name.charAt(0).toUpperCase()}
												</div>
											{/if}
										</div>
									</div>
									<span class="font-medium">{profiles[p.id]?.displayName || p.name}</span>
									{#if p.status === 'dropped'}
										<span class="badge badge-ghost badge-xs">Dropped</span>
									{/if}
								</div>
							</td>
							<td class="text-center font-bold">{p.totalScore}</td>
							<td class="text-center">{p.wins}</td>
							<td class="text-center">{p.losses}</td>
							<td class="text-center">{(p.omw * 100).toFixed(1)}%</td>
							<td class="text-center">{(p.gw * 100).toFixed(1)}%</td>
							<td class="text-center">{(p.ogw * 100).toFixed(1)}%</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>