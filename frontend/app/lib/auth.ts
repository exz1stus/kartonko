import { cookies } from "next/headers";
import { UserData } from "./user";

const API_ORIGIN = process.env.NEXT_PUBLIC_API_LOCAL;

export async function getLoggedUserServer(): Promise<UserData | null> {
    try {
        const cookieStore = cookies();
        const cookieHeader = (await cookieStore)
            .getAll()
            .map((c) => `${c.name}=${c.value}`)
            .join("; ");

        const res = await fetch(`${API_ORIGIN}/me`, {
            headers: {
                cookie: cookieHeader,
            },
            credentials: "include",
            cache: "no-store",
        });

        if (!res.ok) return null;
        const data: UserData = await res.json();
        return data;
    } catch (e) {
        console.error("Failed to fetch user on server:", e);
        return null;
    }
}
