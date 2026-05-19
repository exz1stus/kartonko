async function verifyTurnstile(token: string, ip: string): Promise<boolean> {
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

export default verifyTurnstile;
