export async function apiFetch(url: string, options?: RequestInit): Promise<Response> {
    const res = await fetch(`${process.env.NEXT_PUBLIC_FRONTEND_URI}/api${url}`, options);
    return res;
}
