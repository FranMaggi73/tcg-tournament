<script lang="ts">
	import { goto } from '$app/navigation';
	import { tournamentApi } from '$lib/services/api';

	let { onClose } = $props<{ onClose: () => void }>();
	let name = $state('');
	let format = $state<'BO1' | 'BO3'>('BO3');
	let isLoading = $state(false);
	let errorMessage = $state('');

	async function handleCreate() {
		isLoading = true;
		errorMessage = '';
		try {
			if (!name) throw new Error('El nombre del torneo es obligatorio');
			const result = await tournamentApi.createTournament(name, format);
			// Navigate to the manage page for the new tournament
			goto(`/tournaments/${result.id}/manage`);
		} catch (e: any) {
			errorMessage = e.message;
		} finally {
			isLoading = false;
		}
	}
</script>

<div class="modal modal-open">
	<div class="modal-box bg-base-200 border border-base-300">
		<h3 class="text-lg font-bold text-primary mb-4">Crear Nuevo Torneo</h3>

		<div class="form-control w-full mb-4">
			<label class="label" for="tournament-name">
				<span class="label-text">Nombre del Torneo</span>
			</label>
			<input
				id="tournament-name"
				type="text"
				bind:value={name}
				placeholder="Ej. Regional TCG Spring 2026"
				class="input input-bordered w-full"
			/>
		</div>

		<div class="form-control w-full mb-6">
			<label class="label" for="tournament-format">
				<span class="label-text">Formato de Partida</span>
			</label>
			<select
				id="tournament-format"
				bind:value={format}
				class="select select-bordered w-full"
			>
				<option value="BO1">Best of 1 (Sencilla)</option>
				<option value="BO3">Best of 3 (Al mejor de tres)</option>
			</select>
			<p class="text-xs opacity-50 mt-1">
				{format === 'BO1' ? 'Cada partida es decisiva.' : 'Se requieren 2 juegos ganados para ganar la partida.'}
			</p>
		</div>

		{#if errorMessage}
			<p class="text-error text-sm mt-2">{errorMessage}</p>
		{/if}

		<div class="modal-action">
			<button class="btn btn-ghost" onclick={onClose}>Cancelar</button>
			<button
				class="btn btn-primary"
				onclick={handleCreate}
				disabled={isLoading}
			>
				{#if isLoading}
					<span class="loading loading-spinner"></span>
				{/if}
				Crear
			</button>
		</div>
	</div>
</div>
