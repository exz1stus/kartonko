import ApiError from "./api/error";

export async function assertCaptchaFromRequest(
    headers: Headers,
    formData: FormData,
): Promise<void> {
    const token = formData.get("cf-turnstile-response")?.toString();

    if (!token) {
        throw new ApiError(400, "Missing captcha token");
    }

    const ip = headers.get("x-forwarded-for") ?? headers.get("x-real-ip") ?? "";

    if (!verifyTurnstile(token, ip))
        throw new ApiError(400, "Invalid captcha token");
}

export async function verifyTurnstile(
    token: string,
    ip: string,
): Promise<boolean> {
    const res = await fetch(
        "https://challenges.cloudflare.com/turnstile/v0/siteverify",
        {
            method: "POST",
            headers: { "Content-Type": "application/x-www-form-urlencoded" },
            body: new URLSearchParams({
                secret: process.env.TURNSTILE_SECRET_KEY!,
                response: token,
                remoteip: ip,
            }),
        },
    );

    const data = await res.json();
    return data.success === true;
}
