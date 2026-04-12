import { auth } from '$lib/services/firebase';
import { signOut } from 'firebase/auth';

export async function logout() {
	await signOut(auth);
}
