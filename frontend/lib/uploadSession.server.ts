import "server-only";

import { createHmac, timingSafeEqual } from "node:crypto";
import { getMe } from "@/lib/api/generated/server";
import ApiError from "@/lib/api/error";

const sessionLifetimeSeconds = 8 * 60 * 60;

function signature(payload: string) {
    const secret = process.env.TURNSTILE_SECRET_KEY;
    if (!secret) throw new Error("Turnstile secret is not configured");
    return createHmac("sha256", secret).update(payload).digest("hex");
}

export async function createUploadSession() {
    const user = await getMe();
    const payload = `${user.id}.${Math.floor(Date.now() / 1000) + sessionLifetimeSeconds}`;
    return `${payload}.${signature(payload)}`;
}

export async function assertUploadSession(headers: Headers) {
    const token = headers.get("x-upload-session") ?? "";
    const parts = token.split(".");
    if (parts.length !== 3)
        throw new ApiError(401, { error: "Upload session is missing" });
    const [userID, expires, suppliedSignature] = parts;
    const expiry = Number(expires);
    const expected = Buffer.from(signature(`${userID}.${expires}`), "hex");
    const supplied = Buffer.from(suppliedSignature, "hex");
    if (
        !Number.isSafeInteger(expiry) ||
        expiry <= Date.now() / 1000 ||
        supplied.length !== expected.length ||
        !timingSafeEqual(supplied, expected)
    )
        throw new ApiError(401, { error: "Upload session has expired" });
    const user = await getMe();
    if (String(user.id) !== userID)
        throw new ApiError(403, {
            error: "Upload session belongs to another user",
        });
}
