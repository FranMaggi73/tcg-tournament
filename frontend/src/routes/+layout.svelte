<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import { authStore } from '$lib/stores/auth.svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import Sidebar from '$lib/components/Sidebar.svelte';

	let { children } = $props();
	let sidebarCollapsed = $state(false);

	// Auth guard: redirect to login if not authenticated and not on landing
	$effect(() => {
		if (!authStore.isLoading && !authStore.user && $page.url.pathname !== '/') {
			goto('/');
		}
	});
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

<div class="min-h-screen bg-base-100 text-base-content font-sans">
	{#if authStore.user}
		<Sidebar bind:collapsed={sidebarCollapsed} />
		<main class="min-h-screen transition-all duration-200 {sidebarCollapsed ? 'ml-16' : 'ml-52'}">
			{@render children()}
		</main>
	{:else}
		{@render children()}
	{/if}
</div>