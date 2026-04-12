<script lang="ts">
	import { createTournament } from '$lib/services/tournament';
	import { authStore } from '$lib/stores/auth.svelte';

	let { onClose } = $props<{ onClose: () => void }>();
	let name = $state('');
	let isLoading = $state(false);
	let errorMessage = $state('');

	async function handleCreate() {
		isLoading = true;
		errorMessage = '';
		try {
			if (!name) throw new Error('El nombre del torneo es obligatorio');
			await createTournament(name, authStore.user!.uid);
			onClose();
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
		<div class="form-control w-full">
			<label class="label">
				<span class="label-text">Nombre del Torneo</span>
			</label>
			<input
				type="text"
				bind:value={name}
				placeholder="Ej. Regional TCG Spring 2026"
				class="input input-bordered w-full"
			/>
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
