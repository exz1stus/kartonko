export default class ApiError extends Error {
    constructor(
        public readonly status: number,
        public readonly data: unknown,
    ) {
        let detail = `API request failed with status ${status}`;
        if (typeof data === "object" && data !== null) {
            if ("error" in data && typeof data.error === "string") {
                detail = data.error;
            } else if (
                "errors" in data &&
                Array.isArray(data.errors) &&
                data.errors.every((item) => typeof item === "string")
            ) {
                detail = data.errors.join(", ");
            } else if ("failures" in data && Array.isArray(data.failures)) {
                const failures = data.failures
                    .map((failure) => {
                        if (typeof failure !== "object" || failure === null) return null;
                        const name = "name" in failure ? String(failure.name) : "item";
                        const error = "error" in failure ? String(failure.error) : "failed";
                        return `${name}: ${error}`;
                    })
                    .filter((failure): failure is string => failure !== null);
                if (failures.length > 0) detail = failures.join(", ");
            }
        }
        super(detail);
        this.name = "ApiError";
    }
}
