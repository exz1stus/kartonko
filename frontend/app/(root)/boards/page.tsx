import BoardsList from "@/components/Boards/BoardsList";
import { listBoards } from "@/lib/api/generated/server";

export default async function BoardsPage() {
    const initialBoards = await listBoards({ limit: 12 }, { cache: "no-store" });
    return <BoardsList initialBoards={initialBoards} />;
}
