<script lang="ts">
	import { tournamentApi } from '$lib/services/api';
	import { resolveUserProfiles } from '$lib/services/user';
	import type { Match } from '$lib/types/firebase';

	let { tournamentId, roundDocId, matches, format } = $props<{
		tournamentId: string;
		roundDocId: string;
		matches: Match[];
		format: 'BO1' | 'BO3';
	}>();

	let scores = $state<Record<string, { s1: number, s2: number }>>({});
	let loadingMatches = $state<Record<string, boolean>>({});
	let profiles = $state<Record<string, any>>({});

	// Initialize scores and resolve names
	$effect(() => {
		matches.forEach((m: Match) => {
			if (!scores[m.id]) {
				scores[m.id] = { s1: m.player1Score || 0, s2: m.player2Score || 0 };
			}
		});

		const uids = matches.flatMap((m: Match) => {
			const ids = [m.player1Id];
			if (m.player2Id !== 'BYE') ids.push(m.player2Id);
			return ids;
		});
		resolveUserProfiles(uids).then(resolved => {
			profiles = resolved;
		});
	});

	function isValidScore(format: 'BO1' | 'BO3', s1: number, s2: number): boolean {
		if (format === 'BO3') {
			const isVictory = (s1 === 2 && s2 < 2) || (s2 === 2 && s1 < 2);
			const isDraw = s1 === 1 && s2 === 1;
			return isVictory || isDraw;
		} else {
			return (s1 === 1 && s2 === 0) || (s2 === 1 && s1 === 0);
		}
	}

	async function handleSetResult(match: Match) {
		const { s1, s2 } = scores[match.id];
		if (!isValidScore(format, s1, s2)) return;

		loadingMatches[match.id] = true;
		try {
			await tournamentApi.submitMatchResult(tournamentId, match.id, match.roundId, s1, s2);
		} catch (e: any) {
			alert(`Error: ${e.message}`);
		} finally {
			loadingMatches[match.id] = false;
		}
	}
</script>

<div class="space-y-6">
	<div class="flex justify-between items-center mb-4">
		<h3 class="text-lg font-semibold">Pairings - Round</h3>
		<div class="badge badge-outline">
			{matches.filter((m: Match) => m.status === 'completed').length} / {matches.length} Completadas
		</div>
	</div>

	<div class="grid grid-cols-1 gap-4">
		{#each matches as match}
			<div class="card bg-base-300 shadow-sm border border-base-300 p-4 flex flex-col gap-4">
				<div class="flex justify-between items-center">
					<span class="text-xs font-mono text-base-content/50">Match {match.id.substring(0, 5)}</span>
					{#if match.status === 'completed'}
						<div class="badge badge-success text-xs">Completed</div>
					{:else}
						<div class="badge badge-ghost text-xs">Pending</div>
					{/if}
				</div>

				{#if match.player2Id === 'BYE'}
					<!-- Bye match -->
					<div class="text-center py-2">
						<span class="font-bold text-lg">{profiles[match.player1Id]?.displayName || match.player1Id}</span>
						<span class="badge badge-info ml-2">BYE</span>
					</div>
				{:else}
					<div class="flex items-center justify-between gap-4">
						<!-- Player 1 -->
						<div class="flex flex-col gap-1 flex-1">
							<span class="font-bold text-lg">{profiles[match.player1Id]?.displayName || match.player1Id}</span>
							<div class="flex items-center gap-2">
								<select
									class="select select-bordered h-8 w-16 text-center text-sm"
									bind:value={scores[match.id].s1}
									disabled={match.status === 'completed'}
								>
									{#if format === 'BO3'}
										<option value={0}>0</option>
										<option value={1}>1</option>
										<option value={2}>2</option>
									{:else}
										<option value={0}>0</option>
										<option value={1}>1</option>
									{/if}
								</select>
								<span class="text-xs opacity-50">pts</span>
							</div>
						</div>

						<div class="text-xl font-bold opacity-30">VS</div>

						<!-- Player 2 -->
						<div class="flex flex-col gap-1 flex-1 items-end">
							<span class="font-bold text-lg">{profiles[match.player2Id]?.displayName || match.player2Id}</span>
							<div class="flex items-center gap-2 justify-end">
								<select
									class="select select-bordered h-8 w-16 text-center text-sm"
									bind:value={scores[match.id].s2}
									disabled={match.status === 'completed'}
								>
									{#if format === 'BO3'}
										<option value={0}>0</option>
										<option value={1}>1</option>
										<option value={2}>2</option>
									{:else}
										<option value={0}>0</option>
										<option value={1}>1</option>
									{/if}
								</select>
								<span class="text-xs opacity-50">pts</span>
							</div>
						</div>
					</div>

					<div class="flex justify-center mt-2">
						<button
							class="btn btn-primary btn-xs"
							disabled={match.status === 'completed' || loadingMatches[match.id] || !isValidScore(format, scores[match.id].s1, scores[match.id].s2)}
							onclick={() => handleSetResult(match)}
						>
							{#if loadingMatches[match.id]}
								<span class="loading loading-spinner"></span>
							{/if}
							Guardar Resultado
						</button>
					</div>
				{/if}
			</div>
		{/each}

		{#if matches.length === 0}
			<div class="text-center py-10 text-base-content/50">
				No pairings available for this round.
			</div>
		{/if}
	</div>
</div>