<script lang="ts">
	import { page } from '$app/stores';
	import { authStore } from '$lib/stores/auth.svelte';
	import { userCache } from '$lib/stores/users.svelte';
	import { logout } from '$lib/services/auth-utils';
	import { onMount } from 'svelte';
	import { getUserProfile } from '$lib/services/user';

	let { collapsed = $bindable(false) } = $props();

	// Resolve current user's profile from Firestore cache
	let currentProfile = $derived(userCache.profiles[authStore.user?.uid ?? ''] ?? null);

	// Load profile on mount if not cached
	onMount(async () => {
		if (authStore.user?.uid && !currentProfile) {
			await getUserProfile(authStore.user.uid);
		}
	});

	const navItems = [
		{ href: '/', label: 'Inicio', icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-4 0a1 1 0 01-1-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 01-1 1' },
		{ href: '/tournaments/manage', label: 'Torneos', icon: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10' },
		{ href: '/profile', label: 'Perfil', icon: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z' }
	];

	function isActive(href: string, pathname: string): boolean {
		if (href === '/') return pathname === '/';
		return pathname.startsWith(href);
	}
</script>

<aside class="fixed top-0 left-0 h-screen bg-base-200 border-r border-base-300 z-40 flex flex-col transition-all duration-200 {collapsed ? 'w-16' : 'w-52'}">
	<!-- Header: Logo + toggle -->
	<div class="flex items-center gap-3 px-4 py-4 border-b border-base-300">
		<span class="text-2xl shrink-0">🏆</span>
		{#if !collapsed}
			<span class="text-lg font-bold text-primary whitespace-nowrap">TCG Tournament</span>
		{/if}
		<button
			class="ml-auto p-1 rounded-lg hover:bg-base-300 text-base-content/50 hover:text-base-content transition-colors shrink-0"
			onclick={() => collapsed = !collapsed}
			aria-label={collapsed ? 'Expandir sidebar' : 'Colapsar sidebar'}
		>
			<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 transition-transform {collapsed ? 'rotate-180' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
				<path stroke-linecap="round" stroke-linejoin="round" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
			</svg>
		</button>
	</div>

	<!-- Navigation -->
	<nav class="flex-1 py-4 space-y-1 px-2">
		{#each navItems as item}
			<a
				href={item.href}
				class="flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors {isActive(item.href, $page.url.pathname) ? 'bg-primary text-primary-content' : 'text-base-content/70 hover:bg-base-300 hover:text-base-content'}"
				aria-current={isActive(item.href, $page.url.pathname) ? 'page' : undefined}
				title={collapsed ? item.label : undefined}
			>
				<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d={item.icon} />
				</svg>
				{#if !collapsed}
					<span class="text-sm font-medium">{item.label}</span>
				{/if}
			</a>
		{/each}
	</nav>

	<!-- User section -->
	<div class="border-t border-base-300 p-3">
		<div class="flex items-center gap-3">
			<div class="avatar placeholder shrink-0">
				{#if currentProfile?.photoURL}
					<div class="w-9 rounded-full overflow-hidden ring ring-primary ring-offset-base-100 ring-offset-1">
						<img src={currentProfile.photoURL} alt="Avatar" class="w-full h-full object-cover" />
					</div>
				{:else}
					<div class="bg-neutral text-neutral-content rounded-full w-9 h-9 flex items-center justify-center">
						<span class="text-xs font-bold">{currentProfile?.displayName?.charAt(0).toUpperCase() || authStore.user?.email?.charAt(0).toUpperCase() || 'U'}</span>
					</div>
				{/if}
			</div>
			{#if !collapsed}
				<div class="flex-1 min-w-0">
					<p class="text-sm font-medium truncate">{currentProfile?.displayName || authStore.user?.displayName || 'Jugador'}</p>
					<p class="text-xs opacity-50 truncate">{authStore.user?.email}</p>
				</div>
			{/if}
		</div>
		<button
			class="flex items-center gap-3 w-full px-3 py-2 mt-2 rounded-lg text-sm text-error/70 hover:bg-error/10 hover:text-error transition-colors"
			onclick={() => logout()}
			title={collapsed ? 'Salir' : undefined}
		>
			<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
				<path stroke-linecap="round" stroke-linejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
			</svg>
			{#if !collapsed}
				<span>Salir</span>
			{/if}
		</button>
	</div>
</aside>