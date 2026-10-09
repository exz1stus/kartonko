import { postLogout } from "@/lib/api/generated/server";

export async function POST() {
    await postLogout();
    return new Response(null, { status: 200 });
}
