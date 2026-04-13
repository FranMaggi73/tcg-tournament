<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	// Import auth store to ensure auth observer is initialized
	import '$lib/stores/auth.svelte';

	let { children } = $props();
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

<div class="min-h-screen bg-base-100 text-base-content font-sans">
	{@render children()}
</div>