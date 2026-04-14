import { auth } from '$lib/services/firebase';
import { signOut } from 'firebase/auth';
import { goto } from '$app/navigation';

export async function logout() {
	await signOut(auth);
	goto('/');
}