import ApiError from "./error";
import parseResponse from "./parseResponse";
import { serverFetch } from "./serverFetch";

export async function serverMutator<T>(
    url: string,
    options?: RequestInit,
): Promise<T> {
    const response = await serverFetch(url, options);
    if (!response.ok) {
        const data = await parseResponse(response);
        throw new ApiError(response.status, data);
    }

    return parseResponse<T>(response);
}
