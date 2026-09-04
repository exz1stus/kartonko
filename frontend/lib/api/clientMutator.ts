export async function clientFetch(
    url: string,
    options?: RequestInit,
): Promise<Response> {
    const response = await fetch(`/api${url}`, options);

    if (!response.ok) {
        throw new Error(`API request failed: ${response.status}`);
    }

    return response.json();
}
