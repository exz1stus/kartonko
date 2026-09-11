export default class ApiError extends Error {
    constructor(
        public readonly status: number,
        public readonly data: unknown,
    ) {
        super(`API request failed with status ${status}`);
        this.name = "ApiError";
    }
}
