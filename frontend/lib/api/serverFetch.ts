import { cookies } from "next/headers";

export async function serverFetch(
    url: string,
    options?: RequestInit,
): Promise<Response> {
    const cookieStore = await cookies();

    const cookieHeader = cookieStore
        .getAll()
        .map((cookie) => `${cookie.name}=${cookie.value}`)
        .join("; ");

    return await fetch(`${process.env.API_LOCAL}${url}`, {
        ...options,
        headers: {
            ...options?.headers,
            Cookie: cookieHeader,
        },
        cache: "no-store",
    });
}
