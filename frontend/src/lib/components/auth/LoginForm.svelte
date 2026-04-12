<script lang="ts">
	import { auth } from '$lib/services/firebase';
	import { signInWithEmailAndPassword } from 'firebase/auth';
	import { authStore } from '$lib/stores/auth.svelte';

	let email = $state('');
	let password = $state('');
	let errorMessage = $state('');
	let isLoading = $state(false);

	async function handleLogin() {
		isLoading = true;
		errorMessage = '';
		try {
			await signInWithEmailAndPassword(auth, email, password);
			// The auth observer in layout.svelte will update the authStore
		} catch (e: any) {
			errorMessage = e.message;
		} finally {
			isLoading = false;
		}
	}
</script>

<div class="card w-full max-w-md bg-base-200 shadow-xl border border-base-300">
	<div class="card-body">
		<h2 class="card-title text-primary text-2xl mb-4">Iniciar Sesión</h2>

		<div class="form-control w-full">
			<label class="label">
				<span class="label-text">Correo electrónico</span>
			</label>
			<input
				type="email"
				bind:value={email}
				placeholder="email@ejemplo.com"
				class="input input-bordered w-full"
			/>
		</div>

		<div class="form-control w-full mt-4">
			<label class="label">
				<span class="label-text">Contraseña</span>
			</label>
			<input
				type="password"
				bind:value={password}
				placeholder="••••••••"
				class="input input-bordered w-full"
			/>
		</div>

		{#if errorMessage}
			<div class="alert alert-error mt-4 py-2 text-sm">
				<span>{errorMessage}</span>
			</div>
		{/if}

		<div class="card-actions mt-6">
			<button
				class="btn btn-primary w-full"
				onclick={handleLogin}
				disabled={isLoading}
			>
				{#if isLoading}
					<span class="loading loading-spinner"></span>
				{/if}
				Iniciar Sesión
			</button>
		</div>
	</div>
</div>
