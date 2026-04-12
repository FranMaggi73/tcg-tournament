<script lang="ts">
	import { authStore } from '$lib/stores/auth.svelte';
	import LoginForm from '$lib/components/auth/LoginForm.svelte';
	import SignUpForm from '$lib/components/auth/SignUpForm.svelte';

	let authMode = $state<'login' | 'signup'>('login');
</script>

<div class="p-8 flex flex-col items-center justify-center min-h-screen gap-8">
	{#if authStore.isLoading}
		<div class="flex flex-col items-center gap-4">
			<span class="loading loading-ring loading-lg text-primary"></span>
			<p class="text-base-content/60">Cargando sesión...</p>
		</div>
	{:else if !authStore.user}
		<div class="flex flex-col items-center gap-6 w-full">
			<div class="text-center">
				<h1 class="text-4xl font-bold text-primary mb-2">TCG Tournament</h1>
				<p class="text-base-content/70">Gestiona tus torneos de cartas en tiempo real</p>
			</div>

			<div class="tabs tabs-boxed bg-base-200 p-1">
				<a class="tab {authMode === 'login' ? 'tab-active' : ''}" onclick={() => authMode = 'login'}>Iniciar Sesión</a>
				<a class="tab {authMode === 'signup' ? 'tab-active' : ''}" onclick={() => authMode = 'signup'}>Registrarse</a>
			</div>

			{#if authMode === 'login'}
				<LoginForm />
			{:else}
				<SignUpForm />
			{/if}
		</div>
	{:else}
		<div class="card w-full max-w-md bg-base-200 shadow-xl border border-base-300">
			<div class="card-body items-center text-center">
				<div class="avatar mb-4">
					<div class="w-24 rounded-full ring ring-primary ring-offset-base-100 ring-offset-2">
						<img src={authStore.user?.photoURL || 'https://api.dicebear.com/7.x/avataaars/svg?seed=TCG'} alt="User Avatar" />
					</div>
				</div>
				<h2 class="text-2xl font-bold">¡Bienvenido!</h2>
				<p class="text-base-content/70 mb-4">{authStore.user?.email}</p>
				<script lang="ts">
	import { authStore } from '$lib/stores/auth.svelte';
	import LoginForm from '$lib/components/auth/LoginForm.svelte';
	import SignUpForm from '$lib/components/auth/SignUpForm.svelte';

	let authMode = $state<'login' | 'signup'>('login');
</script>

<div class="p-8 flex flex-col items-center justify-center min-h-screen gap-8">
	{#if authStore.isLoading}
		<div class="flex flex-col items-center gap-4">
			<span class="loading loading-ring loading-lg text-primary"></span>
			<p class="text-base-content/60">Cargando sesión...</p>
		</div>
	{:else if !authStore.user}
		<div class="flex flex-col items-center gap-6 w-full">
			<div class="text-center">
				<h1 class="text-4xl font-bold text-primary mb-2">TCG Tournament</h1>
				<p class="text-base-content/70">Gestiona tus torneos de cartas en tiempo real</p>
			</div>

			<div class="tabs tabs-boxed bg-base-200 p-1">
				<a class="tab {authMode === 'login' ? 'tab-active' : ''}" onclick={() => authMode = 'login'}>Iniciar Sesión</a>
				<a class="tab {authMode === 'signup' ? 'tab-active' : ''}" onclick={() => authMode = 'signup'}>Registrarse</a>
			</div>

			{#if authMode === 'login'}
				<LoginForm />
			{:else}
				<SignUpForm />
			{/if}
		</div>
	{:else}
		<div class="card w-full max-w-md bg-base-200 shadow-xl border border-base-300">
			<div class="card-body items-center text-center">
				<div class="avatar mb-4">
					<div class="w-24 rounded-full ring ring-primary ring-offset-base-100 ring-offset-2">
						<img src={authStore.user?.photoURL || 'https://api.dicebear.com/7.x/avataaars/svg?seed=TCG'} alt="User Avatar" />
					</div>
				</div>
				<h2 class="text-2xl font-bold">¡Bienvenido!</h2>
				<p class="text-base-content/70 mb-4">{authStore.user?.email}</p>
				<div class="card-actions w-full gap-2">
					<a href="/tournaments/manage" class="btn btn-primary flex-1">Mis Torneos</a>
					<button class="btn btn-outline btn-secondary flex-1" onclick={() => import('$lib/services/auth-utils').then(m => m.logout())}>Cerrar Sesión</button>
				</div>
			</div>
		</div>
	{/if}
</div>
			</div>
		</div>
	{/if}
</div>
