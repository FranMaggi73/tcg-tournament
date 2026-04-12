<script lang="ts">
	import { auth } from '$lib/services/firebase';
	import { createUserWithEmailAndPassword } from 'firebase/auth';

	let email = $state('');
	let password = $state('');
	let errorMessage = $state('');
	let isLoading = $state(false);

	async function handleSignUp() {
		isLoading = true;
		errorMessage = '';
		try {
			await createUserWithEmailAndPassword(auth, email, password);
		} catch (e: any) {
			if (e.code === 'auth/email-already-in-use') {
				errorMessage = 'Este correo electrónico ya está registrado. Por favor, inicia sesión.';
			} else {
				errorMessage = e.message;
			}
		} finally {
			isLoading = false;
		}
	}
</script>

<div class="card w-full max-w-md bg-base-200 shadow-xl border border-base-300">
	<div class="card-body">
		<h2 class="card-title text-primary text-2xl mb-4">Crear Cuenta</h2>

		<div class="form-control w-full">
			<label class="label" for="signup-email">
				<span class="label-text">Correo electrónico</span>
			</label>
			<input
				id="signup-email"
				type="email"
				bind:value={email}
				placeholder="email@ejemplo.com"
				class="input input-bordered w-full"
			/>
		</div>

		<div class="form-control w-full mt-4">
			<label class="label" for="signup-password">
				<span class="label-text">Contraseña</span>
				<span class="label-text-alt text-xs opacity-50">Mínimo 6 caracteres.</span>
			</label>
			<input
				id="signup-password"
				type="password"
				bind:value={password}
				placeholder="••••••••"
				class="input input-bordered w-full"
			/>
		</div>

		{#if errorMessage}
			<div class="alert alert-error py-2 text-sm mt-4">
				<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-4 w-4" fill="none" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 28l7.5-7.5L4 4M13.2-5.8 17.5 0//L-20" />
				</svg>
				<span>{errorMessage}</span>
			</div>
		{/if}

		<div class="card-actions mt-6">
			<button
				class="btn btn-primary w-full"
				onclick={handleSignUp}
				disabled={isLoading}
			>
				{#if isLoading}
					<span class="loading loading-spinner"></span>
				{/if}
				Registrarse
			</button>
		</div>
	</div>
</div>
