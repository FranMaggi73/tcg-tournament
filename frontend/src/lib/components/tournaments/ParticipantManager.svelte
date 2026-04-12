<script lang="ts">
	import { updateTournament } from '$lib/services/tournament';
	import { tournamentApi } from '$lib/services/api';
	import { resolveUserProfiles } from '$lib/services/user';
	import type { Tournament, UserProfile } from '$lib/types/firebase';

	let { tournament } = $props<{ tournament: Tournament }>();
	let newUserEmail = $state('');
	let isLoading = $state(false);
	let profiles = $state<Record<string, any>>({});

	// Resolve names whenever the participants list changes
	$effect(() => {
		const uids = tournament.participants;
		resolveUserProfiles(uids).then(resolved => {
			profiles = resolved;
		});
	});

	async function addParticipant() {
		isLoading = true;
		try {
			if (!newUserEmail) return;

			// In a real app, we would look up the user by email.
			// For now, we simulate adding a unique ID based on the email
			const mockUid = `user_${btoa(newUserEmail).substring(0, 8)}`;

			const updatedParticipants = [...tournament.participants, mockUid];
			await updateTournament(tournament.id, { participants: updatedParticipants });
			newUserEmail = '';
		} catch (e: unknown) {
			if (e instanceof Error) {
				console.error('Error adding participant:', e.message);
			} else {
				console.error('An unknown error occurred while adding participant:', e);
			}
		} finally {
			isLoading = false;
		}
	}

	async function removeParticipant(uid: string) {
		const updatedParticipants = tournament.participants.filter((p: string) => p !== uid);
		await updateTournament(tournament.id, { participants: updatedParticipants });
	}

	async function handleDropParticipant(uid: string) {
		try {
			await tournamentApi.dropParticipant(tournament.id, uid);
		} catch (e: any) {
			alert(`Error al marcar como drop: ${e.message}`);
		}
	}
</script>

<div class="space-y-4">
	<div class="flex gap-2 mb-6">
		<input
			type="text"
			bind:value={newUserEmail}
			placeholder="Email del participante..."
			class="input input-bordered flex-1"
		/>
		<button
			class="btn btn-primary"
			onclick={addParticipant}
			disabled={isLoading}
		>
			{#if isLoading}
				<span class="loading loading-spinner"></span>
			{/if}
			Añadir
		</button>
	</div>

	<div class="overflow-x-auto">
		<table class="table table-zebra w-full">
			<thead>
				<tr>
					<th>Jugador</th>
					<th class="text-right">Acción</th>
				</tr>
			</thead>
			<tbody>
				{#each tournament.participants as uid}
					<tr class="hover {profiles[uid]?.status === 'dropped' ? 'opacity-50 grayscale' : ''}">
						<td class="flex items-center gap-3">
							<div class="avatar">
								<div class="w-8 h-8 rounded-full overflow-hidden bg-base-300 ring-1 ring-primary">
									{#if profiles[uid]?.photoURL}
										<img src={profiles[uid]?.photoURL} alt="avatar" />
									{:else}
										<div class="w-full h-full flex items-center justify-center text-xs font-bold">
											{profiles[uid]?.displayName?.charAt(0).toUpperCase() || 'U'}
										</div>
									{/if}
								</div>
							</div>
							<div class="flex flex-col">
								<span class="font-medium">{profiles[uid]?.displayName || uid}</span>
								{#if profiles[uid]?.status === 'dropped'}
									<span class="text-[10px] badge badge-ghost">Dropped</span>
								{/if}
							</div>
						</td>
						<td class="text-right flex justify-end gap-2">
							<button
								class="btn btn-outline btn-warning btn-xs"
								onclick={() => handleDropParticipant(uid)}
							>
								Drop
							</button>
							<button
								class="btn btn-ghost btn-xs text-error"
								onclick={() => removeParticipant(uid)}
							>
								Eliminar
							</button>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</div>
