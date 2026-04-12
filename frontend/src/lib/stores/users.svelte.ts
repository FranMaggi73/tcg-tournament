import type { UserProfile } from '$lib/types/firebase';

interface UserCache {
	profiles: Record<string, UserProfile>;
	isLoading: boolean;
}

export const userCache = $state<UserCache>({
	profiles: {},
	isLoading: false
});

/**
 * Updates the cache with a new profile
 */
export function setCachedProfile(uid: string, profile: UserProfile) {
	userCache.profiles[uid] = profile;
}

/**
 * Gets a profile from the cache or returns null
 */
export function getCachedProfile(uid: string): UserProfile | null {
	return userCache.profiles[uid] || null;
}
