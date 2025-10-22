import { cookies } from "next/headers";

export interface UserData {
    username: string;
    privileage: string;
    picture_url: string;
    joined_at: string;
    last_seen: string;
}

const API_LOCAL = process.env.NEXT_PUBLIC_API_LOCAL;

export async function getUserServer(): Promise<UserData | null> {
    try {
        const cookieStore = cookies();
        const cookieHeader = (await cookieStore)
            .getAll()
            .map((c) => `${c.name}=${c.value}`)
            .join("; ");

        const res = await fetch(`${API_LOCAL}/me`, {
            headers: {
                cookie: cookieHeader,
            },
            credentials: "include",
            cache: "no-store",
        });

        if (!res.ok) return null;
        const data: UserData = await res.json();
        console.log(data);
        return data;
    } catch (e) {
        console.error("Failed to fetch user on server:", e);
        return null;
    }
}
