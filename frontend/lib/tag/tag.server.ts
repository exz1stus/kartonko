import { serverFetch } from "../serverFetch";

export async function addNewTagServer(tag: string): Promise<void> {
    const res = await serverFetch(`/tag`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: tag }),
    });

    if (!res.ok) {
        const { error } = await res
            .json()
            .catch(() => ({ error: "Unknown error" }));
        throw new Error(error);
    }
}

export async function addNewTagsBatchServer(tags: string[]): Promise<void> {
    const res = await serverFetch("/tags/batch", {
        method: "POST",
        credentials: "include",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({ names: tags }),
    });

    if (!res.ok) {
        const { error } = await res
            .json()
            .catch(() => ({ error: "Unknown error" }));

        throw new Error(error);
    }
}
