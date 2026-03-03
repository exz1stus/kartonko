import { UserData } from "@/lib/user";
import { apiFetch } from "@/lib/apiFetch";

const userCache = new Map<number, Promise<UserData | null>>(); //TODO add limit
export async function getUserById(id: number): Promise<UserData | null> {
    if (userCache.has(id)) {
        return userCache.get(id) || null;
    }

    const promise = apiFetch(`/user/id/${id}`).then(
        (response) => {
            if (!response.ok) {
                if (response.status === 404) {
                    return null;
                }
                throw new Error(`Failed to fetch user: ${response.statusText}`);
            }

            return response.json();
        },
        (error) => {
            console.error("Error fetching user:", error);
            return null;
        },
    );

    userCache.set(id, promise);
    return promise;
}
