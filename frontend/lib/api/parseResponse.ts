export default async function parseResponse<T>(response: Response): Promise<T> {
    if (response.status === 204 || response.status === 205) {
        return undefined as T;
    }

    const contentType = response.headers.get("content-type") ?? "";

    if (contentType.includes("application/json")) {
        return response.json();
    }

    if (contentType.startsWith("text/")) {
        return response.text() as T;
    }

    return response.blob() as T;
}
