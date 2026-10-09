import { getBoard, getBoardBySlug, getImagesByQuery, listBoardImageIDs, listBoardImages } from "@/lib/api/generated/server";
import ApiError from "@/lib/api/error";
import { getLoggedUser } from "@/lib/user/user.server";
import { notFound, redirect } from "next/navigation";
import BoardDetail from "@/components/Boards/BoardDetail";

export default async function BoardPage({ params }: { params: Promise<{ slug: string }> }) {
    const { slug } = await params;
    let board;
    try {
        board = await getBoardBySlug(slug);
    } catch (error) {
        if (!(error instanceof ApiError && error.status === 404)) throw error;
        const legacyId = Number(slug);
        if (!Number.isSafeInteger(legacyId) || legacyId <= 0) notFound();
        try {
            const legacyBoard = await getBoard(legacyId);
            redirect(`/board/${encodeURIComponent(legacyBoard.slug)}`);
        } catch (legacyError) {
            if (legacyError instanceof ApiError && legacyError.status === 404) notFound();
            throw legacyError;
        }
    }

    const user = await getLoggedUser();
    const canEdit = !!user && (user.id === board.user_id || user.privilege === "Moderator");
    const [initialItems, initialImageIds, initialPickerImages] = await Promise.all([
        listBoardImages(board.id, { limit: 24 }, { cache: "no-store" }),
        canEdit ? listBoardImageIDs(board.id, { cache: "no-store" }) : Promise.resolve([]),
        canEdit ? getImagesByQuery({ limit: 30 }, { cache: "no-store" }) : Promise.resolve([]),
    ]);

    return <BoardDetail initialBoard={board} initialItems={initialItems} initialImageIds={initialImageIds} initialPickerImages={initialPickerImages} />;
}
