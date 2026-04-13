// Disable SSR — the entire app depends on Firebase Auth (client-side only)
// and Firestore onSnapshot for real-time data.
export const ssr = false;