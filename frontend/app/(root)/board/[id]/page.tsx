import {
    deleteBoard,
    getBoard,
    patchBoard,
} from "@/lib/api/generated/server";
import { getLoggedUser } from "@/lib/user/user.server";
import { isModerator } from "@/lib/user/user";
import { notFound, redirect } from "next/navigation";
import { revalidatePath } from "next/cache";

interface Props {
    id: string;
}

const ImagePage = async ({ params }: { params: Promise<Props> }) => {
    const { id: idParam } = await params;
    const id = Number(idParam);
    if (!Number.isSafeInteger(id) || id <= 0) notFound();

    const board = await getBoard(id);
    const loggedUser = await getLoggedUser();
    const hasEditPermission =
        loggedUser !== null &&
        (isModerator(loggedUser) || loggedUser.id === board.user_id);

    async function updateBoard(formData: FormData) {
        "use server";
        await patchBoard(
            id,
            {
                name: String(formData.get("name") ?? "").trim(),
                description: String(formData.get("description") ?? ""),
            },
            { credentials: "include" },
        );
        revalidatePath(`/board/${id}`);
    }

    async function removeBoard() {
        "use server";
        await deleteBoard(id, { credentials: "include" });
        redirect("/boards");
    }

    return (
        <div className="flex flex-col justify-center items-center w-full h-full">
            <title>{board.name}</title>
            <span>{board.description}</span>
            {hasEditPermission && (
                <details className="mt-4 border rounded-md p-3">
                    <summary className="cursor-pointer">Edit board</summary>
                    <form action={updateBoard} className="flex flex-col gap-2 mt-3">
                        <input
                            name="name"
                            defaultValue={board.name}
                            required
                            className="border rounded-md px-2 py-1 bg-transparent"
                        />
                        <textarea
                            name="description"
                            defaultValue={board.description}
                            className="border rounded-md px-2 py-1 bg-transparent"
                        />
                        <button type="submit" className="border rounded-md px-3 py-1">
                            Save changes
                        </button>
                    </form>
                    <form action={removeBoard} className="mt-2">
                        <button
                            type="submit"
                            className="border border-red-500 rounded-md px-3 py-1 text-red-500"
                        >
                            Delete board
                        </button>
                    </form>
                </details>
            )}
        </div>
    );
};

export default ImagePage;
