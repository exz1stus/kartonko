import { UserData } from "@/lib/user";
import { serverFetch } from "@/lib/serverFetch";

export async function getUserByIdServer(id: number): Promise<UserData | null> {
    try {
        const response = await serverFetch(`/user/id/${id}`);

        if (!response.ok) {
            if (response.status === 404) {
                return null;
            }
            throw new Error(`Failed to fetch user: ${response.statusText}`);
        }

        const userData: UserData = await response.json();
        return userData;
    } catch (error) {
        console.error("Error fetching user:", error);
        return null;
    }
}

export async function getLoggedUserServer(): Promise<UserData | null> {
    const res = await serverFetch("/me");

    if (res.status === 401 || res.status === 403) {
        return null;
    }

    const user: UserData = await res.json();
    return user;
}
