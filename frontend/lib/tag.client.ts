import { apiFetch } from "./apiFetch";

export async function addNewTag(tag: string): Promise<void> {
    const res = await apiFetch(`/tag?name=${tag}`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
    });

    if (!res.ok) {
        const { message } = await res
            .json()
            .catch(() => ({ message: "Unknown error" }));
        throw new Error(message);
    }
}
