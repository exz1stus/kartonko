export async function apiFetch(
    url: string,
    options?: RequestInit,
): Promise<Response> {
    try {
        const res = await fetch(
            `${process.env.NEXT_PUBLIC_FRONTEND_URI}/api${url}`,
            options,
        );
        if (res) return res;
    } catch (e) {
        console.error(`Failed to connect to API: ${e}`);
    }

    return new Response(null, { status: 500 });
}
