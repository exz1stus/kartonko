import { postAuthLogout } from "@/lib/api/generated/server";

export async function POST() {
    const res = await postAuthLogout();

    const setCookie = res.headers.get("set-cookie");
    if (setCookie) {
        res.headers.set("set-cookie", setCookie);
    }

    return res;
}
