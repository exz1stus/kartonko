import ApiError from "./error";
import { clientFetch } from "./clientFetch";
import parseResponse from "./parseResponse";
import { ErrorResponse } from "./generated/model";

export async function clientMutator<T>(
    url: string,
    options?: RequestInit,
): Promise<T> {
    const response = await clientFetch(url, options);
    if (!response.ok) {
        const data = await parseResponse<ErrorResponse>(response);
        console.log(data.error);
        throw new ApiError(response.status, data);
    }

    return parseResponse<T>(response);
}
