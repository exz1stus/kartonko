export async function clientFetch(
    url: string,
    options?: RequestInit,
): Promise<Response> {
    return fetch(`/api${url}`, options);
}
