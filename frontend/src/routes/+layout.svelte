<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	// Import auth store to ensure auth observer is initialized
	import { authStore } from '$lib/stores/auth.svelte';

	let { children } = $props();
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

<div class="min-h-screen bg-base-100 text-base-content font-sans">
	{#if authStore.user}
		<div class="p-4 flex justify-end">
			<a href="/profile" class="btn btn-ghost btn-circle avatar">
				<div class="w-10 rounded-full ring ring-primary ring-offset-base-100 ring-offset-1">
					{#if authStore.user.photoURL}
						<img src={authStore.user.photoURL} alt="User profile" />
					{:else}
						<div class="bg-base-300 w-full h-full flex items-center justify-center text-sm font-bold text-primary">
							{authStore.user.displayName?.charAt(0).toUpperCase() || 'U'}
						</div>
					{/if}
				</div>
			</a>
		</div>
	{/if}
	{@render children()}
</div>