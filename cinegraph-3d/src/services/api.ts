// src/services/api.ts

// Use a safe way to check for Vite env variables
const getApiUrl = () => {
    // Priority 1: Environment variable
    if ((import.meta as any).env?.VITE_API_URL) return (import.meta as any).env.VITE_API_URL;

    // Priority 2: Current origin (useful if hosted on different port but same domain)
    if (typeof window !== 'undefined') {
        const port = window.location.port; // 3000
        const host = window.location.hostname; // localhost or 127.0.0.1
        // If we are on 3000, backend is likely on 4000
        if (port === '3000') return `http://${host}:4000/v1`;
    }

    return 'http://localhost:4000/v1';
};

const API_BASE_URL = getApiUrl();
console.log('%c CineGraph API Base URL:', 'color: #d4a359; font-weight: bold', API_BASE_URL);

let accessToken: string | null = null;

// --- Types ---
export type NodeType = 'Movie' | 'Person' | 'Genre';

export interface Node {
    id: string; // string needed for d3/graph visualization logic
    name: string;
    type: NodeType;
    val: number;
    color?: string;
    image?: string;
    description?: string;
    year?: number;
    rating?: number;
    runtime?: string;
    certificate?: string;
    x?: number;
    y?: number;
    z?: number;
}

export interface Link {
    source: string;
    target: string;
    label: string;
}

export interface GraphData {
    nodes: Node[];
    links: Link[];
}

export interface BackendFilm {
    id: number;
    imdb_id: string;
    title: string;
    year: number;
    runtime: string;
    certificate: string;
    rating: number;
    description: string;
    genres: string[];
    directors: string[];
    actors: string[];
    image: string;
}

export interface WatchlistEntry {
    id: string; // backend might use int, converting to string for frontend
    film_id: string;
    notes: string;
    priority: number;
    watched: boolean;
    rating?: number;
    film: Node;
}

// --- Utils ---
let loginPromise: Promise<void> | null = null;

async function performAutoLogin() {
    const email = 'demo@cinegraph.com';
    const pass = 'demo12345';
    try {
        console.log('API: Attempting login...');
        const data = await api.login(email, pass);
        console.log('API: Login successful. Activated:', data.user?.activated);
        if (data.user && !data.user.activated) {
            throw new Error('Your demo account exists but is NOT activated. Please run the Cypher activation command provided by the assistant.');
        }
    } catch (err: any) {
        if (err.message?.includes('activated')) throw err;

        try {
            console.log('API: Login failed, attempting registration...');
            await api.register(email, pass, 'Demo User');
            console.log('API: Registration + Activation successful. Logging in...');
            await api.login(email, pass);
        } catch (regErr: any) {
            console.error('API: Auto-login sequence failed completely:', regErr);
            throw new Error(`Authentication failed. Detailed error: ${regErr.message || 'Unknown'}`);
        }
    }
}

async function fetchWithAuth(endpoint: string, options: RequestInit = {}) {
    // Simple auto-login for demo purposes if no token exists
    if (!accessToken) {
        if (!loginPromise) {
            loginPromise = performAutoLogin();
        }
        await loginPromise;
    }

    const headers = {
        'Content-Type': 'application/json',
        ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
        ...options.headers,
    };

    const response = await fetch(`${API_BASE_URL}${endpoint}`, { ...options, headers });

    if (!response.ok) {
        const errorText = await response.text().catch(() => 'No error body');
        console.error(`API Fetch Error [${endpoint}]:`, response.status, errorText);
        throw new Error(`API Error: ${response.status} ${response.statusText}`);
    }

    const data = await response.json();
    console.log(`API Success [${endpoint}]:`, data);
    return data;
}

/** 
 * Maps a list of backend films into the Node and Link graph structure.
 */
function buildGraphData(films: BackendFilm[]): GraphData {
    const nodesMap = new Map<string, Node>();
    const links: Link[] = [];

    for (const film of films) {
        const filmNodeId = `m${film.id}`;

        // Movie Node
        if (!nodesMap.has(filmNodeId)) {
            nodesMap.set(filmNodeId, {
                id: filmNodeId,
                name: film.title,
                type: 'Movie',
                val: 2, // Default size
                year: film.year,
                rating: film.rating,
                runtime: film.runtime,
                certificate: film.certificate,
                description: film.description,
                image: film.image
            });
        }

        // Genre Nodes & Links
        for (const genre of film.genres || []) {
            const genreId = `g_${genre.replace(/\s+/g, '_')}`;
            if (!nodesMap.has(genreId)) {
                nodesMap.set(genreId, { id: genreId, name: genre, type: 'Genre', val: 5 });
            }
            links.push({ source: filmNodeId, target: genreId, label: 'IN_GENRE' });
        }

        // Director Nodes & Links
        for (const director of film.directors || []) {
            const dirId = `p_${director.replace(/\s+/g, '_')}`;
            if (!nodesMap.has(dirId)) {
                nodesMap.set(dirId, { id: dirId, name: director, type: 'Person', val: 3 });
            }
            links.push({ source: dirId, target: filmNodeId, label: 'DIRECTED' });
        }

        // Actor Nodes & Links
        for (const actor of film.actors || []) {
            const actId = `p_${actor.replace(/\s+/g, '_')}`;
            if (!nodesMap.has(actId)) {
                nodesMap.set(actId, { id: actId, name: actor, type: 'Person', val: 2.5 });
            }
            links.push({ source: actId, target: filmNodeId, label: 'ACTED_IN' });
        }
    }

    return {
        nodes: Array.from(nodesMap.values()),
        links
    };
}

// --- API Methods ---
export const api = {
    register: async (email: string, password: string, name: string) => {
        // 1. Register
        const res = await fetch(`${API_BASE_URL}/users`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, email, password })
        });
        if (!res.ok) throw new Error('Registration failed');
        const data = await res.json();

        // 2. Activate
        if (data.activation_token?.token) {
            await fetch(`${API_BASE_URL}/users/activate`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ token: data.activation_token.token })
            });
        }
    },

    login: async (email: string, password: string) => {
        const res = await fetch(`${API_BASE_URL}/tokens/authentication`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });
        if (!res.ok) throw new Error('Login failed');
        const data = await res.json();
        accessToken = data.access_token;
        return data;
    },

    getFilmsAsGraph: async (pageSize: number = 200): Promise<GraphData> => {
        const data = await fetchWithAuth(`/films?page=1&page_size=${pageSize}&sort=-rating`);
        return buildGraphData(data.films || []);
    },

    getWatchlist: async (): Promise<WatchlistEntry[]> => {
        const data = await fetchWithAuth(`/watchlist?sort=-priority`);
        const entries = data.watchlist || [];
        return entries.map((w: any) => ({
            ...w,
            id: w.id.toString(),
            film_id: `m${w.film_id}`,
            // Inject Node equivalent for frontend to display
            film: {
                id: `m${w.film.id}`,
                name: w.film.title,
                type: 'Movie',
                val: 2,
                year: w.film.year,
                rating: w.film.rating,
                runtime: w.film.runtime,
                certificate: w.film.certificate,
                description: w.film.description,
                image: w.film.image
            } as Node
        }));
    },

    addToWatchlist: async (film_id: string, notes: string, priority: number) => {
        // film_id looks like "m123". Need to strip 'm'.
        const numericId = parseInt(film_id.replace('m', ''), 10);
        const data = await fetchWithAuth(`/watchlist`, {
            method: 'POST',
            body: JSON.stringify({ film_id: numericId, notes, priority })
        });
        return data;
    },

    rateFilm: async (film_id: string, rating: number, notes: string, priority: number) => {
        const numericId = parseInt(film_id.replace('m', ''), 10);

        // Check if on watchlist first, unfortunately backend rating expects an existing watchlist if we use PATCH...
        // Actually the backend README says: /v1/watchlist POST creates and adding rating automatically marks watched.
        await fetchWithAuth(`/watchlist`, {
            method: 'POST',
            body: JSON.stringify({ film_id: numericId, rating, notes, priority })
        });
    },

    removeFromWatchlist: async (film_id: string, watchlist_id: string) => {
        await fetchWithAuth(`/watchlist/${watchlist_id}`, {
            method: 'DELETE'
        });
    },

    getRecommendations: async (): Promise<Node[]> => {
        const data = await fetchWithAuth(`/recommendations?limit=10`);
        const recs = data.recommendations || [];
        return recs.map((film: BackendFilm) => ({
            id: `m${film.id}`,
            name: film.title,
            type: 'Movie',
            val: 2.5,
            year: film.year,
            rating: film.rating,
            runtime: film.runtime,
            certificate: film.certificate,
            description: film.description,
            image: film.image
        }));
    }
};
