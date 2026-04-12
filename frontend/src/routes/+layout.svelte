<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { initAuthObserver } from '$lib/stores/auth.svelte';

	let { children } = $props();

	// Initialize Firebase Auth observer on app start
	initAuthObserver();
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

<div class="min-h-screen bg-base-100 text-base-content font-sans">
	{@render children()}
</div>
